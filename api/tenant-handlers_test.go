// This file is part of MinIO Operator
// Copyright (c) 2021 MinIO, Inc.
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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/go-openapi/swag"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	xhttp "github.com/minio/operator/pkg/http"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	opClientTenantDeleteMock func(ctx context.Context, namespace string, tenantName string, options metav1.DeleteOptions) error
	opClientTenantGetMock    func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error)
	opClientTenantPatchMock  func(ctx context.Context, namespace string, tenantName string, pt types.PatchType, data []byte, options metav1.PatchOptions) (*miniov2.Tenant, error)
	opClientTenantUpdateMock func(ctx context.Context, tenant *miniov2.Tenant, opts metav1.UpdateOptions) (*miniov2.Tenant, error)
)

var (
	opClientTenantListMock func(ctx context.Context, namespace string, opts metav1.ListOptions) (*miniov2.TenantList, error)
	httpClientGetMock      func(url string) (resp *http.Response, err error)
	httpClientPostMock     func(url, contentType string, body io.Reader) (resp *http.Response, err error)
	httpClientDoMock       func(req *http.Request) (*http.Response, error)
)

// mock function of TenantDelete()
func (ac opClientMock) TenantDelete(ctx context.Context, namespace string, tenantName string, options metav1.DeleteOptions) error {
	return opClientTenantDeleteMock(ctx, namespace, tenantName, options)
}

// mock function of TenantGet()
func (ac opClientMock) TenantGet(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
	return opClientTenantGetMock(ctx, namespace, tenantName, options)
}

// mock function of TenantPatch()
func (ac opClientMock) TenantPatch(ctx context.Context, namespace string, tenantName string, pt types.PatchType, data []byte, options metav1.PatchOptions) (*miniov2.Tenant, error) {
	return opClientTenantPatchMock(ctx, namespace, tenantName, pt, data, options)
}

// mock function of TenantUpdate()
func (ac opClientMock) TenantUpdate(ctx context.Context, tenant *miniov2.Tenant, opts metav1.UpdateOptions) (*miniov2.Tenant, error) {
	return opClientTenantUpdateMock(ctx, tenant, opts)
}

// mock function of TenantList()
func (ac opClientMock) TenantList(ctx context.Context, namespace string, opts metav1.ListOptions) (*miniov2.TenantList, error) {
	return opClientTenantListMock(ctx, namespace, opts)
}

// mock function of get()
func (h httpClientMock) Get(url string) (resp *http.Response, err error) {
	return httpClientGetMock(url)
}

// mock function of post()
func (h httpClientMock) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	return httpClientPostMock(url, contentType, body)
}

// mock function of Do()
func (h httpClientMock) Do(req *http.Request) (*http.Response, error) {
	return httpClientDoMock(req)
}

func Test_TenantInfoTenantAdminClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	kubernetesClient := k8sClientMock{}
	type args struct {
		ctx        context.Context
		client     K8sClientI
		tenant     miniov2.Tenant
		serviceURL string
	}
	tests := []struct {
		name           string
		args           args
		wantErr        bool
		mockGetSecret  func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error)
		mockGetService func(ctx context.Context, namespace, serviceName string, opts metav1.GetOptions) (*corev1.Service, error)
	}{
		{
			name: "Return Tenant Admin, no errors using legacy credentials",
			args: args{
				ctx:    ctx,
				client: kubernetesClient,
				tenant: miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tenant-1",
					},
					Spec: miniov2.TenantSpec{
						CredsSecret: &corev1.LocalObjectReference{
							Name: "secret-name",
						},
					},
				},
				serviceURL: "http://service-1.default.svc.cluster.local:80",
			},
			mockGetSecret: func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
				vals := make(map[string][]byte)
				vals["secretkey"] = []byte("secret")
				vals["accesskey"] = []byte("access")
				sec := &corev1.Secret{
					Data: vals,
				}
				return sec, nil
			},
			mockGetService: func(ctx context.Context, namespace, serviceName string, opts metav1.GetOptions) (*corev1.Service, error) {
				serv := &corev1.Service{
					Spec: corev1.ServiceSpec{
						ClusterIP: "10.1.1.2",
					},
				}
				return serv, nil
			},
			wantErr: false,
		},
		{
			name: "Return Tenant Admin, no errors using credentials from configuration file",
			args: args{
				ctx:    ctx,
				client: kubernetesClient,
				tenant: miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tenant-1",
					},
					Spec: miniov2.TenantSpec{
						CredsSecret: &corev1.LocalObjectReference{
							Name: "secret-name",
						},
						Configuration: &corev1.LocalObjectReference{
							Name: "tenant-configuration",
						},
					},
				},
				serviceURL: "http://service-1.default.svc.cluster.local:80",
			},
			mockGetSecret: func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
				vals := make(map[string][]byte)
				vals["config.env"] = []byte(`
export MINIO_ROOT_USER=minio
export MINIO_ROOT_PASSWORD=minio123
`)
				sec := &corev1.Secret{
					Data: vals,
				}
				return sec, nil
			},
			mockGetService: func(ctx context.Context, namespace, serviceName string, opts metav1.GetOptions) (*corev1.Service, error) {
				serv := &corev1.Service{
					Spec: corev1.ServiceSpec{
						ClusterIP: "10.1.1.2",
					},
				}
				return serv, nil
			},
			wantErr: false,
		},
		{
			name: "Return Tenant Admin, no errors using credentials from configuration file 2",
			args: args{
				ctx:    ctx,
				client: kubernetesClient,
				tenant: miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tenant-1",
					},
					Spec: miniov2.TenantSpec{
						CredsSecret: &corev1.LocalObjectReference{
							Name: "secret-name",
						},
						Configuration: &corev1.LocalObjectReference{
							Name: "tenant-configuration",
						},
					},
				},
				serviceURL: "http://service-1.default.svc.cluster.local:80",
			},
			mockGetSecret: func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
				vals := make(map[string][]byte)
				vals["config.env"] = []byte(`
export MINIO_ACCESS_KEY=minio
export MINIO_SECRET_KEY=minio123
`)
				sec := &corev1.Secret{
					Data: vals,
				}
				return sec, nil
			},
			mockGetService: func(ctx context.Context, namespace, serviceName string, opts metav1.GetOptions) (*corev1.Service, error) {
				serv := &corev1.Service{
					Spec: corev1.ServiceSpec{
						ClusterIP: "10.1.1.2",
					},
				}
				return serv, nil
			},
			wantErr: false,
		},
		{
			name: "Access key not stored on secrets",
			args: args{
				ctx:    ctx,
				client: kubernetesClient,
				tenant: miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tenant-1",
					},
				},
				serviceURL: "http://service-1.default.svc.cluster.local:80",
			},
			mockGetSecret: func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
				vals := make(map[string][]byte)
				vals["secretkey"] = []byte("secret")
				sec := &corev1.Secret{
					Data: vals,
				}
				return sec, nil
			},
			mockGetService: func(ctx context.Context, namespace, serviceName string, opts metav1.GetOptions) (*corev1.Service, error) {
				serv := &corev1.Service{
					Spec: corev1.ServiceSpec{
						ClusterIP: "10.1.1.2",
					},
				}
				return serv, nil
			},
			wantErr: true,
		},
		{
			name: "Secret key not stored on secrets",
			args: args{
				ctx:    ctx,
				client: kubernetesClient,
				tenant: miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tenant-1",
					},
				},
				serviceURL: "http://service-1.default.svc.cluster.local:80",
			},
			mockGetSecret: func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
				vals := make(map[string][]byte)
				vals["accesskey"] = []byte("access")
				sec := &corev1.Secret{
					Data: vals,
				}
				return sec, nil
			},
			mockGetService: func(ctx context.Context, namespace, serviceName string, opts metav1.GetOptions) (*corev1.Service, error) {
				serv := &corev1.Service{
					Spec: corev1.ServiceSpec{
						ClusterIP: "10.1.1.2",
					},
				}
				return serv, nil
			},
			wantErr: true,
		},
		{
			name: "Handle error on getService",
			args: args{
				ctx:    ctx,
				client: kubernetesClient,
				tenant: miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tenant-1",
					},
				},
				serviceURL: "http://service-1.default.svc.cluster.local:80",
			},
			mockGetSecret: func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
				vals := make(map[string][]byte)
				vals["accesskey"] = []byte("access")
				vals["secretkey"] = []byte("secret")
				sec := &corev1.Secret{
					Data: vals,
				}
				return sec, nil
			},
			mockGetService: func(ctx context.Context, namespace, serviceName string, opts metav1.GetOptions) (*corev1.Service, error) {
				return nil, errors.New("error")
			},
			wantErr: true,
		},
		{
			name: "Handle error on getSecret",
			args: args{
				ctx:    ctx,
				client: kubernetesClient,
				tenant: miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tenant-1",
					},
				},
				serviceURL: "http://service-1.default.svc.cluster.local:80",
			},
			mockGetSecret: func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
				return nil, errors.New("error")
			},
			mockGetService: func(ctx context.Context, namespace, serviceName string, opts metav1.GetOptions) (*corev1.Service, error) {
				serv := &corev1.Service{
					Spec: corev1.ServiceSpec{
						ClusterIP: "10.1.1.2",
					},
				}
				return serv, nil
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		k8sclientGetSecretMock = tt.mockGetSecret
		k8sclientGetServiceMock = tt.mockGetService
		t.Run(tt.name, func(t *testing.T) {
			got, err := getTenantAdminClient(tt.args.ctx, tt.args.client, &tt.args.tenant, tt.args.serviceURL)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("getTenantAdminClient() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got == nil {
				t.Errorf("getTenantAdminClient() expected type: *madmin.AdminClient, got: nil")
			}
		})
	}
}

