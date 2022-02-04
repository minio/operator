// This file is part of MinIO Operator
// Copyright (C) 2020, MinIO, Inc.
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

package resources

import (
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewTenantConfigurationSecret will return a new secret a MinIO Tenant
func NewTenantConfigurationSecret(opts *TenantOptions) (*corev1.Secret, error) {
	accessKey, secretKey, err := miniov2.GenerateCredentials()
	if err != nil {
		return nil, err
	}
	tenantConfiguration := map[string]string{
		"MINIO_ROOT_USER":     accessKey,
		"MINIO_ROOT_PASSWORD": secretKey,
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.ConfigurationSecretName,
			Namespace: opts.NS,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: v1.SchemeGroupVersion.Version,
		},
		Data: map[string][]byte{
			"config.env": []byte(miniov2.GenerateTenantConfigurationFile(tenantConfiguration)),
		},
	}, nil
}

// NewUserCredentialsSecret will return a new secret a MinIO Tenant Console
func NewUserCredentialsSecret(opts *TenantOptions, secretName string) (*corev1.Secret, error) {
	accessKey, secretKey, err := miniov2.GenerateCredentials()
	if err != nil {
		return nil, err
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: opts.NS,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: v1.SchemeGroupVersion.Version,
		},
		Data: map[string][]byte{
			"CONSOLE_ACCESS_KEY": []byte(accessKey),
			"CONSOLE_SECRET_KEY": []byte(secretKey),
		},
	}, nil
}
