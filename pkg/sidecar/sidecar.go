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
	"github.com/minio/operator/pkg/common"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	v2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	clientset "github.com/minio/operator/pkg/client/clientset/versioned"
	minioInformers "github.com/minio/operator/pkg/client/informers/externalversions"
	v22 "github.com/minio/operator/pkg/client/informers/externalversions/minio.min.io/v2"
	"github.com/minio/operator/pkg/validator"
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

	controller := NewSideCarController(kubeClient, controllerClient, tenantName, secretName)
	controller.ws = configureWebhookServer(controller)

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
}

// NewSideCarController returns an instance of Controller with the provided clients
func NewSideCarController(kubeClient *kubernetes.Clientset, controllerClient *clientset.Clientset, tenantName string, secretName string) *Controller {
	namespace := v2.GetNSFromFile()

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
			c.regenCfg(newTenant)
		},
	})

	secretInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			oldSecret := old.(*corev1.Secret)

			// ignore any secret in a different namespace
			if oldSecret.Namespace != namespace {
				return
			}
			// ignore anything that is not what we want
			if oldSecret.Name != secretName {
				return
			}
			newSecret := new.(*corev1.Secret)
			if newSecret.ResourceVersion == oldSecret.ResourceVersion {
				// Periodic resync will send update events for all known Tenants.
				// Two different versions of the same Tenant will always have different RVs.
				return
			}
			data := newSecret.Data["config.env"]
			// validate root creds in string
			rootUserFound := false
			rootPassFound := false

			dataStr := string(data)
			if strings.Contains(dataStr, "MINIO_ROOT_USER") || strings.Contains(dataStr, "MINIO_ACCESS_KEY") {
				rootUserFound = true
			}
			if strings.Contains(dataStr, "MINIO_ROOT_PASSWORD") || strings.Contains(dataStr, "MINIO_SECRET_KEY") {
				rootPassFound = true
			}
			if !rootUserFound || !rootPassFound {
				log.Println("Missing root credentials in the configuration.")
				log.Println("MinIO won't start")
				os.Exit(1)
			}
			tenant, err := tenantInformer.Lister().Tenants(namespace).Get(tenantName)

			c.regenCfg(tenantName, namespace, string(data))
		},
	})

	return c
}

func (c *Controller) regenCfg(tenant *v2.Tenant) {
	ctx := context.Background()

	fileContents, err := validator.ReadTmpConfig()
	if err != nil {
		log.Println(err)
		log.Println("MinIO won't start")
		os.Exit(1)
	}

	envVariables, err := c.GetTenantEnvVariables(tenant, fileContents)
	if err != nil {
		log.Println(err)
		log.Println("MinIO won't start")
		os.Exit(1)
	}
	args, err := validator.GetTenantArgs(ctx, c.controllerClient, tenant.Namespace, tenant.Namespace)
	if err != nil {
		log.Println(err)
		return
	}

	envVariables[v2.MinioArgsEnv] = args

	for key, val := range envVariables {
		fileContents = fileContents + fmt.Sprintf("export %s=\"%s\"\n", key, val)
	}

	err = os.WriteFile(v2.CfgFile, []byte(fileContents), 0o644)
	if err != nil {
		log.Println(err)
	}
}

func (c *Controller) regenCfgWithCfg(tenantName string, namespace string, fileContents string) {
	ctx := context.Background()

	args, err := validator.GetTenantArgs(ctx, c.controllerClient, tenantName, namespace)
	if err != nil {
		log.Println(err)
		return
	}

	fileContents = fileContents + fmt.Sprintf("export MINIO_ARGS=\"%s\"\n", args)

	err = os.WriteFile(v2.CfgFile, []byte(fileContents), 0o644)
	if err != nil {
		log.Println(err)
	}
}

// Run starts the informers
func (c *Controller) Run(stopCh chan struct{}) error {
	// Starts all the shared minioInformers that have been created by the factory so
	// far.
	c.minInformerFactory.Start(stopCh)
	c.informerFactory.Start(stopCh)

	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, c.tenantInformer.Informer().HasSynced) {
		return fmt.Errorf("Failed to sync")
	}
	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, c.secretInformer.Informer().HasSynced) {
		return fmt.Errorf("Failed to sync")
	}
	return nil
}

