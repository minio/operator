// This file is part of MinIO Operator
// Copyright (C) 2020, MinIO, Inc.
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
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"github.com/fatih/color"

	"github.com/minio/kubectl-minio/cmd/helpers"
	"github.com/minio/kubectl-minio/cmd/resources"
	"github.com/spf13/cobra"
)

const (
	operatorProxyDesc = `
'proxy' command starts a port-forward with the operator UI.`
	operatorProxyExample = `  kubectl minio proxy`
)

type operatorProxyCmd struct {
	out          io.Writer
	errOut       io.Writer
	operatorOpts resources.OperatorOptions
}

func newProxyCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	o := &operatorProxyCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "proxy",
		Short:   "Open a port-forward to Console UI",
		Long:    operatorProxyDesc,
		Example: operatorProxyExample,
		Args:    cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.run()
			if err != nil {
				klog.Warning(err)
				return err
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVarP(&o.operatorOpts.Namespace, "namespace", "n", helpers.DefaultNamespace, "namespace scope for this request")

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

	secretName := ""

	// Openshift doesn't create the token with the name "console-sa-secret" instead it creates a "console-sa-token-{random id}" secret
	// This section is to  find that token and get the actual secret containing the JWT token to use and authenticate
	secrets, err := client.CoreV1().Secrets(o.operatorOpts.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, secret := range secrets.Items {
		if strings.HasPrefix(secret.Name, "console-sa-token") {
			secretName = secret.Name
		}
	}

	// If no secret was found previously, is a more vanilla kubernetes setup, here we try to find the secret containing the sa token
	if secretName == "" {
		secretName = "console-sa-secret"
		if len(sa.Secrets) > 0 {
			secretName = sa.Secrets[0].Name
		}
	}

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
		path, _ := rootCmd.Flags().GetString(kubeconfig)
		parameters := []string{"port-forward", "--address", "0.0.0.0", "-n", namespace, serviceName, port}
		if path != "" {
			parameters = append([]string{"--kubeconfig", path}, parameters...)
		}
		// command to run
		cmd := exec.CommandContext(ctx, "kubectl", parameters...)
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
		// outStr, errStr := string(stdout), string(stderr)
		// fmt.Printf("\nout:\n%s\nerr:\n%s\n", outStr, errStr)
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
