// Copyright (C) 2020-2023 MinIO, Inc.
//
// This code is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License, version 3,
// as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License, version 3,
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/minio/madmin-go/v2"
	"github.com/minio/operator/pkg/common"
	xcerts "github.com/minio/pkg/certs"

	"github.com/minio/minio-go/v7/pkg/set"
	"github.com/minio/operator/pkg/controller/certificates"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/klog/v2"

	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"

	miniov1 "github.com/minio/operator/pkg/apis/minio.min.io/v1"

	"golang.org/x/time/rate"

	// Workaround for auth import issues refer https://github.com/minio/operator/issues/283
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"

	promclientset "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	queue "k8s.io/client-go/util/workqueue"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	clientset "github.com/minio/operator/pkg/client/clientset/versioned"
	minioscheme "github.com/minio/operator/pkg/client/clientset/versioned/scheme"
	informers "github.com/minio/operator/pkg/client/informers/externalversions/minio.min.io/v2"
	stsInformers "github.com/minio/operator/pkg/client/informers/externalversions/sts.min.io/v1alpha1"
	"github.com/minio/operator/pkg/resources/services"
	"github.com/minio/operator/pkg/resources/statefulsets"
)

const (
	controllerAgentName = "minio-operator"
	// ErrResourceExists is used as part of the Event 'reason' when a Tenant fails
	// to sync due to a StatefulSet of the same name already existing.
	ErrResourceExists = "ErrResourceExists"
	// MessageResourceExists is the message used for Events when a Tenant
	// fails to sync due to a StatefulSet already existing
	MessageResourceExists = "Resource %q already exists and is not managed by MinIO Operator"
)

// Standard Status messages for Tenant
const (
	StatusInitialized                = "Initialized"
	StatusProvisioningCIService      = "Provisioning MinIO Cluster IP Service"
	StatusProvisioningHLService      = "Provisioning MinIO Headless Service"
	StatusProvisioningStatefulSet    = "Provisioning MinIO Statefulset"
	StatusProvisioningConsoleService = "Provisioning Console Service"
	StatusProvisioningKESStatefulSet = "Provisioning KES StatefulSet"
	StatusProvisioningInitialUsers   = "Provisioning initial users"
	StatusProvisioningDefaultBuckets = "Provisioning default buckets"
	StatusWaitingMinIOCert           = "Waiting for MinIO TLS Certificate"
	StatusWaitingMinIOClientCert     = "Waiting for MinIO TLS Client Certificate"
	StatusWaitingKESCert             = "Waiting for KES TLS Certificate"
	StatusUpdatingMinIOVersion       = "Updating MinIO Version"
	StatusUpdatingKES                = "Updating KES"
	StatusNotOwned                   = "Statefulset not controlled by operator"
	StatusFailedAlreadyExists        = "Another MinIO Tenant already exists in the namespace"
	StatusTenantCredentialsNotSet    = "Tenant credentials are not set properly"
	StatusInconsistentMinIOVersions  = "Different versions across MinIO Pools"
	StatusRestartingMinIO            = "Restarting MinIO"
	StatusDecommissioningNotAllowed  = "Pool Decommissioning Not Allowed"
)

// ErrMinIONotReady is the error returned when MinIO is not Ready
var ErrMinIONotReady = fmt.Errorf("MinIO is not ready")

// ErrMinIORestarting is the error returned when MinIO is restarting
var ErrMinIORestarting = fmt.Errorf("MinIO is restarting")

// Controller struct watches the Kubernetes API for changes to Tenant resources
type Controller struct {
	// podName is the identifier of this instance
	podName string
	// namespacesToWatch restricts the action of the opreator to a list of namespaces
	namespacesToWatch set.StringSet
	// kubeClientSet is a standard kubernetes clientset
	kubeClientSet kubernetes.Interface
	// minioClientSet is a clientset for our own API group
	minioClientSet clientset.Interface
	// promClient is a clientset for Prometheus service monitor
	promClient promclientset.Interface
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

	// HTTP Upgrade server instance
	us *http.Server

	// STS API server instance
	sts *http.Server

	// Client transport
	transport *http.Transport

	// monitor pods in the cluster to update the health information
	podInformer cache.SharedIndexInformer

	// healthCheckQueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	healthCheckQueue queue.RateLimitingInterface
	// image being used in the operator deployment
	operatorImage string

	// policyBindingListerSynced returns true if the PolicyBinding shared informer
	// has synced at least once.
	policyBindingListerSynced cache.InformerSynced
}

// EventType is Event type to handle
type EventType int

// Possible values of EventType
const (
	STSServerNotification EventType = iota
	LeaderElection
)

// EventNotification - structure to send messages trough a channel regarding a error event to be handled
type EventNotification struct {
	// Err the error to handle if any, null when is just a message
	Err error
	// Type the event type to handle
	Type EventType
}

