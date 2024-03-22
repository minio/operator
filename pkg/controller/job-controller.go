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
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/minio/minio-go/v7/pkg/set"
	"github.com/minio/operator/pkg/apis/job.min.io/v1alpha1"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	stsv1alpha1 "github.com/minio/operator/pkg/apis/sts.min.io/v1alpha1"
	jobinformers "github.com/minio/operator/pkg/client/informers/externalversions/job.min.io/v1alpha1"
	joblisters "github.com/minio/operator/pkg/client/listers/job.min.io/v1alpha1"
	runtime2 "github.com/minio/operator/pkg/runtime"
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

const (
	commandFilePath = "/temp"
	minioJobName    = "job.min.io/job-name"
	minioJobCRName  = "job.min.io/job-cr-name"
	// DefaultMCImage - job mc image
	DefaultMCImage = "minio/mc:latest"
	// MinioJobPhaseError - error
	MinioJobPhaseError = "Error"
	// MinioJobPhaseSuccess - success
	MinioJobPhaseSuccess = "Success"
	// MinioJobPhaseRunning - running
	MinioJobPhaseRunning = "Running"
	// MinioJobPhaseFailed - failed
	MinioJobPhaseFailed = "Failed"
)

var fileRegexp = regexp.MustCompile(`(\[file\([^)]+\)\])`)

var operationAlias = map[string]string{
	"make-bucket":      "mb",
	"admin/policy/add": "admin/policy/create",
}

// mc operator args...
var jobOperation = map[string]string{
	"mb":                  "[FLAGS] [ALIAS]/{name} --ignore-existing",
	"admin/user/add":      "[ALIAS] {user} {password}",
	"admin/policy/create": "[ALIAS] {name} [file(policy,json)]",
	"admin/policy/attach": "[ALIAS] {policy} --user {user}",
}

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
			newJob := new.(*batchjobv1.Job)
			jobName, ok := newJob.Labels[minioJobName]
			if !ok {
				return
			}
			jobCRName, ok := newJob.Labels[minioJobCRName]
			if !ok {
				return
			}
			val, ok := globalIntervalJobStatus.Load(fmt.Sprintf("%s/%s", newJob.GetNamespace(), jobCRName))
			if ok {
				intervalJob := val.(*MinIOIntervalJob)
				command, ok := intervalJob.CommandMap[jobName]
				if ok {
					if newJob.Status.Succeeded > 0 {
						command.setStatus(true, "")
					} else {
						for _, condition := range newJob.Status.Conditions {
							if condition.Type == batchjobv1.JobFailed {
								command.setStatus(false, condition.Message)
								break
							}
						}
					}
				}
			}
			controller.HandleObject(newJob)
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
		globalIntervalJobStatus.Delete(fmt.Sprintf("%s/%s", jobCR.Namespace, jobCR.Name))
		if errors.IsNotFound(err) {
			return WrapResult(Result{}, nil)
		}
		return WrapResult(Result{}, err)
	}
	// if job is success, do nothing
	if jobCR.Status.Phase == MinioJobPhaseSuccess {
		// delete the job status
		globalIntervalJobStatus.Delete(fmt.Sprintf("%s/%s", jobCR.Namespace, jobCR.Name))
		return WrapResult(Result{}, nil)
	}
	intervalJob, err := checkMinIOJob(&jobCR)
	if err != nil {
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
		jobCR.Status.Phase = MinioJobPhaseError
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
	err = intervalJob.createCommandJob(ctx, c.k8sClient)
	if err != nil {
		jobCR.Status.Phase = MinioJobPhaseError
		jobCR.Status.Message = fmt.Sprintf("Create job error:%v", err)
		err = c.updateJobStatus(ctx, &jobCR)
		return WrapResult(Result{}, err)
	}
	// update status
	jobCR.Status = intervalJob.getMinioJobStatus(ctx)
	err = c.updateJobStatus(ctx, &jobCR)
	return WrapResult(Result{}, err)
}

func (c *JobController) updateJobStatus(ctx context.Context, job *v1alpha1.MinIOJob) error {
	return c.k8sClient.Status().Update(ctx, job)
}

