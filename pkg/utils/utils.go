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
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

// NewPodInformer creates a Shared Index Pod Informer matching the labelSelector string
func NewPodInformer(kubeClientSet kubernetes.Interface, labelSelectorString string) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(_ metav1.ListOptions) (runtime.Object, error) {
				return kubeClientSet.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{
					LabelSelector: labelSelectorString,
				})
			},
			WatchFunc: func(_ metav1.ListOptions) (watch.Interface, error) {
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
