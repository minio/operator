/*
 * Copyright (C) 2020, MinIO, Inc.
 *
 * This code is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, version 3,
 * as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License, version 3,
 * along with this program.  If not, see <http://www.gnu.org/licenses/>
 *
 */

package cluster

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/docker/cli/cli/config/configfile"

	"k8s.io/klog/v2"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	batchinformers "k8s.io/client-go/informers/batch/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	certapi "k8s.io/client-go/kubernetes/typed/certificates/v1beta1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	batchlisters "k8s.io/client-go/listers/batch/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	queue "k8s.io/client-go/util/workqueue"

	"github.com/dgrijalva/jwt-go"
	jwtreq "github.com/dgrijalva/jwt-go/request"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/gorilla/mux"
	"github.com/minio/minio/pkg/auth"
	miniov1 "github.com/minio/operator/pkg/apis/minio.min.io/v1"
	clientset "github.com/minio/operator/pkg/client/clientset/versioned"
	minioscheme "github.com/minio/operator/pkg/client/clientset/versioned/scheme"
	informers "github.com/minio/operator/pkg/client/informers/externalversions/minio.min.io/v1"
	listers "github.com/minio/operator/pkg/client/listers/minio.min.io/v1"
	"github.com/minio/operator/pkg/resources/deployments"
	"github.com/minio/operator/pkg/resources/jobs"
	"github.com/minio/operator/pkg/resources/services"
	"github.com/minio/operator/pkg/resources/statefulsets"
)

const (
	controllerAgentName = "minio-operator"
	// SuccessSynced is used as part of the Event 'reason' when a Tenant is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Tenant fails
	// to sync due to a StatefulSet of the same name already existing.
	ErrResourceExists = "ErrResourceExists"
	// MessageResourceExists is the message used for Events when a Tenant
	// fails to sync due to a StatefulSet already existing
	MessageResourceExists = "Resource %q already exists and is not managed by MinIO Operator"
	// MessageResourceSynced is the message used for an Event fired when a Tenant
	// is synced successfully
	MessageResourceSynced = "Tenant synced successfully"
)

// Standard Status messages for Tenant
const (
	StatusInitialized                   = "Initialized"
	StatusProvisioningCIService         = "Provisioning MinIO Cluster IP Service"
	StatusProvisioningHLService         = "Provisioning MinIO Headless Service"
	StatusProvisioningStatefulSet       = "Provisioning MinIO Statefulset"
	StatusProvisioningConsoleDeployment = "Provisioning Console Deployment"
	StatusProvisioningKESStatefulSet    = "Provisioning KES StatefulSet"
	StatusWaitingForReadyState          = "Waiting for Pods to be ready"
	StatusWaitingMinIOCert              = "Waiting for MinIO TLS Certificate"
	StatusWaitingMinIOClientCert        = "Waiting for MinIO TLS Client Certificate"
	StatusWaitingKESCert                = "Waiting for KES TLS Certificate"
	StatusWaitingConsoleCert            = "Waiting for Console TLS Certificate"
	StatusUpdatingMinIOVersion          = "Updating MinIO Version"
	StatusUpdatingConsoleVersion        = "Updating Console Version"
	StatusUpdatingResourceRequirements  = "Updating Resource Requirements"
	StatusUpdatingAffinity              = "Updating Pod Affinity"
	StatusNotOwned                      = "Statefulset not controlled by operator"
	StatusFailedAlreadyExists           = "Another MinIO Tenant already exists in the namespace"
	StatusInconsistentMinIOVersions     = "Different versions across MinIO Zones"
)

// Controller struct watches the Kubernetes API for changes to Tenant resources
type Controller struct {
	// kubeClientSet is a standard kubernetes clientset
	kubeClientSet kubernetes.Interface
	// minioClientSet is a clientset for our own API group
	minioClientSet clientset.Interface
	// certClient is a clientset for our certficate management
	certClient certapi.CertificatesV1beta1Client
	// statefulSetLister is able to list/get StatefulSets from a shared
	// informer's store.
	statefulSetLister appslisters.StatefulSetLister
	// statefulSetListerSynced returns true if the StatefulSet shared informer
	// has synced at least once.
	statefulSetListerSynced cache.InformerSynced

	// deploymentLister is able to list/get Deployments from a shared
	// informer's store.
	deploymentLister appslisters.DeploymentLister
	// deploymentListerSynced returns true if the Deployment shared informer
	// has synced at least once.
	deploymentListerSynced cache.InformerSynced

	// jobLister is able to list/get Deployments from a shared
	// informer's store.
	jobLister batchlisters.JobLister
	// jobListerSynced returns true if the Deployment shared informer
	// has synced at least once.
	jobListerSynced cache.InformerSynced

	// tenantsLister lists Tenant from a shared informer's
	// store.
	tenantsLister listers.TenantLister
	// tenantsSynced returns true if the StatefulSet shared informer
	// has synced at least once.
	tenantsSynced cache.InformerSynced

	// serviceLister is able to list/get Services from a shared informer's
	// store.
	serviceLister corelisters.ServiceLister
	// serviceListerSynced returns true if the Service shared informer
	// has synced at least once.
	serviceListerSynced cache.InformerSynced

	// queue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue queue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	// Use a go template to render the hosts string
	hostsTemplate string

	// currently running operator version
	operatorVersion string

	// Webhook server instance
	ws *http.Server
}

