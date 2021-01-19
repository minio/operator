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
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fatih/color"

	"github.com/minio/kubectl-minio/cmd/helpers"
	"github.com/minio/kubectl-minio/cmd/resources"
	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/runtime"
)

const (
	operatorProxyDesc = `
'proxy' command starts a port-forward with the operator UI.`
	operatorProxyExample = `  kubectl minio proxy`
)

type operatorProxyCmd struct {
	out          io.Writer
	errOut       io.Writer
	output       bool
	operatorOpts resources.OperatorOptions
	steps        []runtime.Object
}

func newProxyCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	o := &operatorProxyCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "proxy",
		Short:   "Open a port-forward to Console UI",
		Long:    operatorProxyDesc,
		Example: operatorProxyExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("this command does not accept arguments")
			}
			return o.run()
		},
	}
	cmd = helpers.DisableHelp(cmd)
	f := cmd.Flags()
	f.StringVarP(&o.operatorOpts.Image, "image", "i", helpers.DefaultOperatorImage, "operator image")
	f.StringVarP(&o.operatorOpts.Namespace, "namespace", "n", helpers.DefaultNamespace, "namespace scope for this request")
	f.StringVarP(&o.operatorOpts.ClusterDomain, "cluster-domain", "d", helpers.DefaultClusterDomain, "cluster domain of the Kubernetes cluster")
	f.StringVar(&o.operatorOpts.NSToWatch, "namespace-to-watch", "", "namespace where operator looks for MinIO tenants, leave empty for all namespaces")
	f.StringVar(&o.operatorOpts.ImagePullSecret, "image-pull-secret", "", "image pull secret to be used for pulling operator image")
	f.BoolVarP(&o.output, "output", "o", false, "dry run this command and generate requisite yaml")

	return cmd
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (o *operatorProxyCmd) run() error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// kubectl  get secret $(kubectl get serviceaccount console-sa -o jsonpath="{.secrets[0].name}") -o jsonpath="{.data.token}" | base64 --decode | pbcopy

	path, _ := rootCmd.Flags().GetString(kubeconfig)
	client, err := helpers.GetKubeClient(path)
	if err != nil {
		return err
	}

	sa, err := client.CoreV1().ServiceAccounts(o.operatorOpts.Namespace).Get(ctx, "console-sa", metav1.GetOptions{})
	if err != nil {
		return err
	}
	secretName := sa.Secrets[0].Name

	secret, err := client.CoreV1().Secrets(o.operatorOpts.Namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	var jwtToken []byte
	var ok bool

	if jwtToken, ok = secret.Data["token"]; !ok {
		return errors.New("Couldn't determine JWT to connect to console")
	}

	fmt.Println("Starting port forward of the Console UI.")
	fmt.Println("")
	fmt.Println("To connect open a browser and go to http://localhost:9090")
	fmt.Println("")
	fmt.Println("Current JWT to login:", string(jwtToken))
	fmt.Println("")

	consolePFCh := servicePortForwardPort(ctx, o.operatorOpts.Namespace, "console", "9090", color.FgGreen)

	<-consolePFCh

	return nil
}

// run the command inside a goroutine, return a channel that closes then the command dies
func servicePortForwardPort(ctx context.Context, namespace, service, port string, dcolor color.Attribute) chan interface{} {
	ch := make(chan interface{})
	go func() {
		defer close(ch)
		// service we are going to forward
		serviceName := fmt.Sprintf("service/%s", service)
		// command to run
		cmd := exec.CommandContext(ctx, "kubectl", "port-forward", "-n", namespace, serviceName, port)
		// prepare to capture the output
		var errStdout, errStderr error
		stdoutIn, _ := cmd.StdoutPipe()
		stderrIn, _ := cmd.StderrPipe()
		err := cmd.Start()
		if err != nil {
			log.Fatalf("cmd.Start() failed with '%s'\n", err)
		}

		// cmd.Wait() should be called only after we finish reading
		// from stdoutIn and stderrIn.
		// wg ensures that we finish
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			errStdout = copyAndCapture(stdoutIn, dcolor)
			wg.Done()
		}()

		errStderr = copyAndCapture(stderrIn, dcolor)

		wg.Wait()

		err = cmd.Wait()
		if err != nil {
			log.Printf("cmd.Run() failed with %s\n", err.Error())
			return
		}
		if errStdout != nil || errStderr != nil {
			log.Printf("failed to capture stdout or stderr\n")
			return
		}
		//outStr, errStr := string(stdout), string(stderr)
		//fmt.Printf("\nout:\n%s\nerr:\n%s\n", outStr, errStr)
	}()
	return ch
}

// capture and print the output of the command
func copyAndCapture(r io.Reader, dcolor color.Attribute) error {
	var out []byte
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			theColor := color.New(dcolor)
			//_, err := w.Write(d)
			_, err := theColor.Print(string(d))

			if err != nil {
				return err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return err
		}
	}
}
