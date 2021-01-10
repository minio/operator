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

package secrets

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/resources/services"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// generatePassword returns a url-safe base64 encoding of a 32-byte password
func generatePassword() []byte {
	var nonce [32]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		panic("Unable to read enough randomness from the system")
	}
	return []byte(base64.URLEncoding.EncodeToString(nonce[:]))
}

// LogSecret returns a k8s secret object with postgres password
func LogSecret(t *miniov2.Tenant) *corev1.Secret {
	dbAddr := services.GetLogSearchDBAddr(t)
	pgPasswd := generatePassword()
	auditToken := generatePassword()
	queryToken := generatePassword()
	pgConnStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", miniov2.LogPgUser, pgPasswd, dbAddr, miniov2.LogAuditDB)
	return &corev1.Secret{
		Type: "Opaque",
		ObjectMeta: metav1.ObjectMeta{
			Name:            t.LogSecretName(),
			Namespace:       t.Namespace,
			OwnerReferences: t.OwnerRef(),
		},
		Data: map[string][]byte{
			miniov2.LogPgPassKey:     pgPasswd,
			miniov2.LogAuditTokenKey: auditToken,
			miniov2.LogQueryTokenKey: queryToken,
			miniov2.LogPgConnStr:     []byte(pgConnStr),
		},
	}
}
