// This file is part of MinIO Operator
// Copyright (c) 2022 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package utils

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/minio/operator/pkg/common"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

// NewUUID - get a random UUID.
func NewUUID() (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// DecodeBase64 : decoded base64 input into utf-8 text
func DecodeBase64(s string) (string, error) {
	decodedInput, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(decodedInput), nil
}

// Key used for Get/SetReqInfo
type key string

// context keys
const (
	ContextLogKey            = key("console-log")
	ContextRequestID         = key("request-id")
	ContextRequestUserID     = key("request-user-id")
	ContextRequestUserAgent  = key("request-user-agent")
	ContextRequestHost       = key("request-host")
	ContextRequestRemoteAddr = key("request-remote-addr")
	ContextAuditKey          = key("request-audit-entry")
)

// GetOperatorRuntime Retrieves the runtime from env variable
func GetOperatorRuntime() common.Runtime {
	envString := os.Getenv(common.OperatorRuntimeEnv)
	runtimeReturn := common.OperatorRuntimeK8s
	if envString != "" {
		envString = strings.TrimSpace(envString)
		envString = strings.ToUpper(envString)
		if val, ok := common.Runtimes[envString]; ok {
			runtimeReturn = val
		}
	}
	return runtimeReturn
}

// NewPodInformer creates a Shared Index Pod Informer matching the labelSelector string
func NewPodInformer(kubeClientSet kubernetes.Interface, labelSelectorString string) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return kubeClientSet.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{
					LabelSelector: labelSelectorString,
				})
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return kubeClientSet.CoreV1().Pods("").Watch(context.Background(), metav1.ListOptions{
					LabelSelector: labelSelectorString,
				})
			},
		},
		&corev1.Pod{},  // Resource type
		time.Second*30, // Resync period
		cache.Indexers{
			cache.NamespaceIndex: cache.MetaNamespaceIndexFunc, // Index by namespace
		},
	)
}

// LabelSelectorToString gets a string from a labelSelector
func LabelSelectorToString(labelSelector metav1.LabelSelector) (string, error) {
	var matchExpressions []string
	for _, expr := range labelSelector.MatchExpressions {
		// Handle only Exists expressions
		matchExpressions = append(matchExpressions, expr.Key)
	}
	// Join match labels and match expressions into a single string with a comma separator.
	labelSelectorString := strings.Join(matchExpressions, ",")
	// Validate labelSelectorString
	if _, err := labels.Parse(labelSelectorString); err != nil {
		return "", err
	}
	return labelSelectorString, nil
}

// CastObjectToMetaV1 gets a metav1.Object from an interface
func CastObjectToMetaV1(obj interface{}) (metav1.Object, error) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return nil, fmt.Errorf("error decoding object, invalid type")
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			return nil, fmt.Errorf("error decoding object tombstone, invalid type")
		}
		klog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	return object, nil
}
