// +build go1.13

/*
 * Copyright (C) 2019, MinIO, Inc.
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

	"github.com/golang/glog"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	certapi "k8s.io/client-go/kubernetes/typed/certificates/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	clientset "github.com/minio/minio-operator/pkg/client/clientset/versioned"
	informers "github.com/minio/minio-operator/pkg/client/informers/externalversions"
	"github.com/minio/minio-operator/pkg/controller/cluster"
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
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	controllerClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building MinIO clientset: %s", err.Error())
	}

	certClient, err := certapi.NewForConfig(cfg)
	if err != nil {
		glog.Errorf("Error building certificate clientset: %v", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	minioInformerFactory := informers.NewSharedInformerFactory(controllerClient, time.Second*30)

	controller := cluster.NewController(kubeClient, controllerClient, *certClient,
		kubeInformerFactory.Apps().V1().StatefulSets(),
		minioInformerFactory.Min().V1beta1().MinIOInstances(),
		kubeInformerFactory.Core().V1().Services())

	go kubeInformerFactory.Start(stopCh)
	go minioInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
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
