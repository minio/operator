// Copyright (C) 2020 MinIO, Inc.
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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/minio/minio-go/v7/pkg/set"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/klog/v2"

	clientset "github.com/minio/operator/pkg/client/clientset/versioned"
	informers "github.com/minio/operator/pkg/client/informers/externalversions"
	"github.com/minio/operator/pkg/controller/cluster"
	promclientset "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

const (
	// OperatorWatchedNamespaceEnv Env variable name, the namespaces which the operator watches for MinIO tenants. Defaults to "" for all namespaces.
	OperatorWatchedNamespaceEnv = "WATCHED_NAMESPACE"
	// HostnameEnv Host name env variable
	HostnameEnv = "HOSTNAME"
)

// version provides the version of this operator
var version = "DEVELOPMENT.GOGET"

var (
	masterURL     string
	kubeconfig    string
	hostsTemplate string
	checkVersion  bool

	onlyOneSignalHandler = make(chan struct{})
	shutdownSignals      = []os.Signal{os.Interrupt, syscall.SIGTERM}
)

func init() {
	klog.InitFlags(nil)
	klog.LogToStderr(true)
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to a kubeconfig. Only required if out-of-cluster")
	flag.StringVar(&masterURL, "master", "", "the address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster")
	flag.StringVar(&hostsTemplate, "hosts-template", "", "the go template to use for hostname formatting of name fields (StatefulSet, CIService, HLService, Ellipsis, Domain)")
	flag.BoolVar(&checkVersion, "version", false, "print version")
}

func main() {
	klog.Info("Starting MinIO Operator")
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := setupSignalHandler()

	flag.Parse()

	if checkVersion {
		fmt.Println(version)
		return
	}

	// Look for incluster config by default
	cfg, err := rest.InClusterConfig()
	// If config is passed as a flag use that instead
	if kubeconfig != "" {
		cfg, err = clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	}

	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building Kubernetes clientset: %s", err.Error())
	}

	controllerClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building MinIO clientset: %s", err.Error())
	}

	extClient, err := apiextension.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("Error building certificate clientset: %v", err.Error())
	}

	promClient, err := promclientset.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("Error building Prometheus clientset: %v", err.Error())
	}

	// Get a comma separated list of namespaces to watch
	namespacesENv, isNamespaced := os.LookupEnv(OperatorWatchedNamespaceEnv)
	var namespaces set.StringSet
	if isNamespaced {
		namespaces = set.NewStringSet()
		rawNamespaces := strings.Split(namespacesENv, ",")
		for _, nsStr := range rawNamespaces {
			if nsStr != "" {
				namespaces.Add(strings.TrimSpace(nsStr))
			}
		}
		klog.Infof("Watching only namespaces: %s", strings.Join(namespaces.ToSlice(), ","))
	}

	ctx := context.Background()

	// Default kubernetes CA certificate
	caContent := miniov2.GetPodCAFromFile()

	// If ca.crt exists in operator-tls secret load that too, ie: if the cert was issued by cert-manager=
	operatorTLSCert, err := kubeClient.CoreV1().Secrets(miniov2.GetNSFromFile()).Get(context.Background(), cluster.OperatorTLSSecretName, metav1.GetOptions{})
	if err == nil && operatorTLSCert != nil {
		if val, ok := operatorTLSCert.Data["public.crt"]; ok {
			caContent = append(caContent, val...)
		}
		if val, ok := operatorTLSCert.Data["tls.crt"]; ok {
			caContent = append(caContent, val...)
		}
		if val, ok := operatorTLSCert.Data["ca.crt"]; ok {
			caContent = append(caContent, val...)
		}
	}

	// custom ca certificate to be used by operator
	operatorCATLSCert, err := kubeClient.CoreV1().Secrets(miniov2.GetNSFromFile()).Get(ctx, cluster.OperatorCATLSSecretName, metav1.GetOptions{})
	if err == nil && operatorCATLSCert != nil {
		if val, ok := operatorCATLSCert.Data["public.crt"]; ok {
			caContent = append(caContent, val...)
		}
		if val, ok := operatorCATLSCert.Data["tls.crt"]; ok {
			caContent = append(caContent, val...)
		}
		if val, ok := operatorCATLSCert.Data["ca.crt"]; ok {
			caContent = append(caContent, val...)
		}
	}

	// certificate for Operator STS, we need tenants to also trust the Operator STS
	stsCert, err := kubeClient.CoreV1().Secrets(miniov2.GetNSFromFile()).Get(ctx, cluster.STSTLSSecretName, metav1.GetOptions{})
	if err == nil && stsCert != nil {
		if cert, ok := stsCert.Data["public.crt"]; ok {
			caContent = append(caContent, cert...)
		}
		if val, ok := operatorTLSCert.Data["tls.crt"]; ok {
			caContent = append(caContent, val...)
		}
		if val, ok := operatorTLSCert.Data["ca.crt"]; ok {
			caContent = append(caContent, val...)
		}
	}

	if len(caContent) > 0 {
		crd, err := extClient.ApiextensionsV1().CustomResourceDefinitions().Get(context.Background(), "tenants.minio.min.io", metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Error getting CRD for adding caBundle: %v", err.Error())
		} else {
			crd.Spec.Conversion.Webhook.ClientConfig.CABundle = caContent
			crd.Spec.Conversion.Webhook.ClientConfig.Service.Namespace = miniov2.GetNSFromFile()
			_, err := extClient.ApiextensionsV1().CustomResourceDefinitions().Update(context.Background(), crd, metav1.UpdateOptions{})
			if err != nil {
				klog.Errorf("Error updating CRD with caBundle: %v", err.Error())
			}
			klog.Info("caBundle on CRD updated")
		}
	} else {
		klog.Info("WARNING: Could not read ca.crt from the pod")
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	minioInformerFactory := informers.NewSharedInformerFactory(controllerClient, time.Second*30)
	podName := os.Getenv(HostnameEnv)
	if podName == "" {
		klog.Info("Could not determine $%s, defaulting to pod name: operator-pod", HostnameEnv)
		podName = "operator-pod"
	}

	mainController := cluster.NewController(
		podName,
		namespaces,
		kubeClient,
		controllerClient,
		promClient,
		kubeInformerFactory.Apps().V1().StatefulSets(),
		kubeInformerFactory.Apps().V1().Deployments(),
		kubeInformerFactory.Core().V1().Pods(),
		minioInformerFactory.Minio().V2().Tenants(),
		minioInformerFactory.Sts().V1beta1().PolicyBindings(),
		kubeInformerFactory.Core().V1().Services(),
		hostsTemplate,
		version,
	)

	go kubeInformerFactory.Start(stopCh)
	go minioInformerFactory.Start(stopCh)

	if err = mainController.Start(2, stopCh); err != nil {
		klog.Fatalf("Error running mainController: %s", err.Error())
	}

	<-stopCh
	klog.Info("Shutting down the MinIO Operator")
	mainController.Stop()
}

// setupSignalHandler registered for SIGTERM and SIGINT. A stop channel is returned
// which is closed on one of these signals. If a second signal is caught, the program
// is terminated with exit code 1.
func setupSignalHandler() (stopCh <-chan struct{}) {
	// panics when called twice
	close(onlyOneSignalHandler)

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		<-c
		close(stop)
		<-c
		// second signal. Exit directly.
		os.Exit(1)
	}()

	return stop
}
