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
	"fmt"
	"io"
	"io/fs"
	"log"
	"path"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/filesys"

	"github.com/minio/operator/resources"

	"k8s.io/apimachinery/pkg/runtime/schema"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	// Workaround for auth import issues refer https://github.com/minio/operator/issues/283
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/client-go/scale/scheme"

	"github.com/minio/kubectl-minio/cmd/helpers"
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var resourcesFS = resources.GetStaticResources()

func tenantStorage(q resource.Quantity) corev1.ResourceList {
	m := make(corev1.ResourceList, 1)
	m[corev1.ResourceStorage] = q
	return m
}

// Pool returns a Pool object from given values
func Pool(opts *TenantOptions, volumes int32, q resource.Quantity) miniov2.Pool {
	p := miniov2.Pool{
		Name:             opts.PoolName,
		Servers:          opts.Servers,
		VolumesPerServer: volumes,
		VolumeClaimTemplate: &corev1.PersistentVolumeClaim{
			TypeMeta: metav1.TypeMeta{
				Kind:       corev1.ResourcePersistentVolumeClaims.String(),
				APIVersion: corev1.SchemeGroupVersion.Version,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{helpers.MinIOAccessMode},
				Resources: corev1.ResourceRequirements{
					Requests: tenantStorage(q),
				},
			},
		},
	}
	// only pass the storage class if specified
	if opts.StorageClass != "" {
		p.VolumeClaimTemplate.Spec.StorageClassName = storageClass(opts.StorageClass)
	}
	if !opts.DisableAntiAffinity {
		p.Affinity = &corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{{
								Key:      miniov2.TenantLabel,
								Operator: "In",
								Values:   []string{opts.Name},
							}},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		}
	}
	return p
}

// GeneratePoolName Pool Name Generator
func GeneratePoolName(poolNumber int) string {
	return fmt.Sprintf("pool-%d", poolNumber)
}

// GetSchemeDecoder returns a decoder for the scheme's that we use
func GetSchemeDecoder() func(data []byte, defaults *schema.GroupVersionKind, into runtime.Object) (runtime.Object, *schema.GroupVersionKind, error) {
	sch := runtime.NewScheme()
	scheme.AddToScheme(sch)
	apiextensionv1.AddToScheme(sch)
	appsv1.AddToScheme(sch)
	rbacv1.AddToScheme(sch)
	corev1.AddToScheme(sch)
	decode := serializer.NewCodecFactory(sch).UniversalDeserializer().Decode
	return decode
}

// LoadTenantCRD loads tenant crds as k8s runtime object.
func LoadTenantCRD(decode func(data []byte, defaults *schema.GroupVersionKind, into runtime.Object) (runtime.Object, *schema.GroupVersionKind, error)) *apiextensionv1.CustomResourceDefinition {
	contents, err := resourcesFS.Open("base/crds/minio.min.io_tenants.yaml")
	if err != nil {
		log.Fatal(err)
	}
	contentBytes, err := io.ReadAll(contents)
	if err != nil {
		log.Fatal(err)
	}

	obj, _, err := decode(contentBytes, nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	var ok bool
	crdObj, ok := obj.(*apiextensionv1.CustomResourceDefinition)
	if !ok {
		log.Fatal("Unable to locate CustomResourceDefinition object")
	}
	return crdObj
}

// GetResourceFileSys file
func GetResourceFileSys() (filesys.FileSystem, error) {
	inMemSys := filesys.MakeFsInMemory()
	// copy from the resources into the target folder on the in memory FS
	if err := copyDirtoMemFS(".", "operator", inMemSys); err != nil {
		log.Println(err)
		return nil, err
	}
	return inMemSys, nil
}

func copyFileToMemFS(src, dst string, memFS filesys.FileSystem) error {
	// skip .go files
	if strings.HasSuffix(src, ".go") {
		return nil
	}
	var err error
	var srcFileDesc fs.File
	var dstFileDesc filesys.File

	if srcFileDesc, err = resourcesFS.Open(src); err != nil {
		return err
	}
	defer srcFileDesc.Close()

	if dstFileDesc, err = memFS.Create(dst); err != nil {
		return err
	}
	defer dstFileDesc.Close()

	// Note: I had to read the whole string, for some reason io.Copy was not copying the whole content
	input, err := io.ReadAll(srcFileDesc)
	if err != nil {
		return err
	}

	_, err = dstFileDesc.Write(input)
	return err
}

func copyDirtoMemFS(src string, dst string, memFS filesys.FileSystem) error {
	var err error
	var fds []fs.DirEntry

	if err = memFS.MkdirAll(dst); err != nil {
		return err
	}

	if fds, err = resourcesFS.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = copyDirtoMemFS(srcfp, dstfp, memFS); err != nil {
				return err
			}
		} else {
			if err = copyFileToMemFS(srcfp, dstfp, memFS); err != nil {
				return err
			}
		}
	}
	return nil
}
