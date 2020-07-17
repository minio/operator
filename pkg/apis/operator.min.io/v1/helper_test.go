package v1

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEnsureDefaults(t *testing.T) {
	mi := MinIOInstance{}
	mi.EnsureDefaults()

	t.Run("defaults", func(t *testing.T) {
		assert.Equal(t, mi.Spec.Image, DefaultMinIOImage)
		assert.Equal(t, mi.Spec.Mountpath, MinIOVolumeMountPath)
		assert.Equal(t, mi.Spec.Subpath, MinIOVolumeSubPath)
		assert.False(t, mi.AutoCert())
	})

	t.Run("auto cert", func(t *testing.T) {
		mi.Spec.RequestAutoCert = true
		assert.True(t, mi.AutoCert())
		assert.False(t, mi.HasCertConfig())

		mi.EnsureDefaults()

		require.NotNil(t, mi.Spec.CertConfig)
		require.True(t, mi.HasCertConfig())
		oldCertConfig := mi.Spec.CertConfig

		mi.EnsureDefaults()

		assert.Equal(t, oldCertConfig, mi.Spec.CertConfig)
	})

	t.Run("defaults don't override", func(t *testing.T) {
		newImage := "minio/minio:latest"
		mi.Spec.Image = newImage
		mi.EnsureDefaults()

		assert.Equal(t, newImage, mi.Spec.Image)
	})
}

func TestTemplateVariables(t *testing.T) {
	servers := 2
	mi := MinIOInstance{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: MinIOInstanceSpec{
			Zones: "single:" + strconv.Itoa(servers),
		},
	}
	mi.EnsureDefaults()

	t.Run("StatefulSet", func(t *testing.T) {
		hosts := mi.TemplatedMinIOHosts("{{.StatefulSet}}")
		assert.Contains(t, hosts, mi.MinIOStatefulSetName())
	})

	t.Run("CIService", func(t *testing.T) {
		hosts := mi.TemplatedMinIOHosts("{{.CIService}}")
		assert.Contains(t, hosts, mi.MinIOCIServiceName())
	})

	t.Run("HLService", func(t *testing.T) {
		hosts := mi.TemplatedMinIOHosts("{{.HLService}}")
		assert.Contains(t, hosts, mi.MinIOHLServiceName())
	})

	t.Run("Ellipsis", func(t *testing.T) {
		hosts := mi.TemplatedMinIOHosts("{{.Ellipsis}}")
		assert.Contains(t, hosts, ellipsis(0, servers-1))
	})

	t.Run("Domain", func(t *testing.T) {
		hosts := mi.TemplatedMinIOHosts("{{.Domain}}")
		assert.Contains(t, hosts, ClusterDomain)
	})
}
