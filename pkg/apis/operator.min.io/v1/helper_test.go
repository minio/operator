package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureDefaults(t *testing.T) {
	mi := MinIOInstance{}
	mi.EnsureDefaults()

	t.Run("defaults", func(t *testing.T) {
		assert.Equal(t, mi.Spec.Image, DefaultMinIOImage)
		assert.Equal(t, mi.Spec.Mountpath, MinIOVolumeMountPath)
		assert.Equal(t, mi.Spec.Subpath, MinIOVolumeSubPath)
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
		newImage := "minio/minio:latest"
		mi.Spec.Image = newImage
		mi.EnsureDefaults()

		assert.Equal(t, newImage, mi.Spec.Image)
	})
}
