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

package miniojob

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/minio/operator/pkg/apis/job.min.io/v1alpha1"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/runtime"
	batchjobv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// DefaultMCImage - job mc image
	DefaultMCImage = "minio/mc:latest"
	// MinioJobName - job name
	MinioJobName = "job.min.io/job-name"
	// MinioJobCRName - job cr name
	MinioJobCRName = "job.min.io/job-cr-name"
	// CommandFilePath - command file path
	CommandFilePath = "/temp"
	// MinioJobPhaseError - error
	MinioJobPhaseError = "Error"
	// MinioJobPhaseSuccess - Success
	MinioJobPhaseSuccess = "Success"
	// MinioJobPhaseRunning - running
	MinioJobPhaseRunning = "Running"
	// MinioJobPhaseFailed - failed
	MinioJobPhaseFailed = "Failed"
)

var operationAlias = map[string]string{
	"make-bucket":      "mb",
	"admin/policy/add": "admin/policy/create",
}

// JobOperation - job operation
var JobOperation = map[string][]FieldsFunc{
	"mb":                  {FLAGS(), Sanitize(ALIAS(), Static("/"), Key("name")), Static("--ignore-existing")},
	"admin/user/add":      {ALIAS(), Key("user"), Key("password")},
	"admin/policy/create": {ALIAS(), Key("name"), Key("policy")},
	"admin/policy/attach": {ALIAS(), Key("policy"), OneOf(KeyFormat("user", "--user"), KeyFormat("group", "--group"))},
	"admin/config/set":    {ALIAS(), Key("webhookName"), Option(KeyValue("endpoint")), OthersKeyValues()},
	"support/callhome":    {Key("action"), ALIAS(), FLAGS()},
	"license/register":    {ALIAS(), OthersKeyValues()},
}

// OperationAliasToMC - convert operation to mc operation
func OperationAliasToMC(operation string) (op string, found bool) {
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

// MinIOIntervalJobCommand - Job run command
type MinIOIntervalJobCommand struct {
	mutex       sync.RWMutex
	CommandSpec v1alpha1.CommandSpec
	JobName     string
	MCOperation string
	Command     string
	Succeeded   bool
	Message     string
	Created     bool
}

// SetStatus - set job command status
func (jobCommand *MinIOIntervalJobCommand) SetStatus(success bool, message string) {
	if jobCommand == nil {
		return
	}
	jobCommand.mutex.Lock()
	jobCommand.Succeeded = success
	jobCommand.Message = message
	jobCommand.mutex.Unlock()
}

// Success - check job command status
func (jobCommand *MinIOIntervalJobCommand) Success() bool {
	if jobCommand == nil {
		return false
	}
	jobCommand.mutex.Lock()
	defer jobCommand.mutex.Unlock()
	return jobCommand.Succeeded
}

// createJob - create job
func (jobCommand *MinIOIntervalJobCommand) createJob(ctx context.Context, k8sClient client.Client, jobCR *v1alpha1.MinIOJob, stsPort int) (objs []client.Object) {
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
	if len(jobCommand.CommandSpec.Command) == 0 {
		commands := []string{"mc"}
		commands = append(commands, strings.SplitN(jobCommand.MCOperation, "/", -1)...)
		commands = append(commands, strings.SplitN(jobCommand.Command, " ", -1)...)
		for _, command := range commands {
			trimmedCommand := strings.TrimSpace(command)
			if trimmedCommand != "" {
				jobCommands = append(jobCommands, trimmedCommand)
			}
		}
	} else {
		jobCommands = append(jobCommands, jobCommand.CommandSpec.Command...)
	}
	mcImage := jobCR.Spec.MCImage
	if mcImage == "" {
		mcImage = DefaultMCImage
	}
	baseVolumeMounts := []corev1.VolumeMount{
		{
			Name:      "config-dir",
			MountPath: "/.mc",
		},
	}
	baseVolumeMounts = append(baseVolumeMounts, jobCommand.CommandSpec.VolumeMounts...)
	baseVolumes := []corev1.Volume{
		{
			Name: "config-dir",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
	baseVolumes = append(baseVolumes, jobCommand.CommandSpec.Volumes...)
	baseEnvFrom := []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: fmt.Sprintf("%s-job-secret", jobCR.Name),
				},
			},
		},
	}
	baseEnvFrom = append(baseEnvFrom, jobCommand.CommandSpec.EnvFrom...)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-job-secret", jobCR.Name),
			Namespace: jobCR.Namespace,
		},
		StringData: map[string]string{
			"MC_HOST_myminio":                    fmt.Sprintf("https://$(ACCESS_KEY):$(SECRET_KEY)@minio.%s.svc.cluster.local", jobCR.Namespace),
			"MC_STS_ENDPOINT_myminio":            fmt.Sprintf("https://sts.%s.svc.cluster.local:%d/sts/%s", miniov2.GetNSFromFile(), stsPort, jobCR.Namespace),
			"MC_WEB_IDENTITY_TOKEN_FILE_myminio": "/var/run/secrets/kubernetes.io/serviceaccount/token",
		},
	}
	objs = append(objs, secret)
	job := &batchjobv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", jobCR.Name, jobCommand.JobName),
			Namespace: jobCR.Namespace,
			Labels: map[string]string{
				MinioJobName:   jobCommand.JobName,
				MinioJobCRName: jobCR.Name,
			},
			Annotations: map[string]string{
				"job.min.io/operation": jobCommand.MCOperation,
			},
		},
		Spec: batchjobv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						MinioJobName: jobCommand.JobName,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: jobCR.Spec.ServiceAccountName,
					Containers: []corev1.Container{
						{
							Name:            "mc",
							Image:           mcImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env:             jobCommand.CommandSpec.Env,
							EnvFrom:         baseEnvFrom,
							Command:         jobCommands,
							SecurityContext: jobCR.Spec.ContainerSecurityContext,
							VolumeMounts:    baseVolumeMounts,
							Resources:       jobCommand.CommandSpec.Resources,
						},
					},
					SecurityContext: jobCR.Spec.SecurityContext,
					Volumes:         baseVolumes,
				},
			},
		},
	}
	if jobCR.Spec.FailureStrategy == v1alpha1.StopOnFailure {
		job.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyNever
	} else {
		job.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyOnFailure
	}
	objs = append(objs, job)
	return objs
}

