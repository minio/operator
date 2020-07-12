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
	"strconv"

	helpers "github.com/minio/kubectl-minio/cmd/helpers"
	miniov1 "github.com/minio/minio-operator/pkg/apis/operator.min.io/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

func tenantStorage(q resource.Quantity) v1.ResourceList {
	m := make(v1.ResourceList, 1)
	m[v1.ResourceStorage] = q
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

func tenantMCSConfig(tenant, secret string) *miniov1.MCSConfig {
	if secret != "" {
		return &miniov1.MCSConfig{
			Replicas: helpers.MCSReplicas,
			Image:    helpers.DefaultMCSImage,
			MCSSecret: &v1.LocalObjectReference{
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

// NewMinIOInstanceForTenant will return a new minioinstance for a MinIO Operator
func NewMinIOInstanceForTenant(args []string, ns, image, sc, kms, console, cert string, disableTLS bool) (*miniov1.MinIOInstance, error) {
	// validate args
	volumes, err := strconv.Atoi(args[4])
	if err != nil {
		return nil, err
	}
	z, err := helpers.ParseZones(args[3])
	if err != nil {
		return nil, err
	}
	q, err := resource.ParseQuantity(args[5])
	if err != nil {
		return nil, err
	}

	// create the MinIOInstance
	return &miniov1.MinIOInstance{
		Spec: miniov1.MinIOInstanceSpec{
			Metadata: &metav1.ObjectMeta{
				Name:        args[0],
				Labels:      tenantLabels(args[0]),
				Annotations: tenantAnnotations(),
			},
			Image:       image,
			ServiceName: helpers.ServiceName(args[0]),
			CredsSecret: &v1.LocalObjectReference{
				Name: helpers.SecretName(args[0]),
			},
			Zones: []miniov1.Zone{z},
			VolumeClaimTemplate: &v1.PersistentVolumeClaim{
				Spec: v1.PersistentVolumeClaimSpec{
					AccessModes: []v1.PersistentVolumeAccessMode{helpers.MinIOAccessMode},
					Resources: v1.ResourceRequirements{
						Requests: tenantStorage(q),
					},
					StorageClassName: storageClass(sc),
				},
			},
			VolumesPerServer: volumes,
			RequestAutoCert:  !disableTLS,
			CertConfig: &miniov1.CertificateConfig{
				CommonName:       "",
				OrganizationName: []string{},
				DNSNames:         []string{},
			},
			Mountpath:          helpers.MinIOMountPath,
			Liveness:           helpers.DefaultLivenessCheck,
			KES:                tenantKESConfig(args[0], kms),
			MCS:                tenantMCSConfig(args[0], console),
			ExternalCertSecret: externalCertSecret(cert),
		},
	}, nil
}
