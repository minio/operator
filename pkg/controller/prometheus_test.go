// Copyright (C) 2025, MinIO, Inc.
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

package controller

import (
	"context"
	"github.com/minio/operator/pkg/resources/configmaps"
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/fake"
	"gopkg.in/yaml.v2"
	v3 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fake3 "k8s.io/client-go/kubernetes/fake"
	"testing"
	"time"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
)

func Test_checkAndCreatePrometheusAddlConfig(t *testing.T) {
	type except struct {
		afterScrapeNumber  int
		afterScrapePath    []string
		afterScrapeName    []string
		changedTokenNumber int
	}
	type arg struct {
		accessKey     string
		secretKey     string
		beforeScrapes []configmaps.ScrapeConfig
		paths         []string
		excepts       except
	}
	type args struct {
		name string
		arg  arg
	}
	testRunCheckAndCreatePrometheusAddlConfig := func(t *testing.T, arg arg) {
		beforeScrapesBytes, err := yaml.Marshal(arg.beforeScrapes)
		if err != nil {
			t.Fatal(err)
		}
		testNS := "testNamespace"
		t.Setenv("PROMETHEUS_NAMESPACE", testNS)
		prometheus := &v1.Prometheus{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testPrometheus",
				Namespace: testNS,
			},
			Spec: v1.PrometheusSpec{
				CommonPrometheusFields: v1.CommonPrometheusFields{
					AdditionalScrapeConfigs: &v3.SecretKeySelector{
						Key: "",
						LocalObjectReference: v3.LocalObjectReference{
							Name: miniov2.PrometheusAddlScrapeConfigSecret,
						},
					},
				},
			},
		}
		beforeSecret := &v3.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      miniov2.PrometheusAddlScrapeConfigSecret,
				Namespace: testNS,
			},
			Data: map[string][]byte{
				miniov2.PrometheusAddlScrapeConfigKey: beforeScrapesBytes,
			},
		}
		kubeclient := fake3.NewSimpleClientset(beforeSecret)
		controller := Controller{
			promClient:    fake.NewSimpleClientset(prometheus),
			kubeClientSet: kubeclient,
		}
		tenant := &miniov2.Tenant{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testTenant",
				Namespace: testNS,
			},
			Spec: miniov2.TenantSpec{
				PrometheusOperatorScrapeMetricsPaths: arg.paths,
			},
		}
		// bear_token generate by unix timestamp
		time.Sleep(time.Second)
		err = controller.checkAndCreatePrometheusAddlConfig(context.Background(), tenant, arg.accessKey, arg.secretKey)
		if err != nil {
			t.Fatal(err)
		}
		afterSecret, err := kubeclient.CoreV1().Secrets(testNS).Get(context.Background(), miniov2.PrometheusAddlScrapeConfigSecret, metav1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}
		var afterScrapes []configmaps.ScrapeConfig
		err = yaml.Unmarshal(afterSecret.Data[miniov2.PrometheusAddlScrapeConfigKey], &afterScrapes)
		if err != nil {
			t.Fatal(err)
		}
		if len(afterScrapes) != arg.excepts.afterScrapeNumber {
			t.Fatalf("Expected %d scrape configs, got %d", arg.excepts.afterScrapeNumber, len(afterScrapes))
		}
		for i, scrape := range afterScrapes {
			if scrape.JobName != arg.excepts.afterScrapeName[i] {
				t.Fatalf("Expected scrape config name %s, got %s", arg.excepts.afterScrapeName[i], scrape.JobName)
			}
			if scrape.MetricsPath != arg.excepts.afterScrapePath[i] {
				t.Fatalf("Expected scrape config path %s, got %s", arg.excepts.afterScrapePath[i], scrape.MetricsPath)
			}
		}
		changedTokenNumber := 0
		for _, afterScrape := range afterScrapes {
			for _, scrape := range arg.beforeScrapes {
				if scrape.JobName == afterScrape.JobName {
					if scrape.BearerToken != afterScrape.BearerToken {
						changedTokenNumber++
					}
					break
				}
			}
		}
		if changedTokenNumber != arg.excepts.changedTokenNumber {
			t.Fatalf("Expected %d changed tokens, got %d", arg.excepts.changedTokenNumber, changedTokenNumber)
		}

	}
	tests := []args{
		{
			name: "testDefault",
			arg: arg{
				accessKey:     "accessKey",
				secretKey:     "secretKey",
				beforeScrapes: []configmaps.ScrapeConfig{},
				excepts: except{
					afterScrapeNumber:  1,
					afterScrapePath:    []string{"/minio/v2/metrics/cluster"},
					afterScrapeName:    []string{"testTenant-minio-job-0"},
					changedTokenNumber: 0,
				},
			},
		},
		{
			name: "testFirst",
			arg: arg{
				accessKey:     "accessKey",
				secretKey:     "secretKey",
				paths:         []string{"/minio/v2/metrics/cluster", "/minio/metrics/v3/api"},
				beforeScrapes: []configmaps.ScrapeConfig{},
				excepts: except{
					afterScrapeNumber:  2,
					afterScrapePath:    []string{"/minio/v2/metrics/cluster", "/minio/metrics/v3/api"},
					afterScrapeName:    []string{"testTenant-minio-job-0", "testTenant-minio-job-1"},
					changedTokenNumber: 0,
				},
			},
		},
		{
			name: "testAlreadyHave",
			arg: arg{
				accessKey: "accessKey",
				secretKey: "secretKey",
				paths:     []string{"/minio/v2/metrics/cluster", "/minio/metrics/v3/api"},
				beforeScrapes: []configmaps.ScrapeConfig{
					{
						JobName: "testTarget",
						StaticConfigs: []configmaps.StaticConfig{
							{
								Targets: []string{"testTarget"},
							},
						},
					},
				},
				excepts: except{
					afterScrapeNumber:  3,
					afterScrapePath:    []string{"", "/minio/v2/metrics/cluster", "/minio/metrics/v3/api"},
					afterScrapeName:    []string{"testTarget", "testTenant-minio-job-0", "testTenant-minio-job-1"},
					changedTokenNumber: 0,
				},
			},
		},
		{
			name: "testNoChange",
			arg: arg{
				accessKey: "accessKey",
				secretKey: "secretKey",
				paths:     []string{"/minio/v2/metrics/cluster", "/minio/metrics/v3/api"},
				beforeScrapes: []configmaps.ScrapeConfig{
					{
						JobName: "testTarget",
						StaticConfigs: []configmaps.StaticConfig{
							{
								Targets: []string{"testTarget"},
							},
						},
					},
					{
						JobName:     "testTenant-minio-job-0",
						BearerToken: (&miniov2.Tenant{}).GenBearerToken("accessKey", "secretKey"),
						MetricsPath: "/minio/v2/metrics/cluster",
						Scheme:      "https",
						TLSConfig: configmaps.TlsConfig{
							CAFile: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
						},
						StaticConfigs: []configmaps.StaticConfig{
							{
								Targets: []string{"minio.testNamespace.svc.cluster.local:443"},
							},
						},
					},
					{
						JobName:     "testTenant-minio-job-1",
						BearerToken: (&miniov2.Tenant{}).GenBearerToken("accessKey", "secretKey"),
						MetricsPath: "/minio/metrics/v3/api",
						Scheme:      "https",
						TLSConfig: configmaps.TlsConfig{
							CAFile: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
						},
						StaticConfigs: []configmaps.StaticConfig{
							{
								Targets: []string{"minio.testNamespace.svc.cluster.local:443"},
							},
						},
					},
				},
				excepts: except{
					afterScrapeNumber:  3,
					afterScrapePath:    []string{"", "/minio/v2/metrics/cluster", "/minio/metrics/v3/api"},
					afterScrapeName:    []string{"testTarget", "testTenant-minio-job-0", "testTenant-minio-job-1"},
					changedTokenNumber: 0,
				},
			},
		},
		{
			name: "testPasswordChange",
			arg: arg{
				accessKey: "accessKeyChanged",
				secretKey: "secretKeyChanged",
				paths:     []string{"/minio/v2/metrics/cluster", "/minio/metrics/v3/api"},
				beforeScrapes: []configmaps.ScrapeConfig{
					{
						JobName: "testTarget",
						StaticConfigs: []configmaps.StaticConfig{
							{
								Targets: []string{"testTarget"},
							},
						},
					},
					{
						JobName:     "testTenant-minio-job-0",
						BearerToken: (&miniov2.Tenant{}).GenBearerToken("accessKey", "secretKey"),
						MetricsPath: "/minio/v2/metrics/cluster",
						Scheme:      "https",
						TLSConfig: configmaps.TlsConfig{
							CAFile: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
						},
						StaticConfigs: []configmaps.StaticConfig{
							{
								Targets: []string{"minio.testNamespace.svc.cluster.local:443"},
							},
						},
					},
					{
						JobName:     "testTenant-minio-job-1",
						BearerToken: (&miniov2.Tenant{}).GenBearerToken("accessKey", "secretKey"),
						MetricsPath: "/minio/metrics/v3/api",
						Scheme:      "https",
						TLSConfig: configmaps.TlsConfig{
							CAFile: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
						},
						StaticConfigs: []configmaps.StaticConfig{
							{
								Targets: []string{"minio.testNamespace.svc.cluster.local:443"},
							},
						},
					},
				},
				excepts: except{
					afterScrapeNumber:  3,
					afterScrapePath:    []string{"", "/minio/v2/metrics/cluster", "/minio/metrics/v3/api"},
					afterScrapeName:    []string{"testTarget", "testTenant-minio-job-0", "testTenant-minio-job-1"},
					changedTokenNumber: 2,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRunCheckAndCreatePrometheusAddlConfig(t, tt.arg)
		})
	}
}
