// This file is part of MinIO Operator
// Copyright (c) 2020 MinIO, Inc.
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

package controller

//lint:file-ignore ST1005 Incorrectly formatted error string

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/minio/operator/pkg/common"

	"github.com/minio/operator/pkg/apis/sts.min.io/v1alpha1"
	iampolicy "github.com/minio/pkg/iam/policy"

	"github.com/gorilla/mux"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	xhttp "github.com/minio/operator/pkg/internal"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/klog/v2"
)

// Supported remote envs
const (
	updatePath = "/tmp" + common.WebhookAPIUpdate + slashSeparator
)

const contextLogKey = contextKeyType("operatorlog")

// AssumeRoleWithWebIdentityHandler - POST /sts/{tenantNamespace}
// AssumeRoleWithWebIdentity - implementation of AWS STS API.
// Authenticates a Kubernetes Service accounts using a JWT Token
// Evalues a PolicyBinding CRD as Mapping of the Minio Policies that the ServiceAccount can assume on a minio tenant
// Eg:-
// $ curl -k -X POST https://operator:9443/sts/{tenantNamespace} -d "Version=2011-06-15&Action=AssumeRoleWithWebIdentity&WebIdentityToken=<jwt>" -H "Content-Type: application/x-www-form-urlencoded"
func (c *Controller) AssumeRoleWithWebIdentityHandler(w http.ResponseWriter, r *http.Request) {
	routerVars := mux.Vars(r)
	tenantNamespace := ""
	tenantNamespace, err := xhttp.UnescapeQueryPath(routerVars["tenantNamespace"])
	if err != nil {
		writeSTSErrorResponse(w, true, ErrSTSInvalidParameterValue, fmt.Errorf("Unable to unescape tenant namespace: %s", err))
		return
	}

	reqInfo := ReqInfo{
		RequestID:       w.Header().Get(AmzRequestID),
		RemoteHost:      xhttp.GetSourceIPFromHeaders(r),
		Host:            r.Host,
		UserAgent:       r.UserAgent(),
		API:             webIdentity,
		TenantNamespace: tenantNamespace,
	}

	ctx := context.WithValue(r.Context(), contextLogKey, &reqInfo)

	if err != nil {
		writeSTSErrorResponse(w, true, ErrSTSInvalidParameterValue, fmt.Errorf("Tenant namespace is missing:, %s", err))
		return
	}

	// Parse the incoming form data.
	if err := xhttp.ParseForm(r); err != nil {
		writeSTSErrorResponse(w, true, ErrSTSInvalidParameterValue, fmt.Errorf("Error parsing request: %s", err))
		return
	}

	if r.Form.Get(stsVersion) != stsAPIVersion {
		err := fmt.Errorf("invalid STS API version %s, expecting %s", r.Form.Get("Version"), stsAPIVersion)
		writeSTSErrorResponse(w, true, ErrSTSMissingParameter, err)
		return
	}

	action := r.Form.Get(stsAction)
	switch action {
	// For now we only do WebIdentity, leaving it in case we want to implement certificate authentication
	case webIdentity:
	default:
		writeSTSErrorResponse(w, true, ErrSTSInvalidParameterValue, fmt.Errorf("Unsupported action %s", action))
		return
	}

	token := strings.TrimSpace(r.Form.Get(stsWebIdentityToken))

	if token == "" {
		writeSTSErrorResponse(w, true, ErrSTSMissingParameter, fmt.Errorf("Missing parameter '%s'", stsWebIdentityToken))
		return
	}

	// VALIDATE JWT
	accessToken := r.Form.Get(stsWebIdentityToken)
	saAuthResult, err := c.ValidateServiceAccountJWT(&ctx, accessToken)
	if err != nil {
		writeSTSErrorResponse(w, true, ErrSTSInvalidIdentityToken, err)
		return
	}

	if !saAuthResult.Status.Authenticated {
		writeSTSErrorResponse(w, true, ErrSTSAccessDenied, fmt.Errorf("Access denied: Invalid Token"))
		return
	}
	chunks := strings.Split(strings.Replace(saAuthResult.Status.User.Username, "system:serviceaccount:", "", -1), ":")

	if len(chunks) < 2 {
		writeSTSErrorResponse(w, true, ErrSTSInvalidIdentityToken, fmt.Errorf("Error parsing service account name and namespace"))
		return
	}
	// saNamespace Service account Namespace
	saNamespace := chunks[0]
	// saName service account username
	saName := chunks[1]

	// Authorized PolicyBindings for the Service Account
	policyBindings := []v1alpha1.PolicyBinding{}
	pbs, err := c.minioClientSet.StsV1alpha1().PolicyBindings(tenantNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		writeSTSErrorResponse(w, true, ErrSTSInternalError, fmt.Errorf("Error obtaining PolicyBindings: %s", err))
		return
	}

	for _, pb := range pbs.Items {
		if pb.Spec.Application.Namespace == saNamespace && pb.Spec.Application.ServiceAccount == saName {
			policyBindings = append(policyBindings, pb)
		}
	}
	if len(policyBindings) == 0 {
		writeSTSErrorResponse(w, true, ErrSTSAccessDenied, fmt.Errorf("Service account '%s' has no PolicyBindings in namespace '%s'", saAuthResult.Status.User.Username, tenantNamespace))
		return
	}

	tenants, err := c.minioClientSet.MinioV2().Tenants(tenantNamespace).List(ctx, metav1.ListOptions{})

	if err != nil || len(tenants.Items) == 0 {
		if k8serrors.IsNotFound(err) {
			writeSTSErrorResponse(w, true, ErrSTSInvalidParameterValue, fmt.Errorf("No tenant found in namespace '%s'", tenantNamespace))
			return
		}
		writeSTSErrorResponse(w, true, ErrSTSInvalidParameterValue, fmt.Errorf("Error getting tenant in namespace '%s'", tenantNamespace))
		return
	}

	// Only one tenant is allowed in a single namespace, gathering the first tenant in the list
	tenant := tenants.Items[0]

	tenantConfiguration, err := c.getTenantCredentials(ctx, &tenant)
	if err != nil {
		if errors.Is(err, ErrEmptyRootCredentials) {
			writeSTSErrorResponse(w, true, ErrSTSInternalError, fmt.Errorf("Tenant '%s' is missing root credentials", tenant.Name))
			return
		}
		writeSTSErrorResponse(w, true, ErrSTSInternalError, fmt.Errorf("Error getting tenant '%s' root credentials: %s", tenant.Name, err))
		return
	}
	adminClient, err := tenant.NewMinIOAdmin(tenantConfiguration, c.getTransport())
	if err != nil {
		writeSTSErrorResponse(w, true, ErrSTSInternalError, fmt.Errorf("Error communicating with tenant '%s': %s", tenant.Name, err))
		return
	}

	// Session Policy
	sessionPolicyStr := r.Form.Get(stsPolicy)
	var compactedSessionPolicy string
	var sessionPolicy *iampolicy.Policy
	if len(sessionPolicyStr) > 0 {
		compactedSessionPolicy, err = miniov2.CompactJSONString(sessionPolicyStr)
		if err != nil {
			writeSTSErrorResponse(w, true, ErrSTSMalformedPolicyDocument, err)
			return
		}
		sessionPolicy, err = iampolicy.ParseConfig(bytes.NewReader([]byte(compactedSessionPolicy)))
		if err != nil {
			writeSTSErrorResponse(w, true, ErrSTSMalformedPolicyDocument, err)
			return
		}
		// Version in policy must not be empty
		if sessionPolicy.Version == "" {
			writeSTSErrorResponse(w, true, ErrSTSInvalidParameterValue, fmt.Errorf("Invalid session policy version"))
			return
		}
		// The plain text that you use for both inline and managed session
		// policies shouldn't exceed 2048 characters.
		if len(compactedSessionPolicy) > 2048 {
			writeSTSErrorResponse(w, true, ErrSTSPackedPolicyTooLarge, fmt.Errorf("Session policy should not exceed 2048 characters"))
			return
		}
	}

	var bfPolicy iampolicy.Policy
	for _, pb := range policyBindings {
		if sessionPolicy != nil {
			bfPolicy = bfPolicy.Merge(*sessionPolicy)
		}
		for _, policyName := range pb.Spec.Policies {
			policy, err := GetPolicy(ctx, adminClient, policyName)
			if err != nil {
				klog.Error(fmt.Errorf("Invalid policy %s, ignoring: %s", policyName, err))
				continue
			}
			parsedPolicy, err := iampolicy.ParseConfig(bytes.NewReader([]byte(policy.Policy)))
			if err != nil {
				klog.Error(fmt.Errorf("Invalid policy, '%s' isnot parseable ignoring: %s", policyName, err))
				continue
			}
			bfPolicy = bfPolicy.Merge(*parsedPolicy)
		}
	}
	bfJSONPolicy, _ := json.Marshal(bfPolicy)
	bfCompact, err := miniov2.CompactJSONString(string(bfJSONPolicy))
	if err != nil {
		writeSTSErrorResponse(w, true, ErrSTSMalformedPolicyDocument, err)
		return
	}
	if len(bfCompact) > 2048 {
		writeSTSErrorResponse(w, true, ErrSTSPackedPolicyTooLarge, fmt.Errorf("PolicyBinding resulting policy is too long, Policy should not exceed 2048 characters, length %d", len(bfCompact)))
		return
	}

	durationStr := r.Form.Get(stsDurationSeconds)
	var durationInSeconds int
	if durationStr != "" {
		duration, err := strconv.Atoi(durationStr)
		if err != nil {
			writeSTSErrorResponse(w, true, ErrSTSInvalidParameterValue, fmt.Errorf("Invalid token expiry"))
			return
		}

		if duration < 900 || duration > 31536000 {
			writeSTSErrorResponse(w, true, ErrSTSInvalidParameterValue, fmt.Errorf("Invalid token expiry: min 900s, max 31536000s"))
			return
		}
		durationInSeconds = duration
	}

	stsCredentials, err := AssumeRole(ctx, c, &tenant, bfCompact, durationInSeconds)
	if err != nil {
		writeSTSErrorResponse(w, true, ErrSTSInternalError, err)
		return
	}

	assumeRoleResponse := &AssumeRoleWithWebIdentityResponse{
		Result: WebIdentityResult{
			Credentials: Credentials{
				AccessKey:    stsCredentials.AccessKeyID,
				SecretKey:    stsCredentials.SecretAccessKey,
				SessionToken: stsCredentials.SessionToken,
			},
		},
	}

	assumeRoleResponse.ResponseMetadata.RequestID = w.Header().Get(AmzRequestID)
	writeSuccessResponseXML(w, xhttp.EncodeResponse(assumeRoleResponse))
}
