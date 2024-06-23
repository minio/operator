// This file is part of MinIO Operator
// Copyright (c) 2023 MinIO, Inc.
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

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/runtime"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) checkAndCreateServiceAccount(ctx context.Context, tenant *miniov2.Tenant) error {
	// check if service account exits
	sa := &corev1.ServiceAccount{
		ObjectMeta: v1.ObjectMeta{
			Name:      tenant.Spec.ServiceAccountName,
			Namespace: tenant.Namespace,
		},
	}
	_, err := runtime.NewObjectSyncer(ctx, c.k8sClient, tenant, func() error {
		return nil
	}, sa, runtime.SyncTypeCreateOrUpdate).Sync(ctx)
	if err != nil {
		return err
	}

	// check if role exist
	role := getTenantRole(tenant)
	_, err = runtime.NewObjectSyncer(ctx, c.k8sClient, tenant, func() error {
		// set expected rules
		role.Rules = getTenantRole(tenant).Rules
		return nil
	}, role, runtime.SyncTypeCreateOrUpdate).Sync(ctx)
	if err != nil {
		return err
	}

	// check rolebinding
	roleBinding := getRoleBinding(tenant, sa, role)
	_, err = runtime.NewObjectSyncer(ctx, c.k8sClient, tenant, func() error {
		// set expected subjects and roleRef
		roleBinding.Subjects = getRoleBinding(tenant, sa, role).Subjects
		roleBinding.RoleRef = getRoleBinding(tenant, sa, role).RoleRef
		return nil
	}, roleBinding, runtime.SyncTypeCreateOrUpdate).Sync(ctx)
	if err != nil {
		return err
	}
	return nil
}

func getRoleBinding(tenant *miniov2.Tenant, sa *corev1.ServiceAccount, role *rbacv1.Role) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: v1.ObjectMeta{
			Name:      tenant.GetBindingName(),
			Namespace: tenant.Namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      sa.Name,
				Namespace: sa.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     role.Name,
		},
	}
}

func getTenantRole(tenant *miniov2.Tenant) *rbacv1.Role {
	role := rbacv1.Role{
		ObjectMeta: v1.ObjectMeta{
			Name:      tenant.GetRoleName(),
			Namespace: tenant.Namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"secrets",
				},
				Verbs: []string{
					"get",
					"list",
					"watch",
				},
			},
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"services",
				},
				Verbs: []string{
					"create",
					"delete",
					"get",
				},
			},
			{
				APIGroups: []string{
					"minio.min.io",
				},
				Resources: []string{
					"tenants",
				},
				Verbs: []string{
					"get",
					"list",
					"watch",
				},
			},
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"pods",
				},
				Verbs: []string{
					"patch",
				},
			},
		},
	}
	return &role
}
