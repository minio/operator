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
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/klog"

	clientset "github.com/minio/minio-operator/pkg/client/clientset/versioned"
	informers "github.com/minio/minio-operator/pkg/client/informers/externalversions"
	"github.com/minio/minio-operator/pkg/controller/cluster"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	certapi "k8s.io/client-go/kubernetes/typed/certificates/v1beta1"
	"k8s.io/client-go/rest"
)

// Version provides the version of this minio-operator
var Version = "DEVELOPMENT.GOGET"

var (
	masterURL    string
	kubeconfig   string
	checkVersion bool

	onlyOneSignalHandler = make(chan struct{})
	shutdownSignals      = []os.Signal{os.Interrupt, syscall.SIGTERM}
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to a kubeconfig. Only required if out-of-cluster")
	flag.StringVar(&masterURL, "master", "", "the address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster")
	flag.BoolVar(&checkVersion, "version", false, "print version")
}

func main() {
	klog.Info("Starting MinIO Operator")
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := setupSignalHandler()

	flag.Parse()

	if checkVersion {
		fmt.Println(Version)
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

	namespace, isNamespaced := os.LookupEnv("WATCHED_NAMESPACE")

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
		minioInformerFactory.Operator().V1().MinIOInstances(),
		kubeInformerFactory.Core().V1().Services())

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
