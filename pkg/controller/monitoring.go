// Copyright (C) 2021, MinIO, Inc.
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
	"time"

	"github.com/minio/madmin-go/v3"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/v2"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
)

const (
	// HealthUnavailableMessage means MinIO is down
	HealthUnavailableMessage = "Service Unavailable"
	// HealthHealingMessage means MinIO is healing one of more drives
	HealthHealingMessage = "Healing"
	// HealthReduceAvailabilityMessage some drives are offline
	HealthReduceAvailabilityMessage = "Reduced Availability"
)

// recurrentTenantStatusMonitor loop that checks every N minutes for tenants health
func (c *Controller) recurrentTenantStatusMonitor(stopCh <-chan struct{}) {
	// How often will this function run
	interval := miniov2.GetMonitoringInterval()
	ticker := time.NewTicker(time.Duration(interval) * time.Minute)
	defer func() {
		klog.Info("recurrent pod status monitor closed")
	}()
	for {
		select {
		case <-ticker.C:
			if err := c.tenantsHealthMonitor(); err != nil {
				klog.Infof("%v", err)
			}
		case <-stopCh:
			ticker.Stop()
			return
		}
	}
}

func (c *Controller) tenantsHealthMonitor() error {
	// list all tenants and get their cluster health
	tenants, err := c.minioClientSet.MinioV2().Tenants("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, tenant := range tenants.Items {
		if _, err = c.updateHealthStatusForTenant(&tenant); err != nil {
			klog.Errorf("%v", err)
			return err
		}
	}
	return nil
}

