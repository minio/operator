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
	"fmt"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) checkAndCreateServiceAccount(ctx context.Context, tenant *miniov2.Tenant) error {
	// check if service account exits
	sa, err := c.kubeClientSet.CoreV1().ServiceAccounts(tenant.Namespace).Get(ctx, tenant.Spec.ServiceAccountName, v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// create SA
			sa, err = c.kubeClientSet.CoreV1().ServiceAccounts(tenant.Namespace).Create(ctx, &corev1.ServiceAccount{
				ObjectMeta: v1.ObjectMeta{
					Name:      tenant.Spec.ServiceAccountName,
					Namespace: tenant.Namespace,
				},
			}, v1.CreateOptions{})
			if err != nil {
				return err
			}
			c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "SACreated", "Service Account Created")
		} else {
			c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "SAFailed", fmt.Sprintf("Service Account could not be created: %s", err.Error()))
			return err
		}
	}
	// check if role exist
	role, err := c.kubeClientSet.RbacV1().Roles(tenant.Namespace).Get(ctx, tenant.GetRoleName(), v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			role = getTenantRole(tenant)
			role, err = c.kubeClientSet.RbacV1().Roles(tenant.Namespace).Create(ctx, role, v1.CreateOptions{})
			if err != nil {
				return err
			}
			c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "RoleCreated", "Role Created")
		} else {
			c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "RoleFailed", fmt.Sprintf("Role could not be created: %s", err.Error()))
			return err
		}
	}
	// check rolebinding
	_, err = c.kubeClientSet.RbacV1().RoleBindings(tenant.Namespace).Get(ctx, tenant.GetBindingName(), v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			_, err = c.kubeClientSet.RbacV1().RoleBindings(tenant.Namespace).Create(ctx, getRoleBinding(tenant, sa, role), v1.CreateOptions{})
			if err != nil {
				return err
			}
			c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "BindingCreated", "Role Binding Created")
		} else {
			c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "BindingFailed", fmt.Sprintf("Role Binding could not be created: %s", err.Error()))
			return err
		}
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
		},
	}
	return &role
}
