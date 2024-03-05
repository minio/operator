// This file is part of MinIO Operator
// Copyright (c) 2024 MinIO, Inc.

package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7/pkg/set"
	"github.com/minio/operator/pkg/apis/job.min.io/v1alpha1"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	clientset "github.com/minio/operator/pkg/client/clientset/versioned"
	jobinformers "github.com/minio/operator/pkg/client/informers/externalversions/job.min.io/v1alpha1"
	joblisters "github.com/minio/operator/pkg/client/listers/job.min.io/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// JobController struct watches the Kubernetes API for changes to Tenant resources
type JobController struct {
	namespacesToWatch set.StringSet
	lister            joblisters.MinIOJobLister
	hasSynced         cache.InformerSynced
	kubeClientSet     kubernetes.Interface
	statefulSetLister appslisters.StatefulSetLister
	recorder          record.EventRecorder
	workqueue         workqueue.RateLimitingInterface
	minioClientSet    clientset.Interface
	k8sClient         client.Client
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *JobController) runJobWorker() {
	defer runtime.HandleCrash()
	for c.processNextJobWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *JobController) processNextJobWorkItem() bool {
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

		// Run the syncHandler, passing it the namespace/name string of the tenant.
		result, err := c.SyncHandler(key)
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
		}

		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		return nil
	}

	if err := processItem(obj); err != nil {
		runtime.HandleError(err)
		return true
	}
	return true
}

func (c *JobController) enqueueJob(obj interface{}) {
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
	// key = default/mc-job-1
	c.workqueue.AddRateLimited(key)
}

// NewJobController returns a new Operator Controller
func NewJobController(
	jobinformer jobinformers.MinIOJobInformer,
	namespacesToWatch set.StringSet,
	joblister joblisters.MinIOJobLister,
	hasSynced cache.InformerSynced,
	kubeClientSet kubernetes.Interface,
	statefulSetLister appslisters.StatefulSetLister,
	recorder record.EventRecorder,
	workqueue workqueue.RateLimitingInterface,
	minioClientSet clientset.Interface,
	k8sClient client.Client,
) *JobController {
	controller := &JobController{
		namespacesToWatch: namespacesToWatch,
		lister:            joblister,
		hasSynced:         hasSynced,
		kubeClientSet:     kubeClientSet,
		statefulSetLister: statefulSetLister,
		recorder:          recorder,
		workqueue:         workqueue,
		minioClientSet:    minioClientSet,
		k8sClient:         k8sClient,
	}

	// Set up an event handler for when resources change
	jobinformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			controller.enqueueJob(obj)
		},
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueJob(new)
		},
	})
	return controller
}

// HasSynced is to determine if obj is synced
func (c *JobController) HasSynced() cache.InformerSynced {
	return c.hasSynced
}

// HandleObject will take any resource implementing metav1.Object and attempt
// to find the CRD resource that 'owns' it.
func (c *JobController) HandleObject(obj metav1.Object) {
	JobCRDResourceKind := "MinIOJob"
	if ownerRef := metav1.GetControllerOf(obj); ownerRef != nil {
		switch ownerRef.Kind {
		case JobCRDResourceKind:
			job, err := c.lister.MinIOJobs(obj.GetNamespace()).Get(ownerRef.Name)
			if err != nil {
				klog.V(4).Info("Ignore orphaned object", "object", klog.KObj(job), JobCRDResourceKind, ownerRef.Name)
				return
			}
			c.enqueueJob(job)
		default:
			return
		}
		return
	}
}

// SyncHandler compares the current Job state with the desired, and attempts to
// converge the two. It then updates the Status block of the Job resource
// with the current status of the resource.
func (c *JobController) SyncHandler(key string) (Result, error) {
	// Convert the namespace/name string into a distinct namespace and name
	if key == "" {
		runtime.HandleError(fmt.Errorf("Invalid resource key: %s", key))
		return WrapResult(Result{}, nil)
	}
	namespace, jobName := key2NamespaceName(key)
	ctx := context.Background()
	jobCR := v1alpha1.MinIOJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
		},
	}
	err := c.k8sClient.Get(ctx, client.ObjectKeyFromObject(&jobCR), &jobCR)
	if err != nil {
		// job cr have gone
		if errors.IsNotFound(err) {
			return WrapResult(Result{}, nil)
		}
		return WrapResult(Result{}, err)
	}
	// get tenant
	tenant := &miniov2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: jobCR.Spec.TenantRef.Namespace,
			Name:      jobCR.Spec.TenantRef.Name,
		},
	}
	err = c.k8sClient.Get(ctx, client.ObjectKeyFromObject(tenant), tenant)
	if err != nil {
		jobCR.Status.Phase = "Error"
		jobCR.Status.Message = fmt.Sprintf("Get tenant %s/%s error:%v", jobCR.Spec.TenantRef.Namespace, jobCR.Spec.TenantRef.Name, err)
		err = c.updateJobStatus(ctx, &jobCR)
		return WrapResult(Result{}, err)
	}
	if tenant.Status.HealthStatus != miniov2.HealthStatusGreen {
		return WrapResult(Result{RequeueAfter: time.Second * 5}, nil)
	}
	fmt.Println("will do somthing next")
	// Loop through the different supported operations.
	for _, val := range jobCR.Spec.Commands {
		operation := val.Operation
		if operation == "make-bucket" {
			// TODO: Initiate a job to create the bucket(s) if they do not exist and if the Tenant is prepared for it.
		}
	}
	return WrapResult(Result{}, err)
}

func (c *JobController) updateJobStatus(ctx context.Context, job *v1alpha1.MinIOJob) error {
	return c.k8sClient.Status().Update(ctx, job)
}
