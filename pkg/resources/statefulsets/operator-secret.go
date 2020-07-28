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

package statefulsets

import (
	"context"
	"fmt"
	"math/rand"

	v1 "github.com/minio/operator/pkg/apis/minio.min.io/v1"
	"github.com/minio/operator/pkg/server"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	key   = "key"
	token = "token"
)

func CreateOperatorJWTToken(namespace string, tenant string, operator string) (*corev1.Secret, error) {
	immutable := true
	randomKey := []byte(fmt.Sprint(rand.Int()))
	key, err := server.GenerateJWTForTenant(tenant, operator, randomKey)
	if err != nil {
		return nil, err
	}
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind: "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      v1.OperatorJWTSecretName(operator, tenant),
			Namespace: namespace,
		},
		Immutable: &immutable,
		Data: map[string][]byte{
			key: randomKey,
		},
		StringData: map[string]string{
			token: key,
		},
		Type: corev1.SecretTypeOpaque,
	}, nil
}

func GetOperatorJWTAndKey(ctx context.Context, client kubernetes.Interface, namespace string, operator string, tenant string) (string, []byte, error) {
	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, v1.OperatorJWTSecretName(operator, tenant), metav1.GetOptions{})
	if err != nil {
		return "", nil, err
	}
	return secret.StringData[token], secret.Data[key], err
}
