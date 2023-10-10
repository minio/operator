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

package certs

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/cli"
	xcerts "github.com/minio/pkg/certs"
	"github.com/minio/pkg/env"
	"github.com/mitchellh/go-homedir"
)

// ConfigDir - points to a user set directory.
type ConfigDir struct {
	Path string
}

// Get - returns current directory.
func (dir *ConfigDir) Get() string {
	return dir.Path
}

func getDefaultConfigDir() string {
	homeDir, err := homedir.Dir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, DefaultConsoleConfigDir)
}

func getDefaultCertsDir() string {
	return filepath.Join(getDefaultConfigDir(), CertsDir)
}

func getDefaultCertsCADir() string {
	return filepath.Join(getDefaultCertsDir(), CertsCADir)
}

// isFile - returns whether given Path is a file or not.
func isFile(path string) bool {
	if fi, err := os.Stat(path); err == nil {
		return fi.Mode().IsRegular()
	}

	return false
}

var (
	// DefaultCertsDir certs directory.
	DefaultCertsDir = &ConfigDir{Path: getDefaultCertsDir()}
	// DefaultCertsCADir CA directory.
	DefaultCertsCADir = &ConfigDir{Path: getDefaultCertsCADir()}
	// GlobalCertsDir points to current certs directory set by user with --certs-dir
	GlobalCertsDir = DefaultCertsDir
	// GlobalCertsCADir points to relative Path to certs directory and is <value-of-certs-dir>/CAs
	GlobalCertsCADir = DefaultCertsCADir
)

// ParsePublicCertFile - parses public cert into its *x509.Certificate equivalent.
func ParsePublicCertFile(certFile string) (x509Certs []*x509.Certificate, err error) {
	// Read certificate file.
	var data []byte
	if data, err = ioutil.ReadFile(certFile); err != nil {
		return nil, err
	}

	// Trimming leading and tailing white spaces.
	data = bytes.TrimSpace(data)

	// Parse all certs in the chain.
	current := data
	for len(current) > 0 {
		var pemBlock *pem.Block
		if pemBlock, current = pem.Decode(current); pemBlock == nil {
			return nil, fmt.Errorf("could not read PEM block from file %s", certFile)
		}

		var x509Cert *x509.Certificate
		if x509Cert, err = x509.ParseCertificate(pemBlock.Bytes); err != nil {
			return nil, err
		}

		x509Certs = append(x509Certs, x509Cert)
	}

	if len(x509Certs) == 0 {
		return nil, fmt.Errorf("empty public certificate file %s", certFile)
	}

	return x509Certs, nil
}

// MkdirAllIgnorePerm attempts to create all directories, ignores any permission denied errors.
func MkdirAllIgnorePerm(path string) error {
	err := os.MkdirAll(path, 0o700)
	if err != nil {
		// It is possible in kubernetes like deployments this directory
		// is already mounted and is not writable, ignore any write errors.
		if os.IsPermission(err) {
			err = nil
		}
	}
	return err
}

// NewConfigDirFromCtx configuration for dir of certs
func NewConfigDirFromCtx(ctx *cli.Context, option string, getDefaultDir func() string) (*ConfigDir, bool, error) {
	var dir string
	var dirSet bool

	switch {
	case ctx.IsSet(option):
		dir = ctx.String(option)
		dirSet = true
	case ctx.GlobalIsSet(option):
		dir = ctx.GlobalString(option)
		dirSet = true
		// cli package does not expose parent's option option.  Below code is workaround.
		if dir == "" || dir == getDefaultDir() {
			dirSet = false // Unset to false since GlobalIsSet() true is a false positive.
			if ctx.Parent().GlobalIsSet(option) {
				dir = ctx.Parent().GlobalString(option)
				dirSet = true
			}
		}
	default:
		// Neither local nor global option is provided.  In this case, try to use
		// default directory.
		dir = getDefaultDir()
		if dir == "" {
			return nil, false, fmt.Errorf("invalid arguments specified, %s option must be provided", option)
		}
	}

	if dir == "" {
		return nil, false, fmt.Errorf("empty directory, %s directory cannot be empty", option)
	}

	// Disallow relative paths, figure out absolute paths.
	dirAbs, err := filepath.Abs(dir)
	if err != nil {
		return nil, false, fmt.Errorf("%w: Unable to fetch absolute path for %s=%s", err, option, dir)
	}
	if err = MkdirAllIgnorePerm(dirAbs); err != nil {
		return nil, false, fmt.Errorf("%w: Unable to create directory specified %s=%s", err, option, dir)
	}
	return &ConfigDir{Path: dirAbs}, dirSet, nil
}

func getPublicCertFile() string {
	publicCertFile := filepath.Join(GlobalCertsDir.Get(), PublicCertFile)
	TLSCertFile := filepath.Join(GlobalCertsDir.Get(), TLSCertFile)
	if isFile(publicCertFile) {
		return publicCertFile
	}
	return TLSCertFile
}

func getPrivateKeyFile() string {
	privateKeyFile := filepath.Join(GlobalCertsDir.Get(), PrivateKeyFile)
	TLSPrivateKey := filepath.Join(GlobalCertsDir.Get(), TLSKeyFile)
	if isFile(privateKeyFile) {
		return privateKeyFile
	}
	return TLSPrivateKey
}