// NewController returns a new sample controller
func NewController(podName string, namespacesToWatch set.StringSet, kubeClientSet kubernetes.Interface, minioClientSet clientset.Interface, promClient promclientset.Interface, statefulSetInformer appsinformers.StatefulSetInformer, deploymentInformer appsinformers.DeploymentInformer, podInformer coreinformers.PodInformer, tenantInformer informers.TenantInformer, policyBindingInformer stsInformers.PolicyBindingInformer, serviceInformer coreinformers.ServiceInformer, hostsTemplate, operatorVersion string) *Controller {
	// Create event broadcaster
	// Add minio-controller types to the default Kubernetes Scheme so Events can be
	// logged for minio-controller types.
	minioscheme.AddToScheme(scheme.Scheme) //nolint:errcheck
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClientSet.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	// get operator deployment name
	ns := miniov2.GetNSFromFile()
	ctx := context.Background()
	oprImg := DefaultOperatorImage
	oprDep, err := kubeClientSet.AppsV1().Deployments(ns).Get(ctx, DefaultDeploymentName, metav1.GetOptions{})
	if err == nil && oprDep != nil {
		// assume we are the first container, just in case they changed the default name
		if len(oprDep.Spec.Template.Spec.Containers) > 0 {
			oprImg = oprDep.Spec.Template.Spec.Containers[0].Image
		}
		// attempt to iterate in case there's multiple containers
		for _, c := range oprDep.Spec.Template.Spec.Containers {
			if c.Name == "minio-operator" || c.Name == "operator" {
				oprImg = c.Image
			}
		}
	}

	controller := &Controller{
		podName:                   podName,
		namespacesToWatch:         namespacesToWatch,
		kubeClientSet:             kubeClientSet,
		minioClientSet:            minioClientSet,
		promClient:                promClient,
		statefulSetLister:         statefulSetInformer.Lister(),
		statefulSetListerSynced:   statefulSetInformer.Informer().HasSynced,
		podInformer:               podInformer.Informer(),
		deploymentLister:          deploymentInformer.Lister(),
		deploymentListerSynced:    deploymentInformer.Informer().HasSynced,
		tenantsSynced:             tenantInformer.Informer().HasSynced,
		serviceLister:             serviceInformer.Lister(),
		serviceListerSynced:       serviceInformer.Informer().HasSynced,
		workqueue:                 queue.NewNamedRateLimitingQueue(MinIOControllerRateLimiter(), "Tenants"),
		healthCheckQueue:          queue.NewNamedRateLimitingQueue(MinIOControllerRateLimiter(), "TenantsHealth"),
		recorder:                  recorder,
		hostsTemplate:             hostsTemplate,
		operatorVersion:           operatorVersion,
		policyBindingListerSynced: policyBindingInformer.Informer().HasSynced,
		operatorImage:             oprImg,
	}

	// Initialize operator HTTP upgrade server handlers
	controller.us = configureHTTPUpgradeServer()

	// Initialize STS API server handlers
	controller.sts = configureSTSServer(controller)

	klog.Info("Setting up event handlers")
	// Set up an event handler for when Tenant resources change
	tenantInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueTenant,
		UpdateFunc: func(old, new interface{}) {
			oldTenant := old.(*miniov2.Tenant)
			newTenant := new.(*miniov2.Tenant)
			if newTenant.ResourceVersion == oldTenant.ResourceVersion {
				// Periodic resync will send update events for all known Tenants.
				// Two different versions of the same Tenant will always have different RVs.
				return
			}
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

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handlePodChange,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*corev1.Pod)
			oldDepl := old.(*corev1.Pod)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				// Periodic resync will send update events for all known Deployments.
				// Two different versions of the same Deployments will always have different RVs.
				return
			}
			controller.handlePodChange(new)
		},
		DeleteFunc: controller.handlePodChange,
	})

	return controller
}

// startUpgradeServer Starts the Upgrade tenant API server and notifies the start and stop via notificationChannel returned
func (c *Controller) startUpgradeServer(ctx context.Context) <-chan error {
	notificationChannel := make(chan error)
	go func() {
		defer close(notificationChannel)
		klog.Infof("Starting HTTP Upgrade Tenant Image server")
		if err := c.us.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			// only notify on server failure, on http.ErrServerClosed the channel should be already closed
			notificationChannel <- err
		}
	}()
	return notificationChannel
}

// startUpgradeServer Starts the Upgrade tenant API server and notifies the start and stop via notificationChannel
func (c *Controller) startSTSAPIServer(ctx context.Context, notificationChannel chan<- *EventNotification) {
	klog.Infof("Starting STS API server")
	publicCertPath, publicKeyPath := c.waitSTSTLSCert()
	certsManager, err := xcerts.NewManager(ctx, *publicCertPath, *publicKeyPath, LoadX509KeyPair)
	if err != nil {
		klog.Errorf("HTTPS STS API server failed to load CA certificate: %v", err)
		notificationChannel <- &EventNotification{
			Type: STSServerNotification,
			Err:  err,
		}
	}
	serverCertsManager = certsManager
	c.sts.TLSConfig = c.createTLSConfig(serverCertsManager)
	if err := c.sts.ListenAndServeTLS("", ""); !errors.Is(err, http.ErrServerClosed) {
		// only notify on server failure, on http.ErrServerClosed the channel should be already closed
		notificationChannel <- &EventNotification{
			Type: STSServerNotification,
			Err:  err,
		}
	}
}

