// This file is part of MinIO Operator
// Copyright (c) 2021 MinIO, Inc.
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
	"k8s.io/klog/v2"
	"time"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// handlePodChange will handle changes in pods and queue it for processing, pods are already filtered by PodInformer
func (c *Controller) handlePodChange(obj interface{}) {
	// NOTE: currently only Tenant pods are being monitored by the Pod Informer
	object, err := utils.CastObjectToMetaV1(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}

	instanceName, ok := object.GetLabels()[miniov2.TenantLabel]
	if !ok {
		utilruntime.HandleError(fmt.Errorf("label: %s not found in %s", miniov2.TenantLabel, object.GetName()))
		return
	}

	key := fmt.Sprintf("%s/%s", object.GetNamespace(), instanceName)
	c.healthCheckQueue.AddAfter(key, 1*time.Second)
}

// DeletePodsByStatefulSet deletes all pods associated with a statefulset
func (c *Controller) DeletePodsByStatefulSet(ctx context.Context, sts *appsv1.StatefulSet) (err error) {
	listOpt := &client.ListOptions{
		Namespace: sts.Namespace,
	}
	client.MatchingLabels(sts.Spec.Template.Labels).ApplyToList(listOpt)
	podList := &corev1.PodList{}
	err = c.k8sClient.List(ctx, podList, listOpt)
	if err != nil {
		return err
	}
	for _, item := range podList.Items {
		if item.DeletionTimestamp == nil {
			err = c.k8sClient.Delete(ctx, &item)
			// Ignore Not Found
			if client.IgnoreNotFound(err) != nil {
				klog.Infof("unable to restart %s/%s (ignored): %s", item.Namespace, item.Name, err)
			}
		}
	}
	return
}
