// This file is part of MinIO Operator
// Copyright (c) 2022 MinIO, Inc.
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

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/minio/operator/pkg/http"

	"github.com/minio/operator/api/operations/operator_api"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	utils2 "github.com/minio/operator/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// updateTenantAction does an update on the minioTenant by patching the desired changes
func updateTenantAction(ctx context.Context, operatorClient OperatorClientI, clientset K8sClientI, httpCl http.ClientI, namespace string, params operator_api.UpdateTenantParams) error {
	imageToUpdate := params.Body.Image
	imageRegistryReq := params.Body.ImageRegistry

	minInst, err := operatorClient.TenantGet(ctx, namespace, params.Tenant, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// we can take either the `image_pull_secret` of the `image_registry` but not both
	if params.Body.ImagePullSecret != "" {
		minInst.Spec.ImagePullSecret.Name = params.Body.ImagePullSecret
	} else {
		// update the image pull secret content
		if _, err := setImageRegistry(ctx, imageRegistryReq, clientset, namespace, params.Tenant); err != nil {
			LogError("error setting image registry secret: %v", err)
			return err
		}
	}

	// if image to update is empty we'll use the latest image by default
	if strings.TrimSpace(imageToUpdate) != "" {
		minInst.Spec.Image = imageToUpdate
	} else {
		im, err := utils2.GetLatestMinIOImage(httpCl)
		// if we can't get the MinIO image, we won' auto-update it unless it's explicit by name
		if err == nil {
			minInst.Spec.Image = *im
		}
	}

	// Prometheus Annotations
	currentAnnotations := minInst.Annotations
	prometheusAnnotations := map[string]string{
		prometheusPath:   "/minio/prometheus/metrics",
		prometheusPort:   fmt.Sprint(miniov2.MinIOPort),
		prometheusScrape: "true",
	}
	if params.Body.EnablePrometheus && currentAnnotations != nil {
		// add prometheus annotations to the tenant
		minInst.Annotations = addAnnotations(currentAnnotations, prometheusAnnotations)
		// add prometheus annotations to the each pool
		if minInst.Spec.Pools != nil {
			for _, pool := range minInst.Spec.Pools {
				poolAnnotations := pool.VolumeClaimTemplate.GetObjectMeta().GetAnnotations()
				pool.VolumeClaimTemplate.GetObjectMeta().SetAnnotations(addAnnotations(poolAnnotations, prometheusAnnotations))
			}
		}
	} else {
		// remove prometheus annotations to the tenant
		minInst.Annotations = removeAnnotations(currentAnnotations, prometheusAnnotations)
		// add prometheus annotations from each pool
		if minInst.Spec.Pools != nil {
			for _, pool := range minInst.Spec.Pools {
				poolAnnotations := pool.VolumeClaimTemplate.GetObjectMeta().GetAnnotations()
				pool.VolumeClaimTemplate.GetObjectMeta().SetAnnotations(removeAnnotations(poolAnnotations, prometheusAnnotations))
			}
		}
	}

	payloadBytes, err := json.Marshal(minInst)
	if err != nil {
		return err
	}
	_, err = operatorClient.TenantPatch(ctx, namespace, minInst.Name, types.MergePatchType, payloadBytes, metav1.PatchOptions{})
	if err != nil {
		return err
	}
	return nil
}