// leaderRun start the Controller and the API's
// When a new leader is elected this function is ran in the pod
func leaderRun(ctx context.Context, c *Controller, threadiness int, stopCh <-chan struct{}) {
	// we declate the channel to communicate on servers errors
	var upgradeServerChannel <-chan error

	klog.Info("Waiting for Upgrade Server to start")
	upgradeServerChannel = c.startUpgradeServer(ctx)

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Tenant controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.statefulSetListerSynced, c.deploymentListerSynced, c.tenantsSynced, c.policyBindingListerSynced); !ok {
		panic("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process Tenant resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	// Launch a single worker for Health Check reacting to Pod Changes
	go wait.Until(c.runHealthCheckWorker, time.Second, stopCh)

	// Launch a goroutine to monitor all Tenants
	go c.recurrentTenantStatusMonitor(stopCh)

	// 1) we need to make sure we have console TLS certificates (if enabled)
	if isOperatorConsoleTLS() {
		klog.Info("Waiting for Console TLS")
		go func() {
			klog.Infof("Console TLS is enabled, starting console TLS certificate setup")
			err := c.recreateOperatorConsoleCertsIfRequired(ctx)
			if err != nil {
				panic(err)
			}
			klog.Infof("Restarting Console pods")
			err = c.rolloutRestartDeployment(getConsoleDeploymentName())
			if err != nil {
				klog.Errorf("Console deployment didn't restart: %s", err)
			}
		}()
	} else {
		klog.Infof("Console TLS is not enabled")
	}

	// 2) we need to make sure we have STS API certificates (if enabled)
	if IsSTSEnabled() {
		go func() {
			klog.Infof("STS is enabled, starting STS API certificate setup")
			c.generateSTSTLSCert()
		}()
	}

	for {
		select {
		case err := <-upgradeServerChannel:
			if err != http.ErrServerClosed {
				klog.Errorf("Upgrade Server stopped: %v, going to restart", err)
				upgradeServerChannel = c.startUpgradeServer(ctx)
			}
			// webswerver was instructed to stop, do not attempt to restart
			continue
		case <-stopCh:
			return
		}
	}
}

// Start will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Start(threadiness int, stopCh <-chan struct{}) error {
	// use a Go context so we can tell the leaderelection code when we
	// want to step down
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// listen for interrupts or the Linux SIGTERM signal and cancel
	// our context, which the leader election code will observe and
	// step down
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		klog.Info("Received termination, signaling shutdown")
		cancel()
	}()

	leaseLockName := "minio-operator-lock"
	leaseLockNamespace := miniov2.GetNSFromFile()

	// notificationChannel is a channel to notify errors or events
	notificationChannel := make(chan *EventNotification)
	defer close(notificationChannel)

	// Request kubernetes version from Kube ApiServer
	apiCsrVersion := certificates.GetCertificatesAPIVersion(c.kubeClientSet)
	klog.Infof("Using Kubernetes CSR Version: %s", apiCsrVersion)

	// we use the Lease lock type since edits to Leases are less common
	// and fewer objects in the cluster watch "all Leases".
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      leaseLockName,
			Namespace: leaseLockNamespace,
		},
		Client: c.kubeClientSet.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: c.podName,
		},
	}

	if IsSTSEnabled() {
		// runSTS starts the STS API even if the pod is not the leader
		klog.Info("Waiting for STS API to start")
		go c.startSTSAPIServer(ctx, notificationChannel)
	} else {
		klog.Info("STS Api server is not enabled, not starting")
	}

	go func() {
		// start the leader election code loop
		leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
			Lock: lock,
			// IMPORTANT: you MUST ensure that any code you have that
			// is protected by the lease must terminate **before**
			// you call cancel. Otherwise, you could have a background
			// loop still running and another process could
			// get elected before your background loop finished, violating
			// the stated goal of the lease.
			ReleaseOnCancel: true,
			LeaseDuration:   60 * time.Second,
			RenewDeadline:   15 * time.Second,
			RetryPeriod:     5 * time.Second,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(ctx context.Context) {
					// start the controller + API code
					leaderRun(ctx, c, threadiness, stopCh)
				},
				OnStoppedLeading: func() {
					// we can do cleanup here
					klog.Infof("leader lost: %s", c.podName)
				},
				OnNewLeader: func(identity string) {
					// we're notified when new leader elected
					if identity == c.podName {
						klog.Infof("%s: I am the leader, applying leader labels on myself", c.podName)
						// Patch this pod so the main service uses it
						p := []patchAnnotation{{
							Op:    "add",
							Path:  "/metadata/labels/operator",
							Value: "leader",
						}}

						payloadBytes, err := json.Marshal(p)
						if err != nil {
							klog.Errorf("failed to marshal patch: %#v", err)
						} else {
							_, err = c.kubeClientSet.CoreV1().Pods(leaseLockNamespace).Patch(ctx, c.podName, types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
							if err != nil {
								klog.Errorf("failed to patch operator leader pod: %+v", err)
							}
						}
					} else {
						klog.Infof("%s: is the leader, removing any leader labels that I '%s' might have", identity, c.podName)
						// Patch this pod so the main service uses it
						p := []patchAnnotation{{
							Op:   "remove",
							Path: "/metadata/labels/operator",
						}}

						payloadBytes, err := json.Marshal(p)
						if err != nil {
							klog.Errorf("failed to marshal patch: %#v", err)
						} else {
							c.kubeClientSet.CoreV1().Pods(leaseLockNamespace).Patch(ctx, c.podName, types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
						}
					}
				},
			},
		})

		notificationChannel <- &EventNotification{
			Type: LeaderElection,
			Err:  nil,
		}
	}()

	for oerr := range notificationChannel {
		switch oerr.Type {
		case STSServerNotification:
			if !errors.Is(oerr.Err, http.ErrServerClosed) {
				klog.Errorf("STS API Server stopped: %v, going to restart", oerr.Err)
				go c.startSTSAPIServer(ctx, notificationChannel)
			}
		case LeaderElection:
			return nil
		}
	}
	return nil
}

// Stop is called to shutdown the controller
func (c *Controller) Stop() {
	klog.Info("Stopping the minio controller webservers")
	// Wait upto 5 secs and terminate all connections.
	tctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	_ = c.us.Shutdown(tctx)
	_ = c.sts.Shutdown(tctx)
	cancel()

	klog.Info("Stopping the minio controller")
	c.workqueue.ShutDown()
	c.healthCheckQueue.ShutDown()
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	defer runtime.HandleCrash()
	for c.processNextWorkItem() {
	}
}

// runHealthCheckWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// healthCheckQueue.
func (c *Controller) runHealthCheckWorker() {
	defer runtime.HandleCrash()
	for c.processNextHealthCheckItem() {
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

		result, err := c.syncHandler(key)
		switch {
		case err != nil:
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		case result.RequeueAfter > 0:
			// The result.RequeueAfter request will be lost, if it is returned
			// along with a non-nil error. But this is intended as
			// We need to drive to stable reconcile loops before queuing due
			// to result.RequestAfter
			c.workqueue.Forget(obj)
			c.workqueue.AddAfter(key, result.RequeueAfter)
		case result.Requeue:
			c.workqueue.AddRateLimited(key)
		default:
			// Finally, if no error occurs we Forget this item so it does not
			// get queued again until another change happens.
			c.workqueue.Forget(obj)
			klog.V(4).Infof("Successfully synced '%s'", key)
		}
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
func (c *Controller) syncHandler(key string) (Result, error) {
	ctx := context.Background()
	cOpts := metav1.CreateOptions{}
	uOpts := metav1.UpdateOptions{}

	// Convert the namespace/name string into a distinct namespace and name
	if key == "" {
		runtime.HandleError(fmt.Errorf("Invalid resource key: %s", key))
		return WrapResult(Result{}, nil)
	}

	namespace, tenantName := key2NamespaceName(key)

	// Get the Tenant resource with this namespace/name
	tenant, err := c.minioClientSet.MinioV2().Tenants(namespace).Get(context.Background(), tenantName, metav1.GetOptions{})
	if err != nil {
		// The Tenant resource may no longer exist, in which case we stop processing.
		if k8serrors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("Tenant '%s' in work queue no longer exists", key))
			return WrapResult(Result{}, nil)
		}
		// will retry after 5sec
		return WrapResult(Result{RequeueAfter: time.Second * 5}, nil)
	}

	// Check the Sync Version to see if the tenant needs upgrade
	if tenant, err = c.checkForUpgrades(ctx, tenant); err != nil {
		return WrapResult(Result{}, err)
	}

	// Set any required default values and init Global variables
	nsName := types.NamespacedName{Namespace: namespace, Name: tenantName}

	// get combined configurations (tenant.env, tenant.credsSecret and tenant.Configuration) for tenant
	tenantConfiguration, err := c.getTenantCredentials(ctx, tenant)
	if err != nil {
		if errors.Is(err, ErrEmptyRootCredentials) {
			if _, err2 := c.updateTenantStatus(ctx, tenant, err.Error(), 0); err2 != nil {
				klog.V(2).Infof(err2.Error())
			}
			c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "MissingCreds", "Tenant is missing root credentials")
			return WrapResult(Result{}, nil)
		}
		return WrapResult(Result{}, err)
	}
	// get existing configuration from config.env
	skipEnvVars, err := c.getTenantConfiguration(ctx, tenant)
	if err != nil {
		return WrapResult(Result{}, err)
	}

	// Check if we are decommissioning a pool before we ensure defaults, as that would populate a defaulted pool name
	tenant, err = c.checkForPoolDecommission(ctx, key, tenant, tenantConfiguration)
	if err != nil {
		return WrapResult(Result{}, err)
	}

	tenant.EnsureDefaults()

	// Validate the MinIO Tenant
	if err = tenant.Validate(); err != nil {
		klog.V(2).Infof(err.Error())
		var err2 error
		if _, err2 = c.updateTenantStatus(ctx, tenant, err.Error(), 0); err2 != nil {
			klog.V(2).Infof(err2.Error())
		}
		// return nil so we don't re-queue this work item
		return WrapResult(Result{}, nil)
	}

	// AutoCertEnabled verification is used to manage the tenant migration between v1 and v2
	// Previous behavior was that AutoCert is disabled by default if RequestAutoCert is nil
	// New behavior is that AutoCert is enabled by default if RequestAutoCert is nil
	// In the future this support will be dropped
	if tenant.Status.Certificates.AutoCertEnabled == nil {
		autoCertEnabled := true
		if tenant.Spec.RequestAutoCert == nil && tenant.APIVersion != "" {
			// If we get certificate signing requests for MinIO is safe to assume the Tenant v1 was deployed using AutoCert
			// otherwise AutoCert will be false
			if certificates.GetCertificatesAPIVersion(c.kubeClientSet) == certificates.CSRV1 {
				tenantCSR, err := c.kubeClientSet.CertificatesV1().CertificateSigningRequests().Get(ctx, tenant.MinIOCSRName(), metav1.GetOptions{})
				if err != nil || tenantCSR == nil {
					autoCertEnabled = false
				}
			} else {
				tenantCSR, err := c.kubeClientSet.CertificatesV1beta1().CertificateSigningRequests().Get(ctx, tenant.MinIOCSRName(), metav1.GetOptions{})
				if err != nil || tenantCSR == nil {
					autoCertEnabled = false
				}
			}
		} else {
			autoCertEnabled = tenant.AutoCert()
		}
		if tenant, err = c.updateCertificatesStatus(ctx, tenant, autoCertEnabled); err != nil {
			klog.V(2).Infof(err.Error())
		}
	}

	// Custom certificates
	if customCertificates, err := c.getCustomCertificates(ctx, tenant); err == nil {
		if newTenant, err := c.updateCustomCertificatesStatus(ctx, tenant, customCertificates); err != nil {
			klog.V(2).Infof(err.Error())
		} else {
			// Only change tenant if there was no error, otherwise tenant is being deleted
			tenant = newTenant
		}
	} else {
		klog.V(2).Infof(err.Error())
	}

	// validate the minio certificates
	err = c.checkMinIOCertificatesStatus(ctx, tenant, nsName)
	if err != nil {
		klog.V(2).Infof("Error when consolidating tenant service: %v", err)
		// will retry after 5sec
		return WrapResult(Result{RequeueAfter: time.Second * 5}, nil)
	}

	// validate services
	// Check MinIO S3 Endpoint Service
	err = c.checkMinIOSvc(ctx, tenant, nsName)
	if err != nil {
		klog.V(2).Infof("error consolidating minio service: %s", err.Error())
		return WrapResult(Result{}, err)
	}
	// Check Console Endpoint Service
	err = c.checkConsoleSvc(ctx, tenant, nsName)
	if err != nil {
		klog.V(2).Infof("error consolidating console service: %s", err.Error())
		return WrapResult(Result{}, err)
	}

	// Handle the Internal Headless Service for Tenant StatefulSet
	hlSvc, err := c.serviceLister.Services(tenant.Namespace).Get(tenant.MinIOHLServiceName())
	if err != nil {
		if k8serrors.IsNotFound(err) {
			if tenant, err = c.updateTenantStatus(ctx, tenant, StatusProvisioningHLService, 0); err != nil {
				return WrapResult(Result{}, err)
			}
			klog.V(2).Infof("Creating a new Headless Service for cluster %q", nsName)
			// Create the headless service for the tenant
			hlSvc = services.NewHeadlessForMinIO(tenant)
			_, err = c.kubeClientSet.CoreV1().Services(tenant.Namespace).Create(ctx, hlSvc, cOpts)
			if err != nil {
				return WrapResult(Result{}, err)
			}
			c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "SvcCreated", "Headless Service created")
		} else {
			return WrapResult(Result{}, err)
		}
	}

	// List all MinIO Tenants in this namespace.
	li, err := c.minioClientSet.MinioV2().Tenants(tenant.Namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return WrapResult(Result{}, err)
	}

	// Only 1 minio tenant per namespace allowed.
	if len(li.Items) > 1 {
		for _, t := range li.Items {
			if t.Status.CurrentState != StatusInitialized {
				if _, err = c.updateTenantStatus(ctx, &t, StatusFailedAlreadyExists, 0); err != nil {
					return WrapResult(Result{}, err)
				}
				// return nil so we don't re-queue this work item
				return WrapResult(Result{}, err)
			}
		}
	}
	// Create Tenant Services Accoutns for Tenant
	err = c.checkAndCreateServiceAccount(ctx, tenant)
	if err != nil {
		return WrapResult(Result{}, err)
	}

	adminClnt, err := tenant.NewMinIOAdmin(tenantConfiguration, c.getTransport())
	if err != nil {
		if _, uerr := c.updateTenantStatus(ctx, tenant, StatusTenantCredentialsNotSet, 0); uerr != nil {
			return WrapResult(Result{}, uerr)
		}
		klog.Errorf("Error initializing minio admin client: %v", err)
		return WrapResult(Result{}, err)
	}

	// For each pool check if there is a stateful set
	var totalAvailableReplicas int32
	var images []string

	err = c.checkKESStatus(ctx, tenant, totalAvailableReplicas, cOpts, uOpts, nsName)
	if err != nil {
		klog.V(2).Infof("Error checking KES state %v", err)
		return WrapResult(Result{}, err)
	}

	// check if operator-ca-tls has to be updated or re-created in the tenant namespace
	operatorCATLSExists, err := c.checkOperatorCaForTenant(ctx, tenant)
	if err != nil {
		return WrapResult(Result{}, err)
	}

	// consolidate the status of all pools. this is meant to cover for legacy tenants
	// this status value is zero only for new tenants or legacy tenants
	if len(tenant.Status.Pools) == 0 {
		pools, err := c.getAllSSForTenant(tenant)
		if err != nil {
			return WrapResult(Result{}, err)
		}
		for _, pool := range pools {
			if pool != nil {
				tenant.Status.Pools = append(tenant.Status.Pools, miniov2.PoolStatus{
					SSName: pool.Name,
					State:  miniov2.PoolCreated,
				})
			}
		}
		// push updates to status
		if tenant, err = c.updatePoolStatus(ctx, tenant); err != nil {
			return WrapResult(Result{}, err)
		}

		klog.Info("Detected we are updating a legacy tenant deployment")
	}

	// Check if this is fresh setup not an expansion.
	// addingNewPool := len(tenant.Spec.Pools) == len(tenant.Status.Pools)
	addingNewPool := false
	// count the number of initialized pools, if at least 1 is not Initialized, we are still adding a new pool
	for _, poolStatus := range tenant.Status.Pools {
		if poolStatus.State != miniov2.PoolInitialized {
			addingNewPool = true
			break
		}
	}
	if addingNewPool {
		klog.Infof("%s Detected we are adding a new pool", key)
	}

	// Check if we need to create any of the pools. It's important not to update the statefulsets
	// in this loop because we need all the pools "as they are" for the hot-update below
	for i, pool := range tenant.Spec.Pools {
		// Get the StatefulSet with the name specified in Tenant.status.pools[i].SSName

		// if this index is in the status of pools use it, else capture the desired name in the status and store it
		var ssName string
		if len(tenant.Status.Pools) > i {
			ssName = tenant.Status.Pools[i].SSName
		} else {
			ssName = tenant.PoolStatefulsetName(&pool)
			tenant.Status.Pools = append(tenant.Status.Pools, miniov2.PoolStatus{
				SSName: ssName,
				State:  miniov2.PoolNotCreated,
			})
			// push updates to status
			if tenant, err = c.updatePoolStatus(ctx, tenant); err != nil {
				return WrapResult(Result{}, err)
			}
		}
		ss, err := c.statefulSetLister.StatefulSets(tenant.Namespace).Get(ssName)
		if k8serrors.IsNotFound(err) {
			klog.Infof("'%s/%s': Deploying pool %s", tenant.Namespace, tenant.Name, pool.Name)
			if tenant, err = c.updateTenantStatus(ctx, tenant, StatusProvisioningStatefulSet, 0); err != nil {
				return WrapResult(Result{}, err)
			}
			ss = statefulsets.NewPool(&statefulsets.NewPoolArgs{
				Tenant:          tenant,
				SkipEnvVars:     skipEnvVars,
				Pool:            &pool,
				PoolStatus:      &tenant.Status.Pools[i],
				ServiceName:     hlSvc.Name,
				HostsTemplate:   c.hostsTemplate,
				OperatorVersion: c.operatorVersion,
				OperatorCATLS:   operatorCATLSExists,
				OperatorImage:   c.operatorImage,
			})
			ss, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Create(ctx, ss, cOpts)
			if err != nil {
				return WrapResult(Result{}, err)
			}
			c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "PoolCreated", fmt.Sprintf("Tenant pool %s created", pool.Name))
			// Report the pool is properly created
			tenant.Status.Pools[i].State = miniov2.PoolCreated
			// mark we are adding a new pool to the next block can act accordingly
			addingNewPool = true
			// push updates to status
			if tenant, err = c.updatePoolStatus(ctx, tenant); err != nil {
				return WrapResult(Result{}, err)
			}
		}

		// keep track of all replicas
		totalAvailableReplicas += ss.Status.AvailableReplicas
		images = append(images, ss.Spec.Template.Spec.Containers[0].Image)
	}

	var initializedPool miniov2.Pool
	// validate each pool if it's initialized, and mark it if it is.
	for pi, pool := range tenant.Spec.Pools {
		// get a pod for the established statefulset
		if tenant.Status.Pools[pi].State == miniov2.PoolInitialized {
			initializedPool = pool
			continue
		}

		var restarted bool
		// Only restart if there is an existing initialized pool, if there are no initialized
		// pools no need to restart.
		if initializedPool.Name != "" && addingNewPool {
			// Restart services to get new args since we are expanding the deployment here.
			if err := c.restartInitializedPool(ctx, tenant, initializedPool, tenantConfiguration); err != nil {
				klog.Infof("'%s' restart call failed", key)
				return WrapResult(Result{}, err)
			}
			metaNowTime := metav1.Now()
			tenant.Status.WaitingOnReady = &metaNowTime
			tenant.Status.CurrentState = StatusRestartingMinIO
			if tenant, err = c.updatePoolStatus(ctx, tenant); err != nil {
				klog.Infof("'%s' Can't update tenant status: %v", key, err)
				return WrapResult(Result{}, err)
			}
			klog.Infof("'%s' was restarted", key)
			restarted = true
		}

		// Report the pool is initialized.
		tenant.Status.Pools[pi].State = miniov2.PoolInitialized
		// push updates to status
		if tenant, err = c.updatePoolStatus(ctx, tenant); err != nil {
			return WrapResult(Result{}, err)
		}

		if restarted {
			return WrapResult(Result{}, ErrMinIORestarting)
		}
	}

	// wait here until all pools are initialized, so we can continue with updating versions and the existingSS resources.
	for _, poolStatus := range tenant.Status.Pools {
		if poolStatus.State != miniov2.PoolInitialized {
			// at least 1 is not initialized, stop here until they all are.
			return WrapResult(Result{}, errors.New("Waiting for all pools to initialize"))
		}
	}

	// wait here if `waitOnReady` is set to a given time
	if tenant.Status.WaitingOnReady != nil {
		// if it's been longer than the default time 5 minutes, unset the field and continue
		someTimeAgo := time.Now().Add(-5 * time.Minute)
		if tenant.Status.WaitingOnReady.Time.Before(someTimeAgo) {
			tenant.Status.WaitingOnReady = nil
			if tenant, err = c.updatePoolStatus(ctx, tenant); err != nil {
				klog.Infof("'%s' Can't update tenant status: %v", key, err)
			}
		} else {
			// check if MinIO is already online after the previous restart
			if tenant.MinIOHealthCheck(c.getTransport()) {
				tenant.Status.WaitingOnReady = nil
				if _, err = c.updatePoolStatus(ctx, tenant); err != nil {
					klog.Infof("'%s' Can't update tenant status: %v", key, err)
				}
				return WrapResult(Result{}, ErrMinIORestarting)
			}
		}
	}

	// compare all the images across all pools, they should always be the same.
	compareImage := ""
	for i, image := range images {
		if compareImage == "" {
			compareImage = image
		}
		if compareImage != image {
			if _, err = c.updateTenantStatus(ctx, tenant, StatusInconsistentMinIOVersions, totalAvailableReplicas); err != nil {
				return WrapResult(Result{}, err)
			}
			return WrapResult(Result{}, fmt.Errorf("Pool %d is running incorrect image version, all pools are required to be on the same MinIO version. Attempting update of the inconsistent pool", i+1))
		}
	}

	// In loop above we compared all the versions in all pools.
	// So comparing tenant.Spec.Image (version to update to) against one value from images slice is fine.
	ssImages := strings.Split(images[0], ":")
	specImages := strings.Split(tenant.Spec.Image, ":")
	var ssImage string
	var specImage string
	if len(specImages) > 1 {
		specImage = specImages[1]
	}
	if len(ssImages) > 1 {
		ssImage = ssImages[1]
	}
	if specImage != ssImage && tenant.Status.CurrentState != StatusUpdatingMinIOVersion {
		if !tenant.MinIOHealthCheck(c.getTransport()) {
			klog.Infof("%s is not running can't update image online", key)
			return WrapResult(Result{}, ErrMinIONotReady)
		}

		// Images different with the newer state change, continue to verify
		// if upgrade is possible
		tenant, err = c.updateTenantStatus(ctx, tenant, StatusUpdatingMinIOVersion, totalAvailableReplicas)
		if err != nil {
			return WrapResult(Result{}, err)
		}

		klog.V(4).Infof("Collecting artifacts for Tenant '%s' to update MinIO from: %s, to: %s",
			tenantName, images[0], tenant.Spec.Image)

		latest, err := c.fetchArtifacts(tenant)
		defer c.removeArtifacts()
		if err != nil {
			return WrapResult(Result{}, err)
		}
		updateURL, err := tenant.UpdateURL(latest, fmt.Sprintf("http://operator.%s.svc.%s:%s%s",
			miniov2.GetNSFromFile(),
			miniov2.GetClusterDomain(),
			common.UpgradeServerPort,
			common.WebhookAPIUpdate,
		))
		if err != nil {
			err = fmt.Errorf("Unable to get canonical update URL for Tenant '%s', failed with %v", tenantName, err)
			if _, terr := c.updateTenantStatus(ctx, tenant, err.Error(), totalAvailableReplicas); terr != nil {
				return WrapResult(Result{}, terr)
			}

			// Correct URL could not be obtained, not proceeding to update.
			return WrapResult(Result{}, err)
		}

		klog.V(4).Infof("Updating Tenant %s MinIO version from: %s, to: %s -> URL: %s",
			tenantName, tenant.Spec.Image, images[0], updateURL)

		us, err := adminClnt.ServerUpdate(ctx, updateURL)
		if err != nil {
			if madmin.ToErrorResponse(err).Code != "MethodNotAllowed" {
				if _, terr := c.updateTenantStatus(ctx, tenant, err.Error(), totalAvailableReplicas); terr != nil {
					return WrapResult(Result{}, terr)
				}
				// Update failed, nothing needs to be changed in the container
				return WrapResult(Result{}, err)
			}
			c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "Inplace update is disabled, falling back to performing only statefulset update.", fmt.Sprintf("Tenant %s", tenant.Name))
		}
		if err == nil {
			if us.CurrentVersion != us.UpdatedVersion {
				// In case the upgrade is from an older version to RELEASE.2021-07-27T02-40-15Z (which introduced
				// MinIO server integrated with Console), we need to delete the old console deployment and service.
				// We do this only when MinIO server is successfully updated.
				unifiedConsoleReleaseTime, _ := miniov2.ReleaseTagToReleaseTime("RELEASE.2021-07-27T02-40-15Z")
				newVer, err := miniov2.ReleaseTagToReleaseTime(us.UpdatedVersion)
				if err != nil {
					klog.Errorf("Unsupported release tag on new image, server updated but might leave dangling console deployment %v", err)
					return WrapResult(Result{}, err)
				}
				consoleDeployment, err := c.deploymentLister.Deployments(tenant.Namespace).Get(tenant.ConsoleDeploymentName())
				if unifiedConsoleReleaseTime.Before(newVer) && consoleDeployment != nil && err == nil {
					if err := c.deleteOldConsoleDeployment(ctx, tenant, consoleDeployment.Name); err != nil {
						return WrapResult(Result{}, err)
					}
				}
				klog.Infof("Tenant '%s' MinIO updated successfully from: %s, to: %s successfully",
					tenantName, us.CurrentVersion, us.UpdatedVersion)
			} else {
				msg := fmt.Sprintf("Tenant '%s' MinIO is already running the most recent version of %s",
					tenantName,
					us.CurrentVersion)
				klog.Info(msg)
				if _, terr := c.updateTenantStatus(ctx, tenant, msg, totalAvailableReplicas); terr != nil {
					return WrapResult(Result{}, err)
				}
				return WrapResult(Result{}, nil)
			}
		}
		for i, pool := range tenant.Spec.Pools {
			// Now proceed to make the yaml changes for the tenant statefulset.
			ss := statefulsets.NewPool(&statefulsets.NewPoolArgs{
				Tenant:          tenant,
				SkipEnvVars:     skipEnvVars,
				Pool:            &pool,
				PoolStatus:      &tenant.Status.Pools[i],
				ServiceName:     hlSvc.Name,
				HostsTemplate:   c.hostsTemplate,
				OperatorVersion: c.operatorVersion,
				OperatorCATLS:   operatorCATLSExists,
				OperatorImage:   c.operatorImage,
			})
			if _, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Update(ctx, ss, uOpts); err != nil {
				return WrapResult(Result{}, err)
			}
			c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "PoolUpdated", fmt.Sprintf("Tenant pool %s updated", pool.Name))
		}

	}

	// This loop will take care of updating the statefulset for each pool
	for i, pool := range tenant.Spec.Pools {
		// Get the StatefulSet with the name specified in Tenant.status.pools[i].SSName
		// if this index is in the status of pools use it, else capture the desired name in the status and store it
		var ssName string
		if len(tenant.Status.Pools) > i {
			ssName = tenant.Status.Pools[i].SSName
		} else {
			ssName = tenant.PoolStatefulsetName(&pool)
			tenant.Status.Pools = append(tenant.Status.Pools, miniov2.PoolStatus{
				SSName: ssName,
				State:  miniov2.PoolNotCreated,
			})
			// push updates to status
			if tenant, err = c.updatePoolStatus(ctx, tenant); err != nil {
				return WrapResult(Result{}, err)
			}
		}
		existingStatefulSet, err := c.statefulSetLister.StatefulSets(tenant.Namespace).Get(ssName)
		// at this point the existingStatefulSet should already exist, error out
		if k8serrors.IsNotFound(err) {
			klog.Errorf("%s's pool %s doesn't exist: %v", tenant.Name, ssName, err)
			return WrapResult(Result{}, err)
		}
		if pool.Servers != *existingStatefulSet.Spec.Replicas {
			// warn the user that replica count of an existing pool can't be changed
			if tenant, err = c.updateTenantStatus(ctx, tenant, fmt.Sprintf("Can't modify server count for pool %s", pool.Name), 0); err != nil {
				return WrapResult(Result{}, err)
			}
		}
		// generated the expected StatefulSet based on the new tenant configuration
		expectedStatefulSet := statefulsets.NewPool(&statefulsets.NewPoolArgs{
			Tenant:          tenant,
			SkipEnvVars:     skipEnvVars,
			Pool:            &pool,
			PoolStatus:      &tenant.Status.Pools[i],
			ServiceName:     hlSvc.Name,
			HostsTemplate:   c.hostsTemplate,
			OperatorVersion: c.operatorVersion,
			OperatorCATLS:   operatorCATLSExists,
			OperatorImage:   c.operatorImage,
		})
		// Verify if this pool matches the spec on the tenant (resources, affinity, sidecars, etc)
		poolMatchesSS, err := poolSSMatchesSpec(expectedStatefulSet, existingStatefulSet)
		if err != nil {
			return WrapResult(Result{}, err)
		}
		// if the pool doesn't match the spec
		if !poolMatchesSS {
			// for legacy reasons, if the zone label is present in SS we must carry it over
			carryOverLabels := make(map[string]string)
			if val, ok := existingStatefulSet.Spec.Template.ObjectMeta.Labels[miniov1.ZoneLabel]; ok {
				carryOverLabels[miniov1.ZoneLabel] = val
			}

			newStatefulSet := existingStatefulSet.DeepCopy()

			newStatefulSet.Spec.Template = expectedStatefulSet.Spec.Template
			newStatefulSet.Spec.UpdateStrategy = expectedStatefulSet.Spec.UpdateStrategy

			if existingStatefulSet.Spec.Template.ObjectMeta.Labels == nil {
				newStatefulSet.Spec.Template.ObjectMeta.Labels = make(map[string]string)
			}
			for k, v := range carryOverLabels {
				newStatefulSet.Spec.Template.ObjectMeta.Labels[k] = v
			}

			if existingStatefulSet, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Update(ctx, newStatefulSet, uOpts); err != nil {
				klog.Errorf("[Will try again in 5sec] Update tenant %s statefulset %s error %s", tenant.Name, ssName, err)
				return WrapResult(Result{RequeueAfter: time.Second * 5}, nil)
			}
		}

		// If the StatefulSet is not controlled by this Tenant resource, we should log
		// a warning to the event recorder and ret
		if !metav1.IsControlledBy(existingStatefulSet, tenant) {
			if tenant, err = c.updateTenantStatus(ctx, tenant, StatusNotOwned, existingStatefulSet.Status.Replicas); err != nil {
				return WrapResult(Result{}, err)
			}
			msg := fmt.Sprintf(MessageResourceExists, existingStatefulSet.Name)
			c.recorder.Event(tenant, corev1.EventTypeWarning, ErrResourceExists, msg)
			// return nil so we don't re-queue this work item, this error won't get fixed by reprocessing
			return WrapResult(Result{}, nil)
		}
	}

	if tenant.HasPrometheusOperatorEnabled() {
		err := c.checkAndCreatePrometheusAddlConfig(ctx, tenant, string(tenantConfiguration["accesskey"]), string(tenantConfiguration["secretkey"]))
		if err != nil {
			return WrapResult(Result{}, err)
		}
	} else {
		err := c.deletePrometheusAddlConfig(ctx, tenant)
		if err != nil {
			return WrapResult(Result{}, err)
		}
	}

	// Ensure we are only provisioning users one time
	if !tenant.Status.ProvisionedUsers && len(tenant.Spec.Users) > 0 {
		if err := c.createUsers(ctx, tenant, tenantConfiguration); err != nil {
			klog.V(2).Infof("Unable to create MinIO users: %v", err)
			c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "UsersCreatedFailed", fmt.Sprintf("Users creation failed: %s", err))
			// retry after 5sec
			return WrapResult(Result{RequeueAfter: time.Second * 5}, nil)
		}
		c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "UsersCreated", "Users created")
	}

	// Ensure we are only creating the bucket
	if !tenant.Status.ProvisionedBuckets && len(tenant.Spec.Buckets) > 0 {
		if err := c.createBuckets(ctx, tenant, tenantConfiguration); err != nil {
			klog.V(2).Infof("Unable to create MinIO buckets: %v", err)
			c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "BucketsCreatedFailed", fmt.Sprintf("Buckets creation failed: %s", err))
			// retry after 5sec
			return WrapResult(Result{RequeueAfter: time.Second * 5}, err)
		}
		c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "BucketsCreated", "Buckets created")
	}

	// Finally, we update the status block of the Tenant resource to reflect the
	// current state of the world
	_, err = c.updateTenantStatus(ctx, tenant, StatusInitialized, totalAvailableReplicas)
	return WrapResult(Result{}, err)
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
	if !c.namespacesToWatch.IsEmpty() {
		meta, err := meta.Accessor(obj)
		if err != nil {
			runtime.HandleError(err)
			return
		}
		if !c.namespacesToWatch.Contains(meta.GetNamespace()) {
			klog.Infof("Ignoring tenant `%s` in namespace that is not watched by this controller.", key)
			return
		}
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

		tenant, err := c.minioClientSet.MinioV2().Tenants(object.GetNamespace()).Get(context.Background(), ownerRef.Name, metav1.GetOptions{})
		if err != nil {
			klog.V(4).Infof("ignoring orphaned object '%s' of tenant '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueTenant(tenant)
		return
	}
}

// MinIOControllerRateLimiter is a no-arg constructor for a default rate limiter for a workqueue for our controller.
// both overall and per-item rate limiting.  The overall is a token bucket and the per-item is exponential
func MinIOControllerRateLimiter() queue.RateLimiter {
	return queue.NewMaxOfRateLimiter(
		queue.NewItemExponentialFailureRateLimiter(5*time.Second, 60*time.Second),
		// 10 qps, 100 bucket size.  This is only for retry speed and its only the overall factor (not per item)
		&queue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
	)
}

type patchAnnotation struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}
