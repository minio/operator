// Copyright (C) 2022, MinIO, Inc.
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

package certificates

import (
	"context"
	"os"
	"strings"
	"sync"

	certificatesV1 "k8s.io/api/certificates/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const (
	// OperatorCertificatesVersion is the ENV var to force the certificates api version to use.
	OperatorCertificatesVersion = "MINIO_OPERATOR_CERTIFICATES_VERSION"
	// OperatorRuntime tells us which runtime we have. (EKS, Rancher, OpenShift, etc...)
	OperatorRuntime = "MINIO_OPERATOR_RUNTIME"
	// CSRSignerName is the name to use for the CSR Signer, will override the default
	CSRSignerName = "MINIO_OPERATOR_CSR_SIGNER_NAME"
	// EKSCsrSignerName is the signer we should use on EKS after version 1.22
	EKSCsrSignerName = "beta.eks.amazonaws.com/app-serving"
)

// CSRVersion represents the valid types of CSR that can be used
type CSRVersion string

// Valid CSR Versions
const (
	// CSRV1 is the new version to use after k8s 1.21
	CSRV1 CSRVersion = "v1"
	// CSRV1Beta1 is dreprecated and will be removed in k8s 1.22
	CSRV1Beta1 CSRVersion = "v1beta1"
)

var (
	csrVersion               CSRVersion
	certificateVersionOnce   sync.Once
	defaultCsrSignerName     string
	defaultCsrSignerNameOnce sync.Once
	csrSignerName            string
	csrSignerNameOnce        sync.Once
)

func getDefaultCsrSignerName() string {
	defaultCsrSignerNameOnce.Do(func() {
		if os.Getenv(CSRSignerName) != "" {
			defaultCsrSignerName = os.Getenv(CSRSignerName)
		}
		defaultCsrSignerName = certificatesV1.KubeletServingSignerName
	})
	return defaultCsrSignerName
}

// GetCertificatesAPIVersion returns which certificates api version operator will use to generate certificates
func GetCertificatesAPIVersion(clientSet kubernetes.Interface) CSRVersion {
	// we will calculate which CSR version to use only once to avoid having to discover the
	certificateVersionOnce.Do(func() {
		version, _ := os.LookupEnv(OperatorCertificatesVersion)
		csrVersion = CSRV1
		switch version {
		case "v1":
			csrVersion = CSRV1
		case "v1beta1":
			csrVersion = CSRV1Beta1
		default:
			apiVersions, err := clientSet.Discovery().ServerPreferredResources()
			if err != nil {
				// If extension API server is not available, we emit a warning and continue.
				if discovery.IsGroupDiscoveryFailedError(err) {
					klog.Warningf("The Kubernetes server has an orphaned API service. Server reports: %s", err)
					klog.Warningf("To fix this, check related API Server or kubectl delete apiservice <service-name>")
				} else {
					panic(err)
				}
			}
			for _, api := range apiVersions {
				// if certificates v1beta1 is present operator will use that api by default
				// based on: https://github.com/aws/containers-roadmap/issues/1604#issuecomment-1072660824
				if api.GroupVersion == "certificates.k8s.io/v1beta1" {
					csrVersion = CSRV1Beta1
					break
				}
			}
		}
	})
	return csrVersion
}

// GetCSRSignerName returns the signer to be used
func GetCSRSignerName(clientSet kubernetes.Interface) string {
	csrSignerNameOnce.Do(func() {
		// At the moment we will use kubernetes.io/kubelet-serving as the default
		csrSignerName = getDefaultCsrSignerName()
		// only for csr api v1 we will try to detect if we are running inside an EKS cluster and switch to AWS's way to
		// get certificates using their CSRSignerName https://docs.aws.amazon.com/eks/latest/userguide/cert-signing.html
		if GetCertificatesAPIVersion(clientSet) == CSRV1 {
			// if the user specified the EKS runtime, no need to do the check
			if os.Getenv(OperatorRuntime) == "EKS" {
				csrSignerName = EKSCsrSignerName
				return
			}
			nodes, err := clientSet.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
			if err != nil {
				klog.Infof("Could not retrieve nodes to determine if we are in EKS: %v", err)
			}
			// if we find a single node with a kubeletVersion that contains "eks" or a label that starts with
			// "eks.amazonaws.com", we'll start using AWS EKS signer name
			for _, n := range nodes.Items {
				if strings.Contains(n.Status.NodeInfo.KubeletVersion, "eks") {
					csrSignerName = EKSCsrSignerName
					break
				}
				for k := range n.ObjectMeta.Labels {
					if strings.HasPrefix(k, "eks.amazonaws.com") {
						csrSignerName = EKSCsrSignerName
					}
				}
			}
		}
	})
	return csrSignerName
}
