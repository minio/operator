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
	"errors"

	"github.com/minio/kubectl-minio/cmd/helpers"
	operator "github.com/minio/operator/pkg/apis/minio.min.io"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TenantOptions encapsulates the CLI options for a MinIO Tenant
type TenantOptions struct {
	Name                    string
	PoolName                string
	ConfigurationSecretName string
	Servers                 int32
	Volumes                 int32
	Capacity                string
	NS                      string
	Image                   string
	StorageClass            string
	KmsSecret               string
	ConsoleSecret           string
	DisableTLS              bool
	ImagePullSecret         string
	DisableAntiAffinity     bool
	ExposeMinioService      bool
	ExposeConsoleService    bool
	EnableSFTP              bool

	Interactive bool
}

// Validate Tenant Options
func (t TenantOptions) Validate() error {
	if t.Servers <= 0 {
		return errors.New("--servers is required. Specify a value greater than or equal to 1")
	}
	if t.Volumes <= 0 {
		return errors.New("--volumes is required. Specify a positive value")
	}
	if t.Capacity == "" {
		return errors.New("--capacity flag is required")
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

func tenantKESLabels(name string) map[string]string {
	m := make(map[string]string, 1)
	m["app"] = name + "-kes"
	return m
}

func tenantKESConfig(tenant, secret string) *miniov2.KESConfig {
	if secret != "" {
		return &miniov2.KESConfig{
			Replicas: helpers.KESReplicas,
			Image:    helpers.DefaultKESImage,
			Configuration: &v1.LocalObjectReference{
				Name: secret,
			},
			Labels: tenantKESLabels(tenant),
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
func NewTenant(opts *TenantOptions, userSecret *v1.Secret) (*miniov2.Tenant, error) {
	autoCert := !opts.DisableTLS
	volumesPerServer := helpers.VolumesPerServer(opts.Volumes, opts.Servers)
	capacityPerVolume, err := helpers.CapacityPerVolume(opts.Capacity, opts.Volumes)
	if err != nil {
		return nil, err
	}

	t := &miniov2.Tenant{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Tenant",
			APIVersion: operator.GroupName + "/" + miniov2.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.Name,
			Namespace: opts.NS,
		},
		Spec: miniov2.TenantSpec{
			Image: opts.Image,
			Configuration: &v1.LocalObjectReference{
				Name: opts.ConfigurationSecretName,
			},
			ExposeServices: &miniov2.ExposeServices{
				Console: opts.ExposeConsoleService,
				MinIO:   opts.ExposeMinioService,
			},
			Pools:           []miniov2.Pool{Pool(opts, volumesPerServer, *capacityPerVolume)},
			RequestAutoCert: &autoCert,
			Mountpath:       helpers.MinIOMountPath,
			KES:             tenantKESConfig(opts.Name, opts.KmsSecret),
			ImagePullSecret: v1.LocalObjectReference{Name: opts.ImagePullSecret},
			Users: []*v1.LocalObjectReference{
				{
					Name: userSecret.Name,
				},
			},
			Features: &miniov2.Features{
				EnableSFTP: &opts.EnableSFTP,
			},
		},
	}

	if autoCert {
		t.Spec.CertConfig = getAutoCertConfig(opts)
	}

	t.EnsureDefaults()

	return t, t.Validate()
}

func getAutoCertConfig(_ *TenantOptions) *miniov2.CertificateConfig {
	return &miniov2.CertificateConfig{
		CommonName:       "",
		OrganizationName: []string{},
		DNSNames:         []string{},
	}
}