// NewController returns a new sample controller
func NewController(
	kubeClientSet kubernetes.Interface,
	minioClientSet clientset.Interface,
	certClient certapi.CertificatesV1beta1Client,
	statefulSetInformer appsinformers.StatefulSetInformer,
	deploymentInformer appsinformers.DeploymentInformer,
	jobInformer batchinformers.JobInformer,
	tenantInformer informers.TenantInformer,
	serviceInformer coreinformers.ServiceInformer,
	hostsTemplate, operatorVersion string) *Controller {

	// Create event broadcaster
	// Add minio-controller types to the default Kubernetes Scheme so Events can be
	// logged for minio-controller types.
	minioscheme.AddToScheme(scheme.Scheme) //nolint:errcheck
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClientSet.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeClientSet:           kubeClientSet,
		minioClientSet:          minioClientSet,
		certClient:              certClient,
		statefulSetLister:       statefulSetInformer.Lister(),
		statefulSetListerSynced: statefulSetInformer.Informer().HasSynced,
		deploymentLister:        deploymentInformer.Lister(),
		deploymentListerSynced:  deploymentInformer.Informer().HasSynced,
		jobLister:               jobInformer.Lister(),
		jobListerSynced:         jobInformer.Informer().HasSynced,
		tenantsLister:           tenantInformer.Lister(),
		tenantsSynced:           tenantInformer.Informer().HasSynced,
		serviceLister:           serviceInformer.Lister(),
		serviceListerSynced:     serviceInformer.Informer().HasSynced,
		workqueue:               queue.NewNamedRateLimitingQueue(queue.DefaultControllerRateLimiter(), "Tenants"),
		recorder:                recorder,
		hostsTemplate:           hostsTemplate,
		operatorVersion:         operatorVersion,
	}

	// Initialize operator webhook handlers
	controller.ws = configureWebhookServer(controller)

	klog.Info("Setting up event handlers")
	// Set up an event handler for when Tenant resources change
	tenantInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueTenant,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueTenant(new)
		},
	})
	// Set up an event handler for when StatefulSet resources change. This
	// handler will lookup the owner of the given StatefulSet, and if it is
	// owned by a Tenant resource will enqueue that Tenant resource for
	// processing. This way, we don't need to implement custom logic for
	// handling StatefulSet resources. More info on this pattern:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/controllers.md
	statefulSetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*appsv1.StatefulSet)
			oldDepl := old.(*appsv1.StatefulSet)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				// Periodic resync will send update events for all known StatefulSet.
				// Two different versions of the same StatefulSet will always have different RVs.
				return
			}
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
	})

	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*appsv1.Deployment)
			oldDepl := old.(*appsv1.Deployment)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				// Periodic resync will send update events for all known Deployments.
				// Two different versions of the same Deployments will always have different RVs.
				return
			}
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
	})
	return controller
}

func (c *Controller) validateRequest(r *http.Request, secret *v1.Secret) error {
	tokenStr, err := jwtreq.AuthorizationHeaderExtractor.ExtractToken(r)
	if err != nil {
		return err
	}

	stdClaims := &jwt.StandardClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, stdClaims, func(token *jwt.Token) (interface{}, error) {
		return secret.Data[miniov1.WebhookOperatorPassword], nil
	})
	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf(http.StatusText(http.StatusForbidden))
	}
	if stdClaims.Issuer != string(secret.Data[miniov1.WebhookOperatorUsername]) {
		return fmt.Errorf(http.StatusText(http.StatusForbidden))
	}

	return nil
}

func (c *Controller) applyOperatorWebhookSecret(ctx context.Context, tenant *miniov1.Tenant) (*v1.Secret, error) {
	secret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx,
		miniov1.WebhookSecret, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			cred, err := auth.GetNewCredentials()
			if err != nil {
				return nil, err
			}
			secret = &corev1.Secret{
				Type: "Opaque",
				ObjectMeta: metav1.ObjectMeta{
					Name:      miniov1.WebhookSecret,
					Namespace: tenant.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(tenant, schema.GroupVersionKind{
							Group:   miniov1.SchemeGroupVersion.Group,
							Version: miniov1.SchemeGroupVersion.Version,
							Kind:    miniov1.MinIOCRDResourceKind,
						}),
					},
				},
				Data: map[string][]byte{
					miniov1.WebhookOperatorUsername: []byte(cred.AccessKey),
					miniov1.WebhookOperatorPassword: []byte(cred.SecretKey),
					miniov1.WebhookMinIOArgs: []byte(fmt.Sprintf("%s://%s:%s@%s:%s%s/%s/%s",
						"env",
						cred.AccessKey,
						cred.SecretKey,
						fmt.Sprintf("operator.%s.svc.%s",
							miniov1.GetNSFromFile(),
							miniov1.ClusterDomain),
						miniov1.WebhookDefaultPort,
						miniov1.WebhookAPIGetenv,
						tenant.Namespace,
						tenant.Name)),
				},
			}
			return c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Create(ctx, secret, metav1.CreateOptions{})
		}
		return nil, err
	}
	return secret, nil
}

// Supported remote envs
const (
	envMinIOArgs          = "MINIO_ARGS"
	envMinIOServiceTarget = "MINIO_DNS_WEBHOOK_ENDPOINT"
	updatePath            = "/tmp" + miniov1.WebhookAPIUpdate + slashSeparator
)