// CreateJob - create job
func (jobCommand *MinIOIntervalJobCommand) CreateJob(ctx context.Context, k8sClient client.Client, jobCR *v1alpha1.MinIOJob, stsPort int) error {
	for _, obj := range jobCommand.createJob(ctx, k8sClient, jobCR, stsPort) {
		if obj == nil {
			continue
		}
		_, err := runtime.NewObjectSyncer(ctx, k8sClient, jobCR, func() error {
			return nil
		}, obj, runtime.SyncTypeCreateOrUpdate).Sync(ctx)
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

// GetMinioJobStatus - get job status
func (intervalJob *MinIOIntervalJob) GetMinioJobStatus(ctx context.Context) v1alpha1.MinIOJobStatus {
	status := v1alpha1.MinIOJobStatus{}
	failed := false
	running := false
	message := ""
	for _, command := range intervalJob.Command {
		command.mutex.RLock()
		if command.Succeeded {
			status.CommandsStatus = append(status.CommandsStatus, v1alpha1.CommandStatus{
				Name:    command.JobName,
				Result:  "Success",
				Message: command.Message,
			})
		} else {
			failed = true
			message = command.Message
			// if Success is false and message is empty, it means the job is running
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

// CreateCommandJob - create command job
func (intervalJob *MinIOIntervalJob) CreateCommandJob(ctx context.Context, k8sClient client.Client, stsPort int) error {
	for _, command := range intervalJob.Command {
		if len(command.CommandSpec.DependsOn) == 0 {
			err := command.CreateJob(ctx, k8sClient, intervalJob.JobCR, stsPort)
			if err != nil {
				return err
			}
		} else {
			allDepsSuccess := true
			for _, dep := range command.CommandSpec.DependsOn {
				status, found := intervalJob.CommandMap[dep]
				if !found {
					return fmt.Errorf("dependent job %s not found", dep)
				}
				if !status.Success() {
					allDepsSuccess = false
					break
				}
			}
			if allDepsSuccess {
				err := command.CreateJob(ctx, k8sClient, intervalJob.JobCR, stsPort)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// GenerateMinIOIntervalJobCommand - generate command
func GenerateMinIOIntervalJobCommand(commandSpec v1alpha1.CommandSpec, commandIndex int) (*MinIOIntervalJobCommand, error) {
	jobCommand := &MinIOIntervalJobCommand{
		JobName:     commandSpec.Name,
		CommandSpec: commandSpec,
	}
	if len(commandSpec.Command) == 0 {
		mcCommand, found := OperationAliasToMC(commandSpec.Operation)
		if !found {
			return nil, fmt.Errorf("operation %s is not supported", commandSpec.Operation)
		}
		argsFuncs, found := JobOperation[mcCommand]
		if !found {
			return nil, fmt.Errorf("operation %s is not supported", mcCommand)
		}
		commands := []string{}
		for _, argsFunc := range argsFuncs {
			jobArg, err := argsFunc(commandSpec.Args)
			if err != nil {
				return nil, err
			}
			if jobArg.Command != "" {
				commands = append(commands, jobArg.Command)
			}

		}
		jobCommand.MCOperation = mcCommand
		jobCommand.Command = strings.Join(commands, " ")
	}
	// some commands need to have a empty name
	if jobCommand.JobName == "" {
		jobCommand.JobName = fmt.Sprintf("command-%d", commandIndex)
	}
	return jobCommand, nil
}
