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

package api

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type k8sClientMock struct{}

var (
	k8sClientGetResourceQuotaMock func(ctx context.Context, namespace, resource string, opts metav1.GetOptions) (*v1.ResourceQuota, error)
	k8sClientGetNameSpaceMock     func(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Namespace, error)
	k8sClientStorageClassesMock   func(ctx context.Context, opts metav1.ListOptions) (*storagev1.StorageClassList, error)

	k8sClientGetConfigMapMock    func(ctx context.Context, namespace, configMap string, opts metav1.GetOptions) (*corev1.ConfigMap, error)
	k8sClientCreateConfigMapMock func(ctx context.Context, namespace string, cm *corev1.ConfigMap, opts metav1.CreateOptions) (*corev1.ConfigMap, error)
	k8sClientUpdateConfigMapMock func(ctx context.Context, namespace string, cm *corev1.ConfigMap, opts metav1.UpdateOptions) (*corev1.ConfigMap, error)
	k8sClientDeleteConfigMapMock func(ctx context.Context, namespace string, name string, opts metav1.DeleteOptions) error

	k8sClientDeletePodCollectionMock     func(ctx context.Context, namespace string, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	k8sClientDeleteSecretMock            func(ctx context.Context, namespace string, name string, opts metav1.DeleteOptions) error
	k8sClientDeleteSecretsCollectionMock func(ctx context.Context, namespace string, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	k8sClientCreateSecretMock            func(ctx context.Context, namespace string, secret *v1.Secret, opts metav1.CreateOptions) (*v1.Secret, error)
	k8sClientUpdateSecretMock            func(ctx context.Context, namespace string, secret *v1.Secret, opts metav1.UpdateOptions) (*v1.Secret, error)
	k8sclientGetSecretMock               func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error)
	k8sclientGetServiceMock              func(ctx context.Context, namespace, serviceName string, opts metav1.GetOptions) (*corev1.Service, error)
)

func (c k8sClientMock) getResourceQuota(ctx context.Context, namespace, resource string, opts metav1.GetOptions) (*v1.ResourceQuota, error) {
	return k8sClientGetResourceQuotaMock(ctx, namespace, resource, opts)
}

func (c k8sClientMock) getNamespace(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Namespace, error) {
	return k8sClientGetNameSpaceMock(ctx, name, opts)
}

func (c k8sClientMock) getStorageClasses(ctx context.Context, opts metav1.ListOptions) (*storagev1.StorageClassList, error) {
	return k8sClientStorageClassesMock(ctx, opts)
}

func (c k8sClientMock) getConfigMap(ctx context.Context, namespace, configMap string, opts metav1.GetOptions) (*corev1.ConfigMap, error) {
	return k8sClientGetConfigMapMock(ctx, namespace, configMap, opts)
}

func (c k8sClientMock) createConfigMap(ctx context.Context, namespace string, cm *corev1.ConfigMap, opts metav1.CreateOptions) (*corev1.ConfigMap, error) {
	return k8sClientCreateConfigMapMock(ctx, namespace, cm, opts)
}

func (c k8sClientMock) updateConfigMap(ctx context.Context, namespace string, cm *corev1.ConfigMap, opts metav1.UpdateOptions) (*corev1.ConfigMap, error) {
	return k8sClientUpdateConfigMapMock(ctx, namespace, cm, opts)
}

func (c k8sClientMock) deleteConfigMap(ctx context.Context, namespace string, name string, opts metav1.DeleteOptions) error {
	return k8sClientDeleteConfigMapMock(ctx, namespace, name, opts)
}

func (c k8sClientMock) deletePodCollection(ctx context.Context, namespace string, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return k8sClientDeletePodCollectionMock(ctx, namespace, opts, listOpts)
}

func (c k8sClientMock) deleteSecret(ctx context.Context, namespace string, name string, opts metav1.DeleteOptions) error {
	return k8sClientDeleteSecretMock(ctx, namespace, name, opts)
}

func (c k8sClientMock) deleteSecretsCollection(ctx context.Context, namespace string, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return k8sClientDeleteSecretsCollectionMock(ctx, namespace, opts, listOpts)
}

func (c k8sClientMock) createSecret(ctx context.Context, namespace string, secret *v1.Secret, opts metav1.CreateOptions) (*v1.Secret, error) {
	return k8sClientCreateSecretMock(ctx, namespace, secret, opts)
}

func (c k8sClientMock) updateSecret(ctx context.Context, namespace string, secret *v1.Secret, opts metav1.UpdateOptions) (*v1.Secret, error) {
	return k8sClientUpdateSecretMock(ctx, namespace, secret, opts)
}

func (c k8sClientMock) getSecret(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
	return k8sclientGetSecretMock(ctx, namespace, secretName, opts)
}

func (c k8sClientMock) getService(ctx context.Context, namespace, serviceName string, opts metav1.GetOptions) (*corev1.Service, error) {
	return k8sclientGetServiceMock(ctx, namespace, serviceName, opts)
}
