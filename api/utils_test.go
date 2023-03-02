// This file is part of MinIO Operator
// Copyright (c) 2022 MinIO, Inc.
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
	"errors"
	"testing"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetTenantConfiguration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	kubernetesClient := k8sClientMock{}
	type args struct {
		ctx       context.Context
		clientSet K8sClientI
		tenant    *miniov2.Tenant
	}
	tests := []struct {
		name          string
		args          args
		want          map[string]string
		mockGetSecret func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error)
		wantErr       bool
	}{
		{
			name: "error because nil tenant",
			args: args{
				ctx:       ctx,
				clientSet: kubernetesClient,
				tenant:    nil,
			},
			mockGetSecret: func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
				return nil, nil
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "empty configuration map because no configuration secret is present",
			args: args{
				ctx:       ctx,
				clientSet: kubernetesClient,
				tenant:    &miniov2.Tenant{},
			},
			mockGetSecret: func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
				return nil, nil
			},
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name: "empty configuration map because error while retrieving configuration secret",
			args: args{
				ctx:       ctx,
				clientSet: kubernetesClient,
				tenant: &miniov2.Tenant{
					Spec: miniov2.TenantSpec{
						Configuration: &corev1.LocalObjectReference{
							Name: "tenant-configuration-secret",
						},
					},
				},
			},
			mockGetSecret: func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
				return nil, errors.New("an error has occurred")
			},
			want:    map[string]string{},
			wantErr: true,
		},
		{
			name: "parsing tenant configuration from secret file",
			args: args{
				ctx:       ctx,
				clientSet: kubernetesClient,
				tenant: &miniov2.Tenant{
					Spec: miniov2.TenantSpec{
						Configuration: &corev1.LocalObjectReference{
							Name: "tenant-configuration-secret",
						},
					},
				},
			},
			mockGetSecret: func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
				return &corev1.Secret{Data: map[string][]byte{
					"config.env": []byte(`
export MINIO_ROOT_USER=minio
export MINIO_ROOT_PASSWORD=minio123
export MINIO_CONSOLE_ADDRESS=:8080
export MINIO_IDENTITY_LDAP_SERVER_ADDR=localhost:389
export MINIO_IDENTITY_LDAP_LOOKUP_BIND_DN="cn=admin,dc=min,dc=io"
export MINIO_IDENTITY_LDAP_LOOKUP_BIND_PASSWORD="admin"
export MINIO_IDENTITY_LDAP_USER_DN_SEARCH_BASE_DN="dc=min,dc=io"
export MINIO_IDENTITY_LDAP_USER_DN_SEARCH_FILTER="(uid=%s)"
export MINIO_IDENTITY_LDAP_GROUP_SEARCH_BASE_DN="ou=swengg,dc=min,dc=io"
export MINIO_IDENTITY_LDAP_GROUP_SEARCH_FILTER="(&(objectclass=groupOfNames)(member=%d))"
export MINIO_IDENTITY_LDAP_SERVER_INSECURE="on"
`),
				}}, nil
			},
			want: map[string]string{
				"MINIO_ROOT_USER":                            "minio",
				"MINIO_ROOT_PASSWORD":                        "minio123",
				"MINIO_CONSOLE_ADDRESS":                      ":8080",
				"accesskey":                                  "minio",
				"secretkey":                                  "minio123",
				"MINIO_IDENTITY_LDAP_SERVER_INSECURE":        "on",
				"MINIO_IDENTITY_LDAP_GROUP_SEARCH_FILTER":    "(&(objectclass=groupOfNames)(member=%d))",
				"MINIO_IDENTITY_LDAP_GROUP_SEARCH_BASE_DN":   "ou=swengg,dc=min,dc=io",
				"MINIO_IDENTITY_LDAP_USER_DN_SEARCH_FILTER":  "(uid=%s)",
				"MINIO_IDENTITY_LDAP_USER_DN_SEARCH_BASE_DN": "dc=min,dc=io",
				"MINIO_IDENTITY_LDAP_LOOKUP_BIND_PASSWORD":   "admin",
				"MINIO_IDENTITY_LDAP_LOOKUP_BIND_DN":         "cn=admin,dc=min,dc=io",
				"MINIO_IDENTITY_LDAP_SERVER_ADDR":            "localhost:389",
			},
			wantErr: false,
		},
		{
			name: "parsing tenant configuration from secret file and environment variables",
			args: args{
				ctx:       ctx,
				clientSet: kubernetesClient,
				tenant: &miniov2.Tenant{
					Spec: miniov2.TenantSpec{
						Env: []corev1.EnvVar{
							{
								Name:  "MINIO_KMS_SECRET_KEY",
								Value: "my-minio-key:OSMM+vkKUTCvQs9YL/CVMIMt43HFhkUpqJxTmGl6rYw=",
							},
						},
						Configuration: &corev1.LocalObjectReference{
							Name: "tenant-configuration-secret",
						},
					},
				},
			},
			mockGetSecret: func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
				return &corev1.Secret{Data: map[string][]byte{
					"config.env": []byte(`
export MINIO_ROOT_USER=minio
export MINIO_ROOT_PASSWORD=minio123
export MINIO_CONSOLE_ADDRESS=:8080
export MINIO_IDENTITY_LDAP_SERVER_ADDR=localhost:389
export MINIO_IDENTITY_LDAP_LOOKUP_BIND_DN="cn=admin,dc=min,dc=io"
export MINIO_IDENTITY_LDAP_LOOKUP_BIND_PASSWORD="admin"
export MINIO_IDENTITY_LDAP_USER_DN_SEARCH_BASE_DN="dc=min,dc=io"
export MINIO_IDENTITY_LDAP_USER_DN_SEARCH_FILTER="(uid=%s)"
export MINIO_IDENTITY_LDAP_GROUP_SEARCH_BASE_DN="ou=swengg,dc=min,dc=io"
export MINIO_IDENTITY_LDAP_GROUP_SEARCH_FILTER="(&(objectclass=groupOfNames)(member=%d))"
export MINIO_IDENTITY_LDAP_SERVER_INSECURE="on"
`),
				}}, nil
			},
			want: map[string]string{
				"MINIO_ROOT_USER":                            "minio",
				"MINIO_ROOT_PASSWORD":                        "minio123",
				"MINIO_CONSOLE_ADDRESS":                      ":8080",
				"accesskey":                                  "minio",
				"secretkey":                                  "minio123",
				"MINIO_IDENTITY_LDAP_SERVER_INSECURE":        "on",
				"MINIO_IDENTITY_LDAP_GROUP_SEARCH_FILTER":    "(&(objectclass=groupOfNames)(member=%d))",
				"MINIO_IDENTITY_LDAP_GROUP_SEARCH_BASE_DN":   "ou=swengg,dc=min,dc=io",
				"MINIO_IDENTITY_LDAP_USER_DN_SEARCH_FILTER":  "(uid=%s)",
				"MINIO_IDENTITY_LDAP_USER_DN_SEARCH_BASE_DN": "dc=min,dc=io",
				"MINIO_IDENTITY_LDAP_LOOKUP_BIND_PASSWORD":   "admin",
				"MINIO_IDENTITY_LDAP_LOOKUP_BIND_DN":         "cn=admin,dc=min,dc=io",
				"MINIO_IDENTITY_LDAP_SERVER_ADDR":            "localhost:389",
				"MINIO_KMS_SECRET_KEY":                       "my-minio-key:OSMM+vkKUTCvQs9YL/CVMIMt43HFhkUpqJxTmGl6rYw=",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		k8sclientGetSecretMock = tt.mockGetSecret
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTenantConfiguration(tt.args.ctx, tt.args.clientSet, tt.args.tenant)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTenantConfiguration(%v, %v, %v)", tt.args.ctx, tt.args.clientSet, tt.args.tenant)
			}
			assert.Equalf(t, tt.want, got, "GetTenantConfiguration(%v, %v, %v)", tt.args.ctx, tt.args.clientSet, tt.args.tenant)
		})
	}
}

func TestGenerateTenantConfigurationFile(t *testing.T) {
	type args struct {
		configuration map[string]string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "convert configuration map into raw string",
			args: args{
				configuration: map[string]string{
					"MINIO_ROOT_USER": "minio",
				},
			},
			want: `export MINIO_ROOT_USER="minio"
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, GenerateTenantConfigurationFile(tt.args.configuration), "GenerateTenantConfigurationFile(%v)", tt.args.configuration)
		})
	}
}

func Test_stringPtr(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		wantNil bool
	}{
		{
			name: "get a pointer",
			args: args{
				str: "",
			},
			wantNil: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNilf(t, stringPtr(tt.args.str), "stringPtr(%v)", tt.args.str)
		})
	}
}
