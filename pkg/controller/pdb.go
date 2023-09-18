// Copyright (C) 2023, MinIO, Inc.
//
// This code is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License, version 3,
// as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License, version 3,
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package controller

import (
	"context"
	"fmt"
	"strings"
	"sync"

	v2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/runtime"
	v1 "k8s.io/api/policy/v1"
	"k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DeletePDB - delete PDB for tenant
func (c *Controller) DeletePDB(ctx context.Context, t *v2.Tenant) (err error) {
	available := c.GetPDBAvailable()
	if !available.Available() {
		return nil
	}
	listOpt := &client.ListOptions{
		Namespace: t.Namespace,
	}
	client.MatchingLabels{
		v2.TenantLabel: t.Name,
	}.ApplyToList(listOpt)
	if available.V1Available() {
		pdbS := &v1.PodDisruptionBudgetList{}
		err = c.k8sClient.List(ctx, pdbS, listOpt)
		if err != nil {
			return err
		}
		for _, item := range pdbS.Items {
			err = c.k8sClient.Delete(ctx, &item)
			if err != nil {
				return err
			}
		}
	}
	if available.V1BetaAvailable() {
		pdbS := &v1beta1.PodDisruptionBudgetList{}
		err = c.k8sClient.List(ctx, pdbS, listOpt)
		if err != nil {
			return err
		}
		for _, item := range pdbS.Items {
			err = c.k8sClient.Delete(ctx, &item)
			if err != nil {
				return err
			}
		}
	}
	return err
}

// CreateOrUpdatePDB - hold PDB as expected
func (c *Controller) CreateOrUpdatePDB(ctx context.Context, t *v2.Tenant) (err error) {
	available := c.GetPDBAvailable()
	if !available.Available() {
		return nil
	}
	for _, pool := range t.Spec.Pools {
		if strings.TrimSpace(pool.Name) == "" {
			continue
		}
		// No PodDisruptionBudget for minAvailable equal server's numbers.
		if pool.Servers == (int32(pool.Servers/2) + 1) {
			continue
		}
		if available.Available() {
			// check sts status first.
			ssName := t.PoolStatefulsetName(&pool)
			existingStatefulSet, err := c.statefulSetLister.StatefulSets(t.Namespace).Get(ssName)
			if err != nil {
				return err
			}
			if existingStatefulSet.Status.ReadyReplicas != existingStatefulSet.Status.Replicas || existingStatefulSet.Status.Replicas == 0 {
				continue
			}
			if t.Status.CurrentState != StatusInitialized {
				return nil
			}
		}
		var pdbI client.Object
		if available.V1Available() {
			pdbI = &v1.PodDisruptionBudget{}
		} else if available.V1BetaAvailable() {
			pdbI = &v1beta1.PodDisruptionBudget{}
		} else {
			return nil
		}
		pdbI.SetName(t.Name + "-" + pool.Name)
		pdbI.SetNamespace(t.Namespace)
		_, err := runtime.NewObjectSyncer(ctx, c.k8sClient, t, func() error {
			if available.V1Available() {
				pdb := pdbI.(*v1.PodDisruptionBudget)
				minAvailable := intstr.FromInt(int(pool.Servers/2) + 1)
				pdb.Spec.MinAvailable = &minAvailable
				pdb.Labels = map[string]string{
					v2.TenantLabel: t.Name,
					v2.PoolLabel:   pool.Name,
				}
				pdb.Spec.Selector = metav1.SetAsLabelSelector(labels.Set{
					v2.TenantLabel: t.Name,
					v2.PoolLabel:   pool.Name,
				})
			}
			if available.V1BetaAvailable() {
				pdb := pdbI.(*v1beta1.PodDisruptionBudget)
				minAvailable := intstr.FromInt(int(pool.Servers/2) + 1)
				pdb.Spec.MinAvailable = &minAvailable
				pdb.Labels = map[string]string{
					v2.TenantLabel: t.Name,
					v2.PoolLabel:   pool.Name,
				}
				pdb.Spec.Selector = metav1.SetAsLabelSelector(labels.Set{
					v2.TenantLabel: t.Name,
					v2.PoolLabel:   pool.Name,
				})
			}
			return nil
		}, pdbI, runtime.SyncTypeCreateOrUpdate).Sync(ctx)
		if err != nil {
			return err
		}
	}
	if len(t.Spec.Pools) == 0 {
		return fmt.Errorf("%s empty pools", t.Name)
	}
	return nil
}

// PDBAvailable - v1 for v1.PDB and v1beta for v1beta.PDB,flag for support or not
type PDBAvailable struct {
	v1     bool
	v1beta bool
}

// V1Available - show if it supports PDB v1
func (p *PDBAvailable) V1Available() bool {
	return p.v1
}

// V1BetaAvailable - show if it supports PDB v1beta
func (p *PDBAvailable) V1BetaAvailable() bool {
	return p.v1beta
}

// Available - show if it supports PDB
func (p *PDBAvailable) Available() bool {
	return p.v1 || p.v1beta
}

var (
	globalPDBAvailable     PDBAvailable
	globalPDBAvailableOnce sync.Once
)

// GetPDBAvailable - return globalPDBAvailable
// thread safe
func (c *Controller) GetPDBAvailable() PDBAvailable {
	globalPDBAvailableOnce.Do(func() {
		defer func() {
			if globalPDBAvailable.v1 {
				klog.Infof("PodDisruptionBudget: v1")
			} else if globalPDBAvailable.v1beta {
				klog.Infof("PodDisruptionBudget: v1beta")
			} else {
				klog.Infof("PodDisruptionBudget: not supported")
			}
		}()
		resouces, _ := c.kubeClientSet.Discovery().ServerPreferredResources()
		for _, r := range resouces {
			if r.GroupVersion == "policy/v1" {
				for _, api := range r.APIResources {
					if api.Kind == "PodDisruptionBudget" {
						globalPDBAvailable.v1 = true
						return
					}
				}
			}
			if r.GroupVersion == "policy/v1beta" {
				for _, api := range r.APIResources {
					if api.Kind == "PodDisruptionBudget" {
						globalPDBAvailable.v1beta = true
						return
					}
				}
			}
		}
	})
	return globalPDBAvailable
}
