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
	"fmt"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/utils"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
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
	c.healthCheckQueue.Add(key)
}
