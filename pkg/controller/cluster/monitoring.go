/*
 * Copyright (C) 2021, MinIO, Inc.
 *
 * This code is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, version 3,
 * as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License, version 3,
 * along with this program.  If not, see <http://www.gnu.org/licenses/>
 *
 */

package cluster

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/minio/madmin-go"

	corev1 "k8s.io/api/core/v1"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/util/runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	// HealthAboutToLoseQuorumMessage means we are close to losing write capabilities
	HealthAboutToLoseQuorumMessage = "About to lose quorum"
)

// recurrentTenantStatusMonitor loop that checks every 3 minutes for tenants health
func (c *Controller) recurrentTenantStatusMonitor(stopCh <-chan struct{}) {
	// do an initial check, then start the periodic check
	if err := c.tenantsHealthMonitor(); err != nil {
		log.Println(err)
	}
	// How often will this function run
	interval := miniov2.GetMonitoringInterval()
	ticker := time.NewTicker(time.Duration(interval) * time.Minute)
	defer func() {
		log.Println("recurrent pod status monitor closed")
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
		if err = c.updateHealthStatusForTenant(&tenant); err != nil {
			klog.Errorf("%v", err)
			return err
		}
	}
	return nil
}

func (c *Controller) updateHealthStatusForTenant(tenant *miniov2.Tenant) error {
	// don't get the tenant cluster health if it doesn't have at least 1 pool initialized
	oneInitialized := false
	for _, pool := range tenant.Status.Pools {
		if pool.State == miniov2.PoolInitialized {
			oneInitialized = true
		}
	}
	if !oneInitialized {
		klog.Infof("'%s/%s' no pool is initialized", tenant.Namespace, tenant.Name)
		return nil
	}

	// get cluster health for tenant
	healthResult, err := getMinIOHealthStatus(tenant, RegularMode)
	if err != nil {
		// show the error and continue
		klog.Infof("'%s/%s' Failed to get cluster health: %v", tenant.Namespace, tenant.Name, err)
		return nil
	}
	tenantConfiguration, err := c.getTenantCredentials(context.Background(), tenant)
	if err != nil {
		return err
	}
	adminClnt, err := tenant.NewMinIOAdmin(tenantConfiguration)
	if err != nil {
		// show the error and continue
		klog.Infof("'%s/%s': %v", tenant.Namespace, tenant.Name, err)
		return nil
	}

	tenant.Status.DrivesHealing = int32(healthResult.HealingDrives)
	tenant.Status.WriteQuorum = int32(healthResult.WriteQuorumDrives)

	switch healthResult.StatusCode {

	case http.StatusServiceUnavailable:
		tenant.Status.HealthStatus = miniov2.HealthStatusRed
		tenant.Status.HealthMessage = HealthUnavailableMessage
	case http.StatusPreconditionFailed:
		tenant.Status.HealthStatus = miniov2.HealthStatusYellow
		// set message status to show number of drives being healed
		if healthResult.HealingDrives > 0 {
			tenant.Status.HealthMessage = fmt.Sprintf("%s %d Drives", HealthHealingMessage, healthResult.HealingDrives)
		} else {
			tenant.Status.HealthMessage = HealthAboutToLoseQuorumMessage
		}

	case http.StatusOK:
		tenant.Status.HealthStatus = miniov2.HealthStatusGreen
		tenant.Status.HealthMessage = ""
	default:
		tenant.Status.HealthStatus = miniov2.HealthStatusYellow
		tenant.Status.HealthMessage = ""
		log.Printf("tenant's health response code: %d not handled\n", healthResult.StatusCode)
	}

	// check all the tenant pods, if at least 1 is not running, we go yellow
	tenantPods, err := c.kubeClientSet.CoreV1().Pods(tenant.Namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", miniov2.TenantLabel, tenant.Name),
	})
	if err != nil {
		return err
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
		return nil
	}

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

	// get usage from Prometheus Metrics
	accessKey, ok := tenantConfiguration["accesskey"]
	if !ok {
		return errors.New("MinIO server accesskey not set")
	}

	secretKey, ok := tenantConfiguration["secretkey"]
	if !ok {
		return errors.New("MinIO server secretkey not set")
	}
	bearerToken := tenant.GenBearerToken(string(accessKey), string(secretKey))

	metrics, err := getPrometheusMetricsForTenant(tenant, bearerToken)
	if err != nil {
		klog.Infof("'%s/%s' Can't generate tenant prometheus token: %v", tenant.Namespace, tenant.Name, err)
	} else {
		if metrics != nil {
			tenant.Status.Usage.Usage = metrics.Usage
			tenant.Status.Usage.Capacity = metrics.UsableCapacity
			if tenant, err = c.updatePoolStatus(context.Background(), tenant); err != nil {
				klog.Infof("'%s/%s' Can't update tenant status for usage: %v", tenant.Namespace, tenant.Name, err)
			}
		}
	}

	return nil
}

