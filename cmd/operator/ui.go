// This file is part of MinIO Operator
// Copyright (c) 2022 MinIO, Inc.
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

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/minio/operator/api"

	"github.com/minio/operator/pkg/logger"

	"github.com/go-openapi/loads"
	"github.com/jessevdk/go-flags"
	"github.com/minio/cli"
	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/pkg/certs"
)

// starts the server
var uiCmd = cli.Command{
	Name:   "ui",
	Usage:  "Start MinIO Operator UI server",
	Action: startOperatorServer,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "host",
			Value: api.GetHostname(),
			Usage: "bind to a specific HOST, HOST can be an IP or hostname",
		},
		cli.IntFlag{
			Name:  "port",
			Value: api.GetPort(),
			Usage: "bind to specific HTTP port",
		},
		// This is kept here for backward compatibility,
		// hostname's do not have HTTP or HTTPs
		// hostnames are opaque so using --host
		// works for both HTTP and HTTPS setup.
		cli.StringFlag{
			Name:   "tls-host",
			Value:  api.GetHostname(),
			Hidden: true,
		},
		cli.StringFlag{
			Name:  "certs-dir",
			Value: certs.GlobalCertsCADir.Get(),
			Usage: "path to certs directory",
		},
		cli.IntFlag{
			Name:  "tls-port",
			Value: api.GetTLSPort(),
			Usage: "bind to specific HTTPS port",
		},
		cli.StringFlag{
			Name:  "tls-redirect",
			Value: api.GetTLSRedirect(),
			Usage: "toggle HTTP->HTTPS redirect",
		},
		cli.StringFlag{
			Name:   "tls-certificate",
			Value:  "",
			Usage:  "path to TLS public certificate",
			Hidden: true,
		},
		cli.StringFlag{
			Name:   "tls-key",
			Value:  "",
			Usage:  "path to TLS private key",
			Hidden: true,
		},
		cli.StringFlag{
			Name:   "tls-ca",
			Value:  "",
			Usage:  "path to TLS Certificate Authority",
			Hidden: true,
		},
	},
}

func buildOperatorServer() (*api.Server, error) {
	swaggerSpec, err := loads.Embedded(api.SwaggerJSON, api.FlatSwaggerJSON)
	if err != nil {
		return nil, err
	}

	subPath := api.GetSubPath()
	swaggerSpec.Spec().BasePath = subPath + swaggerSpec.Spec().BasePath
	swaggerSpec.OrigSpec().BasePath = subPath + swaggerSpec.OrigSpec().BasePath

	operatorapi := operations.NewOperatorAPI(swaggerSpec)
	operatorapi.Logger = api.LogInfo
	server := api.NewServer(operatorapi)

	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = "MinIO Operator Server"
	parser.LongDescription = swaggerSpec.Spec().Info.Description

	server.ConfigureFlags()

	// register all APIs
	server.ConfigureAPI()

	for _, optsGroup := range operatorapi.CommandLineOptionsGroups {
		_, err := parser.AddGroup(optsGroup.ShortDescription, optsGroup.LongDescription, optsGroup.Options)
		if err != nil {
			return nil, err
		}
	}

	if _, err := parser.Parse(); err != nil {
		return nil, err
	}

	return server, nil
}

// StartServer starts the console service
func startOperatorServer(ctx *cli.Context) error {
	if err := loadAllCerts(ctx); err != nil {
		// Log this as a warning and continue running console without TLS certificates
		api.LogError("Unable to load certs: %v", err)
	}

	xctx := context.Background()
	transport := api.PrepareSTSClientTransport(false)
	if err := logger.InitializeLogger(xctx, transport); err != nil {
		fmt.Println("error InitializeLogger", err)
		logger.CriticalIf(xctx, err)
	}
	// custom error configuration
	api.LogInfo = logger.Info
	api.LogError = logger.Error
	api.LogIf = logger.LogIf

	var rctx api.Context
	if err := rctx.Load(ctx); err != nil {
		api.LogError("argument validation failed: %v", err)
		return err
	}

	server, err := buildOperatorServer()
	if err != nil {
		api.LogError("Unable to initialize console server: %v", err)
		return err
	}

	server.Host = rctx.Host
	server.Port = rctx.HTTPPort
	// set conservative timesout for uploads
	server.ReadTimeout = 1 * time.Hour
	// no timeouts for response for downloads
	server.WriteTimeout = 0
	api.Port = strconv.Itoa(server.Port)
	api.Hostname = server.Host

	if len(api.GlobalPublicCerts) > 0 {
		// If TLS certificates are provided enforce the HTTPS schema, meaning console will redirect
		// plain HTTP connections to HTTPS server
		server.EnabledListeners = []string{"http", "https"}
		server.TLSPort = rctx.HTTPSPort
		// Need to store tls-port, tls-host un config variables so secure.middleware can read from there
		api.TLSPort = strconv.Itoa(server.TLSPort)
		api.Hostname = rctx.Host
		api.TLSRedirect = rctx.TLSRedirect
	}

	defer server.Shutdown()

	return server.Serve()
}

func loadAllCerts(ctx *cli.Context) error {
	var err error
	// Set all certs and CAs directories path
	certs.GlobalCertsDir, _, err = certs.NewConfigDirFromCtx(ctx, "certs-dir", certs.DefaultCertsDir.Get)
	if err != nil {
		return err
	}

	certs.GlobalCertsCADir = &certs.ConfigDir{Path: filepath.Join(certs.GlobalCertsDir.Get(), certs.CertsCADir)}
	// check if certs and CAs directories exists or can be created
	if err = certs.MkdirAllIgnorePerm(certs.GlobalCertsCADir.Get()); err != nil {
		return fmt.Errorf("unable to create certs CA directory at %s: failed with %w", certs.GlobalCertsCADir.Get(), err)
	}

	// load the certificates and the CAs
	api.GlobalRootCAs, api.GlobalPublicCerts, api.GlobalTLSCertsManager, err = certs.GetAllCertificatesAndCAs()
	if err != nil {
		return fmt.Errorf("unable to load certificates at %s: failed with %w", certs.GlobalCertsDir.Get(), err)
	}

	{
		// TLS flags from swagger server, used to support VMware vsphere operator version.
		swaggerServerCertificate := ctx.String("tls-certificate")
		swaggerServerCertificateKey := ctx.String("tls-key")
		swaggerServerCACertificate := ctx.String("tls-ca")
		// load tls cert and key from swagger server tls-certificate and tls-key flags
		if swaggerServerCertificate != "" && swaggerServerCertificateKey != "" {
			if err = api.GlobalTLSCertsManager.AddCertificate(swaggerServerCertificate, swaggerServerCertificateKey); err != nil {
				return err
			}
			x509Certs, err := certs.ParsePublicCertFile(swaggerServerCertificate)
			if err == nil {
				api.GlobalPublicCerts = append(api.GlobalPublicCerts, x509Certs...)
			}
		}

		// load ca cert from swagger server tls-ca flag
		if swaggerServerCACertificate != "" {
			caCert, caCertErr := ioutil.ReadFile(swaggerServerCACertificate)
			if caCertErr == nil {
				api.GlobalRootCAs.AppendCertsFromPEM(caCert)
			}
		}
	}

	if api.GlobalTLSCertsManager != nil {
		api.GlobalTLSCertsManager.ReloadOnSignal(syscall.SIGHUP)
	}

	return nil
}
