// Copyright (C) 2022, MinIO, Inc.
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

package cluster

import (
	"context"
	"errors"
	"fmt"

	"github.com/minio/madmin-go"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/resources/statefulsets"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// updatePoolsFrom starts the process of updating each pool from the current version to the new version specified on the .spec.image field
func (c *Controller) updatePoolsFrom(ctx context.Context, tenant *miniov2.Tenant, adminClient *madmin.AdminClient, operatorWebhookSecret *corev1.Secret, fromImage string, totalReplicas int32) error {
	key := fmt.Sprintf("%s/%s", tenant.Namespace, tenant.Name)
	var err error

	c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "UpdateVersion", fmt.Sprintf("Updating MinIO Version to %s", tenant.Spec.Image))
	if !tenant.MinIOHealthCheck(c.getTransport()) {
		klog.Infof("%s is not running can't update image online", key)
		c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "UpdateFailed", "Tenant is not online, can't update it")
		return ErrMinIONotReady
	}

	// Images different with the newer state change, continue to verify
	// if upgrade is possible
	tenant, err = c.updateTenantStatus(ctx, tenant, StatusUpdatingMinIOVersion, totalReplicas)
	if err != nil {
		return err
	}

	klog.V(4).Infof("Collecting artifacts for Tenant '%s' to update MinIO from: %s, to: %s",
		tenant.Name, fromImage, tenant.Spec.Image)

	latest, err := c.fetchArtifacts(tenant)
	if err != nil {
		_ = c.removeArtifacts()
		return err
	}
	protocol := "https"
	if !isOperatorTLS() {
		protocol = "http"
	}
	updateURL, err := tenant.UpdateURL(latest, fmt.Sprintf("%s://operator.%s.svc.%s:%s%s",
		protocol,
		miniov2.GetNSFromFile(), miniov2.GetClusterDomain(),
		miniov2.WebhookDefaultPort, miniov2.WebhookAPIUpdate,
	))
	if err != nil {
		_ = c.removeArtifacts()

		err = fmt.Errorf("Unable to get canonical update URL for Tenant '%s', failed with %v", tenant.Name, err)
		if _, terr := c.updateTenantStatus(ctx, tenant, err.Error(), totalReplicas); terr != nil {
			return terr
		}

		// Correct URL could not be obtained, not proceeding to update.
		return err
	}

	klog.V(4).Infof("Updating Tenant %s MinIO version from: %s, to: %s -> URL: %s",
		tenant.Name, tenant.Spec.Image, fromImage, updateURL)

	us, err := adminClient.ServerUpdate(ctx, updateURL)
	if err != nil {
		_ = c.removeArtifacts()
		errorMessage := fmt.Sprintf("Tenant '%s' MinIO update failed with %s", tenant.Name, err)

		c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "UpdatedFailed", errorMessage)
		err = errors.New(errorMessage)
		if _, terr := c.updateTenantStatus(ctx, tenant, err.Error(), totalReplicas); terr != nil {
			return terr
		}

		// Update failed, nothing needs to be changed in the container
		return err
	}

	if us.CurrentVersion != us.UpdatedVersion {
		// In case the upgrade is from an older version to RELEASE.2021-07-27T02-40-15Z (which introduced
		// MinIO server integrated with Console), we need to delete the old console deployment and service.
		// We do this only when MinIO server is successfully updated.
		unifiedConsoleReleaseTime, _ := miniov2.ReleaseTagToReleaseTime("RELEASE.2021-07-27T02-40-15Z")
		newVer, err := miniov2.ReleaseTagToReleaseTime(us.UpdatedVersion)
		if err != nil {
			klog.Errorf("Unsupported release tag on new image, server updated but might leave dangling console deployment %v", err)
			return err
		}
		consoleDeployment, err := c.deploymentLister.Deployments(tenant.Namespace).Get(tenant.ConsoleDeploymentName())
		if unifiedConsoleReleaseTime.Before(newVer) && consoleDeployment != nil && err == nil {
			if err := c.deleteOldConsoleDeployment(ctx, tenant, consoleDeployment.Name); err != nil {
				return err
			}
		}
		klog.Infof("Tenant '%s' MinIO updated successfully from: %s, to: %s successfully",
			tenant.Name, us.CurrentVersion, us.UpdatedVersion)
	} else {
		msg := fmt.Sprintf("Tenant '%s' MinIO is already running the most recent version of %s",
			tenant.Name,
			us.CurrentVersion)
		klog.Info(msg)
		if _, terr := c.updateTenantStatus(ctx, tenant, msg, totalReplicas); terr != nil {
			return err
		}
		return nil
	}

	// clean the local directory
	_ = c.removeArtifacts()

	for i, pool := range tenant.Spec.Pools {
		// Now proceed to make the yaml changes for the tenant statefulset.
		ss := statefulsets.NewPool(tenant, operatorWebhookSecret, &pool, &tenant.Status.Pools[i], tenant.MinIOHLServiceName(), c.hostsTemplate, c.operatorVersion, isOperatorTLS())
		if _, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Update(ctx, ss, metav1.UpdateOptions{}); err != nil {
			return err
		}
		c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "PoolUpdated", fmt.Sprintf("Tenant pool %s updated", pool.Name))
	}
	return nil
}
