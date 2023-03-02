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

package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/golang-jwt/jwt/v4"
	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	"github.com/minio/operator/pkg"
	"github.com/minio/pkg/env"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	mpConfigMapDefault = "mp-config"
	mpConfigMapKey     = "MP_CONFIG_KEY"
	mpHostEnvVar       = "MP_HOST"
	defaultMPHost      = "https://marketplace.apps.min.dev"
	mpEUHostEnvVar     = "MP_EU_HOST"
	defaultEUMPHost    = "https://marketplace-eu.apps.min.dev"
	isMPEmailSet       = "isEmailSet"
	emailNotSetMsg     = "Email was not sent in request"
)

func registerMarketplaceHandlers(api *operations.OperatorAPI) {
	api.OperatorAPIGetMPIntegrationHandler = operator_api.GetMPIntegrationHandlerFunc(func(params operator_api.GetMPIntegrationParams, session *models.Principal) middleware.Responder {
		payload, err := getMPIntegrationResponse(session, params)
		if err != nil {
			return operator_api.NewGetMPIntegrationDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetMPIntegrationOK().WithPayload(payload)
	})

	api.OperatorAPIPostMPIntegrationHandler = operator_api.PostMPIntegrationHandlerFunc(func(params operator_api.PostMPIntegrationParams, session *models.Principal) middleware.Responder {
		err := postMPIntegrationResponse(session, params)
		if err != nil {
			return operator_api.NewPostMPIntegrationDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewPostMPIntegrationCreated()
	})
}

func getMPIntegrationResponse(session *models.Principal, params operator_api.GetMPIntegrationParams) (*operator_api.GetMPIntegrationOKBody, *models.Error) {
	clientSet, err := K8sClient(session.STSSessionToken)
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	isMPEmailSet, err := getMPEmail(ctx, &k8sClient{client: clientSet})
	if err != nil {
		return nil, ErrorWithContext(ctx, ErrNotFound)
	}
	return &operator_api.GetMPIntegrationOKBody{
		IsEmailSet: isMPEmailSet,
	}, nil
}

func getMPEmail(ctx context.Context, clientSet K8sClientI) (bool, error) {
	cm, err := clientSet.getConfigMap(ctx, "default", getMPConfigMapKey(mpConfigMapKey), metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	return cm.Data[isMPEmailSet] == "true", nil
}

func postMPIntegrationResponse(session *models.Principal, params operator_api.PostMPIntegrationParams) *models.Error {
	clientSet, err := K8sClient(session.STSSessionToken)
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	if err != nil {
		return ErrorWithContext(ctx, err)
	}
	return setMPIntegration(ctx, params.Body.Email, params.Body.IsInEU, &k8sClient{client: clientSet})
}

func setMPIntegration(ctx context.Context, email string, isInEU bool, clientSet K8sClientI) *models.Error {
	if email == "" {
		return ErrorWithContext(ctx, ErrBadRequest, fmt.Errorf(emailNotSetMsg))
	}
	if _, err := setMPEmail(ctx, email, isInEU, clientSet); err != nil {
		return ErrorWithContext(ctx, err)
	}
	return nil
}

func setMPEmail(ctx context.Context, email string, isInEU bool, clientSet K8sClientI) (*corev1.ConfigMap, error) {
	if err := postEmailToMP(email, isInEU); err != nil {
		return nil, err
	}
	cm := createCM()
	return clientSet.createConfigMap(ctx, "default", cm, metav1.CreateOptions{})
}

func postEmailToMP(email string, isInEU bool) error {
	mpURL, err := getMPURL(isInEU)
	if err != nil {
		return err
	}
	return makePostRequestToMP(mpURL, email)
}

func getMPURL(isInEU bool) (string, error) {
	mpHost := getMPHost(isInEU)
	if mpHost == "" {
		return "", fmt.Errorf("mp host not set")
	}
	return fmt.Sprintf("%s/mp-email", mpHost), nil
}

func getMPHost(isInEU bool) string {
	if isInEU {
		return env.Get(mpEUHostEnvVar, defaultEUMPHost)
	}
	return env.Get(mpHostEnvVar, defaultMPHost)
}

func makePostRequestToMP(url, email string) error {
	request, err := createMPRequest(url, email)
	if err != nil {
		return err
	}
	client := GetConsoleHTTPClient("")
	client.Timeout = 3 * time.Second
	if res, err := client.Do(request); err != nil {
		return err
	} else if res.StatusCode >= http.StatusBadRequest {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("request to %s failed with status code %d and error %s", url, res.StatusCode, string(b))
	}
	return nil
}

func createMPRequest(url, email string) (*http.Request, error) {
	request, err := http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf("{\"email\":\"%s\"}", email)))
	if err != nil {
		return nil, err
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{})
	jwtTokenString, err := jwtToken.SignedString([]byte(pkg.MPSecret))
	if err != nil {
		return nil, err
	}
	request.Header.Add("Cookie", fmt.Sprintf("jwtToken=%s", jwtTokenString))
	request.Header.Add("Content-Type", "application/json")
	return request, nil
}

func createCM() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      getMPConfigMapKey(mpConfigMapKey),
			Namespace: "default",
		},
		Data: map[string]string{isMPEmailSet: "true"},
	}
}

func getMPConfigMapKey(envVar string) string {
	if mp := os.Getenv(envVar); mp != "" {
		return mp
	}
	return mpConfigMapDefault
}
