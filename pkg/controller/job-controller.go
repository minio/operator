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
	stsv1alpha1 "github.com/minio/operator/pkg/apis/sts.min.io/v1alpha1"
	jobinformers "github.com/minio/operator/pkg/client/informers/externalversions/job.min.io/v1alpha1"
	joblisters "github.com/minio/operator/pkg/client/listers/job.min.io/v1alpha1"
	batchjobv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	batchv1 "k8s.io/client-go/informers/batch/v1"
	"k8s.io/client-go/kubernetes"
	k8sjoblisters "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// JobController struct watches the Kubernetes API for changes to Tenant resources
type JobController struct {
	namespacesToWatch set.StringSet
	minioJobLister    joblisters.MinIOJobLister
	minioJobHasSynced cache.InformerSynced
	jobLister         k8sjoblisters.JobLister
	jobHasSynced      cache.InformerSynced
	recorder          record.EventRecorder
	workqueue         workqueue.RateLimitingInterface
	k8sClient         client.Client
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *JobController) runJobWorker() {
	defer runtime.HandleCrash()
	for processNextItem(c.workqueue, c.SyncHandler) {
	}
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
	minioJobInformer jobinformers.MinIOJobInformer,
	jobInformer batchv1.JobInformer,
	namespacesToWatch set.StringSet,
	kubeClientSet kubernetes.Interface,
	recorder record.EventRecorder,
	workqueue workqueue.RateLimitingInterface,
	k8sClient client.Client,
) *JobController {
	controller := &JobController{
		namespacesToWatch: namespacesToWatch,
		minioJobLister:    minioJobInformer.Lister(),
		minioJobHasSynced: minioJobInformer.Informer().HasSynced,
		jobLister:         jobInformer.Lister(),
		jobHasSynced:      jobInformer.Informer().HasSynced,
		recorder:          recorder,
		workqueue:         workqueue,
		k8sClient:         k8sClient,
	}

	// Set up an event handler for when resources change
	minioJobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueJob,
		UpdateFunc: func(old, new interface{}) {
			oldJob := old.(*v1alpha1.MinIOJob)
			newJob := new.(*v1alpha1.MinIOJob)
			if oldJob.ResourceVersion == newJob.ResourceVersion {
				return
			}
			controller.enqueueJob(new)
		},
	})

	jobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			oldJob := old.(*batchjobv1.Job)
			newJob := new.(*batchjobv1.Job)
			if oldJob.ResourceVersion == newJob.ResourceVersion {
				return
			}
			// todo record the job status.
		},
	})
	return controller
}

// HasSynced is to determine if obj is synced
func (c *JobController) HasSynced() cache.InformerSynced {
	return c.minioJobHasSynced
}

// HandleObject will take any resource implementing metav1.Object and attempt
// to find the CRD resource that 'owns' it.
func (c *JobController) HandleObject(obj metav1.Object) {
	JobCRDResourceKind := "MinIOJob"
	if ownerRef := metav1.GetControllerOf(obj); ownerRef != nil {
		switch ownerRef.Kind {
		case JobCRDResourceKind:
			job, err := c.minioJobLister.MinIOJobs(obj.GetNamespace()).Get(ownerRef.Name)
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
	// check sa
	pbs := &stsv1alpha1.PolicyBindingList{}
	err = c.k8sClient.List(ctx, pbs, client.InNamespace(namespace))
	if err != nil {
		return WrapResult(Result{}, err)
	}
	if len(pbs.Items) == 0 {
		return WrapResult(Result{}, fmt.Errorf("no policybinding found"))
	}
	saFound := false
	for _, pb := range pbs.Items {
		if pb.Spec.Application.Namespace == namespace && pb.Spec.Application.ServiceAccount == jobCR.Spec.ServiceAccountName {
			saFound = true
		}
	}
	if !saFound {
		return WrapResult(Result{}, fmt.Errorf("no serviceaccount found"))
	}
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
