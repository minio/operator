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

	"github.com/dustin/go-humanize"
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
	EnableAuditLogs         bool
	AuditLogsDiskSpace      int32
	AuditLogsImage          string
	AuditLogsPGImage        string
	AuditLogsPGInitImage    string
	AuditLogsStorageClass   string
	EnablePrometheus        bool
	PrometheusDiskSpace     int
	PrometheusStorageClass  string
	PrometheusImage         string
	PrometheusSidecarImage  string
	PrometheusInitImage     string
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
	autoCert := true
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
			CredsSecret: &v1.LocalObjectReference{
				Name: opts.Name + "-creds-secret",
			},
			Pools:           []miniov2.Pool{Pool(opts, volumesPerServer, *capacityPerVolume)},
			RequestAutoCert: &autoCert,
			CertConfig: &miniov2.CertificateConfig{
				CommonName:       "",
				OrganizationName: []string{},
				DNSNames:         []string{},
			},
			Mountpath:       helpers.MinIOMountPath,
			KES:             tenantKESConfig(opts.Name, opts.KmsSecret),
			ImagePullSecret: v1.LocalObjectReference{Name: opts.ImagePullSecret},
			Users: []*v1.LocalObjectReference{
				{
					Name: userSecret.Name,
				},
			},
		},
	}
	if opts.EnableAuditLogs {
		t.Spec.Log = getAuditLogConfig(opts)
	}
	if opts.EnablePrometheus {
		t.Spec.Prometheus = getPrometheusConfig(opts)
	}

	t.EnsureDefaults()

	return t, t.Validate()
}

func getAuditLogConfig(opts *TenantOptions) *miniov2.LogConfig {
	diskSpace := int64(opts.AuditLogsDiskSpace) * humanize.GiByte
	var logSearchStorageClass *string
	if opts.AuditLogsStorageClass != "" {
		logSearchStorageClass = &opts.AuditLogsStorageClass
	}
	// the audit max cap cannot be larger than disk size on the DB, else it won't trim the data
	auditMaxCap := 10
	if (diskSpace / humanize.GiByte) < int64(auditMaxCap) {
		auditMaxCap = int(diskSpace / humanize.GiByte)
	}
	logConfig := createLogConfig(diskSpace, auditMaxCap, opts.Name, logSearchStorageClass)
	if opts.AuditLogsImage != "" {
		logConfig.Image = opts.AuditLogsImage
	}
	if opts.AuditLogsPGImage != "" {
		logConfig.Db.Image = opts.AuditLogsPGImage
	}
	if opts.AuditLogsPGInitImage != "" {
		logConfig.Db.InitImage = opts.AuditLogsPGInitImage
	}
	return logConfig
}

func createLogConfig(diskSpace int64, auditMaxCap int, name string, storage *string) *miniov2.LogConfig {
	return &miniov2.LogConfig{
		Audit: &miniov2.AuditConfig{DiskCapacityGB: &auditMaxCap},
		Db: &miniov2.LogDbConfig{
			VolumeClaimTemplate: &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: name + "-log",
				},
				Spec: v1.PersistentVolumeClaimSpec{
					AccessModes: []v1.PersistentVolumeAccessMode{
						v1.ReadWriteOnce,
					},
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceStorage: *resource.NewQuantity(diskSpace, resource.DecimalExponent),
						},
					},
					StorageClassName: storage,
				},
			},
		},
	}
}

func getPrometheusConfig(opts *TenantOptions) *miniov2.PrometheusConfig {
	var prometheusStorageClass *string
	if opts.PrometheusStorageClass != "" {
		prometheusStorageClass = &opts.PrometheusStorageClass
	}
	prometheusConfig := &miniov2.PrometheusConfig{
		DiskCapacityDB:   &opts.PrometheusDiskSpace,
		StorageClassName: prometheusStorageClass,
	}
	if opts.PrometheusImage != "" {
		prometheusConfig.Image = opts.PrometheusImage
	}
	if opts.PrometheusSidecarImage != "" {
		prometheusConfig.SideCarImage = opts.PrometheusSidecarImage
	}
	if opts.PrometheusInitImage != "" {
		prometheusConfig.InitImage = opts.PrometheusInitImage
	}
	return prometheusConfig
}
