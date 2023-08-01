// This file is part of MinIO Operator
// Copyright (C) 2023, MinIO, Inc.
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

package cmd

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/kubectl-minio/cmd/helpers"
	v2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	getCredentialsDesc = `
'get-credentials' command get credentials from MinIO tenant`
	getCredentialsExample = ` kubectl minio tenant get-credentials tenantName`
)

type getCredentialsCmd struct {
	out    io.Writer
	errOut io.Writer
	output bool
	name   string
}

func newTenantGetCredentialsCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	_ = v2.AddToScheme(scheme.Scheme)
	v := &getCredentialsCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "get-credentials <TENANTNAME>",
		Short:   "get credentials from existing tenant",
		Long:    getCredentialsDesc,
		Example: getCredentialsExample,
		Args: func(cmd *cobra.Command, args []string) error {
			return v.validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := v.run()
			if err != nil {
				klog.Warning(err)
				return err
			}
			return nil
		},
	}
	cmd = helpers.DisableHelp(cmd)
	return cmd
}

func (v *getCredentialsCmd) validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("provide the name of the tenant, e.g. 'kubectl minio tenant get-credentials tenantName'")
	}
	// Tenant name should have DNS token restrictions
	if err := helpers.CheckValidTenantName(args[0]); err != nil {
		return err
	}
	v.name = args[0]
	return nil
}

// run initializes local config
// get the tenant first
// get the secret from tenant
func (v *getCredentialsCmd) run() error {
	cfg := config.GetConfigOrDie()
	// If config is passed as a flag use that instead
	mcli, err := client.New(cfg, client.Options{})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	tenantList := &v2.TenantList{}
	err = mcli.List(ctx, tenantList)
	if err != nil {
		return err
	}
	for _, tenant := range tenantList.Items {
		if tenant.Name == v.name {
			if tenant.Spec.Configuration != nil {
				secret := &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      tenant.Spec.Configuration.Name,
						Namespace: tenant.Namespace,
					},
				}
				err = mcli.Get(ctx, client.ObjectKeyFromObject(secret), secret)
				if err != nil {
					return err
				}
				fmt.Println(string(secret.Data["config.env"]))
			}
		}
	}
	return nil
}
