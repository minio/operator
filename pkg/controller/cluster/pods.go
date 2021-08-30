// This file is part of MinIO Console Server
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

package cluster

import (
	"context"
	"fmt"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

// handlePodChange will handle changes in pods and see if they are a MinIO pod, if so, queue it for processing
func (c *Controller) handlePodChange(obj interface{}) {
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
	// if the pod is a MinIO pod, we do process it
	if tenantName, ok := object.GetLabels()[miniov2.TenantLabel]; ok {
		// check this is still a valid tenant
		_, err := c.minioClientSet.MinioV2().Tenants(object.GetNamespace()).Get(context.Background(), tenantName, metav1.GetOptions{})
		if err != nil {
			klog.V(4).Infof("ignoring orphaned object '%s' of tenant '%s'", object.GetSelfLink(), tenantName)
			return
		}
		key := fmt.Sprintf("%s/%s", object.GetNamespace(), tenantName)
		// check if already in queue before re-queueing

		c.healthCheckQueue.Add(key)
		return
	}
}
