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

package controller

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/minio/operator/pkg"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/minio/minio-go/v7/pkg/set"
	"k8s.io/client-go/rest"

	"k8s.io/klog/v2"

	clientset "github.com/minio/operator/pkg/client/clientset/versioned"
	informers "github.com/minio/operator/pkg/client/informers/externalversions"
	promclientset "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

const (
	// OperatorWatchedNamespaceEnv Env variable name, the namespaces which the operator watches for MinIO tenants. Defaults to "" for all namespaces.
	OperatorWatchedNamespaceEnv = "WATCHED_NAMESPACE"
	// HostnameEnv Host name env variable
	HostnameEnv = "HOSTNAME"
)

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

// StartOperator starts the MinIO Operator controller
func StartOperator(kubeconfig string) {
	klog.Info("Starting MinIO Operator")
	// set up signals, so we handle the first shutdown signal gracefully
	stopCh := setupSignalHandler()

	flag.Parse()

	if checkVersion {
		fmt.Println(pkg.Version)
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

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	minioInformerFactory := informers.NewSharedInformerFactory(controllerClient, time.Second*30)
	podName := os.Getenv(HostnameEnv)
	if podName == "" {
		klog.Infof("Could not determine %s, defaulting to pod name: operator-pod", HostnameEnv)
		podName = "operator-pod"
	}

	mainController := NewController(
		podName,
		namespaces,
		kubeClient,
		controllerClient,
		promClient,
		kubeInformerFactory.Apps().V1().StatefulSets(),
		kubeInformerFactory.Apps().V1().Deployments(),
		kubeInformerFactory.Core().V1().Pods(),
		minioInformerFactory.Minio().V2().Tenants(),
		minioInformerFactory.Sts().V1alpha1().PolicyBindings(),
		kubeInformerFactory.Core().V1().Services(),
		hostsTemplate,
		pkg.Version,
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

// Result contains the result of a sync invocation.
type Result struct {
	// Requeue tells the Controller to requeue the reconcile key.  Defaults to false.
	Requeue bool

	// RequeueAfter if greater than 0, tells the Controller to requeue the reconcile key after the Duration.
	// Implies that Requeue is true, there is no need to set Requeue to true at the same time as RequeueAfter.
	RequeueAfter time.Duration
}

// WrapResult is wrap for result.
// We can find where return result.
func WrapResult(result Result, err error) (Result, error) {
	return result, err
}
