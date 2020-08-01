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
	helpers "github.com/minio/kubectl-minio/cmd/helpers"
	miniov1 "github.com/minio/operator/pkg/apis/minio.min.io/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TenantOptions struct {
	Name          string
	SecretName    string
	Servers       int32
	Volumes       int32
	Capacity      string
	NS            string
	Image         string
	StorageClass  string
	KmsSecret     string
	ConsoleSecret string
	CertSecret    string
	DisableTLS    bool
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

func tenantMCSLabels(name string) map[string]string {
	m := make(map[string]string, 1)
	m["app"] = name + "-mcs"
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

func tenantMCSConfig(tenant, secret string) *miniov1.ConsoleConfiguration {
	if secret != "" {
		return &miniov1.ConsoleConfiguration{
			Replicas: helpers.MCSReplicas,
			Image:    helpers.DefaultMCSImage,
			ConsoleSecret: &v1.LocalObjectReference{
				Name: secret,
			},
			Metadata: &metav1.ObjectMeta{
				Labels: tenantMCSLabels(tenant),
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

// NewTenant will return a new minioinstance for a MinIO Operator
func NewTenant(opts *TenantOptions) (*miniov1.Tenant, error) {
	q, err := resource.ParseQuantity(opts.Capacity)
	if err != nil {
		return nil, err
	}
	// create the MinIOInstance
	t := &miniov1.Tenant{
		Spec: miniov1.TenantSpec{
			Metadata: &metav1.ObjectMeta{
				Name:        opts.Name,
				Labels:      tenantLabels(opts.Name),
				Annotations: tenantAnnotations(),
			},
			Image:       opts.Image,
			ServiceName: helpers.ServiceName(opts.Name),
			CredsSecret: &v1.LocalObjectReference{
				Name: opts.SecretName,
			},
			Zones:           []miniov1.Zone{Zone(opts.Servers, opts.Volumes, q, opts.StorageClass)},
			RequestAutoCert: true,
			CertConfig: &miniov1.CertificateConfig{
				CommonName:       "",
				OrganizationName: []string{},
				DNSNames:         []string{},
			},
			Mountpath:          helpers.MinIOMountPath,
			Liveness:           helpers.DefaultLivenessCheck,
			KES:                tenantKESConfig(opts.Name, opts.KmsSecret),
			Console:            tenantMCSConfig(opts.Name, opts.ConsoleSecret),
			ExternalCertSecret: externalCertSecret(opts.CertSecret),
		},
	}
	return t, t.Validate()
}