func operationAliasToMC(operation string) (op string, found bool) {
	for k, v := range operationAlias {
		if k == operation {
			return v, true
		}
		if v == operation {
			return v, true
		}
	}
	// operation like admin/policy/attach match nothing.
	// but it's a valid operation
	if strings.Contains(operation, "/") {
		return operation, true
	}
	// operation like replace match nothing
	// it's not a valid operation
	return "", false
}

// MinIOIntervalJobCommandFile - Job run command need a file such as /temp/policy.json
type MinIOIntervalJobCommandFile struct {
	Name    string
	Ext     string
	Dir     string
	Content string
}

// MinIOIntervalJobCommand - Job run command
type MinIOIntervalJobCommand struct {
	mutex       sync.RWMutex
	JobName     string
	MCOperation string
	Command     string
	DepnedsOn   []string
	Files       []MinIOIntervalJobCommandFile
	Succeeded   bool
	Message     string
	Created     bool
}

func (jobCommand *MinIOIntervalJobCommand) setStatus(success bool, message string) {
	if jobCommand == nil {
		return
	}
	jobCommand.mutex.Lock()
	jobCommand.Succeeded = success
	jobCommand.Message = message
	jobCommand.mutex.Unlock()
}

func (jobCommand *MinIOIntervalJobCommand) success() bool {
	if jobCommand == nil {
		return false
	}
	jobCommand.mutex.Lock()
	defer jobCommand.mutex.Unlock()
	return jobCommand.Succeeded
}

func (jobCommand *MinIOIntervalJobCommand) createJob(ctx context.Context, k8sClient client.Client, jobCR *v1alpha1.MinIOJob) error {
	if jobCommand == nil {
		return nil
	}
	jobCommand.mutex.RLock()
	if jobCommand.Created || jobCommand.Succeeded {
		jobCommand.mutex.RUnlock()
		return nil
	}
	jobCommand.mutex.RUnlock()
	jobCommands := []string{}
	commands := []string{"mc"}
	commands = append(commands, strings.SplitN(jobCommand.MCOperation, "/", -1)...)
	commands = append(commands, strings.SplitN(jobCommand.Command, " ", -1)...)
	for _, command := range commands {
		trimCommand := strings.TrimSpace(command)
		if trimCommand != "" {
			jobCommands = append(jobCommands, trimCommand)
		}
	}
	jobCommands = append(jobCommands, "--insecure")
	objs := []client.Object{}
	mcImage := jobCR.Spec.MCImage
	if mcImage == "" {
		mcImage = DefaultMCImage
	}
	job := &batchjobv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", jobCR.Name, jobCommand.JobName),
			Namespace: jobCR.Namespace,
			Labels: map[string]string{
				minioJobName:   jobCommand.JobName,
				minioJobCRName: jobCR.Name,
			},
			Annotations: map[string]string{
				"job.min.io/operation": jobCommand.MCOperation,
			},
		},
		Spec: batchjobv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						minioJobName: jobCommand.JobName,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: jobCR.Spec.ServiceAccountName,
					Containers: []corev1.Container{
						{
							Name:            "mc",
							Image:           mcImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{
									Name:  "MC_HOST_myminio",
									Value: fmt.Sprintf("https://$(ACCESS_KEY):$(SECRET_KEY)@minio.%s.svc.cluster.local", jobCR.Namespace),
								},
								{
									Name:  "MC_STS_ENDPOINT_myminio",
									Value: fmt.Sprintf("https://sts.%s.svc.cluster.local:4223/sts/%s", miniov2.GetNSFromFile(), jobCR.Namespace),
								},
								{
									Name:  "MC_WEB_IDENTITY_TOKEN_FILE_myminio",
									Value: "/var/run/secrets/kubernetes.io/serviceaccount/token",
								},
							},
							Command: jobCommands,
						},
					},
				},
			},
		},
	}
	if jobCR.Spec.FailureStrategy == v1alpha1.StopOnFailure {
		job.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyNever
	} else {
		job.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyOnFailure
	}
	if len(jobCommand.Files) > 0 {
		cmName := fmt.Sprintf("%s-%s-cm", jobCR.Name, jobCommand.JobName)
		job.Spec.Template.Spec.Containers[0].VolumeMounts = append(job.Spec.Template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      "file-volume",
			ReadOnly:  true,
			MountPath: jobCommand.Files[0].Dir,
		})
		job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: "file-volume",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cmName,
					},
				},
			},
		})
		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cmName,
				Namespace: jobCR.Namespace,
				Labels: map[string]string{
					"job.min.io/name": jobCR.Name,
				},
			},
			Data: map[string]string{},
		}
		for _, file := range jobCommand.Files {
			configMap.Data[fmt.Sprintf("%s.%s", file.Name, file.Ext)] = file.Content
		}
		objs = append(objs, configMap)
	}
	objs = append(objs, job)
	for _, obj := range objs {
		_, err := runtime2.NewObjectSyncer(ctx, k8sClient, jobCR, func() error {
			return nil
		}, obj, runtime2.SyncTypeCreateOrUpdate).Sync(ctx)
		if err != nil {
			return err
		}
	}
	jobCommand.mutex.Lock()
	jobCommand.Created = true
	jobCommand.mutex.Unlock()
	return nil
}

