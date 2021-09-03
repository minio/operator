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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/kustomize/api/resid"

	"sigs.k8s.io/yaml"

	"sigs.k8s.io/kustomize/api/types"

	"k8s.io/klog/v2"

	"github.com/minio/kubectl-minio/cmd/helpers"
	"github.com/minio/kubectl-minio/cmd/resources"
	"github.com/spf13/cobra"

	"sigs.k8s.io/kustomize/api/krusty"
)

const (
	operatorInitDesc = `
'init' command creates MinIO Operator deployment along with all the dependencies.`
	operatorInitExample = `  kubectl minio init`
)

type operatorInitCmd struct {
	out          io.Writer
	errOut       io.Writer
	output       bool
	operatorOpts resources.OperatorOptions
}

func newInitCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	o := &operatorInitCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "init",
		Short:   "Initialize MinIO Operator",
		Long:    operatorInitDesc,
		Example: operatorInitExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("this command does not accept arguments")
			}
			klog.Info("init command started")
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
	f.StringVarP(&o.operatorOpts.Image, "image", "i", helpers.DefaultOperatorImage, "operator image")
	f.StringVarP(&o.operatorOpts.Namespace, "namespace", "n", helpers.DefaultNamespace, "namespace scope for this request")
	f.StringVarP(&o.operatorOpts.ClusterDomain, "cluster-domain", "d", helpers.DefaultClusterDomain, "cluster domain of the Kubernetes cluster")
	f.StringVar(&o.operatorOpts.NSToWatch, "namespace-to-watch", "", "namespace where operator looks for MinIO tenants, leave empty for all namespaces")
	f.StringVar(&o.operatorOpts.ImagePullSecret, "image-pull-secret", "", "image pull secret to be used for pulling operator image")
	f.StringVar(&o.operatorOpts.ConsoleImage, "console-image", "", "console image")
	f.StringVar(&o.operatorOpts.TenantMinIOImage, "default-minio-image", "", "default tenant MinIO image")
	f.StringVar(&o.operatorOpts.TenantConsoleImage, "default-console-image", "", "default tenant Console image")
	f.StringVar(&o.operatorOpts.TenantKesImage, "default-kes-image", "", "default tenant KES image")
	f.BoolVarP(&o.output, "output", "o", false, "dry run this command and generate requisite yaml")

	return cmd
}

type opStr struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

type opInterface struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (o *operatorInitCmd) run(writer io.Writer) error {
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
		PatchesJson6902: []types.Patch{},
	}

	var operatorDepPatches []interface{}
	// create patches for the supplied arguments
	if o.operatorOpts.Image != "" {
		operatorDepPatches = append(operatorDepPatches, opStr{
			Op:    "replace",
			Path:  "/spec/template/spec/containers/0/image",
			Value: o.operatorOpts.Image,
		})
	}
	// create an empty array
	operatorDepPatches = append(operatorDepPatches, opInterface{
		Op:    "add",
		Path:  "/spec/template/spec/containers/0/env",
		Value: []interface{}{},
	})

	if o.operatorOpts.ClusterDomain != "" {
		operatorDepPatches = append(operatorDepPatches, opInterface{
			Op:   "add",
			Path: "/spec/template/spec/containers/0/env/0",
			Value: corev1.EnvVar{
				Name:  "CLUSTER_DOMAIN",
				Value: o.operatorOpts.ClusterDomain,
			},
		})
	}
	if o.operatorOpts.NSToWatch != "" {
		operatorDepPatches = append(operatorDepPatches, opInterface{
			Op:   "add",
			Path: "/spec/template/spec/containers/0/env/0",
			Value: corev1.EnvVar{
				Name:  "WATCHED_NAMESPACE",
				Value: o.operatorOpts.NSToWatch,
			},
		})
	}
	if o.operatorOpts.TenantMinIOImage != "" {
		operatorDepPatches = append(operatorDepPatches, opInterface{
			Op:   "add",
			Path: "/spec/template/spec/containers/0/env/0",
			Value: corev1.EnvVar{
				Name:  "TENANT_MINIO_IMAGE",
				Value: o.operatorOpts.TenantMinIOImage,
			},
		})
	}
	if o.operatorOpts.TenantConsoleImage != "" {
		operatorDepPatches = append(operatorDepPatches, opInterface{
			Op:   "add",
			Path: "/spec/template/spec/containers/0/env/0",
			Value: corev1.EnvVar{
				Name:  "TENANT_CONSOLE_IMAGE",
				Value: o.operatorOpts.TenantConsoleImage,
			},
		})
	}
	if o.operatorOpts.TenantKesImage != "" {
		operatorDepPatches = append(operatorDepPatches, opInterface{
			Op:   "add",
			Path: "/spec/template/spec/containers/0/env/0",
			Value: corev1.EnvVar{
				Name:  "TENANT_KES_IMAGE",
				Value: o.operatorOpts.TenantKesImage,
			},
		})
	}
	if o.operatorOpts.ImagePullSecret != "" {
		operatorDepPatches = append(operatorDepPatches, opInterface{
			Op:    "add",
			Path:  "/spec/template/spec/imagePullSecrets",
			Value: []corev1.LocalObjectReference{{Name: o.operatorOpts.ImagePullSecret}},
		})
	}
	// attach the patches to the kustomization file
	if len(operatorDepPatches) > 0 {
		kustomizationYaml.PatchesJson6902 = append(kustomizationYaml.PatchesJson6902, types.Patch{
			Patch: o.serializeJSONPachOps(operatorDepPatches),
			Target: &types.Selector{
				Gvk: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				},
				Name: "minio-operator",
			},
		})
	}

	if o.operatorOpts.ConsoleImage != "" {

		kustomizationYaml.PatchesJson6902 = append(kustomizationYaml.PatchesJson6902, types.Patch{
			Patch: o.serializeJSONPachOps([]interface{}{
				opStr{
					Op:    "replace",
					Path:  "/spec/template/spec/containers/0/image",
					Value: o.operatorOpts.ConsoleImage,
				},
			}),
			Target: &types.Selector{
				Gvk: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				},
				Name: "console",
			},
		})
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
	cmd := exec.Command("kubectl", "apply", "-f", "-")

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

	// since we did an explicit deployment of resources, let's show a message telling users how to connect to console
	fmt.Println("-----------------")
	fmt.Println("")
	fmt.Println("To open Operator UI, start a port forward using this command:")
	fmt.Println("")
	if o.operatorOpts.Namespace != "" {
		fmt.Printf("kubectl minio proxy -n %s \n", o.operatorOpts.Namespace)
	} else {
		fmt.Println("kubectl minio proxy")
	}

	fmt.Println("")
	fmt.Println("-----------------")

	return nil
}

func (o *operatorInitCmd) serializeJSONPachOps(jp []interface{}) string {
	jpJSON, _ := json.Marshal(jp)
	return string(jpJSON)
}