// HealthResult holds the results from cluster/health query into MinIO
type HealthResult struct {
	StatusCode        int
	HealingDrives     int
	WriteQuorumDrives int
}

// HealthMode type of query we want to perform to MinIO cluster health
type HealthMode string

const (
	// MaintenanceMode query type for when we want to ask MinIO if we can take down 1 server
	MaintenanceMode HealthMode = "MaintenanceMode"
	// RegularMode query type for when we want to ask MinIO the current state of healing/health
	RegularMode = "RegularMode"
)

func getHealthCheckTransport() *http.Transport {
	// Keep TLS config.
	tlsConfig := &tls.Config{
		// Can't use SSLv3 because of POODLE and BEAST
		// Can't use TLSv1.0 because of POODLE and BEAST using CBC cipher
		// Can't use TLSv1.1 because of RC4 cipher usage
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true, // FIXME: use trusted CA
	}
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   2 * time.Second,
			KeepAlive: 10 * time.Second,
		}).DialContext,
		ResponseHeaderTimeout: 5 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 5 * time.Second,
		TLSClientConfig:       tlsConfig,
		// Go net/http automatically unzip if content-type is
		// gzip disable this feature, as we are always interested
		// in raw stream.
		DisableCompression: true,
	}
}

// getMinIOHealthStatus returns the cluster health for a Tenant.
// There's two types of questions we can make to MinIO's cluster/health one asking if the cluster is healthy `RegularMode`
// or if it's acceptable to remove a node `MaintenanceMode`
func getMinIOHealthStatus(tenant *miniov2.Tenant, mode HealthMode) (*HealthResult, error) {
	return getMinIOHealthStatusWithRetry(tenant, mode, 5)
}

// getMinIOHealthStatusWithRetry returns the cluster health for a Tenant.
// There's two types of questions we can make to MinIO's cluster/health one asking if the cluster is healthy `RegularMode`
// or if it's acceptable to remove a node `MaintenanceMode`
func getMinIOHealthStatusWithRetry(tenant *miniov2.Tenant, mode HealthMode, tryCount int) (*HealthResult, error) {
	// build the endpoint to contact the Tenant
	svcURL := tenant.GetTenantServiceURL()

	endpoint := fmt.Sprintf("%s%s", svcURL, "/minio/health/cluster")
	if mode == MaintenanceMode {
		endpoint = fmt.Sprintf("%s?maintenance=true", endpoint)
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		klog.Infof("error request pinging: %v", err)
		return nil, err
	}

	httpClient := &http.Client{
		Transport: getHealthCheckTransport(),
	}
	defer httpClient.CloseIdleConnections()

	resp, err := httpClient.Do(req)
	if err != nil {
		// if we fail due to timeout, retry
		if err, ok := err.(net.Error); ok && err.Timeout() && tryCount > 0 {
			klog.Infof("health check failed, retrying %d, err: %s", tryCount, err)
			time.Sleep(10 * time.Second)
			return getMinIOHealthStatusWithRetry(tenant, mode, tryCount-1)
		}
		klog.Infof("error pinging: %v", err)
		return nil, err
	}
	driveskHealing := 0
	if resp.Header.Get("X-Minio-Healing-Drives") != "" {
		val, err := strconv.Atoi(resp.Header.Get("X-Minio-Healing-Drives"))
		if err != nil {
			klog.Infof("Cannot parse healing drives from health check")
		} else {
			driveskHealing = val
		}
	}
	minDriveWrites := 0
	if resp.Header.Get("X-Minio-Write-Quorum") != "" {
		val, err := strconv.Atoi(resp.Header.Get("X-Minio-Write-Quorum"))
		if err != nil {
			klog.Infof("Cannot parse min write drives from health check")
		} else {
			minDriveWrites = val
		}
	}
	return &HealthResult{StatusCode: resp.StatusCode, HealingDrives: driveskHealing, WriteQuorumDrives: minDriveWrites}, nil
}

// processNextHealthCheckItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextHealthCheckItem() bool {
	obj, shutdown := c.healthCheckQueue.Get()
	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.healthCheckQueue.Done.
	processItem := func(obj interface{}) error {
		// We call Done here so the healthCheckQueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the healthCheckQueue and attempted again after a back-off
		// period.
		defer c.healthCheckQueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the healthCheckQueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// healthCheckQueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// healthCheckQueue.
		if key, ok = obj.(string); !ok {
			// As the item in the healthCheckQueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.healthCheckQueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in healthCheckQueue but got %#v", obj))
			return nil
		}
		klog.V(2).Infof("Key from healthCheckQueue: %s", key)
		// Run the syncHandler, passing it the namespace/name string of the
		// Tenant resource to be synced.
		if err := c.syncHealthCheckHandler(key); err != nil {
			return fmt.Errorf("error checking health check '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.healthCheckQueue.Forget(obj)
		klog.V(4).Infof("Successfully health checked '%s'", key)
		return nil
	}

	if err := processItem(obj); err != nil {
		runtime.HandleError(err)
		return true
	}
	return true
}

// syncHealthCheckHandler acts on work items from the healthCheckQueue
func (c *Controller) syncHealthCheckHandler(key string) error {

	// Convert the namespace/name string into a distinct namespace and name
	if key == "" {
		runtime.HandleError(fmt.Errorf("Invalid resource key: %s", key))
		return nil
	}

	namespace, tenantName := key2NamespaceName(key)

	// Get the Tenant resource with this namespace/name
	tenant, err := c.minioClientSet.MinioV2().Tenants(namespace).Get(context.Background(), tenantName, metav1.GetOptions{})
	if err != nil {
		// The Tenant resource may no longer exist, in which case we stop processing.
		if k8serrors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("Tenant '%s' in work queue no longer exists", key))
			return nil
		}
		return nil
	}

	tenant.EnsureDefaults()

	if err = c.updateHealthStatusForTenant(tenant); err != nil {
		klog.Errorf("%v", err)
		return err
	}

	return nil
}

// MinIOPrometheusMetrics holds metrics pulled from prometheus
type MinIOPrometheusMetrics struct {
	UsableCapacity int64
	Usage          int64
}

func getPrometheusMetricsForTenant(tenant *miniov2.Tenant, bearer string) (*MinIOPrometheusMetrics, error) {
	return getPrometheusMetricsForTenantWithRetry(tenant, bearer, 5)
}
func getPrometheusMetricsForTenantWithRetry(tenant *miniov2.Tenant, bearer string, tryCount int) (*MinIOPrometheusMetrics, error) {
	// build the endpoint to contact the Tenant
	svcURL := tenant.GetTenantServiceURL()

	endpoint := fmt.Sprintf("%s%s", svcURL, "/minio/v2/metrics/cluster")

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		klog.Infof("error request pinging: %v", err)
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bearer))

	httpClient := &http.Client{
		Transport: getHealthCheckTransport(),
	}
	defer httpClient.CloseIdleConnections()

	resp, err := httpClient.Do(req)
	if err != nil {
		// if we fail due to timeout, retry
		if err, ok := err.(net.Error); ok && err.Timeout() && tryCount > 0 {
			klog.Infof("health check failed, retrying %d, err: %s", tryCount, err)
			time.Sleep(10 * time.Second)
			return getPrometheusMetricsForTenantWithRetry(tenant, bearer, tryCount-1)
		}
		klog.Infof("error pinging: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	promMetrics := MinIOPrometheusMetrics{}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		// Usage
		if strings.HasPrefix(line, "minio_bucket_usage_total_bytes") {
			parts := strings.Split(line, " ")
			if len(parts) > 1 {
				usage, err := strconv.ParseFloat(parts[1], 64)
				if err != nil {
					klog.Infof("%s/%s Could not parse usage for tenant: %s", tenant.Namespace, tenant.Name, line)
				}
				promMetrics.Usage = int64(usage)
			}
		}
		// Usable capacity
		if strings.HasPrefix(line, "minio_cluster_capacity_usable_total_bytes") {
			parts := strings.Split(line, " ")
			if len(parts) > 1 {
				usable, err := strconv.ParseFloat(parts[1], 64)
				if err != nil {
					klog.Infof("%s/%s Could not parse usable capacity for tenant: %s", tenant.Namespace, tenant.Name, line)
				}
				promMetrics.UsableCapacity = int64(usable)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		klog.Error(err)
	}
	return &promMetrics, nil
}
