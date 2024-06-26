// This file is part of MinIO Operator
// Copyright (c) 2023 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package sidecar

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	common2 "github.com/minio/operator/sidecar/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	clientset "github.com/minio/operator/pkg/client/clientset/versioned"
	minioInformers "github.com/minio/operator/pkg/client/informers/externalversions"
	v22 "github.com/minio/operator/pkg/client/informers/externalversions/minio.min.io/v2"
	"github.com/minio/operator/sidecar/pkg/validator"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// StartSideCar instantiates kube clients and starts the side-car controller
func StartSideCar(tenantName string, secretName string) {
	log.Println("Starting Sidecar")
	cfg, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building Kubernetes clientset: %s", err.Error())
	}

	controllerClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building MinIO clientset: %s", err.Error())
	}

	namespace := v2.GetNSFromFile()
	// get the only tenant in this namespace
	tenant, err := controllerClient.MinioV2().Tenants(namespace).Get(context.Background(), tenantName, metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}

	tenant.EnsureDefaults()

	controller := NewSideCarController(kubeClient, controllerClient, namespace, tenantName, secretName)
	controller.ws = configureWebhookServer(controller)
	controller.probeServer = configureProbesServer(controller, tenant.TLS())

	stopControllerCh := make(chan struct{})

	defer close(stopControllerCh)
	err = controller.Run(stopControllerCh)
	if err != nil {
		klog.Fatal(err)
	}

	go func() {
		if err = controller.ws.ListenAndServe(); err != nil {
			// if the web server exits,
			klog.Error(err)
			close(stopControllerCh)
		}
	}()

	go func() {
		if err = controller.probeServer.ListenAndServe(); err != nil {
			// if the web server exits,
			klog.Error(err)
			close(stopControllerCh)
		}
	}()

	<-stopControllerCh
}

// Controller is the controller holding the informers used to monitor args and tenant structure
type Controller struct {
	kubeClient         *kubernetes.Clientset
	controllerClient   *clientset.Clientset
	tenantName         string
	secretName         string
	minInformerFactory minioInformers.SharedInformerFactory
	secretInformer     coreinformers.SecretInformer
	tenantInformer     v22.TenantInformer
	namespace          string
	informerFactory    informers.SharedInformerFactory
	ws                 *http.Server
	probeServer        *http.Server
}

// NewSideCarController returns an instance of Controller with the provided clients
func NewSideCarController(kubeClient *kubernetes.Clientset, controllerClient *clientset.Clientset, namespace string, tenantName string, secretName string) *Controller {
	factory := informers.NewSharedInformerFactoryWithOptions(kubeClient, time.Hour*1, informers.WithNamespace(namespace))
	secretInformer := factory.Core().V1().Secrets()

	minioInformerFactory := minioInformers.NewSharedInformerFactoryWithOptions(controllerClient, time.Hour*1, minioInformers.WithNamespace(namespace))
	tenantInformer := minioInformerFactory.Minio().V2().Tenants()

	c := &Controller{
		kubeClient:         kubeClient,
		controllerClient:   controllerClient,
		tenantName:         tenantName,
		namespace:          namespace,
		secretName:         secretName,
		minInformerFactory: minioInformerFactory,
		informerFactory:    factory,
		tenantInformer:     tenantInformer,
		secretInformer:     secretInformer,
	}

	tenantInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			oldTenant := old.(*v2.Tenant)
			newTenant := new.(*v2.Tenant)
			if newTenant.ResourceVersion == oldTenant.ResourceVersion {
				// Periodic resync will send update events for all known Tenants.
				// Two different versions of the same Tenant will always have different RVs.
				return
			}
			c.regenCfg(tenantName, namespace)
		},
	})

	secretInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			oldSecret := old.(*corev1.Secret)
			// ignore anything that is not what we want
			if oldSecret.Name != secretName {
				return
			}
			log.Printf("Config secret '%s' sync", secretName)
			newSecret := new.(*corev1.Secret)
			if newSecret.ResourceVersion == oldSecret.ResourceVersion {
				// Periodic resync will send update events for all known Tenants.
				// Two different versions of the same Tenant will always have different RVs.
				return
			}
			data := newSecret.Data["config.env"]
			// validate root creds in string
			rootUserFound := false
			rootPwdFound := false

			dataStr := string(data)
			if strings.Contains(dataStr, "MINIO_ROOT_USER") {
				rootUserFound = true
			}
			if strings.Contains(dataStr, "MINIO_ACCESS_KEY") {
				rootUserFound = true
			}
			if strings.Contains(dataStr, "MINIO_ROOT_PASSWORD") {
				rootPwdFound = true
			}
			if strings.Contains(dataStr, "MINIO_SECRET_KEY") {
				rootPwdFound = true
			}
			if !rootUserFound || !rootPwdFound {
				log.Println("Missing root credentials in the configuration.")
				log.Println("MinIO won't start")
				os.Exit(1)
			}

			if !strings.HasSuffix(dataStr, "\n") {
				dataStr = dataStr + "\n"
			}
			c.regenCfgWithCfg(tenantName, namespace, dataStr)
		},
	})

	return c
}

func (c *Controller) regenCfg(tenantName string, namespace string) {
	rootUserFound, rootPwdFound, fileContents, err := validator.ReadTmpConfig()
	if err != nil {
		log.Println(err)
		return
	}
	if !rootUserFound || !rootPwdFound {
		log.Println("Missing root credentials in the configuration.")
		log.Println("MinIO won't start")
		os.Exit(1)
	}
	c.regenCfgWithCfg(tenantName, namespace, fileContents)
}

func (c *Controller) regenCfgWithCfg(tenantName string, namespace string, fileContents string) {
	ctx := context.Background()

	tenant, err := c.controllerClient.MinioV2().Tenants(namespace).Get(ctx, tenantName, metav1.GetOptions{})
	if err != nil {
		log.Println("could not get tenant", err)
		return
	}
	tenant.EnsureDefaults()

	fileContents = common2.AttachGeneratedConfig(tenant, fileContents)

	err = os.WriteFile(v2.CfgFile, []byte(fileContents), 0o644)
	if err != nil {
		log.Println(err)
	}
}

// Run starts the informers
func (c *Controller) Run(stopCh chan struct{}) error {
	// Starts all the shared minioInformers that have been created by the factory so far.
	go c.minInformerFactory.Start(stopCh)
	go c.informerFactory.Start(stopCh)

	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, c.tenantInformer.Informer().HasSynced, c.secretInformer.Informer().HasSynced) {
		return fmt.Errorf("failed to sync")
	}
	return nil
}