// EnvCertPassword is the environment variable which contains the password used
// to decrypt the TLS private key. It must be set if the TLS private key is
// password protected.
const EnvCertPassword = "CONSOLE_CERT_PASSWD"

// LoadX509KeyPair - load an X509 key pair (private key , certificate)
// from the provided paths. The private key may be encrypted and is
// decrypted using the ENV_VAR: MINIO_CERT_PASSWD.
func LoadX509KeyPair(certFile, keyFile string) (tls.Certificate, error) {
	certPEMBlock, err := ioutil.ReadFile(certFile)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyPEMBlock, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return tls.Certificate{}, err
	}
	key, rest := pem.Decode(keyPEMBlock)
	if len(rest) > 0 {
		return tls.Certificate{}, errors.New("the private key contains additional data")
	}
	if x509.IsEncryptedPEMBlock(key) {
		password := env.Get(EnvCertPassword, "")
		if len(password) == 0 {
			return tls.Certificate{}, errors.New("no password")
		}
		decryptedKey, decErr := x509.DecryptPEMBlock(key, []byte(password))
		if decErr != nil {
			return tls.Certificate{}, decErr
		}
		keyPEMBlock = pem.EncodeToMemory(&pem.Block{Type: key.Type, Bytes: decryptedKey})
	}
	return tls.X509KeyPair(certPEMBlock, keyPEMBlock)
}

// GetTLSConfig returns the TLS config for the server
func GetTLSConfig() (x509Certs []*x509.Certificate, manager *xcerts.Manager, err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if !(isFile(getPublicCertFile()) && isFile(getPrivateKeyFile())) {
		return nil, nil, nil
	}

	if x509Certs, err = ParsePublicCertFile(getPublicCertFile()); err != nil {
		return nil, nil, err
	}

	manager, err = xcerts.NewManager(ctx, getPublicCertFile(), getPrivateKeyFile(), LoadX509KeyPair)
	if err != nil {
		return nil, nil, err
	}

	// Console has support for multiple certificates. It expects the following structure:
	// certs/
	//  │
	//  ├─ public.crt
	//  ├─ private.key
	//  │
	//  ├─ example.com/
	//  │   │
	//  │   ├─ public.crt
	//  │   └─ private.key
	//  └─ foobar.org/
	//     │
	//     ├─ public.crt
	//     └─ private.key
	//  ...
	//
	// Therefore, we read all filenames in the cert directory and check
	// for each directory whether it contains a public.crt and private.key.
	// If so, we try to add it to certificate manager.
	root, err := os.Open(GlobalCertsDir.Get())
	if err != nil {
		return nil, nil, err
	}
	defer root.Close()

	files, err := root.Readdir(-1)
	if err != nil {
		return nil, nil, err
	}
	for _, file := range files {
		// Ignore all
		// - regular files
		// - "CAs" directory
		// - any directory which starts with ".."
		if file.Mode().IsRegular() || file.Name() == "CAs" || strings.HasPrefix(file.Name(), "..") {
			continue
		}
		if file.Mode()&os.ModeSymlink == os.ModeSymlink {
			file, err = os.Stat(filepath.Join(root.Name(), file.Name()))
			if err != nil {
				// not accessible ignore
				continue
			}
			if !file.IsDir() {
				continue
			}
		}

		var (
			certFile = filepath.Join(root.Name(), file.Name(), PublicCertFile)
			keyFile  = filepath.Join(root.Name(), file.Name(), PrivateKeyFile)
		)
		if !isFile(certFile) || !isFile(keyFile) {
			continue
		}
		if err = manager.AddCertificate(certFile, keyFile); err != nil {
			return nil, nil, fmt.Errorf("unable to load TLS certificate '%s,%s': %w", certFile, keyFile, err)
		}
	}
	return x509Certs, manager, nil
}

// GetAllCertificatesAndCAs returns all certs and cas
func GetAllCertificatesAndCAs() (*x509.CertPool, []*x509.Certificate, *xcerts.Manager, error) {
	// load all CAs from ~/.console/certs/CAs
	rootCAs, err := xcerts.GetRootCAs(GlobalCertsCADir.Get())
	if err != nil {
		return nil, nil, nil, err
	}
	// load all certs from ~/.console/certs
	publicCerts, certsManager, err := GetTLSConfig()
	if err != nil {
		return nil, nil, nil, err
	}
	if rootCAs == nil {
		rootCAs = &x509.CertPool{}
	}
	// Add the public crts as part of root CAs to trust self.
	for _, publicCrt := range publicCerts {
		rootCAs.AddCert(publicCrt)
	}
	return rootCAs, publicCerts, certsManager, nil
}

// EnsureCertAndKey checks if both client certificate and key paths are provided
func EnsureCertAndKey(clientCert, clientKey string) error {
	if (clientCert != "" && clientKey == "") ||
		(clientCert == "" && clientKey != "") {
		return errors.New("cert and key must be specified as a pair")
	}
	return nil
}
