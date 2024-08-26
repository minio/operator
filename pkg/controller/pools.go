// This file is part of MinIO Operator
// Copyright (c) 2020 MinIO, Inc.
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
	"errors"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
)

func (c *Controller) getSSForPool(tenant *miniov2.Tenant, pool *miniov2.Pool) (*appsv1.StatefulSet, error) {
	ss, err := c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Get(context.Background(), tenant.PoolStatefulsetName(pool), metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, err
		}
		// check if there are legacy statefulsets
		ss, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Get(context.Background(), tenant.LegacyStatefulsetName(pool), metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
	}
	return ss, nil
}

func (c *Controller) getAllSSForTenant(tenant *miniov2.Tenant) (map[int]*appsv1.StatefulSet, error) {
	poolDir := make(map[int]*appsv1.StatefulSet)
	// TODO: Load all statefulsets by using the tenant label in a single list call
	for i := range tenant.Spec.Pools {
		ss, err := c.getSSForPool(tenant, &tenant.Spec.Pools[i])
		if err != nil && !k8serrors.IsNotFound(err) {
			return nil, err
		}
		if ss != nil {
			poolDir[i] = ss
		}
	}
	return poolDir, nil
}

// poolSSMatchesSpec checks if the statefulset for the pool matches what is expected and described from the Tenant
func poolSSMatchesSpec(expectedStatefulSet, existingStatefulSet *appsv1.StatefulSet) (bool, error) {
	if expectedStatefulSet == nil || existingStatefulSet == nil {
		return false, errors.New("cannot process an empty MinIO StatefulSet")
	}
	// Try to detect changes in the labels or annotations
	expectedMetadata := expectedStatefulSet.ObjectMeta
	if !equality.Semantic.DeepEqual(expectedMetadata.Labels, existingStatefulSet.ObjectMeta.Labels) {
		return false, nil
	}
	expectedAnnotations := map[string]string{}
	for k, v := range expectedMetadata.Annotations {
		expectedAnnotations[k] = v
	}
	currentAnnotations := existingStatefulSet.ObjectMeta.Annotations
	delete(expectedAnnotations, corev1.LastAppliedConfigAnnotation)
	delete(currentAnnotations, corev1.LastAppliedConfigAnnotation)
	if !equality.Semantic.DeepEqual(expectedAnnotations, currentAnnotations) {
		return false, nil
	}
	if miniov2.IsContainersEnvUpdated(existingStatefulSet.Spec.Template.Spec.Containers, expectedStatefulSet.Spec.Template.Spec.Containers) {
		return false, nil
	}
	if !equality.Semantic.DeepEqual(expectedStatefulSet.Spec, existingStatefulSet.Spec) {
		// some field set by the operator has changed
		return false, nil
	}
	return true, nil
}

// getSystemCfgFromStatefulSet gets the MinIO environment variables from a statefulset
func (c *Controller) getSystemCfgFromStatefulSet(ctx context.Context, sts *appsv1.StatefulSet) (systemCfg map[string]string, err error) {
	systemCfg = make(map[string]string)
	for _, container := range sts.Spec.Template.Spec.Containers {
		if container.Name == miniov2.MinIOServerName {
			for _, e := range container.Env {
				if strings.HasPrefix(e.Name, "MINIO_") {
					switch {
					case e.Value != "":
						systemCfg[e.Name] = e.Value
					case e.ValueFrom != nil && e.ValueFrom.SecretKeyRef != nil:
						secret, err := c.kubeClientSet.CoreV1().Secrets(sts.Namespace).Get(ctx, e.ValueFrom.SecretKeyRef.Name, metav1.GetOptions{})
						if err != nil {
							return nil, err
						}
						value, ok := secret.Data[e.ValueFrom.SecretKeyRef.Key]
						if !ok {
							return nil, fmt.Errorf("secret %s does not have key %s", e.ValueFrom.SecretKeyRef.Name, e.ValueFrom.SecretKeyRef.Key)
						}
						systemCfg[e.Name] = string(value)
					case e.ValueFrom != nil && e.ValueFrom.ConfigMapKeyRef != nil:
						configMap, err := c.kubeClientSet.CoreV1().ConfigMaps(sts.Namespace).Get(ctx, e.ValueFrom.ConfigMapKeyRef.Name, metav1.GetOptions{})
						if err != nil {
							return nil, err
						}
						value, ok := configMap.Data[e.ValueFrom.ConfigMapKeyRef.Key]
						if !ok {
							return nil, fmt.Errorf("configmap %s does not have key %s", e.ValueFrom.ConfigMapKeyRef.Name, e.ValueFrom.ConfigMapKeyRef.Key)
						}
						systemCfg[e.Name] = value
					default:
						return nil, fmt.Errorf("unsupported env var %s", e.Name)
					}
				}
			}
		}
	}
	return
}

// restartInitializedPool restarts a pool that is assumed to have been initialized
func (c *Controller) restartInitializedPool(ctx context.Context, tenant *miniov2.Tenant, pool miniov2.Pool, tenantConfiguration map[string][]byte) error {
	// get a new admin client that points to a pod of an already initialized pool (ie: pool-0)
	livePods, err := c.kubeClientSet.CoreV1().Pods(tenant.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", miniov2.PoolLabel, pool.Name),
	})
	if err != nil {
		klog.Warning("Could not validate state of statefulset for pool", err)
	}
	if len(livePods.Items) == 0 {
		livePods, err = c.kubeClientSet.CoreV1().Pods(tenant.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", miniov2.ZoneLabel, pool.Name),
		})
		if err != nil {
			klog.Warning("Could not validate state of statefulset for zone", err)
			return err
		}
	}
	var livePod *corev1.Pod
	for _, p := range livePods.Items {
		if p.Status.Phase == corev1.PodRunning {
			livePod = &p
			break
		}
	}
	if livePod == nil {
		return fmt.Errorf("no running pods found for statefulsets %s", pool.Name)
	}

	livePodAddress := fmt.Sprintf("%s:9000", tenant.MinIOHLPodHostname(livePod.Name))
	livePodAdminClnt, err := tenant.NewMinIOAdminForAddress(livePodAddress, tenantConfiguration, c.getTransport())
	if err != nil {
		return err
	}

	// Now tell MinIO to restart
	if err = livePodAdminClnt.ServiceRestart(ctx); err != nil {
		klog.Infof("We failed to restart MinIO to adopt the new pool: %v", err)
		return err
	}

	return nil
}
