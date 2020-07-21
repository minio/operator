package v1

import (
	"testing"

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
		assert.False(t, mt.AutoCert())
	})

	t.Run("auto cert", func(t *testing.T) {
		mt.Spec.RequestAutoCert = true
		assert.True(t, mt.AutoCert())
		assert.False(t, mt.HasCertConfig())

		mt.EnsureDefaults()

		require.NotNil(t, mt.Spec.CertConfig)
		require.True(t, mt.HasCertConfig())
		oldCertConfig := mt.Spec.CertConfig

		mt.EnsureDefaults()

		assert.Equal(t, oldCertConfig, mt.Spec.CertConfig)
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
			Zones: []Zone{{"single", int32(servers)}},
		},
	}
	mt.EnsureDefaults()

	t.Run("StatefulSet", func(t *testing.T) {
		hosts := mt.TemplatedMinIOHosts("{{.StatefulSet}}")
		assert.Contains(t, hosts, mt.MinIOStatefulSetName())
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
		assert.Contains(t, hosts, ellipsis(0, servers-1))
	})

	t.Run("Domain", func(t *testing.T) {
		hosts := mt.TemplatedMinIOHosts("{{.Domain}}")
		assert.Contains(t, hosts, ClusterDomain)
	})
}
