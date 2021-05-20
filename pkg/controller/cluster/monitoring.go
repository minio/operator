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
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"k8s.io/klog/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/labels"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
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
				log.Println(err)
			}
		case <-stopCh:
			ticker.Stop()
			return
		}
	}

}

func (c *Controller) tenantsHealthMonitor() error {
	// list all tenants and get their cluster health
	tenants, err := c.tenantsLister.Tenants("").List(labels.NewSelector())
	if err != nil {
		return err
	}
	for _, tenant := range tenants {
		// don't get the tenant cluster health if it doesn't have at least 1 pool initialized
		oneInitialized := false
		for _, pool := range tenant.Status.Pools {
			if pool.State == miniov2.PoolInitialized {
				oneInitialized = true
			}
		}
		if !oneInitialized {
			continue
		}

		// get cluster health for tenant
		healthResult, err := getMinIOHealthStatus(tenant, RegularMode)
		if err != nil {
			// show the error and continue
			klog.V(2).Infof(err.Error())
			continue
		}

		// get mc admin info
		minioSecretName := tenant.Spec.CredsSecret.Name
		minioSecret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(context.Background(), minioSecretName, metav1.GetOptions{})
		if err != nil {
			// show the error and continue
			klog.V(2).Infof(err.Error())
			continue
		}

		adminClnt, err := tenant.NewMinIOAdmin(minioSecret.Data)
		if err != nil {
			// show the error and continue
			klog.V(2).Infof(err.Error())
			continue
		}

		srvInfoCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		storageInfo, err := adminClnt.StorageInfo(srvInfoCtx)
		if err != nil {
			// show the error and continue
			klog.V(2).Infof(err.Error())
			continue
		}

		onlineDisks := 0
		offlineDisks := 0
		for _, d := range storageInfo.Disks {
			if d.State == "ok" {
				onlineDisks++
			} else {
				offlineDisks++
			}
		}

		tenant.Status.DrivesHealing = int32(healthResult.HealingDrives)
		tenant.Status.WriteQuorum = int32(healthResult.WriteQuorumDrives)

		tenant.Status.DrivesOnline = int32(onlineDisks)
		tenant.Status.DrivesOffline = int32(offlineDisks)

		tenant.Status.HealthStatus = miniov2.HealthStatusGreen

		if tenant.Status.DrivesOffline > 0 || tenant.Status.DrivesHealing > 0 {
			tenant.Status.HealthStatus = miniov2.HealthStatusYellow
		}
		if tenant.Status.DrivesOnline < tenant.Status.WriteQuorum {
			tenant.Status.HealthStatus = miniov2.HealthStatusRed
		}

		if _, err = c.updatePoolStatus(context.Background(), tenant); err != nil {
			klog.V(2).Infof(err.Error())
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
		log.Println("error request pinging", err)
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
			log.Printf("health check failed, retrying %d, err: %s", tryCount, err)
			time.Sleep(10 * time.Second)
			return getMinIOHealthStatusWithRetry(tenant, mode, tryCount-1)
		}
		log.Println("error pinging", err)
		return nil, err
	}
	driveskHealing := 0
	if resp.Header.Get("X-Minio-Healing-Drives") != "" {
		val, err := strconv.Atoi(resp.Header.Get("X-Minio-Healing-Drives"))
		if err != nil {
			log.Println("Cannot parse healing drives from health check")
		} else {
			driveskHealing = val
		}
	}
	minDriveWrites := 0
	if resp.Header.Get("X-Minio-Write-Quorum") != "" {
		val, err := strconv.Atoi(resp.Header.Get("X-Minio-Write-Quorum"))
		if err != nil {
			log.Println("Cannot parse min write drives from health check")
		} else {
			minDriveWrites = val
		}
	}
	return &HealthResult{StatusCode: resp.StatusCode, HealingDrives: driveskHealing, WriteQuorumDrives: minDriveWrites}, nil
}
