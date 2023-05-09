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
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/minio/operator/api/operations/operator_api"

	"github.com/minio/operator/models"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_tenantUpdateCertificates(t *testing.T) {
	k8sClient := k8sClientMock{}
	opClient := opClientMock{}
	crt := "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUUzRENDQTBTZ0F3SUJBZ0lRS0pDalNyK0xaMnJrYVo5ODAvZnV5akFOQmdrcWhraUc5dzBCQVFzRkFEQ0IKdFRFZU1Cd0dBMVVFQ2hNVmJXdGpaWEowSUdSbGRtVnNiM0J0Wlc1MElFTkJNVVV3UXdZRFZRUUxERHhoYkdWMgpjMnRBVEdWdWFXNXpMVTFoWTBKdmIyc3RVSEp2TG14dlkyRnNJQ2hNWlc1cGJpQkJiR1YyYzJ0cElFaDFaWEowCllTQkJjbWxoY3lreFREQktCZ05WQkFNTVEyMXJZMlZ5ZENCaGJHVjJjMnRBVEdWdWFXNXpMVTFoWTBKdmIyc3QKVUhKdkxteHZZMkZzSUNoTVpXNXBiaUJCYkdWMmMydHBJRWgxWlhKMFlTQkJjbWxoY3lrd0hoY05NVGt3TmpBeApNREF3TURBd1doY05NekF3T0RBeU1ERTFNekEzV2pCaU1TY3dKUVlEVlFRS0V4NXRhMk5sY25RZ1pHVjJaV3h2CmNHMWxiblFnWTJWeWRHbG1hV05oZEdVeE56QTFCZ05WQkFzTUxtRnNaWFp6YTBCMGFXWmhMbXh2WTJGc0lDaE0KWlc1cGJpQkJiR1YyYzJ0cElFaDFaWEowWVNCQmNtbGhjeWt3Z2dFaU1BMEdDU3FHU0liM0RRRUJBUVVBQTRJQgpEd0F3Z2dFS0FvSUJBUUM2dlhLVyswWmhJQUNycmpxTU9QZ3VSdGpSemk1L0VxK2JvZTJkQ1BpT3djdXo0WFlhCm1rVlcxSkhBS3VTc0I1UHE0QnFSRXFueUhYT0ZROWQ1bEFjRGNDYmhlTFRVb2h2ZkhOWW9SdWRrYkZ3RzAyOFQKdVMxTFNtcHU5VjhFU2Q0Q3BiOGRvUkcvUW8vRTF4RGk5STFhNVhoeUdHTTNsUFlYQU1HU1ZkWUNNNzh0M2ZLTwp0NFVlcEt6d2dFQ2NKRVg4MVFoem1ZT25ELysyazNxSVppZU9nOGFkbDNFUWV2eFBISXlXQ1JkaDJZTkZsRi9rCmEwTzQyRVl2NVNUT1N0dzI2TzMwTVRrcEg0dzRRZm9VV3BnZU93MDZyYTluRHdzV3VhMUZIaHlKZmxOanR1USsKU0NlaDdqTVlVUThYV1JkbXVHbzFRT0g5cjdDKy82L1JhMlZIQWdNQkFBR2pnYmt3Z2JZd0RnWURWUjBQQVFILwpCQVFEQWdXZ01CTUdBMVVkSlFRTU1Bb0dDQ3NHQVFVRkJ3TUJNQXdHQTFVZEV3RUIvd1FDTUFBd0h3WURWUjBqCkJCZ3dGb0FVNHg4NUIyeGVlQzg0UEdvM1dZVWwxRGVvZlFNd1lBWURWUjBSQkZrd1Y0SWpLaTV0YVc1cGJ5MWgKYkdWMmMyc3Ribk11YzNaakxtTnNkWE4wWlhJdWJHOWpZV3lDTUNvdWFHOXVaWGwzWld4c0xXaHNMbTFwYm1sdgpMV0ZzWlhaemF5MXVjeTV6ZG1NdVkyeDFjM1JsY2k1c2IyTmhiREFOQmdrcWhraUc5dzBCQVFzRkFBT0NBWUVBCnZnMFlIVGtzd3Z4NEE5ZXl0b1V3eTBOSFVjQXpuNy9BVCt3WEt6d3Fqa0RiS3hVYlFTMFNQVVI4SkVDMkpuUUgKU3pTNGFCNXRURVRYZUcrbHJZVU02b0RNcXlIalpUN1NJaW83OUNxOFloWWxORU9qMkhvaXdxalczZm10UU9kVApoVG1tS01lRnZleHo2cnRxbHp0cVdKa3kvOGd1MkMrMWw5UDFFUmhFNDZZY0puVmJ5REFTSGNvV2tiZVhFUGErCjNpNFd4bU1yMDlNZXFTNkFUb2NKanFBeEtYcURYWmlFZjRUVEp1ZTRBQ2s4WmhqaEtUUEV6RnQ3WllVaHY3dVoKdlZCOXhla2FqRzdMRU5kSHZncXVMRUxUV3czWkpBaXpSTU5KUUpzZU11LzdQN1NIeXRKTVJYdlVwd2R2dWhDOApKNm54aGRmUS91QVlQY1lHdlU5NUZPZ1NjWVZqNW1WQktXM0ZHbkl2YzZuamQ1OFhBSTE3dlk0K0ZZTnY4M2UxCm9mOGlxRFdWNTFyT1FlbG9FZFdmdHkvZTI2bXVWUUQzQlVjY2Z5SWpGYy9SeGNHdm5maUEwUm1uRDNkWFcyZ0oKUHFTd01ZZ3hxQndqMm1Qck5jdkFWY3BOeWRjTWJKOTFIOHRWY0ttRHMrY1NiV0tnblpmKzUvZm5yaTdmU3FLUQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
	key := "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUV2Z0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktnd2dnU2tBZ0VBQW9JQkFRQzZ2WEtXKzBaaElBQ3IKcmpxTU9QZ3VSdGpSemk1L0VxK2JvZTJkQ1BpT3djdXo0WFlhbWtWVzFKSEFLdVNzQjVQcTRCcVJFcW55SFhPRgpROWQ1bEFjRGNDYmhlTFRVb2h2ZkhOWW9SdWRrYkZ3RzAyOFR1UzFMU21wdTlWOEVTZDRDcGI4ZG9SRy9Rby9FCjF4RGk5STFhNVhoeUdHTTNsUFlYQU1HU1ZkWUNNNzh0M2ZLT3Q0VWVwS3p3Z0VDY0pFWDgxUWh6bVlPbkQvKzIKazNxSVppZU9nOGFkbDNFUWV2eFBISXlXQ1JkaDJZTkZsRi9rYTBPNDJFWXY1U1RPU3R3MjZPMzBNVGtwSDR3NApRZm9VV3BnZU93MDZyYTluRHdzV3VhMUZIaHlKZmxOanR1UStTQ2VoN2pNWVVROFhXUmRtdUdvMVFPSDlyN0MrCi82L1JhMlZIQWdNQkFBRUNnZ0VCQUlyTlVFUnJSM2ZmK3IraGhJRS93ekZhbGNUMUZWaDh3aXpUWXJQcnZCMFkKYlZvcVJzZ2xUVTdxTjkvM3dmc2dzdEROZk5IQ1pySEJOR0drK0orMDZMV2tnakhycjdXeFBUaE16ZDRvUGN4RwpRdTBMOGE5ZVlBMXJwY3NOOVc5Um5JU3BRSEk4aTkxM0V6Z0RoOWk2WCt0bFQyNjNNK0JYaDhlM1Z5cDNSTmhpCjZZUTQwcWJsNlQ0TUlyLzRSMGJmcFExZWVMNVNnbHB6Z1d6ZGs4WGtmc0E5YnRiU1RoMjZKRlBPUU1tMm5adkcKbjBlZm85TDZtaktwRW9rRWpTY1hWYTRZNHJaREZFYTVIbkpUSDBHblByQzU1cDhBb3BFN1c1M2RDT3lxd29CcQo4SWtZb2grcm1qTUZrOGc5VlRVYlcwVE9YVTFSY1E1Z0gxWS9jam5uVTRrQ2dZRUE4amw5ZEQyN2JaQ2ZaTW9jCjJYRThiYkJTaFVXTjRvaklKc1lHY2xyTjJDUEtUNmExaVhvS3BrTXdGNUdsODEzQURGR1pTbWtUSUFVQXRXQU8KNzZCcEpGUlVCZ1hmUXhSU0gyS1RaRldxbE5yekZPQjNsT3h2bFJ1amw5eE9ueStyWUhJWC9BOWNqQlp3a2orSAo3MDZRTExvS1ZFL2lMYy9DMlI5VzRLRVI3R3NDZ1lFQXhWd3FyM1JnUXhHUmpueG1zbDZtUW5DQ3k2MXFoUzUxCkJicDJzWldraFNTV0w0YzFDY0NDMnZ2ektPSmJkZ2NZQU43L05sTHgzWjlCZUFjTVZMUEVoc2NPalB6YUlneEMKczJ2UkdwQUFtYnRjWi9MVlpKaERWTjIrSHowRHhnRUtCa1JzQU5XTG51cGRoUWJBdlZuTkVsY29YVlo1cC9qcwp3U1BCelFoSklaVUNnWUJqTUlHY0dTOW9SWUhRRHlmVEx4aVV2bEI4ZktnR2JRYXhRZ1FmemVsZktnRE5yekhGCnN6RXJOblk2SUkxNVpCbWhzY1I1QVNBd3kzdW55a2N6ZjFldTVjMW1qZjhJQkFsQkN1ZmFmVzRWK0xiMEJKdFQKWTZLcHg2Q3RMaTBQNk1CZ0JUaW5JazgrbW0zTXBiRnZvSmRQaVh0elhTYjhwWWhmeXdLVGg4SEVNd0tCZ1FDUwpvR0VPTFpYKy9pUjRDYkI2d0pzaExWbmZYSjJSQ096a0xwNVVYV3IzaURFVWFvMWJDMjJzcUJjRnZ2WllmL2l6ClhQbWJNSkNGS1BhSTZDT2ZJbGZXRWptYlFaZ0dSN21lZDNISkhFZDE3NTg5azBvN0RHeXB0bnl6MUs3akFvNmkKRFY5NFZ5NytCLzBuQWRkY1ZrVm5aTjJXU3RMam1xcTY2NGZtZmt0bTZRS0JnQzBsMk5OVlFWK2N0OFFRRFpINQpBRFIrSGM3Qk0wNDhURGhNTmhoR3JHc2tWNngwMCtMZTdISGdpT3h2NXJudkNlTlY2M001YUZHdFVWbllTN1VoCkE1NndaNVlZeFFnQ0xzNi9PRmZhK3NiTngrSjdnSjRjNXdMZVlJMXVPMTlzZHBHa2VHZ25vK3dXVmxDSzFCbW0KRGM0TXA2STRiUTVtdy93YVpLQnpjRTJLCi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0K"
	badCrt := "ASD"
	badKey := "ASD"
	type args struct {
		ctx                     context.Context
		opClient                OperatorClientI
		clientSet               K8sClientI
		namespace               string
		params                  operator_api.TenantUpdateCertificateParams
		mockTenantGet           func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error)
		mockDeleteSecret        func(ctx context.Context, namespace string, name string, opts metav1.DeleteOptions) error
		mockCreateSecret        func(ctx context.Context, namespace string, secret *v1.Secret, opts metav1.CreateOptions) (*v1.Secret, error)
		mockDeletePodCollection func(ctx context.Context, namespace string, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "error getting tenant information",
			args: args{
				ctx:       context.Background(),
				opClient:  opClient,
				clientSet: k8sClient,
				namespace: "",
				params:    operator_api.TenantUpdateCertificateParams{},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return nil, errors.New("invalid tenant")
				},
			},
			wantErr: true,
		},
		{
			name: "error replacing external certs for tenant because of empty keypair values",
			args: args{
				ctx:       context.Background(),
				opClient:  opClient,
				clientSet: k8sClient,
				namespace: "",
				params: operator_api.TenantUpdateCertificateParams{
					Body: &models.TLSConfiguration{
						MinioServerCertificates: []*models.KeyPairConfiguration{
							{
								Crt: nil,
								Key: nil,
							},
						},
					},
				},
				mockDeletePodCollection: func(ctx context.Context, namespace string, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
					return nil
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{
						Spec: miniov2.TenantSpec{
							ExternalCertSecret: []*miniov2.LocalCertificateReference{
								{
									Name: "secret",
								},
							},
						},
					}, nil
				},
			},
			wantErr: true,
		},
		{
			name: "error replacing external certs for tenant because of malformed encoded certs",
			args: args{
				ctx:       context.Background(),
				opClient:  opClient,
				clientSet: k8sClient,
				namespace: "",
				params: operator_api.TenantUpdateCertificateParams{
					Body: &models.TLSConfiguration{
						MinioServerCertificates: []*models.KeyPairConfiguration{
							{
								Crt: &badCrt,
								Key: &badKey,
							},
						},
					},
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{
						Spec: miniov2.TenantSpec{
							ExternalCertSecret: []*miniov2.LocalCertificateReference{
								{
									Name: "secret",
								},
							},
						},
					}, nil
				},
				mockDeleteSecret: func(ctx context.Context, namespace string, name string, opts metav1.DeleteOptions) error {
					return nil
				},
			},
			wantErr: true,
		},
		{
			name: "error replacing external certs for tenant because of error during secret creation",
			args: args{
				ctx:       context.Background(),
				opClient:  opClient,
				clientSet: k8sClient,
				namespace: "",
				params: operator_api.TenantUpdateCertificateParams{
					Body: &models.TLSConfiguration{
						MinioServerCertificates: []*models.KeyPairConfiguration{
							{
								Crt: &crt,
								Key: &key,
							},
						},
					},
				},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return &miniov2.Tenant{
						Spec: miniov2.TenantSpec{
							ExternalCertSecret: []*miniov2.LocalCertificateReference{
								{
									Name: "secret",
								},
							},
						},
					}, nil
				},
				mockDeleteSecret: func(ctx context.Context, namespace string, name string, opts metav1.DeleteOptions) error {
					return nil
				},
				mockCreateSecret: func(ctx context.Context, namespace string, secret *v1.Secret, opts metav1.CreateOptions) (*v1.Secret, error) {
					return nil, errors.New("error creating secret")
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		opClientTenantGetMock = tt.args.mockTenantGet
		k8sClientDeleteSecretMock = tt.args.mockDeleteSecret
		k8sClientCreateSecretMock = tt.args.mockCreateSecret
		k8sClientDeletePodCollectionMock = tt.args.mockDeletePodCollection
		t.Run(tt.name, func(t *testing.T) {
			if err := tenantUpdateCertificates(tt.args.ctx, tt.args.opClient, tt.args.clientSet, tt.args.namespace, tt.args.params); (err != nil) != tt.wantErr {
				t.Errorf("tenantUpdateCertificates() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_tenantUpdateEncryption(t *testing.T) {
	k8sClient := k8sClientMock{}
	opClient := opClientMock{}
	type args struct {
		ctx                     context.Context
		opClient                OperatorClientI
		clientSet               K8sClientI
		namespace               string
		params                  operator_api.TenantUpdateEncryptionParams
		mockTenantGet           func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error)
		mockDeleteSecret        func(ctx context.Context, namespace string, name string, opts metav1.DeleteOptions) error
		mockCreateSecret        func(ctx context.Context, namespace string, secret *v1.Secret, opts metav1.CreateOptions) (*v1.Secret, error)
		mockDeletePodCollection func(ctx context.Context, namespace string, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "error updating encryption configuration because of error getting tenant",
			args: args{
				ctx:       context.Background(),
				opClient:  opClient,
				clientSet: k8sClient,
				namespace: "",
				params:    operator_api.TenantUpdateEncryptionParams{},
				mockTenantGet: func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
					return nil, errors.New("invalid tenant")
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opClientTenantGetMock = tt.args.mockTenantGet
			k8sClientDeleteSecretMock = tt.args.mockDeleteSecret
			k8sClientCreateSecretMock = tt.args.mockCreateSecret
			k8sClientDeletePodCollectionMock = tt.args.mockDeletePodCollection
			if err := tenantUpdateEncryption(tt.args.ctx, tt.args.opClient, tt.args.clientSet, tt.args.namespace, tt.args.params); (err != nil) != tt.wantErr {
				t.Errorf("tenantUpdateEncryption() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_createOrReplaceKesConfigurationSecrets(t *testing.T) {
	k8sClient := k8sClientMock{}
	crt := "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUUzRENDQTBTZ0F3SUJBZ0lRS0pDalNyK0xaMnJrYVo5ODAvZnV5akFOQmdrcWhraUc5dzBCQVFzRkFEQ0IKdFRFZU1Cd0dBMVVFQ2hNVmJXdGpaWEowSUdSbGRtVnNiM0J0Wlc1MElFTkJNVVV3UXdZRFZRUUxERHhoYkdWMgpjMnRBVEdWdWFXNXpMVTFoWTBKdmIyc3RVSEp2TG14dlkyRnNJQ2hNWlc1cGJpQkJiR1YyYzJ0cElFaDFaWEowCllTQkJjbWxoY3lreFREQktCZ05WQkFNTVEyMXJZMlZ5ZENCaGJHVjJjMnRBVEdWdWFXNXpMVTFoWTBKdmIyc3QKVUhKdkxteHZZMkZzSUNoTVpXNXBiaUJCYkdWMmMydHBJRWgxWlhKMFlTQkJjbWxoY3lrd0hoY05NVGt3TmpBeApNREF3TURBd1doY05NekF3T0RBeU1ERTFNekEzV2pCaU1TY3dKUVlEVlFRS0V4NXRhMk5sY25RZ1pHVjJaV3h2CmNHMWxiblFnWTJWeWRHbG1hV05oZEdVeE56QTFCZ05WQkFzTUxtRnNaWFp6YTBCMGFXWmhMbXh2WTJGc0lDaE0KWlc1cGJpQkJiR1YyYzJ0cElFaDFaWEowWVNCQmNtbGhjeWt3Z2dFaU1BMEdDU3FHU0liM0RRRUJBUVVBQTRJQgpEd0F3Z2dFS0FvSUJBUUM2dlhLVyswWmhJQUNycmpxTU9QZ3VSdGpSemk1L0VxK2JvZTJkQ1BpT3djdXo0WFlhCm1rVlcxSkhBS3VTc0I1UHE0QnFSRXFueUhYT0ZROWQ1bEFjRGNDYmhlTFRVb2h2ZkhOWW9SdWRrYkZ3RzAyOFQKdVMxTFNtcHU5VjhFU2Q0Q3BiOGRvUkcvUW8vRTF4RGk5STFhNVhoeUdHTTNsUFlYQU1HU1ZkWUNNNzh0M2ZLTwp0NFVlcEt6d2dFQ2NKRVg4MVFoem1ZT25ELysyazNxSVppZU9nOGFkbDNFUWV2eFBISXlXQ1JkaDJZTkZsRi9rCmEwTzQyRVl2NVNUT1N0dzI2TzMwTVRrcEg0dzRRZm9VV3BnZU93MDZyYTluRHdzV3VhMUZIaHlKZmxOanR1USsKU0NlaDdqTVlVUThYV1JkbXVHbzFRT0g5cjdDKy82L1JhMlZIQWdNQkFBR2pnYmt3Z2JZd0RnWURWUjBQQVFILwpCQVFEQWdXZ01CTUdBMVVkSlFRTU1Bb0dDQ3NHQVFVRkJ3TUJNQXdHQTFVZEV3RUIvd1FDTUFBd0h3WURWUjBqCkJCZ3dGb0FVNHg4NUIyeGVlQzg0UEdvM1dZVWwxRGVvZlFNd1lBWURWUjBSQkZrd1Y0SWpLaTV0YVc1cGJ5MWgKYkdWMmMyc3Ribk11YzNaakxtTnNkWE4wWlhJdWJHOWpZV3lDTUNvdWFHOXVaWGwzWld4c0xXaHNMbTFwYm1sdgpMV0ZzWlhaemF5MXVjeTV6ZG1NdVkyeDFjM1JsY2k1c2IyTmhiREFOQmdrcWhraUc5dzBCQVFzRkFBT0NBWUVBCnZnMFlIVGtzd3Z4NEE5ZXl0b1V3eTBOSFVjQXpuNy9BVCt3WEt6d3Fqa0RiS3hVYlFTMFNQVVI4SkVDMkpuUUgKU3pTNGFCNXRURVRYZUcrbHJZVU02b0RNcXlIalpUN1NJaW83OUNxOFloWWxORU9qMkhvaXdxalczZm10UU9kVApoVG1tS01lRnZleHo2cnRxbHp0cVdKa3kvOGd1MkMrMWw5UDFFUmhFNDZZY0puVmJ5REFTSGNvV2tiZVhFUGErCjNpNFd4bU1yMDlNZXFTNkFUb2NKanFBeEtYcURYWmlFZjRUVEp1ZTRBQ2s4WmhqaEtUUEV6RnQ3WllVaHY3dVoKdlZCOXhla2FqRzdMRU5kSHZncXVMRUxUV3czWkpBaXpSTU5KUUpzZU11LzdQN1NIeXRKTVJYdlVwd2R2dWhDOApKNm54aGRmUS91QVlQY1lHdlU5NUZPZ1NjWVZqNW1WQktXM0ZHbkl2YzZuamQ1OFhBSTE3dlk0K0ZZTnY4M2UxCm9mOGlxRFdWNTFyT1FlbG9FZFdmdHkvZTI2bXVWUUQzQlVjY2Z5SWpGYy9SeGNHdm5maUEwUm1uRDNkWFcyZ0oKUHFTd01ZZ3hxQndqMm1Qck5jdkFWY3BOeWRjTWJKOTFIOHRWY0ttRHMrY1NiV0tnblpmKzUvZm5yaTdmU3FLUQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
	key := "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUV2Z0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktnd2dnU2tBZ0VBQW9JQkFRQzZ2WEtXKzBaaElBQ3IKcmpxTU9QZ3VSdGpSemk1L0VxK2JvZTJkQ1BpT3djdXo0WFlhbWtWVzFKSEFLdVNzQjVQcTRCcVJFcW55SFhPRgpROWQ1bEFjRGNDYmhlTFRVb2h2ZkhOWW9SdWRrYkZ3RzAyOFR1UzFMU21wdTlWOEVTZDRDcGI4ZG9SRy9Rby9FCjF4RGk5STFhNVhoeUdHTTNsUFlYQU1HU1ZkWUNNNzh0M2ZLT3Q0VWVwS3p3Z0VDY0pFWDgxUWh6bVlPbkQvKzIKazNxSVppZU9nOGFkbDNFUWV2eFBISXlXQ1JkaDJZTkZsRi9rYTBPNDJFWXY1U1RPU3R3MjZPMzBNVGtwSDR3NApRZm9VV3BnZU93MDZyYTluRHdzV3VhMUZIaHlKZmxOanR1UStTQ2VoN2pNWVVROFhXUmRtdUdvMVFPSDlyN0MrCi82L1JhMlZIQWdNQkFBRUNnZ0VCQUlyTlVFUnJSM2ZmK3IraGhJRS93ekZhbGNUMUZWaDh3aXpUWXJQcnZCMFkKYlZvcVJzZ2xUVTdxTjkvM3dmc2dzdEROZk5IQ1pySEJOR0drK0orMDZMV2tnakhycjdXeFBUaE16ZDRvUGN4RwpRdTBMOGE5ZVlBMXJwY3NOOVc5Um5JU3BRSEk4aTkxM0V6Z0RoOWk2WCt0bFQyNjNNK0JYaDhlM1Z5cDNSTmhpCjZZUTQwcWJsNlQ0TUlyLzRSMGJmcFExZWVMNVNnbHB6Z1d6ZGs4WGtmc0E5YnRiU1RoMjZKRlBPUU1tMm5adkcKbjBlZm85TDZtaktwRW9rRWpTY1hWYTRZNHJaREZFYTVIbkpUSDBHblByQzU1cDhBb3BFN1c1M2RDT3lxd29CcQo4SWtZb2grcm1qTUZrOGc5VlRVYlcwVE9YVTFSY1E1Z0gxWS9jam5uVTRrQ2dZRUE4amw5ZEQyN2JaQ2ZaTW9jCjJYRThiYkJTaFVXTjRvaklKc1lHY2xyTjJDUEtUNmExaVhvS3BrTXdGNUdsODEzQURGR1pTbWtUSUFVQXRXQU8KNzZCcEpGUlVCZ1hmUXhSU0gyS1RaRldxbE5yekZPQjNsT3h2bFJ1amw5eE9ueStyWUhJWC9BOWNqQlp3a2orSAo3MDZRTExvS1ZFL2lMYy9DMlI5VzRLRVI3R3NDZ1lFQXhWd3FyM1JnUXhHUmpueG1zbDZtUW5DQ3k2MXFoUzUxCkJicDJzWldraFNTV0w0YzFDY0NDMnZ2ektPSmJkZ2NZQU43L05sTHgzWjlCZUFjTVZMUEVoc2NPalB6YUlneEMKczJ2UkdwQUFtYnRjWi9MVlpKaERWTjIrSHowRHhnRUtCa1JzQU5XTG51cGRoUWJBdlZuTkVsY29YVlo1cC9qcwp3U1BCelFoSklaVUNnWUJqTUlHY0dTOW9SWUhRRHlmVEx4aVV2bEI4ZktnR2JRYXhRZ1FmemVsZktnRE5yekhGCnN6RXJOblk2SUkxNVpCbWhzY1I1QVNBd3kzdW55a2N6ZjFldTVjMW1qZjhJQkFsQkN1ZmFmVzRWK0xiMEJKdFQKWTZLcHg2Q3RMaTBQNk1CZ0JUaW5JazgrbW0zTXBiRnZvSmRQaVh0elhTYjhwWWhmeXdLVGg4SEVNd0tCZ1FDUwpvR0VPTFpYKy9pUjRDYkI2d0pzaExWbmZYSjJSQ096a0xwNVVYV3IzaURFVWFvMWJDMjJzcUJjRnZ2WllmL2l6ClhQbWJNSkNGS1BhSTZDT2ZJbGZXRWptYlFaZ0dSN21lZDNISkhFZDE3NTg5azBvN0RHeXB0bnl6MUs3akFvNmkKRFY5NFZ5NytCLzBuQWRkY1ZrVm5aTjJXU3RMam1xcTY2NGZtZmt0bTZRS0JnQzBsMk5OVlFWK2N0OFFRRFpINQpBRFIrSGM3Qk0wNDhURGhNTmhoR3JHc2tWNngwMCtMZTdISGdpT3h2NXJudkNlTlY2M001YUZHdFVWbllTN1VoCkE1NndaNVlZeFFnQ0xzNi9PRmZhK3NiTngrSjdnSjRjNXdMZVlJMXVPMTlzZHBHa2VHZ25vK3dXVmxDSzFCbW0KRGM0TXA2STRiUTVtdy93YVpLQnpjRTJLCi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0K"
	badCrt := "ASD"
	badKey := "ASD"
	appRole := "ASD"
	appSecret := "ASD"
	endpoint := "ASD"
	type args struct {
		ctx                        context.Context
		clientSet                  K8sClientI
		ns                         string
		encryptionCfg              *models.EncryptionConfiguration
		kesConfigurationSecretName string
		kesClientCertSecretName    string
		kesImage                   string
		tenantName                 string
		mockDeleteSecret           func(ctx context.Context, namespace string, name string, opts metav1.DeleteOptions) error
		mockCreateSecret           func(ctx context.Context, namespace string, secret *v1.Secret, opts metav1.CreateOptions) (*v1.Secret, error)
	}
	tests := []struct {
		name    string
		args    args
		want    *v1.LocalObjectReference
		want1   *miniov2.LocalCertificateReference
		wantErr bool
	}{
		{
			name: "error decoding the client certificate",
			args: args{
				ctx:       context.Background(),
				clientSet: k8sClient,
				encryptionCfg: &models.EncryptionConfiguration{
					MinioMtls: &models.KeyPairConfiguration{
						Crt: &badCrt,
						Key: &badKey,
					},
				},
				ns:                         "default",
				kesImage:                   "minio/kes:v0.22.3",
				kesConfigurationSecretName: "test-secret",
				kesClientCertSecretName:    "test-client-secret",
				tenantName:                 "test",
				mockDeleteSecret: func(ctx context.Context, namespace string, name string, opts metav1.DeleteOptions) error {
					return nil
				},
			},
			want:    nil,
			want1:   nil,
			wantErr: true,
		},
		{
			name: "error because of malformed decoded certificate",
			args: args{
				ctx:       context.Background(),
				clientSet: k8sClient,
				encryptionCfg: &models.EncryptionConfiguration{
					MinioMtls: &models.KeyPairConfiguration{
						Crt: &key, // will cause an error because we are passing a private key as the public key
						Key: &key,
					},
				},
				ns:                         "default",
				kesImage:                   "minio/kes:v0.22.3",
				kesConfigurationSecretName: "test-secret",
				kesClientCertSecretName:    "test-client-secret",
				tenantName:                 "test",
				mockDeleteSecret: func(ctx context.Context, namespace string, name string, opts metav1.DeleteOptions) error {
					return nil
				},
			},
			want:    nil,
			want1:   nil,
			wantErr: true,
		},
		{
			name: "error because of malformed decoded certificate",
			args: args{
				ctx:       context.Background(),
				clientSet: k8sClient,
				encryptionCfg: &models.EncryptionConfiguration{
					MinioMtls: &models.KeyPairConfiguration{
						Crt: &crt,
						Key: &key,
					},
					KmsMtls: &models.EncryptionConfigurationAO1KmsMtls{
						Ca:  crt,
						Crt: crt,
						Key: key,
					},
					Vault: &models.VaultConfiguration{
						Approle: &models.VaultConfigurationApprole{
							Engine: "",
							ID:     &appRole,
							Retry:  0,
							Secret: &appSecret,
						},
						Endpoint:  &endpoint,
						Engine:    "",
						Namespace: "",
						Prefix:    "",
						Status:    nil,
					},
				},
				ns:                         "default",
				kesImage:                   "minio/kes:v0.22.3",
				kesConfigurationSecretName: "test-secret",
				kesClientCertSecretName:    "test-client-secret",
				tenantName:                 "test",
				mockDeleteSecret: func(ctx context.Context, namespace string, name string, opts metav1.DeleteOptions) error {
					return nil
				},
				mockCreateSecret: func(ctx context.Context, namespace string, secret *v1.Secret, opts metav1.CreateOptions) (*v1.Secret, error) {
					return &v1.Secret{}, nil
				},
			},
			want: &v1.LocalObjectReference{
				Name: "test-secret",
			},
			want1: &miniov2.LocalCertificateReference{
				Name: "test-client-secret",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8sClientDeleteSecretMock = tt.args.mockDeleteSecret
			k8sClientCreateSecretMock = tt.args.mockCreateSecret
			got, got1, err := createOrReplaceKesConfigurationSecrets(tt.args.ctx, tt.args.clientSet, tt.args.ns, tt.args.encryptionCfg, tt.args.kesConfigurationSecretName, tt.args.kesClientCertSecretName, tt.args.tenantName, tt.args.kesImage)
			if (err != nil) != tt.wantErr {
				t.Errorf("createOrReplaceKesConfigurationSecrets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createOrReplaceKesConfigurationSecrets() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("createOrReplaceKesConfigurationSecrets() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_GetConfigVersion(t *testing.T) {
	type args struct {
		kesImage string
	}
	tests := []struct {
		name        string
		args        args
		wantversion string
		wantErr     bool
	}{
		{
			name: "error unexpected KES config version",
			args: args{
				kesImage: "minio/kes:v0.22.0",
			},
			wantversion: KesConfigVersion2,
			wantErr:     false,
		},
		{
			name: "error unexpected KES config version",
			args: args{
				kesImage: "minio/kes:v0.21.0",
			},
			wantversion: KesConfigVersion1,
			wantErr:     false,
		},
		{
			name: "error unexpected KES config version",
			args: args{
				kesImage: "minio/kes:2023-02-15T14-54-37Z",
			},
			wantversion: KesConfigVersion2,
			wantErr:     false,
		},
		{
			name: "error unexpected KES config version",
			args: args{
				kesImage: "minio/kes:2023-04-03T16-41-28Z",
			},
			wantversion: KesConfigVersion2,
			wantErr:     false,
		},
		{
			name: "error unexpected KES config version",
			args: args{
				kesImage: "minio/kes:2023-05-02T22-48-10Z",
			},
			wantversion: KesConfigVersion2,
			wantErr:     false,
		},
		{
			name: "error unexpected KES config version",
			args: args{
				kesImage: "minio/kes:2023-05-02T22-48-10Z-arm64",
			},
			wantversion: KesConfigVersion2,
			wantErr:     false,
		},
		{
			name: "error unexpected KES config version",
			args: args{
				kesImage: "minio/kes:edge",
			},
			wantversion: KesConfigVersion2,
			wantErr:     false,
		},
		{
			name: "error unexpected KES config version",
			args: args{
				kesImage: "minio/kes:latest",
			},
			wantversion: KesConfigVersion2,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getKesConfigVersion(tt.args.kesImage)
			if (err != nil) != tt.wantErr {
				t.Errorf("getKesConfigVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantversion) {
				t.Errorf("getKesConfigVersion() got = %v, want %v", got, tt.wantversion)
			}
		})
	}
}