func (c *Controller) updateHealthStatusForTenant(tenant *miniov2.Tenant) (*miniov2.Tenant, error) {
	// don't get the tenant cluster health if it doesn't have at least 1 pool initialized
	oneInitialized := false
	for _, pool := range tenant.Status.Pools {
		if pool.State == miniov2.PoolInitialized {
			oneInitialized = true
		}
	}
	if !oneInitialized {
		klog.Infof("'%s/%s' no pool is initialized", tenant.Namespace, tenant.Name)
		return tenant, nil
	}

	tenantConfiguration, err := c.getTenantCredentials(context.Background(), tenant)
	if err != nil {
		return nil, err
	}

	adminClnt, err := tenant.NewMinIOAdmin(tenantConfiguration, c.getTransport())
	if err != nil {
		klog.Errorf("Error instantiating adminClnt '%s/%s': %v", tenant.Namespace, tenant.Name, err)
		return nil, err
	}

	aClnt, err := madmin.NewAnonymousClient(tenant.MinIOServerHostAddress(), tenant.TLS())
	if err != nil {
		// show the error and continue
		klog.Infof("'%s/%s': %v", tenant.Namespace, tenant.Name, err)
		return tenant, nil
	}
	aClnt.SetCustomTransport(c.getTransport())

	hctx, hcancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer hcancel()

	// get cluster health for tenant
	healthResult, err := aClnt.Healthy(hctx, madmin.HealthOpts{})
	if err != nil {
		// show the error and continue
		klog.Infof("'%s/%s' Failed to get cluster health: %v", tenant.Namespace, tenant.Name, err)
		return tenant, nil
	}

	tenant.Status.DrivesHealing = int32(healthResult.HealingDrives)
	tenant.Status.WriteQuorum = int32(healthResult.WriteQuorum)

	if healthResult.Healthy {
		tenant.Status.HealthStatus = miniov2.HealthStatusGreen
		tenant.Status.HealthMessage = ""
	} else {
		tenant.Status.HealthStatus = miniov2.HealthStatusRed
		tenant.Status.HealthMessage = HealthUnavailableMessage
	}

	// check all the tenant pods, if at least 1 is not running, we go yellow
	tenantPods, err := c.kubeClientSet.CoreV1().Pods(tenant.Namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", miniov2.TenantLabel, tenant.Name),
	})
	if err != nil {
		return nil, err
	}

	allPodsRunning := true
	for _, pod := range tenantPods.Items {
		if pod.Status.Phase != corev1.PodRunning {
			allPodsRunning = false
		}
	}
	if !allPodsRunning && tenant.Status.HealthStatus != miniov2.HealthStatusRed {
		tenant.Status.HealthStatus = miniov2.HealthStatusYellow
	}

	// partial status update, since the storage info might take a while
	if tenant, err = c.updatePoolStatus(context.Background(), tenant); err != nil {
		klog.Infof("'%s/%s' Can't update tenant status: %v", tenant.Namespace, tenant.Name, err)
	}

	srvInfoCtx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	storageInfo, err := adminClnt.StorageInfo(srvInfoCtx)
	if err != nil {
		// show the error and continue
		klog.Infof("'%s/%s' Failed to get storage info: %v", tenant.Namespace, tenant.Name, err)
		return tenant, nil
	}

	// Add back "Usable Capacity" & "Internal" values in Tenant Status and in the UI
	// How much is available: "Usable Capacity" in UI comes from "tenant.Status.Usage.Capacity"
	// How much is used: "Internal" in UI comes from "tenant.Status.Usage.Usage"
	UsedSpace := int64(0)      // How much is used
	AvailableSpace := int64(0) // How much is available
	for _, disk := range storageInfo.Disks {
		UsedSpace = UsedSpace + int64(disk.UsedSpace)
		AvailableSpace = AvailableSpace + int64(disk.AvailableSpace)
	}
	tenant.Status.Usage.Usage = UsedSpace
	tenant.Status.Usage.Capacity = AvailableSpace

	var rawUsage uint64

	var onlineDisks int32
	var offlineDisks int32
	for _, d := range storageInfo.Disks {
		if d.State == madmin.DriveStateOk {
			onlineDisks++
		} else {
			offlineDisks++
		}
		rawUsage = rawUsage + d.UsedSpace
	}

	tenant.Status.DrivesOnline = onlineDisks
	tenant.Status.DrivesOffline = offlineDisks

	tenant.Status.Usage.RawUsage = int64(rawUsage)

	var rawCapacity int64
	for _, p := range tenant.Spec.Pools {
		rawCapacity = rawCapacity + (int64(p.Servers) * int64(p.VolumesPerServer) * p.VolumeClaimTemplate.Spec.Resources.Requests.Storage().Value())
	}
	tenant.Status.Usage.RawCapacity = rawCapacity

	if tenant.Status.DrivesOffline > 0 || tenant.Status.DrivesHealing > 0 {
		tenant.Status.HealthStatus = miniov2.HealthStatusYellow
		if tenant.Status.DrivesHealing > 0 {
			tenant.Status.HealthMessage = HealthHealingMessage
		} else {
			tenant.Status.HealthMessage = HealthReduceAvailabilityMessage
		}
	}
	if tenant.Status.DrivesOnline < tenant.Status.WriteQuorum {
		tenant.Status.HealthStatus = miniov2.HealthStatusRed
		tenant.Status.HealthMessage = HealthUnavailableMessage
	}

	// only if no disks are offline and we are not healing, we are green
	if tenant.Status.DrivesOffline == 0 && tenant.Status.DrivesHealing == 0 {
		tenant.Status.HealthStatus = miniov2.HealthStatusGreen
		tenant.Status.HealthMessage = ""
	}

	if tenant, err = c.updatePoolStatus(context.Background(), tenant); err != nil {
		klog.Infof("'%s/%s' Can't update tenant status: %v", tenant.Namespace, tenant.Name, err)
	}

	// Store the usage reported by the tiers
	tiersStatsCtx, cancelTiers := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelTiers()
	tInfos, err := adminClnt.TierStats(tiersStatsCtx)
	if err != nil {
		klog.Infof("'%s/%s' Can't retrieve tenant tiers: %v", tenant.Namespace, tenant.Name, err)
	}
	if tInfos != nil {
		var tiersUsage []miniov2.TierUsage
		for _, tier := range tInfos {
			tiersUsage = append(tiersUsage, miniov2.TierUsage{
				Name:      tier.Name,
				Type:      tier.Type,
				TotalSize: int64(tier.Stats.TotalSize),
			})
		}
		tenant.Status.Usage.Tiers = tiersUsage
		if tenant, err = c.updatePoolStatus(context.Background(), tenant); err != nil {
			klog.Infof("'%s/%s' Can't update tenant status with tiers: %v", tenant.Namespace, tenant.Name, err)
		}
	}

	return tenant, nil
}

// HealthResult holds the results from cluster/health query into MinIO
type HealthResult struct {
	StatusCode        int
	HealingDrives     int
	WriteQuorumDrives int
}

// syncHealthCheckHandler acts on work items from the healthCheckQueue
func (c *Controller) syncHealthCheckHandler(key string) (Result, error) {
	// Convert the namespace/name string into a distinct namespace and name
	if key == "" {
		runtime.HandleError(fmt.Errorf("Invalid resource key: %s", key))
		return WrapResult(Result{}, nil)
	}

	namespace, tenantName := key2NamespaceName(key)

	// Get the Tenant resource with this namespace/name
	tenant, err := c.minioClientSet.MinioV2().Tenants(namespace).Get(context.Background(), tenantName, metav1.GetOptions{})
	if err != nil {
		// The Tenant resource may no longer exist, in which case we stop processing.
		if k8serrors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("Tenant '%s' in work queue no longer exists", key))
			return WrapResult(Result{}, nil)
		}
		return WrapResult(Result{}, err)
	}

	tenant.EnsureDefaults()

	tenant, err = c.updateHealthStatusForTenant(tenant)
	if err != nil {
		klog.Errorf("%v", err)
		return WrapResult(Result{}, err)
	}

	// Add tenant to the health check queue again until is green again
	if tenant != nil && tenant.Status.HealthStatus != miniov2.HealthStatusGreen {
		c.healthCheckQueue.Add(key)
	}

	return WrapResult(Result{}, nil)
}
