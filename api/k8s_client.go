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

package api

import (
	"context"

	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// K8sClientI interface with all functions to be implemented
// by mock when testing, it should include all K8sClientI respective api calls
// that are used within this project.
type K8sClientI interface {
	getResourceQuota(ctx context.Context, namespace, resource string, opts metav1.GetOptions) (*v1.ResourceQuota, error)
	getNamespace(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Namespace, error)
	getStorageClasses(ctx context.Context, opts metav1.ListOptions) (*storagev1.StorageClassList, error)
	getSecret(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*v1.Secret, error)
	getService(ctx context.Context, namespace, serviceName string, opts metav1.GetOptions) (*v1.Service, error)
	deletePodCollection(ctx context.Context, namespace string, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	deleteSecret(ctx context.Context, namespace string, name string, opts metav1.DeleteOptions) error
	deleteSecretsCollection(ctx context.Context, namespace string, deleteOpts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	createSecret(ctx context.Context, namespace string, secret *v1.Secret, opts metav1.CreateOptions) (*v1.Secret, error)
	updateSecret(ctx context.Context, namespace string, secret *v1.Secret, opts metav1.UpdateOptions) (*v1.Secret, error)
	getPVC(ctx context.Context, namespace string, pvcName string, opts metav1.GetOptions) (*v1.PersistentVolumeClaim, error)
	getConfigMap(ctx context.Context, namespace, configMap string, opts metav1.GetOptions) (*v1.ConfigMap, error)
	createConfigMap(ctx context.Context, namespace string, cm *v1.ConfigMap, opts metav1.CreateOptions) (*v1.ConfigMap, error)
	updateConfigMap(ctx context.Context, namespace string, cm *v1.ConfigMap, opts metav1.UpdateOptions) (*v1.ConfigMap, error)
	deleteConfigMap(ctx context.Context, namespace string, name string, opts metav1.DeleteOptions) error
}

// Interface implementation
//
// Define the structure of a k8s client and define the functions that are actually used
type k8sClient struct {
	client *kubernetes.Clientset
}

func (c *k8sClient) getResourceQuota(ctx context.Context, namespace, resource string, opts metav1.GetOptions) (*v1.ResourceQuota, error) {
	return c.client.CoreV1().ResourceQuotas(namespace).Get(ctx, resource, opts)
}

func (c *k8sClient) getSecret(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*v1.Secret, error) {
	return c.client.CoreV1().Secrets(namespace).Get(ctx, secretName, opts)
}

func (c *k8sClient) getService(ctx context.Context, namespace, serviceName string, opts metav1.GetOptions) (*v1.Service, error) {
	return c.client.CoreV1().Services(namespace).Get(ctx, serviceName, opts)
}

func (c *k8sClient) deletePodCollection(ctx context.Context, namespace string, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return c.client.CoreV1().Pods(namespace).DeleteCollection(ctx, opts, listOpts)
}

func (c *k8sClient) deleteSecret(ctx context.Context, namespace string, name string, opts metav1.DeleteOptions) error {
	return c.client.CoreV1().Secrets(namespace).Delete(ctx, name, opts)
}

func (c *k8sClient) deleteSecretsCollection(ctx context.Context, namespace string, deleteOpts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return c.client.CoreV1().Secrets(namespace).DeleteCollection(ctx, deleteOpts, listOpts)
}

func (c *k8sClient) createSecret(ctx context.Context, namespace string, secret *v1.Secret, opts metav1.CreateOptions) (*v1.Secret, error) {
	return c.client.CoreV1().Secrets(namespace).Create(ctx, secret, opts)
}

func (c *k8sClient) updateSecret(ctx context.Context, namespace string, secret *v1.Secret, opts metav1.UpdateOptions) (*v1.Secret, error) {
	return c.client.CoreV1().Secrets(namespace).Update(ctx, secret, opts)
}

func (c *k8sClient) getNamespace(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Namespace, error) {
	return c.client.CoreV1().Namespaces().Get(ctx, name, opts)
}

func (c *k8sClient) getStorageClasses(ctx context.Context, opts metav1.ListOptions) (*storagev1.StorageClassList, error) {
	return c.client.StorageV1().StorageClasses().List(ctx, opts)
}

func (c *k8sClient) getPVC(ctx context.Context, namespace string, pvcName string, opts metav1.GetOptions) (*v1.PersistentVolumeClaim, error) {
	return c.client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvcName, opts)
}

func (c *k8sClient) getConfigMap(ctx context.Context, namespace, name string, opts metav1.GetOptions) (*v1.ConfigMap, error) {
	return c.client.CoreV1().ConfigMaps(namespace).Get(ctx, name, opts)
}

func (c *k8sClient) createConfigMap(ctx context.Context, namespace string, cm *v1.ConfigMap, opts metav1.CreateOptions) (*v1.ConfigMap, error) {
	return c.client.CoreV1().ConfigMaps(namespace).Create(ctx, cm, opts)
}

func (c *k8sClient) updateConfigMap(ctx context.Context, namespace string, cm *v1.ConfigMap, opts metav1.UpdateOptions) (*v1.ConfigMap, error) {
	return c.client.CoreV1().ConfigMaps(namespace).Update(ctx, cm, opts)
}

func (c *k8sClient) deleteConfigMap(ctx context.Context, namespace string, name string, opts metav1.DeleteOptions) error {
	return c.client.CoreV1().ConfigMaps(namespace).Delete(ctx, name, opts)
}
