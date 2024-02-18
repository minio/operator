// This file is part of MinIO Operator
// Copyright (c) 2024 MinIO, Inc.

package controller

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7/pkg/set"
	"k8s.io/apimachinery/pkg/api/meta"

	jobinformers "github.com/minio/operator/pkg/client/informers/externalversions/job.min.io/v1alpha1"
	joblisters "github.com/minio/operator/pkg/client/listers/job.min.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	queue "k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

// JobController struct watches the Kubernetes API for changes to Tenant resources
type JobController struct {
	namespacesToWatch set.StringSet
	lister            joblisters.MinIOJobLister
	hasSynced         cache.InformerSynced
	kubeClientSet     kubernetes.Interface
	statefulSetLister appslisters.StatefulSetLister
	recorder          record.EventRecorder
	workqueue         queue.RateLimitingInterface
}

// JobControllerInterface is an interface for the controller with the methods supported by it.
type JobControllerInterface interface {
	WorkQueue() workqueue.RateLimitingInterface
	KeyFunc() cache.KeyFunc
	HasSynced() cache.InformerSynced
	SyncHandler(ctx context.Context, name, namespace string) error
	HandleObject(obj metav1.Object)
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
		klog.V(2).Infof("Key from workqueue: %s", key)

		c.SyncHandler(key)
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.V(4).Infof("Successfully synced '%s'", key)
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
	workqueue queue.RateLimitingInterface,
) *JobController {
	controller := &JobController{
		namespacesToWatch: namespacesToWatch,
		lister:            joblister,
		hasSynced:         hasSynced,
		kubeClientSet:     kubeClientSet,
		statefulSetLister: statefulSetLister,
		recorder:          recorder,
		workqueue:         workqueue,
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
func (c *JobController) SyncHandler(key string) error {
	klog.Info("Job Controller Loop!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")

	return nil
}
