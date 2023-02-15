package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var (
	masterURL  string
	kubeconfig string
)

// GetNSFromFile assumes the operator is running inside a k8s pod and extract the
// current namespace from the /var/run/secrets/kubernetes.io/serviceaccount/namespace file
func GetNSFromFile() string {
	namespace, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "minio-operator"
	}
	return string(namespace)
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to a kubeconfig. Only required if out-of-cluster")
}

func main() {
	fmt.Println("Look for incluster config by default")
	cfg, err := rest.InClusterConfig()
	// If config is passed as a flag use that instead
	if kubeconfig != "" {
		cfg, err = clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building Kubernetes clientset: %s", err.Error())
	}

	ctx := context.TODO()
	secret, err := kubeClient.CoreV1().Secrets("openshift-kube-controller-manager-operator").Get(
		ctx, "csr-signer", metav1.GetOptions{})
	klog.Info("Checking if this is OpenShift Environment...")
	if err != nil {
		klog.Errorf("failed to get secret: %#v", err)
		if k8serrors.IsNotFound(err) {
			// Do nothing special, because this is maybe k8s vanilla
			klog.Info("This is NOT OpenShift because csr-signer secret was NOT found")
		}
	} else {
		// Do something special, create the secret to trust the tenant spec.
		klog.Info("This is OpenShift because csr-signer secret was found")
		cpData := *&secret.Data
		var tlsCrt []byte
		for k, v := range cpData {
			if k == "tls.crt" {
				tlsCrt = v
			}
		}
		// To get minio-operator namespace without hardcoding the value in case
		// it comes from OperatorHub I think...
		namespace := GetNSFromFile()
		newSecret := &corev1.Secret{
			Type: "Opaque",
			ObjectMeta: metav1.ObjectMeta{
				Name:      "minio-operator-openshift-signer",
				Namespace: namespace,
			},
			Data: map[string][]byte{
				"tls.crt": tlsCrt,
			},
		}
		_, err := kubeClient.CoreV1().Secrets(namespace).Create(
			ctx, newSecret, metav1.CreateOptions{})
		if err != nil {
			klog.Errorf("failed to create secret: %#v", err)
		}
	}
}