func (c *Controller) GetTenantEnvVariables(tenant *v2.Tenant, fileContents string) (map[string]string, error) {

	prevConfiguration := v2.ParseRawConfiguration([]byte(fileContents))

	var tenantConfiguration map[string]string
	for k, v := range prevConfiguration {
		tenantConfiguration[k] = string(v)
	}

	// Enable `mc admin update` style updates to MinIO binaries
	// within the container, only operator is supposed to perform
	// these operations.
	tenantConfiguration[v2.MinioUpdateEnv] = "on"
	tenantConfiguration[v2.MinioSignPubKey] = "RWTx5Zr1tiHQLwG9keckT0c45M3AGeHD6IvimQHpyRywVWGbP1aVSGav"
	// tenantConfiguration[v2.MinioOperatorVersionEnv] = c.operatorVersion TODO:// determinate the source of truth of the Operator version, a constant?
	tenantConfiguration[v2.MinioPrometheusJobIdEnv] = tenant.PrometheusConfigJobName()

	var domains []string
	// Set Bucket DNS domain only if enabled
	if tenant.BucketDNS() {
		domains = append(domains, tenant.MinIOBucketBaseDomain())
		sidecarBucketURL := fmt.Sprintf("http://127.0.0.1:%s%s/%s/%s",
			common.WebhookDefaultPort,
			common.WebhookAPIBucketService,
			tenant.Namespace,
			tenant.Name)
		tenantConfiguration[common.BucketDNSEnv] = sidecarBucketURL
	}
	// Check if any domains are configured
	if tenant.HasMinIODomains() {
		domains = append(domains, tenant.GetDomainHosts()...)
	}
	// tell MinIO about all the domains meant to hit it if they are not passed manually via .spec.env
	if len(domains) > 0 {
		tenantConfiguration[v2.MinIODomain] = strings.Join(domains, ",")
	}
	// If no specific server URL is specified we will specify the internal k8s url, but if a list of domains was
	// provided we will use the first domain.
	serverURL := tenant.MinIOServerEndpoint()
	if tenant.HasMinIODomains() {
		// Infer schema from tenant TLS, if not explicit
		if !strings.HasPrefix(tenant.Spec.Features.Domains.Minio[0], "http") {
			useSchema := "http"
			if tenant.TLS() {
				useSchema = "https"
			}
			serverURL = fmt.Sprintf("%s://%s", useSchema, tenant.Spec.Features.Domains.Minio[0])
		} else {
			serverURL = tenant.Spec.Features.Domains.Minio[0]
		}
	}
	tenantConfiguration[v2.MinIOServerURL] = serverURL

	// Set the redirect url for console
	if tenant.HasConsoleDomains() {
		consoleDomain := tenant.Spec.Features.Domains.Console
		// Infer schema from tenant TLS, if not explicit
		if !strings.HasPrefix(consoleDomain, "http") {
			useSchema := "http"
			if tenant.TLS() {
				useSchema = "https"
			}
			consoleDomain = fmt.Sprintf("%s://%s", useSchema, consoleDomain)
		}
		tenantConfiguration[v2.MinIOBrowserRedirectURL] = consoleDomain
	}

	if tenant.HasKESEnabled() {
		tenantConfiguration[v2.MinioKMSKESEndpointEnv] = tenant.KESServiceEndpoint()
		tenantConfiguration[v2.MinioKMSKESCertFileEnv] = v2.MinIOCertPath + "/client.crt"
		tenantConfiguration[v2.MinioKMSKESKeyFileEnv] = v2.MinIOCertPath + "/client.key"
		tenantConfiguration[v2.MinioKMSKESCAPathEnv] = v2.MinIOCertPath + "/CAs/kes.crt"
		tenantConfiguration[v2.MinioKMSKESKeyNameEnv] = tenant.Spec.KES.KeyName
	}

	// Set the env variables in the tenant.spec.env field
	// User defined environment variables will take precedence over default environment variables
	envVars := tenant.GetEnvVars()
	for _, ev := range envVars {
		tenantConfiguration[ev.Name] = ev.Value
	}

	return tenantConfiguration, nil
}
