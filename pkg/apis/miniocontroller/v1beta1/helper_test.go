package v1beta1

import (
	"testing"

	constants "github.com/minio/minio-operator/pkg/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureDefaults(t *testing.T) {
	mi := MinIOInstance{}
	mi.EnsureDefaults()

	t.Run("defaults", func(t *testing.T) {
		assert.Equal(t, mi.Spec.Replicas, int32(constants.DefaultReplicas))
		assert.Equal(t, mi.Spec.ClusterDomain, constants.DefaultClusterDomain)
		assert.Equal(t, mi.Spec.Image, constants.DefaultMinIOImage)
		assert.Equal(t, mi.Spec.Mountpath, constants.MinIOVolumeMountPath)
		assert.Equal(t, mi.Spec.Subpath, constants.MinIOVolumeSubPath)
		assert.False(t, mi.RequiresAutoCertSetup())
	})

	t.Run("auto cert", func(t *testing.T) {
		mi.Spec.RequestAutoCert = true
		assert.True(t, mi.RequiresAutoCertSetup())
		assert.False(t, mi.HasCertConfig())

		mi.EnsureDefaults()

		require.NotNil(t, mi.Spec.CertConfig)
		require.True(t, mi.HasCertConfig())
		oldCertConfig := mi.Spec.CertConfig

		mi.EnsureDefaults()

		assert.Equal(t, oldCertConfig, mi.Spec.CertConfig)
	})

	t.Run("defaults don't override", func(t *testing.T) {
		newClusterDomain := "k8s.example.com"
		newImage := "minio/minio:latest"
		newReplicas := int32(99)
		mi.Spec.ClusterDomain = newClusterDomain
		mi.Spec.Image = newImage
		mi.Spec.Replicas = newReplicas

		mi.EnsureDefaults()

		assert.Equal(t, newClusterDomain, mi.Spec.ClusterDomain)
		assert.Equal(t, newImage, mi.Spec.Image)
		assert.Equal(t, newReplicas, mi.Spec.Replicas)
	})
}
