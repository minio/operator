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
	"io/ioutil"
	"log"
	"regexp"

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
	apiextensionv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

var fs = resources.GetStaticResources()

func tenantStorage(q resource.Quantity) corev1.ResourceList {
	m := make(corev1.ResourceList, 1)
	m[corev1.ResourceStorage] = q
	return m
}

// Pool returns a Pool object from given values
func Pool(opts *TenantOptions, volumes int32, q resource.Quantity) miniov2.Pool {
	p := miniov2.Pool{
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
				StorageClassName: storageClass(opts.StorageClass),
			},
		},
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

// GetSchemeDecoder returns a decoder for the scheme's that we use
func GetSchemeDecoder() func(data []byte, defaults *schema.GroupVersionKind, into runtime.Object) (runtime.Object, *schema.GroupVersionKind, error) {

	sch := runtime.NewScheme()
	scheme.AddToScheme(sch)
	apiextensionv1.AddToScheme(sch)
	apiextensionv1beta1.AddToScheme(sch)
	appsv1.AddToScheme(sch)
	rbacv1.AddToScheme(sch)
	corev1.AddToScheme(sch)
	decode := serializer.NewCodecFactory(sch).UniversalDeserializer().Decode
	return decode
}

func LoadTenantCRD(decode func(data []byte, defaults *schema.GroupVersionKind, into runtime.Object) (runtime.Object, *schema.GroupVersionKind, error)) *apiextensionv1.CustomResourceDefinition {
	contents, err := fs.Open("base/crds/minio.min.io_tenants.yaml")
	if err != nil {
		log.Fatal(err)
	}
	contentBytes, err := ioutil.ReadAll(contents)
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

func LoadClusterRole(decode func(data []byte, defaults *schema.GroupVersionKind, into runtime.Object) (runtime.Object, *schema.GroupVersionKind, error)) *rbacv1.ClusterRole {
	contents, err := fs.Open("cluster-role.yaml")
	if err != nil {
		log.Fatal(err)
	}
	contentBytes, err := ioutil.ReadAll(contents)
	if err != nil {
		log.Fatal(err)
	}

	obj, _, err := decode(contentBytes, nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	var ok bool
	resourceObj, ok := obj.(*rbacv1.ClusterRole)
	if !ok {
		log.Fatal("Unable to locate CustomResourceDefinition object")
	}
	return resourceObj
}

func LoadConsoleUI(decode func(data []byte, defaults *schema.GroupVersionKind, into runtime.Object) (runtime.Object, *schema.GroupVersionKind, error), opts *OperatorOptions) []runtime.Object {
	contents, err := fs.Open("console-ui.yaml")
	if err != nil {
		log.Fatal(err)
	}
	contentBytes, err := ioutil.ReadAll(contents)
	if err != nil {
		log.Fatal(err)
	}
	contentsString := string(contentBytes)

	regex := regexp.MustCompile("\n---")
	chunks := regex.Split(contentsString, -1)
	var consoleUIChunks []runtime.Object
	for _, chunk := range chunks {
		// ignore empty
		if chunk == "\n" || chunk == "" {
			continue
		}
		obj, _, err := decode([]byte(chunk), nil, nil)
		if err != nil {
			log.Fatal(err)
		}

		if opts != nil {
			switch obj.(type) {
			case *appsv1.Deployment:
				if resourceObj, ok := obj.(*appsv1.Deployment); ok {
					resourceObj.Namespace = opts.Namespace
					// console image
					if opts.ConsoleImage != "" {
						resourceObj.Spec.Template.Spec.Containers[0].Image = opts.ConsoleImage
					}
					// push down image pull secrets
					if opts.ImagePullSecret != "" {
						resourceObj.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{Name: opts.ImagePullSecret}}
					}
				}
			case *corev1.Service:
				if resourceObj, ok := obj.(*corev1.Service); ok {
					resourceObj.Namespace = opts.Namespace
				}
			case *corev1.ConfigMap:
				if resourceObj, ok := obj.(*corev1.ConfigMap); ok {
					resourceObj.Namespace = opts.Namespace
				}
			case *corev1.ServiceAccount:
				if resourceObj, ok := obj.(*corev1.ServiceAccount); ok {
					resourceObj.Namespace = opts.Namespace
				}
			case *rbacv1.ClusterRoleBinding:
				if resourceObj, ok := obj.(*rbacv1.ClusterRoleBinding); ok {
					updatedSubjects := []rbacv1.Subject{}
					for _, sub := range resourceObj.Subjects {
						sub.Namespace = opts.Namespace
						// store modified subject
						updatedSubjects = append(updatedSubjects, sub)
					}
					// update subjects with modified array
					resourceObj.Subjects = updatedSubjects
				}
			default:
				// fmt.Println("Unhandled kind:", obj.GetObjectKind())
			}
		}

		consoleUIChunks = append(consoleUIChunks, obj)
	}

	return consoleUIChunks
}