// MinIOIntervalJob - Interval Job
type MinIOIntervalJob struct {
	// to see if that change
	JobCR      *v1alpha1.MinIOJob
	Command    []*MinIOIntervalJobCommand
	CommandMap map[string]*MinIOIntervalJobCommand
}

func (intervalJob *MinIOIntervalJob) getMinioJobStatus(ctx context.Context) v1alpha1.MinIOJobStatus {
	status := v1alpha1.MinIOJobStatus{}
	failed := false
	running := false
	message := ""
	for _, command := range intervalJob.Command {
		command.mutex.RLock()
		if command.Succeeded {
			status.CommandsStatus = append(status.CommandsStatus, v1alpha1.CommandStatus{
				Name:    command.JobName,
				Result:  "success",
				Message: command.Message,
			})
		} else {
			failed = true
			message = command.Message
			// if success is false and message is empty, it means the job is running
			if command.Message == "" {
				running = true
				status.CommandsStatus = append(status.CommandsStatus, v1alpha1.CommandStatus{
					Name:    command.JobName,
					Result:  "running",
					Message: command.Message,
				})
			} else {
				status.CommandsStatus = append(status.CommandsStatus, v1alpha1.CommandStatus{
					Name:    command.JobName,
					Result:  "failed",
					Message: command.Message,
				})
			}
		}
		command.mutex.RUnlock()
	}
	if running {
		status.Phase = MinioJobPhaseRunning
	} else {
		if failed {
			status.Phase = MinioJobPhaseFailed
			status.Message = message
		} else {
			status.Phase = MinioJobPhaseSuccess
		}
	}
	return status
}