func Test_TenantInfo(t *testing.T) {
	testTimeStamp := metav1.Now()
	type args struct {
		minioTenant *miniov2.Tenant
	}
	tests := []struct {
		name string
		args args
		want *models.Tenant
	}{
		{
			name: "Get tenant Info",
			args: args{
				minioTenant: &miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						CreationTimestamp: testTimeStamp,
						Name:              "tenant1",
						Namespace:         "minio-ns",
					},
					Spec: miniov2.TenantSpec{
						Pools: []miniov2.Pool{
							{
								Name:             "pool1",
								Servers:          int32(2),
								VolumesPerServer: 4,
								VolumeClaimTemplate: &corev1.PersistentVolumeClaim{
									Spec: corev1.PersistentVolumeClaimSpec{
										Resources: corev1.ResourceRequirements{
											Requests: map[corev1.ResourceName]resource.Quantity{
												corev1.ResourceStorage: resource.MustParse("1Mi"),
											},
										},
										StorageClassName: swag.String("standard"),
									},
								},
								RuntimeClassName: swag.String(""),
							},
						},

						Image: "minio/minio:RELEASE.2020-06-14T18-32-17Z",
					},
					Status: miniov2.TenantStatus{
						CurrentState: "ready",
					},
				},
			},
			want: &models.Tenant{
				CreationDate: testTimeStamp.Format(time.RFC3339),
				Name:         "tenant1",
				TotalSize:    int64(8388608),
				CurrentState: "ready",
				Pools: []*models.Pool{
					{
						Name: "pool1",
						SecurityContext: &models.SecurityContext{
							RunAsGroup:   nil,
							RunAsNonRoot: nil,
							RunAsUser:    nil,
						},
						Servers:          swag.Int64(int64(2)),
						VolumesPerServer: swag.Int32(4),
						VolumeConfiguration: &models.PoolVolumeConfiguration{
							StorageClassName: "standard",
							Size:             swag.Int64(1024 * 1024),
						},
					},
				},
				Namespace: "minio-ns",
				Image:     "minio/minio:RELEASE.2020-06-14T18-32-17Z",
			},
		},
		{
			// If console image is set, it should be returned on tenant info
			name: "Get tenant Info, Console image set",
			args: args{
				minioTenant: &miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						CreationTimestamp: testTimeStamp,
						Name:              "tenant1",
						Namespace:         "minio-ns",
						Annotations: map[string]string{
							prometheusPath:   "some/path",
							prometheusScrape: "other/path",
						},
					},
					Spec: miniov2.TenantSpec{
						Pools: []miniov2.Pool{},
						Image: "minio/minio:RELEASE.2020-06-14T18-32-17Z",
					},
					Status: miniov2.TenantStatus{
						CurrentState: "ready",
					},
				},
			},
			want: &models.Tenant{
				CreationDate: testTimeStamp.Format(time.RFC3339),
				Name:         "tenant1",
				CurrentState: "ready",
				Namespace:    "minio-ns",
				Image:        "minio/minio:RELEASE.2020-06-14T18-32-17Z",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getTenantInfo(tt.args.minioTenant)
			if !reflect.DeepEqual(got, tt.want) {
				ji, _ := json.Marshal(got)
				vi, _ := json.Marshal(tt.want)
				t.Errorf("got %s want %s", ji, vi)
			}
		})
	}
}

