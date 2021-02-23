// +build go1.13

/*
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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/klog/v2"

	clientset "github.com/minio/operator/pkg/client/clientset/versioned"
	informers "github.com/minio/operator/pkg/client/informers/externalversions"
	"github.com/minio/operator/pkg/controller/cluster"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	certapi "k8s.io/client-go/kubernetes/typed/certificates/v1beta1"
	"k8s.io/client-go/rest"
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

	certClient, err := certapi.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("Error building certificate clientset: %v", err.Error())
	}

	extClient, err := apiextension.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("Error building certificate clientset: %v", err.Error())
	}

	namespace, isNamespaced := os.LookupEnv("WATCHED_NAMESPACE")

	ctx := context.Background()
	var caContent []byte
	operatorCATLSCert, err := kubeClient.CoreV1().Secrets(miniov2.GetNSFromFile()).Get(ctx, "operator-ca-tls", metav1.GetOptions{})
	// if custom ca.crt is not present in kubernetes secrets use the one stored in the pod
	if err != nil {
		caContent = miniov2.GetPodCAFromFile()
	} else {
		if val, ok := operatorCATLSCert.Data["ca.crt"]; ok {
			caContent = val
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

	var kubeInformerFactory kubeinformers.SharedInformerFactory
	var minioInformerFactory informers.SharedInformerFactory
	if isNamespaced {
		kubeInformerFactory = kubeinformers.NewSharedInformerFactoryWithOptions(kubeClient, time.Second*30, kubeinformers.WithNamespace(namespace))
		minioInformerFactory = informers.NewSharedInformerFactoryWithOptions(controllerClient, time.Second*30, informers.WithNamespace(namespace))
	} else {
		kubeInformerFactory = kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
		minioInformerFactory = informers.NewSharedInformerFactory(controllerClient, time.Second*30)
	}

	mainController := cluster.NewController(kubeClient, controllerClient, *certClient,
		kubeInformerFactory.Apps().V1().StatefulSets(),
		kubeInformerFactory.Apps().V1().Deployments(),
		kubeInformerFactory.Batch().V1().Jobs(),
		minioInformerFactory.Minio().V2().Tenants(),
		kubeInformerFactory.Core().V1().Services(),
		hostsTemplate,
		version)

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
