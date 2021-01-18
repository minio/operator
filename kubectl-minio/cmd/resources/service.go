/*
 * This file is part of MinIO Operator
 * Copyright (C) 2021, MinIO, Inc.
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

package resources

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewServiceForOperator will return a new service for a MinIO Operator webhook server
func NewServiceForOperator(opts OperatorOptions) *corev1.Service {
	operatorWebhookHTTPPort := corev1.ServicePort{Port: 4222, Name: "http"}
	operatorWebhookHTTPSPort := corev1.ServicePort{Port: 4233, Name: "https"}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    operatorLabels(),
			Name:      "operator",
			Namespace: opts.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				operatorWebhookHTTPPort,
				operatorWebhookHTTPSPort,
			},
			Selector: operatorLabels(),
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
}