func Test_deleteTenantAction(t *testing.T) {
	opClient := opClientMock{}
	type args struct {
		ctx              context.Context
		operatorClient   OperatorClientI
		tenant           *miniov2.Tenant
		deletePvcs       bool
		objs             []runtime.Object
		mockTenantDelete func(ctx context.Context, namespace string, tenantName string, options metav1.DeleteOptions) error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Success",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				tenant: &miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "default",
						Namespace: "minio-tenant",
					},
				},
				deletePvcs: false,
				mockTenantDelete: func(ctx context.Context, namespace string, tenantName string, options metav1.DeleteOptions) error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			name: "Error",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				tenant: &miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "default",
						Namespace: "minio-tenant",
					},
				},
				deletePvcs: false,
				mockTenantDelete: func(ctx context.Context, namespace string, tenantName string, options metav1.DeleteOptions) error {
					return errors.New("something happened")
				},
			},
			wantErr: true,
		},
		{
			// Delete only PVCs of the defined tenant on the specific namespace
			name: "Delete PVCs on Tenant Deletion",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				tenant: &miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant1",
						Namespace: "minio-tenant",
					},
				},
				deletePvcs: true,
				objs: []runtime.Object{
					&corev1.PersistentVolumeClaim{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "PVC1",
							Namespace: "minio-tenant",
							Labels: map[string]string{
								miniov2.TenantLabel: "tenant1",
								miniov2.PoolLabel:   "pool-1",
							},
						},
					},
				},
				mockTenantDelete: func(ctx context.Context, namespace string, tenantName string, options metav1.DeleteOptions) error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			// Do not delete underlying pvcs
			name: "Don't Delete PVCs on Tenant Deletion",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				tenant: &miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant1",
						Namespace: "minio-tenant",
					},
				},
				deletePvcs: false,
				objs: []runtime.Object{
					&corev1.PersistentVolumeClaim{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "PVC1",
							Namespace: "minio-tenant",
							Labels: map[string]string{
								miniov2.TenantLabel: "tenant1",
								miniov2.PoolLabel:   "pool-1",
							},
						},
					},
				},
				mockTenantDelete: func(ctx context.Context, namespace string, tenantName string, options metav1.DeleteOptions) error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			// If error is different than NotFound, PVC deletion should not continue
			name: "Don't delete pvcs if error Deleting Tenant, return",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				tenant: &miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant1",
						Namespace: "minio-tenant",
					},
				},
				deletePvcs: true,
				objs: []runtime.Object{
					&corev1.PersistentVolumeClaim{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "PVC1",
							Namespace: "minio-tenant",
							Labels: map[string]string{
								miniov2.TenantLabel: "tenant1",
								miniov2.PoolLabel:   "pool-1",
							},
						},
					},
				},
				mockTenantDelete: func(ctx context.Context, namespace string, tenantName string, options metav1.DeleteOptions) error {
					return errors.New("error returned")
				},
			},
			wantErr: true,
		},
		{
			// If error is NotFound while trying to Delete Tenant, PVC deletion should continue
			name: "Delete pvcs if tenant not found",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				tenant: &miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant1",
						Namespace: "minio-tenant",
					},
				},
				deletePvcs: true,
				objs: []runtime.Object{
					&corev1.PersistentVolumeClaim{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "PVC1",
							Namespace: "minio-tenant",
							Labels: map[string]string{
								miniov2.TenantLabel: "tenant1",
								miniov2.PoolLabel:   "pool-1",
							},
						},
					},
				},
				mockTenantDelete: func(ctx context.Context, namespace string, tenantName string, options metav1.DeleteOptions) error {
					return k8sErrors.NewNotFound(schema.GroupResource{}, "tenant1")
				},
			},
			wantErr: false,
		},
		{
			// If error is NotFound while trying to Delete Tenant and pvcdeletion=false,
			// error should be returned
			name: "Don't delete pvcs and return error if tenant not found",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				tenant: &miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant1",
						Namespace: "minio-tenant",
					},
				},
				deletePvcs: false,
				objs: []runtime.Object{
					&corev1.PersistentVolumeClaim{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "PVC1",
							Namespace: "minio-tenant",
							Labels: map[string]string{
								miniov2.TenantLabel: "tenant1",
								miniov2.PoolLabel:   "pool-1",
							},
						},
					},
				},
				mockTenantDelete: func(ctx context.Context, namespace string, tenantName string, options metav1.DeleteOptions) error {
					return k8sErrors.NewNotFound(schema.GroupResource{}, "tenant1")
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		opClientTenantDeleteMock = tt.args.mockTenantDelete
		kubeClient := fake.NewSimpleClientset(tt.args.objs...)
		t.Run(tt.name, func(t *testing.T) {
			if err := deleteTenantAction(tt.args.ctx, tt.args.operatorClient, kubeClient.CoreV1(), tt.args.tenant, tt.args.deletePvcs); (err != nil) != tt.wantErr {
				t.Errorf("deleteTenantAction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_TenantAddPool(t *testing.T) {
	opClient := opClientMock{}

	type args struct {
		ctx             context.Context
		operatorClient  OperatorClientI
		nameSpace       string
		mockTenantPatch func(ctx context.Context, namespace string, tenantName string, pt types.PatchType, data []byte, options metav1.PatchOptions) (*miniov2.Tenant, error)
		mockTenantGet   func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error)
		params          operator_api.TenantAddPoolParams
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Add pool, no errors",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				nameSpace:      "default",
				mockTenantPatch: func(ctx context.Context, namespace string, tenantName string, pt types.PatchType, data []byte, options metav1.PatchOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				params: operator_api.TenantAddPoolParams{
					Body: &models.Pool{
						Name:    "pool-1",
						Servers: swag.Int64(int64(4)),
						VolumeConfiguration: &models.PoolVolumeConfiguration{
							Size:             swag.Int64(2147483648),
							StorageClassName: "standard",
						},
						VolumesPerServer: swag.Int32(4),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Add pool, error size",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				nameSpace:      "default",
				mockTenantPatch: func(ctx context.Context, namespace string, tenantName string, pt types.PatchType, data []byte, options metav1.PatchOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				params: operator_api.TenantAddPoolParams{
					Body: &models.Pool{
						Name:    "pool-1",
						Servers: swag.Int64(int64(4)),
						VolumeConfiguration: &models.PoolVolumeConfiguration{
							Size:             swag.Int64(0),
							StorageClassName: "standard",
						},
						VolumesPerServer: swag.Int32(4),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Add pool, error servers negative",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				nameSpace:      "default",
				mockTenantPatch: func(ctx context.Context, namespace string, tenantName string, pt types.PatchType, data []byte, options metav1.PatchOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				params: operator_api.TenantAddPoolParams{
					Body: &models.Pool{
						Name:    "pool-1",
						Servers: swag.Int64(int64(-1)),
						VolumeConfiguration: &models.PoolVolumeConfiguration{
							Size:             swag.Int64(2147483648),
							StorageClassName: "standard",
						},
						VolumesPerServer: swag.Int32(4),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Add pool, error volumes per server negative",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				nameSpace:      "default",
				mockTenantPatch: func(ctx context.Context, namespace string, tenantName string, pt types.PatchType, data []byte, options metav1.PatchOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				params: operator_api.TenantAddPoolParams{
					Body: &models.Pool{
						Name:    "pool-1",
						Servers: swag.Int64(int64(4)),
						VolumeConfiguration: &models.PoolVolumeConfiguration{
							Size:             swag.Int64(2147483648),
							StorageClassName: "standard",
						},
						VolumesPerServer: swag.Int32(-1),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Error on patch, handle error",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				nameSpace:      "default",
				mockTenantPatch: func(ctx context.Context, namespace string, tenantName string, pt types.PatchType, data []byte, options metav1.PatchOptions) (*miniov2.Tenant, error) {
					return nil, errors.New("errors")
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				params: operator_api.TenantAddPoolParams{
					Body: &models.Pool{
						Name:    "pool-1",
						Servers: swag.Int64(int64(4)),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Error on get, handle error",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				nameSpace:      "default",
				mockTenantPatch: func(ctx context.Context, namespace string, tenantName string, pt types.PatchType, data []byte, options metav1.PatchOptions) (*miniov2.Tenant, error) {
					return nil, errors.New("errors")
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return nil, errors.New("errors")
				},
				params: operator_api.TenantAddPoolParams{
					Body: &models.Pool{
						Name:    "pool-1",
						Servers: swag.Int64(int64(4)),
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		opClientTenantGetMock = tt.args.mockTenantGet
		opClientTenantPatchMock = tt.args.mockTenantPatch
		t.Run(tt.name, func(t *testing.T) {
			if err := addTenantPool(tt.args.ctx, tt.args.operatorClient, tt.args.params); (err != nil) != tt.wantErr {
				t.Errorf("addTenantPool() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_UpdateTenantAction(t *testing.T) {
	opClient := opClientMock{}
	httpClientM := httpClientMock{}

	type args struct {
		ctx               context.Context
		operatorClient    OperatorClientI
		httpCl            xhttp.ClientI
		nameSpace         string
		tenantName        string
		mockTenantPatch   func(ctx context.Context, namespace string, tenantName string, pt types.PatchType, data []byte, options metav1.PatchOptions) (*miniov2.Tenant, error)
		mockTenantGet     func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error)
		mockHTTPClientGet func(url string) (resp *http.Response, err error)
		params            operator_api.UpdateTenantParams
	}
	tests := []struct {
		name    string
		args    args
		objs    []runtime.Object
		wantErr bool
	}{
		{
			name: "Update minio version no errors",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				httpCl:         httpClientM,
				nameSpace:      "default",
				tenantName:     "myminio",
				mockTenantPatch: func(ctx context.Context, namespace string, tenantName string, pt types.PatchType, data []byte, options metav1.PatchOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				mockHTTPClientGet: func(url string) (resp *http.Response, err error) {
					return &http.Response{}, nil
				},
				params: operator_api.UpdateTenantParams{
					Body: &models.UpdateTenantRequest{
						Image: "minio/minio:RELEASE.2023-01-06T18-11-18Z",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Error occurs getting minioTenant",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				httpCl:         httpClientM,
				nameSpace:      "default",
				tenantName:     "myminio",
				mockTenantPatch: func(ctx context.Context, namespace string, tenantName string, pt types.PatchType, data []byte, options metav1.PatchOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return nil, errors.New("error-get")
				},
				mockHTTPClientGet: func(url string) (resp *http.Response, err error) {
					return &http.Response{}, nil
				},
				params: operator_api.UpdateTenantParams{
					Body: &models.UpdateTenantRequest{
						Image: "minio/minio:RELEASE.2023-01-06T18-11-18Z",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Error occurs patching minioTenant",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				httpCl:         httpClientM,
				nameSpace:      "default",
				tenantName:     "myminio",
				mockTenantPatch: func(ctx context.Context, namespace string, tenantName string, pt types.PatchType, data []byte, options metav1.PatchOptions) (*miniov2.Tenant, error) {
					return nil, errors.New("error-get")
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				mockHTTPClientGet: func(url string) (resp *http.Response, err error) {
					return &http.Response{}, nil
				},
				params: operator_api.UpdateTenantParams{
					Tenant: "myminio",
					Body: &models.UpdateTenantRequest{
						Image: "minio/minio:RELEASE.2023-01-06T18-11-18Z",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Empty image should patch correctly with latest image",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				httpCl:         httpClientM,
				nameSpace:      "default",
				tenantName:     "myminio",
				mockTenantPatch: func(ctx context.Context, namespace string, tenantName string, pt types.PatchType, data []byte, options metav1.PatchOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				mockHTTPClientGet: func(url string) (resp *http.Response, err error) {
					r := ioutil.NopCloser(bytes.NewReader([]byte(`./minio.RELEASE.2020-06-18T02-23-35Z"`)))
					return &http.Response{
						Body: r,
					}, nil
				},
				params: operator_api.UpdateTenantParams{
					Tenant: "myminio",
					Body: &models.UpdateTenantRequest{
						Image: "",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Empty image input Error retrieving latest image, nothing happens",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				httpCl:         httpClientM,
				nameSpace:      "default",
				tenantName:     "myminio",
				mockTenantPatch: func(ctx context.Context, namespace string, tenantName string, pt types.PatchType, data []byte, options metav1.PatchOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				mockHTTPClientGet: func(url string) (resp *http.Response, err error) {
					return nil, errors.New("error")
				},
				params: operator_api.UpdateTenantParams{
					Tenant: "myminio",
					Body: &models.UpdateTenantRequest{
						Image: "",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Update minio image pull secrets no errors",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				httpCl:         httpClientM,
				nameSpace:      "default",
				tenantName:     "myminio",
				mockTenantPatch: func(ctx context.Context, namespace string, tenantName string, pt types.PatchType, data []byte, options metav1.PatchOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				mockHTTPClientGet: func(url string) (resp *http.Response, err error) {
					return nil, errors.New("use default minio")
				},
				params: operator_api.UpdateTenantParams{
					Body: &models.UpdateTenantRequest{
						ImagePullSecret: "minio-regcred",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		opClientTenantGetMock = tt.args.mockTenantGet
		opClientTenantPatchMock = tt.args.mockTenantPatch
		httpClientGetMock = tt.args.mockHTTPClientGet
		cnsClient := k8sClientMock{}
		t.Run(tt.name, func(t *testing.T) {
			if err := updateTenantAction(tt.args.ctx, tt.args.operatorClient, cnsClient, tt.args.httpCl, tt.args.nameSpace, tt.args.params); (err != nil) != tt.wantErr {
				t.Errorf("updateTenantAction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_UpdateDomainsResponse(t *testing.T) {
	opClient := opClientMock{}

	type args struct {
		ctx              context.Context
		operatorClient   OperatorClientI
		nameSpace        string
		tenantName       string
		mockTenantUpdate func(ctx context.Context, tenant *miniov2.Tenant, options metav1.UpdateOptions) (*miniov2.Tenant, error)
		mockTenantGet    func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error)
		domains          *models.DomainsConfiguration
	}
	tests := []struct {
		name    string
		args    args
		objs    []runtime.Object
		wantErr bool
	}{
		{
			name: "Update console & minio domains OK",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				nameSpace:      "default",
				tenantName:     "myminio",
				mockTenantUpdate: func(ctx context.Context, tenant *miniov2.Tenant, options metav1.UpdateOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				domains: &models.DomainsConfiguration{
					Console: "http://console.min.io",
					Minio:   []string{"http://domain1.min.io", "http://domain2.min.io", "http://domain3.min.io"},
				},
			},
			wantErr: false,
		},
		{
			name: "Error occurs getting minioTenant",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				nameSpace:      "default",
				tenantName:     "myminio",
				mockTenantUpdate: func(ctx context.Context, tenant *miniov2.Tenant, options metav1.UpdateOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return nil, errors.New("error-getting-tenant-info")
				},
				domains: &models.DomainsConfiguration{
					Console: "http://console.min.io",
					Minio:   []string{"http://domain1.min.io", "http://domain2.min.io", "http://domain3.min.io"},
				},
			},
			wantErr: true,
		},
		{
			name: "Tenant already has domains",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				nameSpace:      "default",
				tenantName:     "myminio",
				mockTenantUpdate: func(ctx context.Context, tenant *miniov2.Tenant, options metav1.UpdateOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					domains := miniov2.TenantDomains{
						Console: "http://onerandomdomain.min.io",
						Minio: []string{
							"http://oneDomain.min.io",
							"http://twoDomains.min.io",
						},
					}

					features := miniov2.Features{
						Domains: &domains,
					}

					return &miniov2.Tenant{
						Spec: miniov2.TenantSpec{Features: &features},
					}, nil
				},
				domains: &models.DomainsConfiguration{
					Console: "http://console.min.io",
					Minio:   []string{"http://domain1.min.io", "http://domain2.min.io", "http://domain3.min.io"},
				},
			},
			wantErr: false,
		},
		{
			name: "Tenant features only have BucketDNS",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				nameSpace:      "default",
				tenantName:     "myminio",
				mockTenantUpdate: func(ctx context.Context, tenant *miniov2.Tenant, options metav1.UpdateOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{}, nil
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					features := miniov2.Features{
						BucketDNS: true,
					}

					return &miniov2.Tenant{
						Spec: miniov2.TenantSpec{Features: &features},
					}, nil
				},
				domains: &models.DomainsConfiguration{
					Console: "http://console.min.io",
					Minio:   []string{"http://domain1.min.io", "http://domain2.min.io", "http://domain3.min.io"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		opClientTenantGetMock = tt.args.mockTenantGet
		opClientTenantUpdateMock = tt.args.mockTenantUpdate
		t.Run(tt.name, func(t *testing.T) {
			if err := updateTenantDomains(tt.args.ctx, tt.args.operatorClient, tt.args.nameSpace, tt.args.tenantName, tt.args.domains); (err != nil) != tt.wantErr {
				t.Errorf("updateTenantDomains() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_parseTenantCertificates(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	kubernetesClient := k8sClientMock{}
	type args struct {
		ctx       context.Context
		clientSet K8sClientI
		namespace string
		secrets   []*miniov2.LocalCertificateReference
	}
	tests := []struct {
		name          string
		args          args
		want          []*models.CertificateInfo
		mockGetSecret func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error)
		wantErr       bool
	}{
		{
			name: "empty secrets list",
			args: args{
				ctx:       ctx,
				clientSet: kubernetesClient,
				secrets:   nil,
			},
			mockGetSecret: func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
				return nil, nil
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "error getting secret",
			args: args{
				ctx:       ctx,
				clientSet: kubernetesClient,
				secrets: []*miniov2.LocalCertificateReference{
					{
						Name: "certificate-1",
					},
				},
			},
			mockGetSecret: func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
				return nil, errors.New("error getting secret")
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error getting certificate because of missing public key",
			args: args{
				ctx:       ctx,
				clientSet: kubernetesClient,
				secrets: []*miniov2.LocalCertificateReference{
					{
						Name: "certificate-1",
						Type: "Opaque",
					},
				},
			},
			mockGetSecret: func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
				certificateSecret := &corev1.Secret{
					Data: map[string][]byte{
						"eaeaeae": []byte(`
-----BEGIN CERTIFICATE-----
MIIBUDCCAQKgAwIBAgIRALdFZh8hLU348ho9wYzlZbAwBQYDK2VwMBIxEDAOBgNV
BAoTB0FjbWUgQ28wHhcNMjIwODE4MjAxMzUzWhcNMjMwODE4MjAxMzUzWjASMRAw
DgYDVQQKEwdBY21lIENvMCowBQYDK2VwAyEAct5c3dzzbNOTi+C62w7QHoSivEWD
MYAheDXZWHC55tGjbTBrMA4GA1UdDwEB/wQEAwIChDATBgNVHSUEDDAKBggrBgEF
BQcDATAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBTs0At8sTSLCjiM24AZhxFY
a2CswjAUBgNVHREEDTALgglsb2NhbGhvc3QwBQYDK2VwA0EABan+d16CeN8UD+QF
a8HBhPAiOpaZeEF6+EqTlq9VfL3eSVd7CLRI+/KtY7ptwomuTeYzuV73adKdE9N2
ZrJuAw==
-----END CERTIFICATE-----
						`),
					},
					Type: "Opaque",
				}
				return certificateSecret, nil
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "return certificate from existing secret",
			args: args{
				ctx:       ctx,
				clientSet: kubernetesClient,
				secrets: []*miniov2.LocalCertificateReference{
					{
						Name: "certificate-1",
						Type: "Opaque",
					},
				},
			},
			mockGetSecret: func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
				certificateSecret := &corev1.Secret{
					Data: map[string][]byte{
						"public.crt": []byte(`
-----BEGIN CERTIFICATE-----
MIIBUDCCAQKgAwIBAgIRALdFZh8hLU348ho9wYzlZbAwBQYDK2VwMBIxEDAOBgNV
BAoTB0FjbWUgQ28wHhcNMjIwODE4MjAxMzUzWhcNMjMwODE4MjAxMzUzWjASMRAw
DgYDVQQKEwdBY21lIENvMCowBQYDK2VwAyEAct5c3dzzbNOTi+C62w7QHoSivEWD
MYAheDXZWHC55tGjbTBrMA4GA1UdDwEB/wQEAwIChDATBgNVHSUEDDAKBggrBgEF
BQcDATAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBTs0At8sTSLCjiM24AZhxFY
a2CswjAUBgNVHREEDTALgglsb2NhbGhvc3QwBQYDK2VwA0EABan+d16CeN8UD+QF
a8HBhPAiOpaZeEF6+EqTlq9VfL3eSVd7CLRI+/KtY7ptwomuTeYzuV73adKdE9N2
ZrJuAw==
-----END CERTIFICATE-----
						`),
					},
					Type: "Opaque",
				}
				return certificateSecret, nil
			},
			want: []*models.CertificateInfo{
				{
					SerialNumber: "243609062983998893460787085129017550256",
					Name:         "certificate-1",
					Expiry:       "2023-08-18T20:13:53Z",
					Domains:      []string{"localhost"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		k8sclientGetSecretMock = tt.mockGetSecret
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTenantCertificates(tt.args.ctx, tt.args.clientSet, tt.args.namespace, tt.args.secrets)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTenantCertificates(%v, %v, %v, %v)error = %v, wantErr %v", tt.args.ctx, tt.args.clientSet, tt.args.namespace, tt.args.secrets, err, tt.wantErr)
			}
			assert.Equalf(t, tt.want, got, "parseTenantCertificates(%v, %v, %v, %v)", tt.args.ctx, tt.args.clientSet, tt.args.namespace, tt.args.secrets)
		})
	}
}

func Test_getTenant(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	opClient := opClientMock{}
	type args struct {
		ctx            context.Context
		operatorClient OperatorClientI
		namespace      string
		tenantName     string
		mockTenantGet  func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error)
	}
	tests := []struct {
		name    string
		args    args
		want    *miniov2.Tenant
		wantErr bool
	}{
		{
			name: "error getting tenant information",
			args: args{
				ctx:            ctx,
				operatorClient: opClient,
				namespace:      "default",
				tenantName:     "test",
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return nil, errors.New("error getting tenant information")
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "success getting tenant information",
			args: args{
				ctx:            ctx,
				operatorClient: opClient,
				namespace:      "default",
				tenantName:     "test",
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "default",
						},
					}, nil
				},
			},
			want: &miniov2.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		opClientTenantGetMock = tt.args.mockTenantGet
		t.Run(tt.name, func(t *testing.T) {
			got, err := getTenant(tt.args.ctx, tt.args.operatorClient, tt.args.namespace, tt.args.tenantName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getTenant(%v, %v, %v, %v)", tt.args.ctx, tt.args.operatorClient, tt.args.namespace, tt.args.tenantName)
			}
			assert.Equalf(t, tt.want, got, "getTenant(%v, %v, %v, %v)", tt.args.ctx, tt.args.operatorClient, tt.args.namespace, tt.args.tenantName)
		})
	}
}

func Test_updateTenantConfigurationFile(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	opClient := opClientMock{}
	kubernetesClient := k8sClientMock{}
	type args struct {
		ctx                     context.Context
		operatorClient          OperatorClientI
		client                  K8sClientI
		namespace               string
		params                  operator_api.UpdateTenantConfigurationParams
		mockTenantGet           func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error)
		mockGetSecret           func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error)
		mockUpdateSecret        func(ctx context.Context, namespace string, secret *corev1.Secret, opts metav1.UpdateOptions) (*corev1.Secret, error)
		mockDeletePodCollection func(ctx context.Context, namespace string, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "error getting tenant information",
			wantErr: true,
			args: args{
				ctx:            ctx,
				operatorClient: opClient,
				client:         kubernetesClient,
				namespace:      "default",
				params: operator_api.UpdateTenantConfigurationParams{
					Namespace: "default",
					Tenant:    "test",
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return nil, errors.New("error getting tenant")
				},
				mockGetSecret: func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
					return nil, nil
				},
				mockUpdateSecret: func(ctx context.Context, namespace string, secret *corev1.Secret, opts metav1.UpdateOptions) (*corev1.Secret, error) {
					return nil, nil
				},
				mockDeletePodCollection: func(ctx context.Context, namespace string, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
					return nil
				},
			},
		},
		{
			name:    "error during getting tenant configuration",
			wantErr: true,
			args: args{
				ctx:            ctx,
				operatorClient: opClient,
				client:         kubernetesClient,
				namespace:      "default",
				params: operator_api.UpdateTenantConfigurationParams{
					Namespace: "default",
					Tenant:    "test",
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      "test",
						},
						Spec: miniov2.TenantSpec{
							Configuration: &corev1.LocalObjectReference{
								Name: "tenant-configuration-secret",
							},
						},
					}, nil
				},
				mockGetSecret: func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
					return nil, errors.New("error getting tenant configuration")
				},
				mockUpdateSecret: func(ctx context.Context, namespace string, secret *corev1.Secret, opts metav1.UpdateOptions) (*corev1.Secret, error) {
					return nil, nil
				},
				mockDeletePodCollection: func(ctx context.Context, namespace string, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
					return nil
				},
			},
		},
		{
			name:    "error updating tenant configuration because of missing configuration secret",
			wantErr: true,
			args: args{
				ctx:            ctx,
				operatorClient: opClient,
				client:         kubernetesClient,
				namespace:      "default",
				params: operator_api.UpdateTenantConfigurationParams{
					Namespace: "default",
					Tenant:    "test",
					Body:      &models.UpdateTenantConfigurationRequest{},
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      "test",
						},
					}, nil
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
				mockUpdateSecret: func(ctx context.Context, namespace string, secret *corev1.Secret, opts metav1.UpdateOptions) (*corev1.Secret, error) {
					return nil, nil
				},
				mockDeletePodCollection: func(ctx context.Context, namespace string, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
					return nil
				},
			},
		},
		{
			name:    "error because tenant configuration secret is nil",
			wantErr: true,
			args: args{
				ctx:            ctx,
				operatorClient: opClient,
				client:         kubernetesClient,
				namespace:      "default",
				params: operator_api.UpdateTenantConfigurationParams{
					Namespace: "default",
					Tenant:    "test",
					Body:      &models.UpdateTenantConfigurationRequest{},
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      "test",
						},
						Spec: miniov2.TenantSpec{
							Configuration: &corev1.LocalObjectReference{
								Name: "tenant-configuration-secret",
							},
						},
					}, nil
				},
				mockGetSecret: func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
					return nil, nil
				},
				mockUpdateSecret: func(ctx context.Context, namespace string, secret *corev1.Secret, opts metav1.UpdateOptions) (*corev1.Secret, error) {
					return nil, nil
				},
				mockDeletePodCollection: func(ctx context.Context, namespace string, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
					return nil
				},
			},
		},
		{
			name:    "error updating tenant configuration because of k8s issue",
			wantErr: true,
			args: args{
				ctx:            ctx,
				operatorClient: opClient,
				client:         kubernetesClient,
				namespace:      "default",
				params: operator_api.UpdateTenantConfigurationParams{
					Namespace: "default",
					Tenant:    "test",
					Body:      &models.UpdateTenantConfigurationRequest{},
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      "test",
						},
						Spec: miniov2.TenantSpec{
							Configuration: &corev1.LocalObjectReference{
								Name: "tenant-configuration-secret",
							},
						},
					}, nil
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
				mockUpdateSecret: func(ctx context.Context, namespace string, secret *corev1.Secret, opts metav1.UpdateOptions) (*corev1.Secret, error) {
					return nil, errors.New("error updating configuration secret")
				},
				mockDeletePodCollection: func(ctx context.Context, namespace string, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
					return nil
				},
			},
		},
		{
			name:    "error during deleting pod collection",
			wantErr: true,
			args: args{
				ctx:            ctx,
				operatorClient: opClient,
				client:         kubernetesClient,
				namespace:      "default",
				params: operator_api.UpdateTenantConfigurationParams{
					Namespace: "default",
					Tenant:    "test",
					Body:      &models.UpdateTenantConfigurationRequest{},
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      "test",
						},
						Spec: miniov2.TenantSpec{
							Configuration: &corev1.LocalObjectReference{
								Name: "tenant-configuration-secret",
							},
						},
					}, nil
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
				mockUpdateSecret: func(ctx context.Context, namespace string, secret *corev1.Secret, opts metav1.UpdateOptions) (*corev1.Secret, error) {
					return nil, nil
				},
				mockDeletePodCollection: func(ctx context.Context, namespace string, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
					return errors.New("error deleting minio pods")
				},
			},
		},
		{
			name:    "success updating tenant configuration secret",
			wantErr: false,
			args: args{
				ctx:            ctx,
				operatorClient: opClient,
				client:         kubernetesClient,
				namespace:      "default",
				params: operator_api.UpdateTenantConfigurationParams{
					Namespace: "default",
					Tenant:    "test",
					Body:      &models.UpdateTenantConfigurationRequest{},
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      "test",
						},
						Spec: miniov2.TenantSpec{
							Configuration: &corev1.LocalObjectReference{
								Name: "tenant-configuration-secret",
							},
						},
					}, nil
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
				mockUpdateSecret: func(ctx context.Context, namespace string, secret *corev1.Secret, opts metav1.UpdateOptions) (*corev1.Secret, error) {
					return nil, nil
				},
				mockDeletePodCollection: func(ctx context.Context, namespace string, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
					return nil
				},
			},
		},
	}
	for _, tt := range tests {
		k8sclientGetSecretMock = tt.args.mockGetSecret
		opClientTenantGetMock = tt.args.mockTenantGet
		k8sClientUpdateSecretMock = tt.args.mockUpdateSecret
		k8sClientDeletePodCollectionMock = tt.args.mockDeletePodCollection
		t.Run(tt.name, func(t *testing.T) {
			err := updateTenantConfigurationFile(tt.args.ctx, tt.args.operatorClient, tt.args.client, tt.args.namespace, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateTenantConfigurationFile(%v, %v, %v, %v, %v)", tt.args.ctx, tt.args.operatorClient, tt.args.client, tt.args.namespace, tt.args.params)
			}
		})
	}
}
