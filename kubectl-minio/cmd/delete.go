/*
 * This file is part of MinIO Operator
 * Copyright (C) 2020, MinIO, Inc.
 *
 * This code is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, version 3,
 * as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License, version 3,
 * along with this program.  If not, see <http://www.gnu.org/licenses/>
 *
 */

package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"

	"k8s.io/klog/v2"

	"github.com/minio/kubectl-minio/cmd/helpers"
	"github.com/minio/kubectl-minio/cmd/resources"
	"github.com/spf13/cobra"
)

const (
	deleteDesc = `
'delete' command delete MinIO Operator along with all the tenants.`
	deleteExample = `  kubectl minio delete`
)

type deleteCmd struct {
	out          io.Writer
	errOut       io.Writer
	output       bool
	operatorOpts resources.OperatorOptions
}

func newDeleteCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	o := &deleteCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete MinIO Operator",
		Long:    deleteDesc,
		Example: deleteExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !helpers.Ask("Are you sure you want to delete ALL the MinIO Tenants and MinIO Operator?") {
				return fmt.Errorf(Bold("Aborting Operator deletion\n"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("delete command does not accept arguments")
			}
			klog.Info("delete command started")
			err := o.run(out)
			if err != nil {
				klog.Warning(err)
				return err
			}
			return nil
		},
	}
	cmd = helpers.DisableHelp(cmd)
	f := cmd.Flags()
	f.StringVarP(&o.operatorOpts.Namespace, "namespace", "n", helpers.DefaultNamespace, "namespace scope for this request")
	return cmd
}

func (o *deleteCmd) run(writer io.Writer) error {
	inMemSys, err := resources.GetResourceFileSys()
	if err != nil {
		return err
	}

	// write the kustomization file

	kustomizationYaml := types.Kustomization{
		TypeMeta: types.TypeMeta{
			Kind:       "Kustomization",
			APIVersion: "kustomize.config.k8s.io/v1beta1",
		},
		Resources: []string{
			"operator/",
		},
	}

	if o.operatorOpts.Namespace != "" {
		kustomizationYaml.Namespace = o.operatorOpts.Namespace
	}

	// Compile the kustomization to a file and create on the in memory filesystem
	kustYaml, err := yaml.Marshal(kustomizationYaml)
	if err != nil {
		return err
	}

	kustFile, err := inMemSys.Create("kustomization.yaml")
	if err != nil {
		return err
	}

	_, err = kustFile.Write(kustYaml)
	if err != nil {
		return err
	}

	// kustomize build the target location
	k := krusty.MakeKustomizer(
		krusty.MakeDefaultOptions(),
	)

	m, err := k.Run(inMemSys, ".")
	if err != nil {
		return err
	}

	yml, err := m.AsYaml()
	if err != nil {
		return err
	}

	if o.output {
		_, err = writer.Write(yml)
		return err
	}

	// do kubectl apply
	cmd := exec.Command("kubectl", "delete", "-f", "-")

	cmd.Stdin = strings.NewReader(string(yml))

	stdoutReader, _ := cmd.StdoutPipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	go func() {
		for stdoutScanner.Scan() {
			fmt.Println(stdoutScanner.Text())
		}
	}()
	stderrReader, _ := cmd.StderrPipe()
	stderrScanner := bufio.NewScanner(stderrReader)
	go func() {
		for stderrScanner.Scan() {
			fmt.Println(stderrScanner.Text())
		}
	}()
	err = cmd.Start()
	if err != nil {
		fmt.Printf("Error : %v \n", err)
		os.Exit(1)
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("Error: %v \n", err)
		os.Exit(1)
	}

	return nil
}
