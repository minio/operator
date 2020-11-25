/*
 * This file is part of MinIO Operator
 * Copyright (C) 2020, MinIO, Inc.
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

package resources

import (
	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func tenantSecretData() map[string][]byte {
	m := make(map[string][]byte, 2)
	m["accesskey"] = []byte(uuid.New().String())
	m["secretkey"] = []byte(uuid.New().String())
	return m
}

func consoleSecretData() map[string][]byte {
	m := make(map[string][]byte, 5)
	m["CONSOLE_ACCESS_KEY"] = []byte("admin")
	m["CONSOLE_SECRET_KEY"] = []byte(uuid.New().String())
	m["CONSOLE_PBKDF_PASSPHRASE"] = []byte(uuid.New().String())
	m["CONSOLE_PBKDF_SALT"] = []byte(uuid.New().String())
	return m
}

// NewSecretForTenant will return a new secret a MinIO Tenant
func NewSecretForTenant(opts *TenantOptions) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.SecretName,
			Namespace: opts.NS,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: v1.SchemeGroupVersion.Version,
		},
		Data: tenantSecretData(),
	}
}

// NewSecretForConsole will return a new secret a MinIO Tenant Console
func NewSecretForConsole(opts *TenantOptions) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.ConsoleSecret,
			Namespace: opts.NS,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: v1.SchemeGroupVersion.Version,
		},
		Data: consoleSecretData(),
	}
}
