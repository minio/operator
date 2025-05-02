// This file is part of MinIO Operator
// Copyright (c) 2024 MinIO, Inc.
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

package controller

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/minio/minio-go/v7/pkg/set"
	"github.com/minio/operator/pkg/apis/job.min.io/v1alpha1"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	stsv1beta1 "github.com/minio/operator/pkg/apis/sts.min.io/v1beta1"
	jobinformers "github.com/minio/operator/pkg/client/informers/externalversions/job.min.io/v1alpha1"
	joblisters "github.com/minio/operator/pkg/client/listers/job.min.io/v1alpha1"
	"github.com/minio/operator/pkg/utils/miniojob"
	batchjobv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
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
	_ kubernetes.Interface,
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
		UpdateFunc: func(old, newObject interface{}) {
			oldJob := old.(*v1alpha1.MinIOJob)
			newJob := newObject.(*v1alpha1.MinIOJob)
			if oldJob.ResourceVersion == newJob.ResourceVersion {
				return
			}
			controller.enqueueJob(newObject)
		},
		DeleteFunc: controller.enqueueJob,
	})

	jobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(_, newObject interface{}) {
			newJob := newObject.(*batchjobv1.Job)
			jobName, ok := newJob.Labels[miniojob.MinioJobName]
			if !ok {
				return
			}
			jobCRName, ok := newJob.Labels[miniojob.MinioJobCRName]
			if !ok {
				return
			}
			val, ok := globalIntervalJobStatus.Load(fmt.Sprintf("%s/%s", newJob.GetNamespace(), jobCRName))
			if ok {
				intervalJob := val.(*miniojob.MinIOIntervalJob)
				command, ok := intervalJob.CommandMap[jobName]
				if ok {
					if newJob.Status.Succeeded > 0 {
						command.SetStatus(true, "")
					} else {
						for _, condition := range newJob.Status.Conditions {
							if condition.Type == batchjobv1.JobFailed {
								command.SetStatus(false, condition.Message)
								break
							}
						}
					}
				}
			}
			controller.HandleObject(newJob)
		},
		DeleteFunc: controller.enqueueJob,
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
func (c *JobController) SyncHandler(key string) (_ Result, err error) {
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
	err = c.k8sClient.Get(ctx, client.ObjectKeyFromObject(&jobCR), &jobCR)
	if err != nil {
		// job cr have gone
		globalIntervalJobStatus.Delete(fmt.Sprintf("%s/%s", jobCR.Namespace, jobCR.Name))
		if errors.IsNotFound(err) {
			return WrapResult(Result{}, nil)
		}
		return WrapResult(Result{}, err)
	}

	// if job cr is Success, do nothing
	if jobCR.Status.Phase == miniojob.MinioJobPhaseSuccess {
		// delete the job status
		globalIntervalJobStatus.Delete(fmt.Sprintf("%s/%s", jobCR.Namespace, jobCR.Name))
		return WrapResult(Result{}, nil)
	}

	defer func() {
		if err != nil {
			if jobCR.Status.Phase != miniojob.MinioJobPhaseSuccess {
				jobCR.Status.Phase = miniojob.MinioJobPhaseError
				jobCR.Status.Message = err.Error()
				err = c.updateJobStatus(ctx, &jobCR)
			}
		}
	}()

	if !IsSTSEnabled() {
		c.recorder.Eventf(&jobCR, corev1.EventTypeWarning, "STSDisabled", "JobCR cannot work with STS disabled")
		return WrapResult(Result{}, fmt.Errorf("JobCR cannot work with STS disabled"))
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
		return WrapResult(Result{}, fmt.Errorf("get tenant %s/%s error: %w", jobCR.Spec.TenantRef.Namespace, jobCR.Spec.TenantRef.Name, err))
	}
	if tenant.Status.HealthStatus != miniov2.HealthStatusGreen {
		return WrapResult(Result{RequeueAfter: time.Second * 5}, fmt.Errorf("get tenant %s/%s error: %w", jobCR.Spec.TenantRef.Namespace, jobCR.Spec.TenantRef.Name, err))
	}
	// check sa
	pbs := &stsv1beta1.PolicyBindingList{}
	err = c.k8sClient.List(ctx, pbs, client.InNamespace(namespace))
	if err != nil {
		return WrapResult(Result{}, fmt.Errorf("list policybinding error: %w", err))
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
	intervalJob, err := checkMinIOJob(&jobCR)
	if err != nil {
		return WrapResult(Result{}, err)
	}
	err = intervalJob.CreateCommandJob(ctx, c.k8sClient, STSDefaultPort, tenant.TLS())
	if err != nil {
		return WrapResult(Result{}, fmt.Errorf("create job error: %w", err))
	}
	// update status
	jobCR.Status = intervalJob.GetMinioJobStatus(ctx)
	err = c.updateJobStatus(ctx, &jobCR)
	return WrapResult(Result{}, err)
}

func (c *JobController) updateJobStatus(ctx context.Context, job *v1alpha1.MinIOJob) error {
	return c.k8sClient.Status().Update(ctx, job)
}

func checkMinIOJob(jobCR *v1alpha1.MinIOJob) (intervalJob *miniojob.MinIOIntervalJob, err error) {
	defer func() {
		if err != nil {
			globalIntervalJobStatus.Delete(fmt.Sprintf("%s/%s", jobCR.Namespace, jobCR.Name))
		}
	}()
	val, found := globalIntervalJobStatus.Load(fmt.Sprintf("%s/%s", jobCR.Namespace, jobCR.Name))
	if found {
		intervalJob = val.(*miniojob.MinIOIntervalJob)
		if reflect.DeepEqual(intervalJob.JobCR.Spec, jobCR.Spec) {
			intervalJob.JobCR.UID = jobCR.UID
			return intervalJob, nil
		}
	}
	intervalJob = &miniojob.MinIOIntervalJob{
		JobCR:      jobCR.DeepCopy(),
		Command:    []*miniojob.MinIOIntervalJobCommand{},
		CommandMap: map[string]*miniojob.MinIOIntervalJobCommand{},
	}
	if jobCR.Spec.TenantRef.Namespace == "" {
		return intervalJob, fmt.Errorf("tenant namespace is empty")
	}
	if jobCR.Spec.TenantRef.Name == "" {
		return intervalJob, fmt.Errorf("tenant name is empty")
	}
	if jobCR.Spec.ServiceAccountName == "" {
		return intervalJob, fmt.Errorf("serviceaccount name is empty")
	}
	for index, val := range jobCR.Spec.Commands {
		jobCommand, err := miniojob.GenerateMinIOIntervalJobCommand(val, index)
		if err != nil {
			return intervalJob, err
		}
		intervalJob.Command = append(intervalJob.Command, jobCommand)
		intervalJob.CommandMap[jobCommand.JobName] = jobCommand
	}
	// check all dependon
	for _, command := range intervalJob.Command {
		for _, dep := range command.CommandSpec.DependsOn {
			_, found := intervalJob.CommandMap[dep]
			if !found {
				return intervalJob, fmt.Errorf("dependent job %s not found", dep)
			}
		}
	}
	globalIntervalJobStatus.Store(fmt.Sprintf("%s/%s", jobCR.Namespace, jobCR.Name), intervalJob)
	return intervalJob, nil
}

var globalIntervalJobStatus = sync.Map{}
