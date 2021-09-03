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

package cluster

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	jwtreq "github.com/golang-jwt/jwt/request"

	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
)

// Used for registering with rest handlers (have a look at registerStorageRESTHandlers for usage example)
// If it is passed ["aaaa", "bbbb"], it returns ["aaaa", "{aaaa:.*}", "bbbb", "{bbbb:.*}"]
func restQueries(keys ...string) []string {
	var accumulator []string
	for _, key := range keys {
		accumulator = append(accumulator, key, "{"+key+":.*}")
	}
	return accumulator
}

func configureWebhookServer(c *Controller) *http.Server {
	router := mux.NewRouter().SkipClean(true).UseEncodedPath()

	router.Methods(http.MethodGet).
		Path(miniov2.WebhookAPIGetenv + "/{namespace}/{name:.+}").
		HandlerFunc(c.GetenvHandler).
		Queries(restQueries("key")...)

	router.Methods(http.MethodPost).
		Path(miniov2.WebhookAPIBucketService + "/{namespace}/{name:.+}").
		HandlerFunc(c.BucketSrvHandler).
		Queries(restQueries("bucket")...)

	router.Methods(http.MethodGet).
		PathPrefix(miniov2.WebhookAPIUpdate).
		Handler(http.StripPrefix(miniov2.WebhookAPIUpdate, http.FileServer(http.Dir(updatePath))))

	// CRD Conversion
	router.Methods(http.MethodPost).
		Path(miniov2.WebhookCRDConversaion).
		HandlerFunc(c.CRDConversionHandler)

	router.NotFoundHandler = http.NotFoundHandler()

	s := &http.Server{
		Addr:           ":" + miniov2.WebhookDefaultPort,
		Handler:        router,
		ReadTimeout:    time.Minute,
		WriteTimeout:   time.Minute,
		MaxHeaderBytes: 1 << 20,
	}

	return s
}

func (c *Controller) validateRequest(r *http.Request, secret *v1.Secret) error {
	tokenStr, err := jwtreq.AuthorizationHeaderExtractor.ExtractToken(r)
	if err != nil {
		return err
	}

	stdClaims := &jwt.StandardClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, stdClaims, func(token *jwt.Token) (interface{}, error) {
		return secret.Data[miniov2.WebhookOperatorPassword], nil
	})
	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf(http.StatusText(http.StatusForbidden))
	}
	if stdClaims.Issuer != string(secret.Data[miniov2.WebhookOperatorUsername]) {
		return fmt.Errorf(http.StatusText(http.StatusForbidden))
	}

	return nil
}

func generateRandomKey(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

func (c *Controller) applyOperatorWebhookSecret(ctx context.Context, tenant *miniov2.Tenant) (*v1.Secret, error) {
	secret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx,
		miniov2.WebhookSecret, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			secret = getSecretForTenant(tenant, generateRandomKey(20), generateRandomKey(40))
			return c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Create(ctx, secret, metav1.CreateOptions{})
		}
		return nil, err
	}
	// check the secret has the desired values
	minioArgs := string(secret.Data[miniov2.WebhookMinIOArgs])
	if strings.Contains(minioArgs, "env://") && isOperatorTLS() {
		// update the secret
		minioArgs = strings.ReplaceAll(minioArgs, "env://", "env+tls://")
		secret.Data[miniov2.WebhookMinIOArgs] = []byte(minioArgs)
		secret, err = c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Update(ctx, secret, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		}
		// update the revision of the tenant to force a rolling restart across all statefulsets
		t2, err := c.increaseTenantRevision(ctx, tenant)
		if err != nil {
			return nil, err
		}
		*tenant = *t2
	}

	return secret, nil
}

func secretData(tenant *miniov2.Tenant, accessKey, secretKey string) []byte {
	scheme := "env"
	if isOperatorTLS() {
		scheme = "env+tls"
	}
	return []byte(fmt.Sprintf("%s://%s:%s@%s:%s%s/%s/%s",
		scheme,
		accessKey,
		secretKey,
		fmt.Sprintf("operator.%s.svc.%s",
			miniov2.GetNSFromFile(),
			miniov2.GetClusterDomain()),
		miniov2.WebhookDefaultPort,
		miniov2.WebhookAPIGetenv,
		tenant.Namespace,
		tenant.Name))
}
