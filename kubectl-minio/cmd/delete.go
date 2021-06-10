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
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/klog/v2"

	rbacv1 "k8s.io/api/rbac/v1"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"

	"github.com/minio/kubectl-minio/cmd/helpers"
	"github.com/minio/kubectl-minio/cmd/resources"
	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/runtime"
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
	steps        []runtime.Object
}

func newDeleteCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	o := &deleteCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete MinIO Operator",
		Long:    deleteDesc,
		Example: deleteExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !helpers.Ask(fmt.Sprintf("Are you sure you want to delete ALL the MinIO Tenants and MinIO Operator?")) {
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
	kustYaml, _ := yaml.Marshal(kustomizationYaml)
	kustFile, err := inMemSys.Create("kustomization.yaml")
	_, err = kustFile.Write(kustYaml)
	if err != nil {
		log.Println(err)
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
		//done
		return nil
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

func deleteConsoleResources(opts resources.OperatorOptions, clientset *apiextension.Clientset, dynClient dynamic.Interface, consoleResources []runtime.Object) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	groupResources, err := restmapper.GetAPIGroupResources(clientset.Discovery())
	if err != nil {
		klog.Info(err)
		return errors.New("Cannot get group resources")
	}
	rm := restmapper.NewDiscoveryRESTMapper(groupResources)

	for _, obj := range consoleResources {
		gvk := obj.GetObjectKind().GroupVersionKind()
		gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}

		mapping, err := rm.RESTMapping(gk, gvk.Version)

		// convert the runtime.Object to unstructured.Unstructured
		unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return err
		}
		var resourceName string
		if metaobj, ok := unstructuredObj["metadata"]; ok {
			mtobj := metaobj.(map[string]interface{})
			if name, ok2 := mtobj["name"]; ok2 {
				resourceName = name.(string)
			}
		}

		switch obj.(type) {
		case *rbacv1.ClusterRoleBinding:
			if err := clusterScopeDelete(dynClient, mapping, ctx, resourceName); err != nil {
				return err
			}
		case *rbacv1.ClusterRole:
			if err := clusterScopeDelete(dynClient, mapping, ctx, resourceName); err != nil {
				return err
			}
		default:
			if err := namespaceScopeDelete(opts, dynClient, mapping, ctx, resourceName); err != nil {
				return err
			}
		}
	}
	return nil
}

func clusterScopeDelete(dynClient dynamic.Interface, mapping *meta.RESTMapping, ctx context.Context, name string) error {
	if err := dynClient.Resource(mapping.Resource).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}

func namespaceScopeDelete(opts resources.OperatorOptions, dynClient dynamic.Interface, mapping *meta.RESTMapping, ctx context.Context, name string) error {
	if err := dynClient.Resource(mapping.Resource).Namespace(opts.Namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}