func (intervalJob *MinIOIntervalJob) createCommandJob(ctx context.Context, k8sClient client.Client) error {
	for _, command := range intervalJob.Command {
		if len(command.DepnedsOn) == 0 {
			err := command.createJob(ctx, k8sClient, intervalJob.JobCR)
			if err != nil {
				return err
			}
		} else {
			allDepsSuccess := true
			for _, dep := range command.DepnedsOn {
				status, found := intervalJob.CommandMap[dep]
				if !found {
					return fmt.Errorf("dependent job %s not found", dep)
				}
				if !status.success() {
					allDepsSuccess = false
					break
				}
			}
			if allDepsSuccess {
				err := command.createJob(ctx, k8sClient, intervalJob.JobCR)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func checkMinIOJob(jobCR *v1alpha1.MinIOJob) (intervalJob *MinIOIntervalJob, err error) {
	defer func() {
		if err != nil {
			globalIntervalJobStatus.Delete(fmt.Sprintf("%s/%s", jobCR.Namespace, jobCR.Name))
		}
	}()
	val, found := globalIntervalJobStatus.Load(fmt.Sprintf("%s/%s", jobCR.Namespace, jobCR.Name))
	if found {
		intervalJob = val.(*MinIOIntervalJob)
		if reflect.DeepEqual(intervalJob.JobCR.Spec, jobCR.Spec) {
			intervalJob.JobCR.UID = jobCR.UID
			return intervalJob, nil
		}
	}
	intervalJob = &MinIOIntervalJob{
		JobCR:      jobCR.DeepCopy(),
		Command:    []*MinIOIntervalJobCommand{},
		CommandMap: map[string]*MinIOIntervalJobCommand{},
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
		mcCommand, found := operationAliasToMC(val.Operation)
		if !found {
			return intervalJob, fmt.Errorf("operation %s is not supported", val.Operation)
		}
		var command string
		argsTpl, found := jobOperation[mcCommand]
		if !found {
			return intervalJob, fmt.Errorf("operation %s is not supported", mcCommand)
		}
		// [ALIAS]
		command = strings.ReplaceAll(argsTpl, "[ALIAS]", "myminio")
		flags := []string{}
		for flagName, flagVal := range val.Args {
			if strings.HasPrefix(flagName, "-") {
				// args:
				//   --with-locks: ""
				//   --region: us-west-2
				if flagVal == "" {
					flags = append(flags, flagName)
				}
				if strings.Contains(flagVal, ",") {
					for _, flagsVals := range strings.Split(flagVal, ",") {
						flags = append(flags, fmt.Sprintf("%s=%s", flagName, flagsVals))
					}
				}
			} else {
				// args:
				// 	 users: user1,user2
				if strings.Contains(flagVal, ",") {
					flagVal = strings.ReplaceAll(flagVal, ",", " ")
				}
				command = strings.ReplaceAll(command, fmt.Sprintf("{%s}", flagName), flagVal)
			}
		}

		// [FLAGS]
		command = strings.ReplaceAll(command, "[FLAGS]", strings.Join(flags, " "))
		// should replace all template
		if strings.Contains(command, "{") || strings.Contains(command, "}") {
			return intervalJob, fmt.Errorf("invalid args for %s", val.Operation)
		}
		jobCommand := MinIOIntervalJobCommand{
			JobName:     val.Name,
			MCOperation: mcCommand,
			Command:     command,
			DepnedsOn:   val.DependsOn,
			Files:       []MinIOIntervalJobCommandFile{},
		}
		// some commands need to have a empty name
		if jobCommand.JobName == "" {
			jobCommand.JobName = fmt.Sprintf("command-%d", index)
		}
		var matchString, name, ext string
		cmdFound := true
		for cmdFound {
			matchString, name, ext, cmdFound = getFileNameAndExt(command)
			if cmdFound {
				if _, ok := val.Args[name]; !ok {
					return intervalJob, fmt.Errorf("args %s not found", name)
				}
				jobCommandFile := MinIOIntervalJobCommandFile{
					Name:    name,
					Ext:     ext,
					Dir:     commandFilePath,
					Content: val.Args[name],
				}
				command = strings.ReplaceAll(command, matchString, fmt.Sprintf("%s/%s.%s", commandFilePath, name, ext))
				jobCommand.Command = command
				jobCommand.Files = append(jobCommand.Files, jobCommandFile)
			}
		}
		intervalJob.Command = append(intervalJob.Command, &jobCommand)
		intervalJob.CommandMap[jobCommand.JobName] = &jobCommand
	}
	// check all dependon
	for _, command := range intervalJob.Command {
		for _, dep := range command.DepnedsOn {
			_, found := intervalJob.CommandMap[dep]
			if !found {
				return intervalJob, fmt.Errorf("dependent job %s not found", dep)
			}
		}
	}
	globalIntervalJobStatus.Store(fmt.Sprintf("%s/%s", jobCR.Namespace, jobCR.Name), intervalJob)
	return intervalJob, nil
}

func getFileNameAndExt(commond string) (matchString, name, ext string, found bool) {
	match := fileRegexp.MatchString(commond)
	if !match {
		return "", "", "", false
	}
	matchStrings := fileRegexp.FindAllStringSubmatch(commond, -1)
	for _, val := range matchStrings {
		if len(val) == 2 {
			// val[1] must be `[file(policy,json)]`
			if strings.Contains(val[1], ",") {
				args := strings.Replace(val[1], "[file(", "", 1)
				args = strings.Replace(args, ")]", "", 1)
				fileArgs := strings.SplitN(args, ",", 2)
				return val[1], fileArgs[0], fileArgs[1], true
			}
		}
	}
	return "", "", "", false
}

var globalIntervalJobStatus = sync.Map{}
