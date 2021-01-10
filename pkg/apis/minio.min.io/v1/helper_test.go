package v1

import (
	"testing"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"

	corev1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEnsureDefaults(t *testing.T) {
	mt := Tenant{}
	mt.EnsureDefaults()

	t.Run("defaults", func(t *testing.T) {
		assert.Equal(t, mt.Spec.Image, DefaultMinIOImage)
		assert.Equal(t, mt.Spec.Mountpath, MinIOVolumeMountPath)
		assert.Equal(t, mt.Spec.Subpath, MinIOVolumeSubPath)
		// default behavior is autoCert will be enabled
		assert.True(t, mt.AutoCert())
		require.True(t, mt.HasCertConfig())
		require.NotNil(t, mt.Spec.CertConfig)
	})

	t.Run("enable and disable autoCert", func(t *testing.T) {
		// disable autoCert explicitly
		autoCertEnabled := false
		mt.Spec.RequestAutoCert = &autoCertEnabled

		mt.EnsureDefaults()

		assert.False(t, mt.AutoCert())
		assert.False(t, mt.HasCertConfig())
		require.Nil(t, mt.Spec.CertConfig)

		// enable autoCert explicitly
		autoCertEnabled = true
		mt.Spec.RequestAutoCert = &autoCertEnabled

		mt.EnsureDefaults()

		assert.True(t, mt.AutoCert())
		assert.True(t, mt.HasCertConfig())
		require.NotNil(t, mt.Spec.CertConfig)
	})

	t.Run("defaults don't override", func(t *testing.T) {
		newImage := "mtnio/mtnio:latest"
		mt.Spec.Image = newImage
		mt.EnsureDefaults()

		assert.Equal(t, newImage, mt.Spec.Image)
	})
}

func TestTemplateVariables(t *testing.T) {
	servers := 2
	mt := Tenant{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: TenantSpec{
			Zones: []Zone{
				{
					Name:                "single",
					Servers:             int32(servers),
					VolumesPerServer:    4,
					VolumeClaimTemplate: nil,
					Resources:           corev1.ResourceRequirements{},
					NodeSelector:        nil,
					Affinity:            nil,
					Tolerations:         nil,
				},
			},
		},
	}
	mt.EnsureDefaults()

	t.Run("StatefulSet", func(t *testing.T) {
		hosts := mt.TemplatedMinIOHosts("{{.StatefulSet}}")
		assert.Contains(t, hosts, mt.MinIOStatefulSetNameForZone(&mt.Spec.Zones[0]))
	})

	t.Run("CIService", func(t *testing.T) {
		hosts := mt.TemplatedMinIOHosts("{{.CIService}}")
		assert.Contains(t, hosts, mt.MinIOCIServiceName())
	})

	t.Run("HLService", func(t *testing.T) {
		hosts := mt.TemplatedMinIOHosts("{{.HLService}}")
		assert.Contains(t, hosts, mt.MinIOHLServiceName())
	})

	t.Run("Ellipsis", func(t *testing.T) {
		hosts := mt.TemplatedMinIOHosts("{{.Ellipsis}}")
		assert.Contains(t, hosts, genEllipsis(0, servers-1))
	})

	t.Run("Domain", func(t *testing.T) {
		hosts := mt.TemplatedMinIOHosts("{{.Domain}}")
		assert.Contains(t, hosts, ClusterDomain)
	})
}

func TestTenant_KESServiceEndpoint(t1 *testing.T) {
	type fields struct {
		TypeMeta   metav1.TypeMeta
		ObjectMeta metav1.ObjectMeta
		Scheduler  miniov2.TenantScheduler
		Spec       TenantSpec
		Status     miniov2.TenantStatus
	}
	ClusterDomain = "cluster.local"
	autoCertEnabled := true
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Success",
			fields: fields{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kes",
					Namespace: "namespace",
				},
				Spec: TenantSpec{
					RequestAutoCert: &autoCertEnabled,
				},
			},
			want: "https://kes" + KESHLSvcNameSuffix + ".namespace.svc.cluster.local:7373",
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Tenant{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Scheduler:  tt.fields.Scheduler,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			if got := t.KESServiceEndpoint(); got != tt.want {
				t1.Errorf("KESServiceEndpoint() = %v, want %v", got, tt.want)
			}
		})
	}
}
