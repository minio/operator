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
//

package api

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/minio/cli"
)

var (
	infoLog  = log.New(os.Stdout, "I: ", log.LstdFlags)
	errorLog = log.New(os.Stdout, "E: ", log.LstdFlags)
)

func logInfo(msg string, data ...interface{}) {
	infoLog.Printf(msg+"\n", data...)
}

func logError(msg string, data ...interface{}) {
	errorLog.Printf(msg+"\n", data...)
}

func logIf(ctx context.Context, err error, errKind ...interface{}) {
}

// globally changeable logger styles
var (
	LogInfo  = logInfo
	LogError = logError
	LogIf    = logIf
)

// Context captures all command line flags values
type Context struct {
	Host                string
	HTTPPort, HTTPSPort int
	TLSRedirect         string
	// Legacy options, TODO: remove in future
	TLSCertificate, TLSKey, TLSca string
}

// Load loads restapi Context from command line context.
func (c *Context) Load(ctx *cli.Context) error {
	*c = Context{
		Host:        ctx.String("host"),
		HTTPPort:    ctx.Int("port"),
		HTTPSPort:   ctx.Int("tls-port"),
		TLSRedirect: ctx.String("tls-redirect"),
		// Legacy options to be removed.
		TLSCertificate: ctx.String("tls-certificate"),
		TLSKey:         ctx.String("tls-key"),
		TLSca:          ctx.String("tls-ca"),
	}
	if c.HTTPPort > 65535 {
		return errors.New("invalid argument --port out of range - ports can range from 1-65535")
	}
	if c.HTTPSPort > 65535 {
		return errors.New("invalid argument --tls-port out of range - ports can range from 1-65535")
	}
	if c.TLSRedirect != "on" && c.TLSRedirect != "off" {
		return errors.New("invalid argument --tls-redirect only accepts either 'on' or 'off'")
	}
	return nil
}
