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
	"sync"
	"time"

	"github.com/minio/operator/sidecar/pkg/configuration"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	clientset "github.com/minio/operator/pkg/client/clientset/versioned"
	minioInformers "github.com/minio/operator/pkg/client/informers/externalversions"
	v22 "github.com/minio/operator/pkg/client/informers/externalversions/minio.min.io/v2"
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
func StartSideCar(tenantName string) {
	log.Println("Starting Sidecar")
	var cfg *rest.Config
	var err error

	if os.Getenv("DEV_NAMESPACE") != "" {
		klog.Info("DEV_NAMESPACE present, running dev mode")
		cfg = &rest.Config{
			Host:            "http://localhost:8001",
			TLSClientConfig: rest.TLSClientConfig{Insecure: true},
			APIPath:         "/",
		}
	} else {
		// Look for incluster config by default
		cfg, err = rest.InClusterConfig()
	}

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

	controller := NewSideCarController(kubeClient, controllerClient, tenant)
	controller.ws = configureWebhookServer(controller)
	controller.probeServer = configureProbesServer(tenant)
	controller.sidecar = configureSidecarServer(controller)

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

	go func() {
		if err = controller.sidecar.ListenAndServe(); err != nil {
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
	minInformerFactory minioInformers.SharedInformerFactory
	secretInformer     coreinformers.SecretInformer
	configMapInformer  coreinformers.ConfigMapInformer
	tenantInformer     v22.TenantInformer
	informerFactory    informers.SharedInformerFactory
	lock               sync.Mutex
	tenant             *v2.Tenant
	configMaps         map[string]*corev1.ConfigMap
	secrets            map[string]*corev1.Secret
	ws                 *http.Server
	probeServer        *http.Server
	sidecar            *http.Server
}

// NewSideCarController returns an instance of Controller with the provided clients
func NewSideCarController(kubeClient *kubernetes.Clientset, controllerClient *clientset.Clientset, tenant *v2.Tenant) *Controller {
	factory := informers.NewSharedInformerFactoryWithOptions(kubeClient, time.Hour*1, informers.WithNamespace(tenant.Namespace))
	secretInformer := factory.Core().V1().Secrets()
	configMapInformer := factory.Core().V1().ConfigMaps()

	minioInformerFactory := minioInformers.NewSharedInformerFactoryWithOptions(controllerClient, time.Hour*1, minioInformers.WithNamespace(tenant.Namespace))
	tenantInformer := minioInformerFactory.Minio().V2().Tenants()

	c := &Controller{
		kubeClient:         kubeClient,
		controllerClient:   controllerClient,
		minInformerFactory: minioInformerFactory,
		informerFactory:    factory,
		tenantInformer:     tenantInformer,
		secretInformer:     secretInformer,
		configMapInformer:  configMapInformer,
		tenant:             tenant,
	}

	_, err := tenantInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			oldTenant := old.(*v2.Tenant)
			newTenant := new.(*v2.Tenant)
			if newTenant.ResourceVersion == oldTenant.ResourceVersion {
				// Periodic resync will send update events for all known Tenants.
				// Two different versions of the same Tenant will always have different RVs.
				return
			}
			c.lock.Lock()
			defer c.lock.Unlock()

			log.Println("tenant was updated, regenerating configuration")

			c.tenant = newTenant
			c.regenCfg()
		},
	})
	if err != nil {
		log.Println("could not add event handler for tenant informer", err)
		return nil
	}

	_, err = secretInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			oldSecret := old.(*corev1.Secret)
			newSecret := new.(*corev1.Secret)
			if oldSecret.ResourceVersion == newSecret.ResourceVersion {
				// Periodic resync will send update events for all known secrets.
				// Two different versions of the same secret will always have different RVs.
				return
			}
			c.lock.Lock()
			defer c.lock.Unlock()

			log.Printf("secret %s was updated, regenerating configuration", newSecret.Name)

			if _, ok := c.secrets[oldSecret.Name]; !ok {
				// Not interested in secrets that we don't use
				return
			}
			c.regenCfg()
		},
	})
	if err != nil {
		log.Println("could not add event handler for secret informer", err)
		return nil
	}

	_, err = configMapInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			oldConfigMap := old.(*corev1.ConfigMap)
			newConfigMap := new.(*corev1.ConfigMap)
			if oldConfigMap.ResourceVersion == newConfigMap.ResourceVersion {
				// Periodic resync will send update events for all known config maps.
				// Two different versions of the same config map will always have different RVs.
				return
			}
			c.lock.Lock()
			defer c.lock.Unlock()

			log.Printf("configmap %s was updated, regenerating configuration", newConfigMap.Name)

			if _, ok := c.configMaps[oldConfigMap.Name]; !ok {
				// Not interested in configmaps that we don't use
				return
			}
			c.regenCfg()
		},
	})
	if err != nil {
		log.Println("could not add event handler for secret informer", err)
		return nil
	}

	return c
}

func (c *Controller) getSecret(_ context.Context, name string) (*corev1.Secret, error) {
	return c.secretInformer.Lister().Secrets(c.tenant.Namespace).Get(name)
}

func (c *Controller) getConfigMap(_ context.Context, name string) (*corev1.ConfigMap, error) {
	return c.configMapInformer.Lister().ConfigMaps(c.tenant.Namespace).Get(name)
}

func (c *Controller) regenCfg() {
	// get the tenant secret
	c.tenant.EnsureDefaults()

	// determine the configmaps and secrets to watch
	configMaps, secrets, err := configuration.TenantResources(context.Background(), c.tenant, c.getConfigMap, c.getSecret)
	if err != nil {
		log.Println(err)
		return
	}

	// update secrets and configmaps that should be watched
	c.secrets = secrets
	c.configMaps = configMaps

	// obtain the full tenant configuration
	fileContents, rootUserFound, rootPwdFound := configuration.GetFullTenantConfig(c.tenant, c.configMaps, c.secrets)

	if !rootUserFound || !rootPwdFound {
		log.Println("Missing root credentials in the configuration.")
		log.Println("MinIO won't start")
		os.Exit(1)
	}

	tmpFile := v2.CfgFile + ".tmp"
	defer os.Remove(tmpFile)

	err = os.WriteFile(tmpFile, []byte(fileContents), 0o644)
	if err != nil {
		log.Println(err)
	}
	err = os.Rename(tmpFile, v2.CfgFile)
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
	if !cache.WaitForCacheSync(stopCh, c.tenantInformer.Informer().HasSynced, c.configMapInformer.Informer().HasSynced, c.secretInformer.Informer().HasSynced) {
		return fmt.Errorf("failed to sync")
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	c.regenCfg()
	return nil
}