// BucketSrvHandler - POST /webhook/v1/bucketsrv/{namespace}/{name}?bucket={bucket}
func (c *Controller) BucketSrvHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	v := r.URL.Query()

	namespace := vars["namespace"]
	bucket := vars["bucket"]
	name := vars["name"]
	deleteBucket := v.Get("delete")

	secret, err := c.kubeClientSet.CoreV1().Secrets(namespace).Get(r.Context(),
		miniov1.WebhookSecret, metav1.GetOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	if err = c.validateRequest(r, secret); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	ok, err := strconv.ParseBool(deleteBucket)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	if ok {
		if err = c.kubeClientSet.CoreV1().Services(namespace).Delete(r.Context(), bucket, metav1.DeleteOptions{}); err != nil {
			klog.Errorf("failed to delete service:%s for tenant:%s/%s, err:%s", name, namespace, name, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		return
	}

	// Find the tenant
	tenant, err := c.tenantsLister.Tenants(namespace).Get(name)
	if err != nil {
		klog.Errorf("Unable to lookup tenant:%s/%s for the bucket:%s request. err:%s", namespace, name, bucket, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tenant.EnsureDefaults()
	miniov1.InitGlobals(tenant)

	// Validate the MinIO Tenant
	if err = tenant.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// Create the service for the bucket name
	service := services.ServiceForBucket(tenant, bucket)
	_, err = c.kubeClientSet.CoreV1().Services(namespace).Create(r.Context(), service, metav1.CreateOptions{})
	if err != nil && k8serrors.IsAlreadyExists(err) {
		klog.Infof("Bucket:%s already exists for tenant:%s/%s err:%s ", bucket, namespace, name, err)
		// This might be a previously failed bucket creation. The service is expected to the be the same as the one
		// already in place so clear the error.
		err = nil
	}
	if err != nil {
		klog.Errorf("Unable to create service for tenant:%s/%s for the bucket:%s request. err:%s", namespace, name, bucket, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

// GetenvHandler - GET /webhook/v1/getenv/{namespace}/{name}?key={env}
func (c *Controller) GetenvHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	name := vars["name"]
	key := vars["key"]

	secret, err := c.kubeClientSet.CoreV1().Secrets(namespace).Get(r.Context(),
		miniov1.WebhookSecret, metav1.GetOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if err = c.validateRequest(r, secret); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// Get the Tenant resource with this namespace/name
	tenant, err := c.tenantsLister.Tenants(namespace).Get(name)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// The Tenant resource may no longer exist, in which case we stop processing.
			http.Error(w, fmt.Sprintf("Tenant '%s' in work queue no longer exists", key), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	tenant.EnsureDefaults()
	miniov1.InitGlobals(tenant)

	// Validate the MinIO Tenant
	if err = tenant.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	switch key {
	case envMinIOArgs:
		args := strings.Join(statefulsets.GetContainerArgs(tenant, c.hostsTemplate), " ")
		klog.Infof("%s value is %s", key, args)

		_, _ = w.Write([]byte(args))
		w.(http.Flusher).Flush()
	case envMinIOServiceTarget:
		target := fmt.Sprintf("%s://%s:%s%s/%s/%s",
			"http",
			fmt.Sprintf("operator.%s.svc.%s",
				miniov1.GetNSFromFile(),
				miniov1.ClusterDomain),
			miniov1.WebhookDefaultPort,
			miniov1.WebhookAPIBucketService,
			tenant.Namespace,
			tenant.Name)
		klog.Infof("%s value is %s", key, target)

		_, _ = w.Write([]byte(target))
	default:
		http.Error(w, fmt.Sprintf("%s env key is not supported yet", key), http.StatusBadRequest)
		return
	}
}

func (c *Controller) fetchTag(path string) (string, error) {
	cmd := exec.Command(path, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	op := strings.Fields(out.String())
	if len(op) != 3 {
		return "", fmt.Errorf("incorrect output while fetching tag value - %d", len((op)))
	}
	return op[2], nil
}

// minioKeychain implements Keychain to pass custom credentials
type minioKeychain struct {
	authn.Keychain
	Username      string
	Password      string
	Auth          string
	IdentityToken string
	RegistryToken string
}

// Resolve implements Keychain.
func (mk *minioKeychain) Resolve(_ authn.Resource) (authn.Authenticator, error) {
	return authn.FromConfig(authn.AuthConfig{
		Username:      mk.Username,
		Password:      mk.Password,
		Auth:          mk.Auth,
		IdentityToken: mk.IdentityToken,
		RegistryToken: mk.RegistryToken,
	}), nil
}

// getKeychainForTenant attempts to build a new authn.Keychain from the image pull secret on the Tenant
func (c *Controller) getKeychainForTenant(ctx context.Context, ref name.Reference, tenant *miniov1.Tenant) (authn.Keychain, error) {
	// Get the secret
	secret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, tenant.Spec.ImagePullSecret.Name, metav1.GetOptions{})
	if err != nil {
		return authn.DefaultKeychain, errors.New("can't retrieve the tenant image pull secret")
	}
	// if we can't find .dockerconfigjson, error out
	dockerConfigJSON, ok := secret.Data[".dockerconfigjson"]
	if !ok {
		return authn.DefaultKeychain, fmt.Errorf("unable to find `.dockerconfigjson` in image pull secret")
	}
	var config configfile.ConfigFile
	if err = json.Unmarshal(dockerConfigJSON, &config); err != nil {
		return authn.DefaultKeychain, fmt.Errorf("Unable to decode docker config secrets %w", err)
	}
	cfg, ok := config.AuthConfigs[ref.Context().RegistryStr()]
	if !ok {
		return authn.DefaultKeychain, fmt.Errorf("unable to locate auth config registry context %s", ref.Context().RegistryStr())
	}
	return &minioKeychain{
		Username:      cfg.Username,
		Password:      cfg.Password,
		Auth:          cfg.Auth,
		IdentityToken: cfg.IdentityToken,
		RegistryToken: cfg.RegistryToken,
	}, nil
}

// Attempts to fetch given image and then extracts and keeps relevant files
// (minio, minio.sha256sum & minio.minisig) at a pre-defined location (/tmp/webhook/v1/update)
func (c *Controller) fetchArtifacts(tenant *miniov1.Tenant) (latest time.Time, err error) {
	basePath := updatePath

	if err = os.MkdirAll(basePath, 1777); err != nil {
		return latest, err
	}

	ref, err := name.ParseReference(tenant.Spec.Image)
	if err != nil {
		return latest, err
	}

	keychain := authn.DefaultKeychain

	// if the tenant has imagePullSecret use that for pulling the image, but if we fail to extract the secret or we
	// can't find the expected registry in the secret we will continue with the default keychain. This is because the
	// needed pull secret could be attached to the service-account.
	if tenant.Spec.ImagePullSecret.Name != "" {
		// Get the secret
		keychain, err = c.getKeychainForTenant(context.Background(), ref, tenant)
		if err != nil {
			klog.Info(err)
		}
	}

	img, err := remote.Image(ref, remote.WithAuthFromKeychain(keychain))
	if err != nil {
		return latest, err
	}

	ls, err := img.Layers()
	if err != nil {
		return latest, err
	}

	// Find the file with largest size among all layers.
	// This is the tar file with all minio relevant files.
	maxSizeHash, _ := ls[0].Digest()
	maxSize, _ := ls[0].Size()
	for i := range ls {
		s, _ := ls[i].Size()
		if s > maxSize {
			maxSize, _ = ls[i].Size()
			maxSizeHash, _ = ls[i].Digest()
		}
	}

	f, err := os.OpenFile(basePath+"image.tar", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return latest, err
	}
	defer func() {
		_ = f.Close()
	}()

	// Tarball writes a file called image.tar
	// This file in turn has each container layer present inside in the form `<layer-hash>.tar.gz`
	if err = tarball.Write(ref, img, f); err != nil {
		return latest, err
	}

	// Extract the <layer-hash>.tar.gz file that has minio contents from `image.tar`
	fileNameToExtract := strings.Split(maxSizeHash.String(), ":")[1] + ".tar.gz"
	if err = miniov1.ExtractTar([]string{fileNameToExtract}, basePath, "image.tar"); err != nil {
		return latest, err
	}

	// Extract the minio update related files (minio, minio.sha256sum and minio.minisig) from `<layer-hash>.tar.gz`
	if err = miniov1.ExtractTar([]string{"usr/bin/minio", "usr/bin/minio.sha256sum", "usr/bin/minio.minisig"}, basePath, fileNameToExtract); err != nil {
		return latest, err
	}

	srcBinary := "minio"
	srcShaSum := "minio.sha256sum"
	srcSig := "minio.minisig"

	tag, err := c.fetchTag(basePath + srcBinary)
	if err != nil {
		return latest, err
	}

	latest, err = miniov1.ReleaseTagToReleaseTime(tag)
	if err != nil {
		return latest, err
	}

	destBinary := "minio." + tag
	destShaSum := "minio." + tag + ".sha256sum"
	destSig := "minio." + tag + ".minisig"
	filesToRename := map[string]string{srcBinary: destBinary, srcShaSum: destShaSum, srcSig: destSig}

	// rename all files to add tag specific values in the name.
	// this is because minio updater looks for files in this name format.
	for s, d := range filesToRename {
		if err = os.Rename(basePath+s, basePath+d); err != nil {
			return latest, err
		}
	}
	return latest, nil
}

// Remove all the files created during upload process
func (c *Controller) removeArtifacts() error {
	return os.RemoveAll(updatePath)
}

// Start will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Start(threadiness int, stopCh <-chan struct{}) error {
	go func() {
		if err := c.ws.ListenAndServe(); err != http.ErrServerClosed {
			klog.Infof("HTTP server ListenAndServe: %v", err)
			return
		}
	}()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Tenant controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.statefulSetListerSynced, c.deploymentListerSynced, c.tenantsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process Tenant resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	return nil
}

// Stop is called to shutdown the controller
func (c *Controller) Stop() {
	klog.Info("Stopping the minio controller webhook")
	// Wait upto 5 secs and terminate all connections.
	tctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	_ = c.ws.Shutdown(tctx)
	cancel()

	klog.Info("Stopping the minio controller")
	c.workqueue.ShutDown()
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	defer runtime.HandleCrash()
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	processItem := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		klog.V(2).Infof("Key from workqueue: %s", key)
		// Run the syncHandler, passing it the namespace/name string of the
		// Tenant resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}

	if err := processItem(obj); err != nil {
		runtime.HandleError(err)
		return true
	}
	return true
}

const slashSeparator = "/"

func key2NamespaceName(key string) (namespace, name string) {
	key = strings.TrimPrefix(key, slashSeparator)
	m := strings.Index(key, slashSeparator)
	if m < 0 {
		return "", key
	}
	return key[:m], key[m+len(slashSeparator):]
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Tenant resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	ctx := context.Background()
	cOpts := metav1.CreateOptions{}
	uOpts := metav1.UpdateOptions{}
	gOpts := metav1.GetOptions{}

	var consoleDeployment *appsv1.Deployment

	// Convert the namespace/name string into a distinct namespace and name
	if key == "" {
		runtime.HandleError(fmt.Errorf("Invalid resource key: %s", key))
		return nil
	}

	namespace, tenantName := key2NamespaceName(key)

	// Get the Tenant resource with this namespace/name
	tenant, err := c.tenantsLister.Tenants(namespace).Get(tenantName)
	if err != nil {
		// The Tenant resource may no longer exist, in which case we stop processing.
		if k8serrors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("Tenant '%s' in work queue no longer exists", key))
			return nil
		}
		return nil
	}
	// Set any required default values and init Global variables
	nsName := types.NamespacedName{Namespace: namespace, Name: tenantName}

	tenant.EnsureDefaults()
	miniov1.InitGlobals(tenant)

	// Validate the MinIO Tenant
	if err = tenant.Validate(); err != nil {
		klog.V(2).Infof(err.Error())
		var err2 error
		if _, err2 = c.updateTenantStatus(ctx, tenant, err.Error(), 0); err2 != nil {
			klog.V(2).Infof(err2.Error())
		}
		return err
	}

	secret, err := c.applyOperatorWebhookSecret(ctx, tenant)
	if err != nil {
		return err
	}

	// check if both auto certificate creation and external secret with certificate is passed,
	// this is an error as only one of this is allowed in one Tenant
	if tenant.AutoCert() && (tenant.ExternalCert() || tenant.ExternalClientCert() || tenant.KESExternalCert() || tenant.ConsoleExternalCert()) {
		msg := "Please set either externalCertSecret or requestAutoCert in Tenant config"
		klog.V(2).Infof(msg)
		if _, err = c.updateTenantStatus(ctx, tenant, msg, 0); err != nil {
			klog.V(2).Infof(err.Error())
		}
		return fmt.Errorf(msg)
	}

	// TLS is mandatory if KES is enabled
	// AutoCert if enabled takes care of MinIO and KES certs
	if tenant.HasKESEnabled() && !tenant.AutoCert() {
		// if AutoCert is not enabled, user needs to provide external secrets for
		// KES and MinIO pods
		if !(tenant.ExternalCert() && tenant.ExternalClientCert() && tenant.KESExternalCert()) {
			msg := "Please provide certificate secrets for MinIO and KES, since automatic TLS is disabled"
			klog.V(2).Infof(msg)
			if _, err = c.updateTenantStatus(ctx, tenant, msg, 0); err != nil {
				klog.V(2).Infof(err.Error())
			}
			return fmt.Errorf(msg)
		}
	}

	// Handle the Internal ClusterIP Service for Tenant
	svc, err := c.serviceLister.Services(tenant.Namespace).Get(tenant.MinIOCIServiceName())
	if err != nil {
		if k8serrors.IsNotFound(err) {
			if tenant, err = c.updateTenantStatus(ctx, tenant, StatusProvisioningCIService, 0); err != nil {
				return err
			}
			klog.V(2).Infof("Creating a new Cluster IP Service for cluster %q", nsName)
			// Create the clusterIP service for the Tenant
			svc = services.NewClusterIPForMinIO(tenant)
			_, err = c.kubeClientSet.CoreV1().Services(tenant.Namespace).Create(ctx, svc, cOpts)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Handle the Internal Headless Service for Tenant StatefulSet
	hlSvc, err := c.serviceLister.Services(tenant.Namespace).Get(tenant.MinIOHLServiceName())
	if err != nil {
		if k8serrors.IsNotFound(err) {
			if tenant, err = c.updateTenantStatus(ctx, tenant, StatusProvisioningHLService, 0); err != nil {
				return err
			}
			klog.V(2).Infof("Creating a new Headless Service for cluster %q", nsName)
			// Create the headless service for the tenant
			hlSvc = services.NewHeadlessForMinIO(tenant)
			_, err = c.kubeClientSet.CoreV1().Services(tenant.Namespace).Create(ctx, hlSvc, cOpts)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// List all MinIO Tenants in this namespace.
	li, err := c.tenantsLister.Tenants(tenant.Namespace).List(labels.NewSelector())
	if err != nil {
		return err
	}

	// Only 1 minio tenant per namespace allowed.
	if len(li) > 1 {
		for _, t := range li {
			if t.Status.CurrentState != StatusInitialized {
				if _, err = c.updateTenantStatus(ctx, t, StatusFailedAlreadyExists, 0); err != nil {
					return err
				}
				return fmt.Errorf("Failed creating MinIO Tenant '%s' because another MinIO Tenant already exists in the namespace '%s'", t.Name, tenant.Namespace)
			}
		}
	}

	// For each zone check it's stateful set
	minioSecretName := tenant.Spec.CredsSecret.Name
	minioSecret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, minioSecretName, gOpts)
	if err != nil {
		return err
	}

	adminClnt, err := tenant.NewMinIOAdmin(minioSecret.Data)
	if err != nil {
		return err
	}

	// For each zone check if it's a stateful set
	var totalReplicas int32
	var images []string
	for _, zone := range tenant.Spec.Zones {
		// Get the StatefulSet with the name specified in Tenant.spec
		ss, err := c.statefulSetLister.StatefulSets(tenant.Namespace).Get(tenant.ZoneStatefulsetName(&zone))
		if err != nil {
			if k8serrors.IsNotFound(err) {
				// If auto cert is enabled, create certificates for MinIO and
				// optionally KES
				if tenant.AutoCert() {
					// Client cert is needed only with KES for mTLS authentication
					if err = c.checkAndCreateMinIOCSR(ctx, nsName, tenant, tenant.HasKESEnabled()); err != nil {
						return err
					}
					if tenant.HasKESEnabled() {
						if err = c.checkAndCreateKESCSR(ctx, nsName, tenant); err != nil {
							return err
						}
					}
					if tenant.HasConsoleEnabled() {
						if err = c.checkAndCreateConsoleCSR(ctx, nsName, tenant); err != nil {
							return err
						}
					}
				}
				if tenant, err = c.updateTenantStatus(ctx, tenant, StatusProvisioningStatefulSet, 0); err != nil {
					return err
				}

				ss = statefulsets.NewForMinIOZone(tenant, secret, &zone, hlSvc.Name, c.hostsTemplate, c.operatorVersion)
				ss, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Create(ctx, ss, cOpts)
				if err != nil {
					return err
				}

				// Restart the services to fetch the new args, ignore any error.
				_ = adminClnt.ServiceRestart(ctx)
			} else {
				return err
			}
		} else {
			if zone.Servers != *ss.Spec.Replicas {
				// warn the user that replica count of an existing zone can't be changed
				if tenant, err = c.updateTenantStatus(ctx, tenant, fmt.Sprintf("Can't modify server count for zone %s", zone.Name), 0); err != nil {
					return err
				}
			}

			if zone.Resources.String() != ss.Spec.Template.Spec.Containers[0].Resources.String() {
				if tenant, err = c.updateTenantStatus(ctx, tenant, StatusUpdatingResourceRequirements, ss.Status.Replicas); err != nil {
					return err
				}
				klog.V(4).Infof("resource requirements updates for zone %s", zone.Name)
				ss = statefulsets.NewForMinIOZone(tenant, secret, &zone, hlSvc.Name, c.hostsTemplate, c.operatorVersion)
				if ss, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Update(ctx, ss, uOpts); err != nil {
					return err
				}
			}

			if zone.Affinity.String() != ss.Spec.Template.Spec.Affinity.String() {
				if tenant, err = c.updateTenantStatus(ctx, tenant, StatusUpdatingAffinity, ss.Status.Replicas); err != nil {
					return err
				}
				klog.V(4).Infof("affinity update for zone %s", zone.Name)
				ss = statefulsets.NewForMinIOZone(tenant, secret, &zone, hlSvc.Name, c.hostsTemplate, c.operatorVersion)
				if ss, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Update(ctx, ss, uOpts); err != nil {
					return err
				}
			}

		}

		// If the StatefulSet is not controlled by this Tenant resource, we should log
		// a warning to the event recorder and ret
		if !metav1.IsControlledBy(ss, tenant) {
			if tenant, err = c.updateTenantStatus(ctx, tenant, StatusNotOwned, ss.Status.Replicas); err != nil {
				return err
			}
			msg := fmt.Sprintf(MessageResourceExists, ss.Name)
			c.recorder.Event(tenant, corev1.EventTypeWarning, ErrResourceExists, msg)
			return fmt.Errorf(msg)
		}

		// keep track of all replicas
		totalReplicas += ss.Status.Replicas
		images = append(images, ss.Spec.Template.Spec.Containers[0].Image)
	}

	// compare all the images across all zones, they should always be the same.
	for _, image := range images {
		for i := 0; i < len(images); i++ {
			if image != images[i] {
				if _, err = c.updateTenantStatus(ctx, tenant, StatusInconsistentMinIOVersions, totalReplicas); err != nil {
					return err
				}
				return fmt.Errorf("Zone %d is running incorrect image version, all zones are required to be on the same MinIO version. Attempting update of the inconsistent zone",
					i+1)
			}
		}
	}

	// In loop above we compared all the versions in all zones.
	// So comparing tenant.Spec.Image (version to update to) against one value from images slice is fine.
	if tenant.Spec.Image != images[0] && tenant.Status.CurrentState != StatusUpdatingMinIOVersion {
		if !tenant.MinIOHealthCheck() {
			return fmt.Errorf("MinIO doesn't seem to have enough quorum to proceed with binary update")
		}

		// Images different with the newer state change, continue to verify
		// if upgrade is possible
		tenant, err = c.updateTenantStatus(ctx, tenant, StatusUpdatingMinIOVersion, totalReplicas)
		if err != nil {
			return err
		}

		klog.V(4).Infof("Collecting artifacts for Tenant '%s' to update MinIO from: %s, to: %s",
			tenantName, images[0], tenant.Spec.Image)

		latest, err := c.fetchArtifacts(tenant)
		if err != nil {
			_ = c.removeArtifacts()
			return err
		}

		// FIXME: we can make operator TLS configurable here.
		updateURL, err := tenant.UpdateURL(latest, fmt.Sprintf("http://operator.%s.svc.%s:%s%s",
			miniov1.GetNSFromFile(), miniov1.ClusterDomain,
			miniov1.WebhookDefaultPort, miniov1.WebhookAPIUpdate,
		))
		if err != nil {
			_ = c.removeArtifacts()

			err = fmt.Errorf("Unable to get canonical update URL for Tenant '%s', failed with %v", tenantName, err)
			if _, terr := c.updateTenantStatus(ctx, tenant, err.Error(), totalReplicas); terr != nil {
				return terr
			}

			// Correct URL could not be obtained, not proceeding to update.
			return err
		}

		klog.V(4).Infof("Updating Tenant %s MinIO version from: %s, to: %s -> URL: %s",
			tenantName, tenant.Spec.Image, images[0], updateURL)

		us, err := adminClnt.ServerUpdate(ctx, updateURL)
		if err != nil {
			_ = c.removeArtifacts()

			err = fmt.Errorf("Tenant '%s' MinIO update failed with %w", tenantName, err)
			if _, terr := c.updateTenantStatus(ctx, tenant, err.Error(), totalReplicas); terr != nil {
				return terr
			}

			// Update failed, nothing needs to be changed in the container
			return err
		}

		if us.CurrentVersion != us.UpdatedVersion {
			klog.Infof("Tenant '%s' MinIO updated successfully from: %s, to: %s successfully",
				tenantName, us.CurrentVersion, us.UpdatedVersion)
		} else {
			msg := fmt.Sprintf("Tenant '%s' MinIO is already running the most recent version of %s",
				tenantName,
				us.CurrentVersion)
			klog.Info(msg)
			if _, terr := c.updateTenantStatus(ctx, tenant, msg, totalReplicas); terr != nil {
				return err
			}
			return nil
		}

		// clean the local directory
		_ = c.removeArtifacts()

		for _, zone := range tenant.Spec.Zones {
			// Now proceed to make the yaml changes for the tenant statefulset.
			ss := statefulsets.NewForMinIOZone(tenant, secret, &zone, hlSvc.Name, c.hostsTemplate, c.operatorVersion)
			if _, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Update(ctx, ss, uOpts); err != nil {
				return err
			}
		}

	}

	if tenant.HasConsoleEnabled() {
		// Get the Deployment with the name specified in MirrorInstace.spec
		if consoleDeployment, err = c.deploymentLister.Deployments(tenant.Namespace).Get(tenant.ConsoleDeploymentName()); err != nil {
			if k8serrors.IsNotFound(err) {
				if !tenant.HasCredsSecret() || !tenant.HasConsoleSecret() {
					msg := "Please set the credentials"
					klog.V(2).Infof(msg)
					return fmt.Errorf(msg)
				}

				consoleSecretName := tenant.Spec.Console.ConsoleSecret.Name
				consoleSecret, sErr := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, consoleSecretName, gOpts)
				if sErr != nil {
					return sErr
				}

				// Make sure that MinIO is up and running to enable MinIO console user.
				if tenant.MinIOHealthCheck() {
					if tenant, err = c.updateTenantStatus(ctx, tenant, StatusProvisioningConsoleDeployment, totalReplicas); err != nil {
						return err
					}

					if pErr := tenant.CreateConsoleUser(adminClnt, consoleSecret.Data); pErr != nil {
						klog.V(2).Infof(pErr.Error())
						return pErr
					}

					// Create Console Deployment
					consoleDeployment = deployments.NewConsole(tenant)
					_, err = c.kubeClientSet.AppsV1().Deployments(tenant.Namespace).Create(ctx, consoleDeployment, cOpts)
					if err != nil {
						klog.V(2).Infof(err.Error())
						return err
					}
					// Create Console service
					consoleSvc := services.NewClusterIPForConsole(tenant)
					_, err = c.kubeClientSet.CoreV1().Services(svc.Namespace).Create(ctx, consoleSvc, cOpts)
					if err != nil {
						klog.V(2).Infof(err.Error())
						return err
					}
				} else {
					if tenant, err = c.updateTenantStatus(ctx, tenant, StatusWaitingForReadyState, totalReplicas); err != nil {
						return err
					}
				}
			} else {
				return err
			}
		} else {
			if consoleDeployment != nil && !tenant.Spec.Console.EqualImage(consoleDeployment.Spec.Template.Spec.Containers[0].Image) {
				if tenant, err = c.updateTenantStatus(ctx, tenant, StatusUpdatingConsoleVersion, totalReplicas); err != nil {
					return err
				}
				klog.V(2).Infof("Updating Tenant %s console version %s, to: %s", tenantName,
					tenant.Spec.Console.Image, consoleDeployment.Spec.Template.Spec.Containers[0].Image)
				consoleDeployment = deployments.NewConsole(tenant)
				_, err = c.kubeClientSet.AppsV1().Deployments(tenant.Namespace).Update(ctx, consoleDeployment, uOpts)
				// If an error occurs during Update, we'll requeue the item so we can
				// attempt processing again later. This could have been caused by a
				// temporary network failure, or any other transient reason.
				if err != nil {
					return err
				}
			}
		}
	}

	if tenant.HasKESEnabled() && (tenant.AutoCert() || tenant.ExternalCert()) {
		if tenant.ExternalClientCert() {
			// Since we're using external secret, store the identity for later use
			miniov1.KESIdentity, err = c.getCertIdentity(tenant.Namespace, tenant.Spec.ExternalClientCertSecret)
			if err != nil {
				return err
			}
		}

		svc, err := c.serviceLister.Services(tenant.Namespace).Get(tenant.KESHLServiceName())
		if err != nil {
			if k8serrors.IsNotFound(err) {
				klog.V(2).Infof("Creating a new Headless Service for cluster %q", nsName)
				svc = services.NewHeadlessForKES(tenant)
				if _, err = c.kubeClientSet.CoreV1().Services(svc.Namespace).Create(ctx, svc, cOpts); err != nil {
					return err
				}
			} else {
				return err
			}
		}

		// Get the StatefulSet with the name specified in spec
		_, err = c.statefulSetLister.StatefulSets(tenant.Namespace).Get(tenant.KESStatefulSetName())
		if err != nil {
			if k8serrors.IsNotFound(err) {
				if tenant, err = c.updateTenantStatus(ctx, tenant, StatusProvisioningKESStatefulSet, 0); err != nil {
					return err
				}

				ks := statefulsets.NewForKES(tenant, svc.Name)
				klog.V(2).Infof("Creating a new StatefulSet for cluster %q", nsName)
				if _, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Create(ctx, ks, cOpts); err != nil {
					klog.V(2).Infof(err.Error())
					return err
				}
			} else {
				return err
			}
		}

		// After KES and MinIO are deployed successfully, create the MinIO Key on KES KMS Backend
		_, err = c.jobLister.Jobs(tenant.Namespace).Get(tenant.KESJobName())
		if err != nil {
			if k8serrors.IsNotFound(err) {
				j := jobs.NewForKES(tenant)
				klog.V(2).Infof("Creating a new Job for cluster %q", nsName)
				if _, err = c.kubeClientSet.BatchV1().Jobs(tenant.Namespace).Create(ctx, j, cOpts); err != nil {
					klog.V(2).Infof(err.Error())
					return err
				}
			} else {
				return err
			}
		}
	}

	// Finally, we update the status block of the Tenant resource to reflect the
	// current state of the world
	_, err = c.updateTenantStatus(ctx, tenant, StatusInitialized, totalReplicas)
	return err
}

func (c *Controller) checkAndCreateMinIOCSR(ctx context.Context, nsName types.NamespacedName, tenant *miniov1.Tenant, createClientCert bool) error {
	if _, err := c.certClient.CertificateSigningRequests().Get(ctx, tenant.MinIOCSRName(), metav1.GetOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			if tenant, err = c.updateTenantStatus(ctx, tenant, StatusWaitingMinIOCert, 0); err != nil {
				return err
			}
			klog.V(2).Infof("Creating a new Certificate Signing Request for MinIO Server Certs, cluster %q", nsName)
			if err = c.createCSR(ctx, tenant); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if createClientCert {
		if _, err := c.certClient.CertificateSigningRequests().Get(ctx, tenant.MinIOClientCSRName(), metav1.GetOptions{}); err != nil {
			if k8serrors.IsNotFound(err) {
				if tenant, err = c.updateTenantStatus(ctx, tenant, StatusWaitingMinIOClientCert, 0); err != nil {
					return err
				}
				klog.V(2).Infof("Creating a new Certificate Signing Request for MinIO Client Certs, cluster %q", nsName)
				if err = c.createMinIOClientTLSCSR(ctx, tenant); err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}
	return nil
}

func (c *Controller) checkAndCreateKESCSR(ctx context.Context, nsName types.NamespacedName, tenant *miniov1.Tenant) error {
	if _, err := c.certClient.CertificateSigningRequests().Get(ctx, tenant.KESCSRName(), metav1.GetOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			if tenant, err = c.updateTenantStatus(ctx, tenant, StatusWaitingKESCert, 0); err != nil {
				return err
			}
			klog.V(2).Infof("Creating a new Certificate Signing Request for KES Server Certs, cluster %q", nsName)
			if err = c.createKESTLSCSR(ctx, tenant); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func (c *Controller) getCertIdentity(ns string, cert *miniov1.LocalCertificateReference) (string, error) {
	var certbytes []byte
	secret, err := c.kubeClientSet.CoreV1().Secrets(ns).Get(context.Background(), cert.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	// Store the Identity to be used later during KES container creation
	if secret.Type == "kubernetes.io/tls" || secret.Type == "cert-manager.io/v1alpha2" {
		certbytes = secret.Data["tls.crt"]
	} else {
		certbytes = secret.Data["public.crt"]
	}

	// parse the certificate here to generate the identity for this certifcate.
	// This is later used to update the identity in KES Server Config File
	h := sha256.New()
	parsedCert, err := parseCertificate(bytes.NewReader(certbytes))
	if err != nil {
		klog.Errorf("Unexpected error during the parsing the secret/%s: %v", cert.Name, err)
		return "", err
	}

	_, err = h.Write(parsedCert.RawSubjectPublicKeyInfo)
	if err != nil {
		klog.Errorf("Unexpected error during the parsing the secret/%s: %v", cert.Name, err)
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func (c *Controller) updateTenantStatus(ctx context.Context, tenant *miniov1.Tenant, currentState string, availableReplicas int32) (*miniov1.Tenant, error) {
	opts := metav1.UpdateOptions{}
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	tenantCopy := tenant.DeepCopy()
	tenantCopy.Status.AvailableReplicas = availableReplicas
	tenantCopy.Status.CurrentState = currentState
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Tenant resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	t, err := c.minioClientSet.MinioV1().Tenants(tenant.Namespace).UpdateStatus(ctx, tenantCopy, opts)
	t.EnsureDefaults()
	return t, err
}

// enqueueTenant takes a Tenant resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Tenant.
func (c *Controller) enqueueTenant(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the Tenant resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that Tenant resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		klog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	klog.V(4).Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a Tenant, we should not do anything more
		// with it.
		if ownerRef.Kind != "Tenant" {
			return
		}

		tenant, err := c.tenantsLister.Tenants(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			klog.V(4).Infof("ignoring orphaned object '%s' of tenant '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueTenant(tenant)
		return
	}
}

func (c *Controller) checkAndCreateConsoleCSR(ctx context.Context, nsName types.NamespacedName, tenant *miniov1.Tenant) error {
	if _, err := c.certClient.CertificateSigningRequests().Get(ctx, tenant.ConsoleCSRName(), metav1.GetOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			if tenant, err = c.updateTenantStatus(ctx, tenant, StatusWaitingConsoleCert, 0); err != nil {
				return err
			}
			klog.V(2).Infof("Creating a new Certificate Signing Request for Console Server Certs, cluster %q", nsName)
			if err = c.createConsoleTLSCSR(ctx, tenant); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}
