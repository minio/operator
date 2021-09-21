// This file is part of MinIO Console Server
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

package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/cli/cli/config/configfile"

	"k8s.io/klog/v2"

	// Workaround for auth import issues refer https://github.com/minio/operator/issues/283
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
)

func (c *Controller) fetchTag(path string) (string, error) {
	cmd := exec.Command(path, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	op := strings.Fields(out.String())
	if len(op) != 3 {
		return "", fmt.Errorf("incorrect output while fetching tag value - %d", len((op)))
	}
	return op[2], nil
}

// minioKeychain implements Keychain to pass custom credentials
type minioKeychain struct {
	authn.Keychain
	Username      string
	Password      string
	Auth          string
	IdentityToken string
	RegistryToken string
}

// Resolve implements Keychain.
func (mk *minioKeychain) Resolve(_ authn.Resource) (authn.Authenticator, error) {
	return authn.FromConfig(authn.AuthConfig{
		Username:      mk.Username,
		Password:      mk.Password,
		Auth:          mk.Auth,
		IdentityToken: mk.IdentityToken,
		RegistryToken: mk.RegistryToken,
	}), nil
}

// getKeychainForTenant attempts to build a new authn.Keychain from the image pull secret on the Tenant
func (c *Controller) getKeychainForTenant(ctx context.Context, ref name.Reference, tenant *miniov2.Tenant) (authn.Keychain, error) {
	// Get the secret
	secret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, tenant.Spec.ImagePullSecret.Name, metav1.GetOptions{})
	if err != nil {
		return authn.DefaultKeychain, errors.New("can't retrieve the tenant image pull secret")
	}
	// if we can't find .dockerconfigjson, error out
	dockerConfigJSON, ok := secret.Data[".dockerconfigjson"]
	if !ok {
		return authn.DefaultKeychain, fmt.Errorf("unable to find `.dockerconfigjson` in image pull secret")
	}
	var config configfile.ConfigFile
	if err = json.Unmarshal(dockerConfigJSON, &config); err != nil {
		return authn.DefaultKeychain, fmt.Errorf("Unable to decode docker config secrets %w", err)
	}
	cfg, ok := config.AuthConfigs[ref.Context().RegistryStr()]
	if !ok {
		return authn.DefaultKeychain, fmt.Errorf("unable to locate auth config registry context %s", ref.Context().RegistryStr())
	}
	return &minioKeychain{
		Username:      cfg.Username,
		Password:      cfg.Password,
		Auth:          cfg.Auth,
		IdentityToken: cfg.IdentityToken,
		RegistryToken: cfg.RegistryToken,
	}, nil
}

// Attempts to fetch given image and then extracts and keeps relevant files
// (minio, minio.sha256sum & minio.minisig) at a pre-defined location (/tmp/webhook/v1/update)
func (c *Controller) fetchArtifacts(tenant *miniov2.Tenant) (latest time.Time, err error) {
	basePath := updatePath

	if err = os.MkdirAll(basePath, 1777); err != nil {
		return latest, err
	}

	ref, err := name.ParseReference(tenant.Spec.Image)
	if err != nil {
		return latest, err
	}

	keychain := authn.DefaultKeychain

	// if the tenant has imagePullSecret use that for pulling the image, but if we fail to extract the secret or we
	// can't find the expected registry in the secret we will continue with the default keychain. This is because the
	// needed pull secret could be attached to the service-account.
	if tenant.Spec.ImagePullSecret.Name != "" {
		// Get the secret
		keychain, err = c.getKeychainForTenant(context.Background(), ref, tenant)
		if err != nil {
			klog.Info(err)
		}
	}

	img, err := remote.Image(ref, remote.WithAuthFromKeychain(keychain))
	if err != nil {
		return latest, err
	}

	ls, err := img.Layers()
	if err != nil {
		return latest, err
	}

	// Find the file with largest size among all layers.
	// This is the tar file with all minio relevant files.
	start := 0
	if len(ls) >= 2 { // skip the base layer
		start = 1
	}
	maxSizeHash, _ := ls[start].Digest()
	maxSize, _ := ls[start].Size()
	for i := range ls {
		if i < start {
			continue
		}
		s, _ := ls[i].Size()
		if s > maxSize {
			maxSize, _ = ls[i].Size()
			maxSizeHash, _ = ls[i].Digest()
		}
	}

	f, err := os.OpenFile(basePath+"image.tar", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return latest, err
	}
	defer func() {
		_ = f.Close()
	}()

	// Tarball writes a file called image.tar
	// This file in turn has each container layer present inside in the form `<layer-hash>.tar.gz`
	if err = tarball.Write(ref, img, f); err != nil {
		return latest, err
	}

	// Extract the <layer-hash>.tar.gz file that has minio contents from `image.tar`
	fileNameToExtract := strings.Split(maxSizeHash.String(), ":")[1] + ".tar.gz"
	if err = miniov2.ExtractTar([]string{fileNameToExtract}, basePath, "image.tar"); err != nil {
		return latest, err
	}

	latestAssets := []string{"opt/bin/minio", "opt/bin/minio.sha256sum", "opt/bin/minio.minisig"}
	legacyAssets := []string{"usr/bin/minio", "usr/bin/minio.sha256sum", "usr/bin/minio.minisig"}

	// Extract the minio update related files (minio, minio.sha256sum and minio.minisig) from `<layer-hash>.tar.gz`
	if err = miniov2.ExtractTar(latestAssets, basePath, fileNameToExtract); err != nil {
		// attempt legacy if latest failed to extract artifacts
		if err = miniov2.ExtractTar(legacyAssets, basePath, fileNameToExtract); err != nil {
			return latest, err
		}
	}

	srcBinary := "minio"
	srcShaSum := "minio.sha256sum"
	srcSig := "minio.minisig"

	tag, err := c.fetchTag(basePath + srcBinary)
	if err != nil {
		return latest, err
	}

	latest, err = miniov2.ReleaseTagToReleaseTime(tag)
	if err != nil {
		return latest, err
	}

	destBinary := "minio." + tag
	destShaSum := "minio." + tag + ".sha256sum"
	destSig := "minio." + tag + ".minisig"
	filesToRename := map[string]string{srcBinary: destBinary, srcShaSum: destShaSum, srcSig: destSig}

	// rename all files to add tag specific values in the name.
	// this is because minio updater looks for files in this name format.
	for s, d := range filesToRename {
		if err = os.Rename(basePath+s, basePath+d); err != nil {
			return latest, err
		}
	}
	return latest, nil
}

// Remove all the files created during upload process
func (c *Controller) removeArtifacts() error {
	return os.RemoveAll(updatePath)
}
