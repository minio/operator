// This file is part of MinIO Operator
// Copyright (c) 2024 MinIO, Inc.

package controller

import (
	"context"
	"fmt"
	"time"

	corelisters "k8s.io/client-go/listers/core/v1"

	informers "github.com/minio/operator/pkg/client/informers/externalversions/job.min.io/v1alpha1"
	listers "github.com/minio/operator/pkg/client/listers/job.min.io/v1alpha1"
	"golang.org/x/time/rate"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

type jobController struct {
	lister            listers.MinIOJobLister
	hasSynced         cache.InformerSynced
	workqueue         workqueue.RateLimitingInterface
	kubeClientSet     kubernetes.Interface
	statefulSetLister appslisters.StatefulSetLister
	recorder          record.EventRecorder
}

type controllerConfig struct {
	serviceLister     corelisters.ServiceLister
	kubeClientSet     kubernetes.Interface
	statefulSetLister appslisters.StatefulSetLister
	deploymentLister  appslisters.DeploymentLister
	recorder          record.EventRecorder
}

// JobControllerInterface is an interface for the controller with the methods supported by it.
type JobControllerInterface interface {
	WorkQueue() workqueue.RateLimitingInterface
	KeyFunc() cache.KeyFunc
	HasSynced() cache.InformerSynced
	SyncHandler(ctx context.Context, name, namespace string) error
	HandleObject(obj metav1.Object)
}

func enqueue(c JobControllerInterface, obj interface{}) {
	var key string
	var err error
	if key, err = c.KeyFunc()(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.WorkQueue().Add(key)
}

func newJobController(informer informers.MinIOJobInformer, config controllerConfig) *jobController {
	rateLimiter := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 1000*time.Second),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(50), 300)},
	)
	jobController := &jobController{
		lister:            informer.Lister(),
		hasSynced:         informer.Informer().HasSynced,
		workqueue:         workqueue.NewRateLimitingQueue(rateLimiter),
		kubeClientSet:     config.kubeClientSet,
		statefulSetLister: config.statefulSetLister,
		recorder:          config.recorder,
	}
	// Set up an event handler for when resources change
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			enqueue(jobController, obj)
		},
		UpdateFunc: func(old, new interface{}) {
			enqueue(jobController, new)
		},
	})
	return jobController
}

func (c *jobController) WorkQueue() workqueue.RateLimitingInterface {
	return c.workqueue
}

func (c *jobController) KeyFunc() cache.KeyFunc {
	return cache.MetaNamespaceKeyFunc
}

func (c *jobController) HasSynced() cache.InformerSynced {
	return c.hasSynced
}

func (c *jobController) HandleObject(obj metav1.Object) {
	JobCRDResourceKind := "MinIOJob"
	if ownerRef := metav1.GetControllerOf(obj); ownerRef != nil {
		switch ownerRef.Kind {
		case JobCRDResourceKind:
			job, err := c.lister.MinIOJobs(obj.GetNamespace()).Get(ownerRef.Name)
			if err != nil {
				klog.V(4).Info("Ignore orphaned object", "object", klog.KObj(job), JobCRDResourceKind, ownerRef.Name)
				return
			}
			enqueue(c, job)
		default:
			return
		}
		return
	}
}

// syncJobHandler compares the current Job state with the desired, and attempts to
// converge the two. It then updates the Status block of the Job resource
// with the current status of the resource.
func (c *jobController) SyncHandler(ctx context.Context, name, namespace string) error {
	// Get the Job resource with this namespace/name
	_, err := c.lister.MinIOJobs(namespace).Get(name)
	if err != nil {
		// The Job resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("job '%s' in work queue no longer exists: %+v", name, err))
			return nil
		}

		return err
	}

	return nil
}
