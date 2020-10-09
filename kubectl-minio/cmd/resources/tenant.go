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
	"errors"

	helpers "github.com/minio/kubectl-minio/cmd/helpers"
	operator "github.com/minio/operator/pkg/apis/minio.min.io"
	miniov1 "github.com/minio/operator/pkg/apis/minio.min.io/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TenantOptions encapsulates the CLI options for a MinIO Tenant
type TenantOptions struct {
	Name            string
	SecretName      string
	Servers         int32
	Volumes         int32
	Capacity        string
	NS              string
	Image           string
	StorageClass    string
	KmsSecret       string
	ConsoleSecret   string
	CertSecret      string
	DisableTLS      bool
	ImagePullSecret string
}

// Validate Tenant Options
func (t TenantOptions) Validate() error {
	if t.Name == "" {
		return errors.New("--name flag is required for tenant creation")
	}
	if t.Servers == 0 {
		return errors.New("--servers flag is required for tenant creation")
	}
	if t.Servers < 0 {
		return errors.New("servers can not be negative")
	}
	if t.Volumes == 0 {
		return errors.New("--volumes flag is required for tenant creation")
	}
	if t.Volumes < 0 {
		return errors.New("volumes can not be negative")
	}
	if t.Capacity == "" {
		return errors.New("--capacity flag is required for tenant creation")
	}
	_, err := resource.ParseQuantity(t.Capacity)
	if err != nil {
		if err == resource.ErrFormatWrong {
			return errors.New("--capacity flag is incorrectly formatted. Please use suffix like 'T' or 'Ti' only")
		}
		return err
	}
	if t.Volumes%t.Servers != 0 {
		return errors.New("--volumes should be a multiple of --servers")
	}
	return nil
}

func tenantLabels(name string) map[string]string {
	m := make(map[string]string, 1)
	m["app"] = name
	return m
}

func tenantKESLabels(name string) map[string]string {
	m := make(map[string]string, 1)
	m["app"] = name + "-kes"
	return m
}

func tenantConsoleLabels(name string) map[string]string {
	m := make(map[string]string, 1)
	m["app"] = name + "-console"
	return m
}

func tenantAnnotations() map[string]string {
	m := make(map[string]string, 3)
	m["prometheus.io/path"] = helpers.MinIOPrometheusPath
	m["prometheus.io/port"] = helpers.MinIOPrometheusPort
	m["prometheus.io/scrape"] = "true"
	return m
}

func tenantKESConfig(tenant, secret string) *miniov1.KESConfig {
	if secret != "" {
		return &miniov1.KESConfig{
			Replicas: helpers.KESReplicas,
			Image:    helpers.DefaultKESImage,
			Configuration: &v1.LocalObjectReference{
				Name: secret,
			},
			Metadata: &metav1.ObjectMeta{
				Labels: tenantKESLabels(tenant),
			},
		}
	}
	return nil
}

func tenantConsoleConfig(tenant, secret string) *miniov1.ConsoleConfiguration {
	if secret != "" {
		return &miniov1.ConsoleConfiguration{
			Metadata: &metav1.ObjectMeta{
				Name: tenant,
			},
			Replicas: helpers.ConsoleReplicas,
			Image:    helpers.DefaultConsoleImage,
			ConsoleSecret: &v1.LocalObjectReference{
				Name: secret,
			},
		}
	}
	return nil
}

func externalCertSecret(secret string) *miniov1.LocalCertificateReference {
	if secret != "" {
		return &miniov1.LocalCertificateReference{
			Name: secret,
		}
	}
	return nil
}

func storageClass(sc string) *string {
	if sc != "" {
		return &sc
	}
	return nil
}

// NewTenant will return a new Tenant for a MinIO Operator
func NewTenant(opts *TenantOptions) (*miniov1.Tenant, error) {
	volumesPerServer := helpers.VolumesPerServer(opts.Volumes, opts.Servers)
	capacityPerVolume, err := helpers.CapacityPerVolume(opts.Capacity, opts.Volumes)
	if err != nil {
		return nil, err
	}

	t := &miniov1.Tenant{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Tenant",
			APIVersion: operator.GroupName + "/" + miniov1.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.Name,
			Namespace: opts.NS,
		},
		Spec: miniov1.TenantSpec{
			Image:       opts.Image,
			ServiceName: helpers.ServiceName(opts.Name),
			CredsSecret: &v1.LocalObjectReference{
				Name: opts.SecretName,
			},
			Zones:           []miniov1.Zone{Zone(opts.Servers, volumesPerServer, *capacityPerVolume, opts.StorageClass)},
			RequestAutoCert: true,
			CertConfig: &miniov1.CertificateConfig{
				CommonName:       "",
				OrganizationName: []string{},
				DNSNames:         []string{},
			},
			Mountpath:          helpers.MinIOMountPath,
			KES:                tenantKESConfig(opts.Name, opts.KmsSecret),
			Console:            tenantConsoleConfig(opts.Name, opts.ConsoleSecret),
			ExternalCertSecret: externalCertSecret(opts.CertSecret),
			ImagePullSecret:    v1.LocalObjectReference{Name: opts.ImagePullSecret},
		},
	}
	return t, t.Validate()
}
