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

	"github.com/blang/semver/v4"
	"github.com/hashicorp/go-version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
)

const (
	version500 = "v5.0.0"
	version600 = "v6.0.0"
	// currentVersion will point to the latest released update version
	currentVersion = version600
)

// Legacy const
const (
	WebhookSecret = "operator-webhook-secret"
)

type upgradeFunction func(ctx context.Context, tenant *miniov2.Tenant) (*miniov2.Tenant, error)

// checkForUpgrades verifies if the tenant definition needs any upgrades
func (c *Controller) checkForUpgrades(ctx context.Context, tenant *miniov2.Tenant) (*miniov2.Tenant, error) {
	var upgradesToDo []string
	upgrades := map[string]upgradeFunction{
		version600: c.upgrade600,
	}

	// if tenant has no version we mark it with latest version upgrade released
	if tenant.Status.SyncVersion == "" {
		tenant.Status.SyncVersion = currentVersion
		return c.updateTenantSyncVersion(ctx, tenant, version600)
	}

	// if the version is empty, upgrades might not been applied, we apply them all
	if tenant.Status.SyncVersion != "" {
		currentSyncVersion, err := version.NewVersion(tenant.Status.SyncVersion)
		if err != nil {
			return tenant, err
		}
		// when processing the version below 5.0.0, give a hint to manually upgrade
		if currentSyncVersion.LessThan(version.Must(version.NewVersion(version500))) {
			return tenant, fmt.Errorf("Tenant version %s is too old. Please upgrade to latest v5 operator first, before upgrading to the this operator version.", tenant.Status.SyncVersion)
		}
		// check which upgrades we need to do
		versionsThatNeedUpgrades := []string{
			version600,
		}
		for _, v := range versionsThatNeedUpgrades {
			vp, _ := version.NewVersion(v)
			if currentSyncVersion.LessThan(vp) {
				upgradesToDo = append(upgradesToDo, v)
			}
		}

	}

	for _, u := range upgradesToDo {
		klog.Infof("Upgrading %s", u)
		if tenant, err := upgrades[u](ctx, tenant); err != nil {
			klog.V(2).Infof("'%s/%s' Error upgrading tenant: %v", tenant.Namespace, tenant.Name, err.Error())
			return tenant, err
		}
	}

	return tenant, nil
}

// Method to compare two versions.
// Returns 1 if v2 is smaller, -1
// if v1 is smaller, 0 if equal
func versionCompare(version1 string, version2 string) int {
	vs1, err := semver.ParseTolerant(version1)
	if err != nil {
		klog.Errorf("Error parsing version %s: %v", version1, err)
		return -1
	}
	vs2, err := semver.ParseTolerant(version2)
	if err != nil {
		klog.Errorf("Error parsing version %s: %v", version2, err)
		return -1
	}
	return vs1.Compare(vs2)
}

// Upgrades the sync version to v6.0.0
// since we are adding `publishNotReadyAddresses` to the headless service, we need to restart all pods
func (c *Controller) upgrade600(ctx context.Context, tenant *miniov2.Tenant) (*miniov2.Tenant, error) {
	nsName := types.NamespacedName{Namespace: tenant.Namespace, Name: tenant.Name}
	// Check MinIO Headless Service used for internode communication
	err := c.checkMinIOHLSvc(ctx, tenant, nsName)
	if err != nil {
		klog.V(2).Infof("error consolidating headless service: %s", err.Error())
		return nil, err
	}
	// restart all pods for this tenant
	listOpts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", miniov2.TenantLabel, tenant.Name),
	}
	err = c.kubeClientSet.CoreV1().Pods(tenant.Namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, listOpts)
	if err != nil {
		klog.V(2).Infof("error deleting pods: %s", err.Error())
		return nil, err
	}
	return c.updateTenantSyncVersion(ctx, tenant, version600)
}
