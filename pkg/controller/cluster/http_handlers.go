// This file is part of MinIO Console Server
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

package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	miniov1 "github.com/minio/operator/pkg/apis/minio.min.io/v1"
	"github.com/minio/operator/pkg/apis/sts.min.io/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/minio/operator/pkg/resources/statefulsets"
	iampolicy "github.com/minio/pkg/iam/policy"

	"github.com/gorilla/mux"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	xhttp "github.com/minio/operator/pkg/internal"
	"github.com/minio/operator/pkg/resources/services"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/klog/v2"
)

// Supported remote envs
const (
	envMinIOArgs          = "MINIO_ARGS"
	envMinIOServiceTarget = "MINIO_DNS_WEBHOOK_ENDPOINT"
	updatePath            = "/tmp" + miniov2.WebhookAPIUpdate + slashSeparator
)

const contextLogKey = contextKeyType("operatorlog")

// BucketSrvHandler - POST /webhook/v1/bucketsrv/{namespace}/{name}?bucket={bucket}
func (c *Controller) BucketSrvHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	v := r.URL.Query()

	namespace := vars["namespace"]
	bucket := vars["bucket"]
	name := vars["name"]
	deleteBucket := v.Get("delete")

	secret, err := c.kubeClientSet.CoreV1().Secrets(namespace).Get(r.Context(),
		miniov2.WebhookSecret, metav1.GetOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	if err = c.validateRequest(r, secret); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	ok, err := strconv.ParseBool(deleteBucket)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	if ok {
		if err = c.kubeClientSet.CoreV1().Services(namespace).Delete(r.Context(), bucket, metav1.DeleteOptions{}); err != nil {
			klog.Errorf("failed to delete service:%s for tenant:%s/%s, err:%s", name, namespace, name, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		return
	}

	// Find the tenant
	tenant, err := c.minioClientSet.MinioV2().Tenants(namespace).Get(r.Context(), name, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Unable to lookup tenant:%s/%s for the bucket:%s request. err:%s", namespace, name, bucket, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tenant.EnsureDefaults()

	// Validate the MinIO Tenant
	if err = tenant.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	ok, error := validateBucketName(bucket)
	if !ok {
		http.Error(w, error.Error(), http.StatusBadRequest)
		return
	}
	// Create the service for the bucket name
	service := services.ServiceForBucket(tenant, bucket)
	_, err = c.kubeClientSet.CoreV1().Services(namespace).Create(r.Context(), service, metav1.CreateOptions{})
	if err != nil && k8serrors.IsAlreadyExists(err) {
		klog.Infof("Bucket:%s already exists for tenant:%s/%s err:%s ", bucket, namespace, name, err)
		// This might be a previously failed bucket creation. The service is expected to the be the same as the one
		// already in place so clear the error.
		err = nil
	}
	if err != nil {
		klog.Errorf("Unable to create service for tenant:%s/%s for the bucket:%s request. err:%s", namespace, name, bucket, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func validateBucketName(bucket string) (bool, error) {
	// Additional check on top of existing checks done by minio due to limitation of service creation in k8s
	if strings.Contains(bucket, ".") {
		return false, fmt.Errorf("invalid bucket name: . in bucket name: %s", bucket)
	}
	return true, nil
}

// GetenvHandler - GET /webhook/v1/getenv/{namespace}/{name}?key={env}
func (c *Controller) GetenvHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	name := vars["name"]
	key := vars["key"]

	secret, err := c.kubeClientSet.CoreV1().Secrets(namespace).Get(r.Context(),
		miniov2.WebhookSecret, metav1.GetOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if err = c.validateRequest(r, secret); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// Get the Tenant resource with this namespace/name
	tenant, err := c.minioClientSet.MinioV2().Tenants(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// The Tenant resource may no longer exist, in which case we stop processing.
			http.Error(w, fmt.Sprintf("Tenant '%s' in work queue no longer exists", key), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	tenant.EnsureDefaults()

	// Validate the MinIO Tenant
	if err = tenant.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	// correct all statefulset names by loading them, this will fix their name on the tenant pool names
	_, err = c.getAllSSForTenant(tenant)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	switch key {
	case envMinIOArgs:
		args := strings.Join(statefulsets.GetContainerArgs(tenant, c.hostsTemplate), " ")
		klog.Infof("%s value is %s", key, args)

		_, _ = w.Write([]byte(args))
		w.(http.Flusher).Flush()
	case envMinIOServiceTarget:
		schema := "https"
		if !isOperatorTLS() {
			schema = "http"
		}
		target := fmt.Sprintf("%s://%s:%s%s/%s/%s",
			schema,
			fmt.Sprintf("operator.%s.svc.%s",
				miniov2.GetNSFromFile(),
				miniov2.GetClusterDomain()),
			miniov2.WebhookDefaultPort,
			miniov2.WebhookAPIBucketService,
			tenant.Namespace,
			tenant.Name)
		klog.Infof("%s value is %s", key, target)

		_, _ = w.Write([]byte(target))
	default:
		http.Error(w, fmt.Sprintf("%s env key is not supported yet", key), http.StatusBadRequest)
		return
	}
}

// CRDConversionHandler - POST /webhook/v1/crd-conversion
func (c *Controller) CRDConversionHandler(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)

	var req v1.ConversionReview

	if err := dec.Decode(&req); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, obj := range req.Request.Objects {

		cr := unstructured.Unstructured{}
		if err := cr.UnmarshalJSON(obj.Raw); err != nil {
			log.Println(err)
			w.WriteHeader(500)
			req.Response.Result.Status = metav1.StatusFailure
			rawResp, _ := json.Marshal(req)
			if _, err = w.Write(rawResp); err != nil {
				log.Println(err)
			}
			return
		}

		switch cr.GetObjectKind().GroupVersionKind().Version {
		case "v1":

			// cast to v1
			// tenantV1 := obj.Object.(*miniov1.Tenant)
			tenantV1 := miniov1.Tenant{}

			// convert the runtime.Object to unstructured.Unstructured
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(cr.Object, &tenantV1)
			if err != nil {
				log.Println(err)
				w.WriteHeader(500)
				req.Response.Result.Status = metav1.StatusFailure
				rawResp, _ := json.Marshal(req)
				if _, err = w.Write(rawResp); err != nil {
					log.Println(err)
				}
				return
			}

			switch req.Request.DesiredAPIVersion {
			case "minio.min.io/v1":
				req.Response.ConvertedObjects = append(req.Response.ConvertedObjects, runtime.RawExtension{Object: &tenantV1})
			case "minio.min.io/v2":
				tenantV2 := miniov2.Tenant{}
				// convert to v2
				if err := tenantV1.ConvertTo(&tenantV2); err != nil {
					log.Println(err)
					w.WriteHeader(500)
					req.Response.Result.Status = metav1.StatusFailure
					rawResp, _ := json.Marshal(req)
					if _, err = w.Write(rawResp); err != nil {
						log.Println(err)
					}
					return
				}
				req.Response.ConvertedObjects = append(req.Response.ConvertedObjects, runtime.RawExtension{Object: &tenantV2})
			}
		case "v2":
			// cast to v2
			tenantV2 := miniov2.Tenant{}

			// convert the runtime.Object to unstructured.Unstructured
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(cr.Object, &tenantV2)
			if err != nil {
				log.Println(err)
				w.WriteHeader(500)
				req.Response.Result.Status = metav1.StatusFailure
				rawResp, _ := json.Marshal(req)
				if _, err = w.Write(rawResp); err != nil {
					log.Println(err)
				}
				return
			}

			switch req.Request.DesiredAPIVersion {
			case "minio.min.io/v1":
				// convert to v1
				var tenantV1 miniov1.Tenant
				if err := tenantV1.ConvertFrom(&tenantV2); err != nil {
					log.Println(err)
					w.WriteHeader(500)
					req.Response.Result.Status = metav1.StatusFailure
					rawResp, _ := json.Marshal(req)
					if _, err = w.Write(rawResp); err != nil {
						log.Println(err)
					}
					return
				}
				req.Response.ConvertedObjects = append(req.Response.ConvertedObjects, runtime.RawExtension{Object: &tenantV1})
			case "minio.min.io/v2":
				req.Response.ConvertedObjects = append(req.Response.ConvertedObjects, runtime.RawExtension{Object: &tenantV2})
			}
		}
	}
	// prepare to reply
	req.Response.UID = req.Request.UID
	req.Response.Result.Status = metav1.StatusSuccess
	req.Request = &apiextensionsv1.ConversionRequest{}

	rawResp, _ := json.Marshal(req)
	if _, err := w.Write(rawResp); err != nil {
		log.Println(err)
	}
}

// AssumeRoleWithWebIdentityHandler - POST /sts/{tenantNamespace}
// AssumeRoleWithWebIdentity - implementation of AWS STS API.
// Authenticates a Kubernetes Service accounts using a JWT Token
// Evalues a PolicyBinding CRD as Mapping of the Minio Policies that the ServiceAccount can assume on a minio tenant
// Eg:-
// $ curl -k -X POST https://operator:9443/sts/{tenantNamespace} -d "Action=AssumeRoleWithWebIdentity&WebIdentityToken=<jwt>" -H "Content-Type: application/x-www-form-urlencoded"
func (c *Controller) AssumeRoleWithWebIdentityHandler(w http.ResponseWriter, r *http.Request) {
	routerVars := mux.Vars(r)
	tenantNamespace := ""
	tenantNamespace, err := xhttp.UnescapeQueryPath(routerVars["tenantNamespace"])

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
		writeSTSErrorResponse(ctx, w, true, ErrSTSInvalidParameterValue, fmt.Errorf("tenant namespace is missing"))
	}

	// Parse the incoming form data.
	if err := xhttp.ParseForm(r); err != nil {
		writeSTSErrorResponse(ctx, w, true, ErrSTSInvalidParameterValue, err)
		return
	}

	if r.Form.Get(stsVersion) != stsAPIVersion {
		err := fmt.Errorf("invalid STS API version %s, expecting %s", r.Form.Get("Version"), stsAPIVersion)
		writeSTSErrorResponse(ctx, w, true, ErrSTSMissingParameter, err)
		return
	}

	action := r.Form.Get(stsAction)
	switch action {
	// For now we only do WebIdentity, leaving it in case we want to implement certificate authentication
	case webIdentity:
	default:
		writeSTSErrorResponse(ctx, w, true, ErrSTSInvalidParameterValue, fmt.Errorf("unsupported action %s", action))
		return
	}

	token := strings.TrimSpace(r.Form.Get(stsWebIdentityToken))

	if token == "" {
		writeSTSErrorResponse(ctx, w, true, ErrSTSMissingParameter, fmt.Errorf("missing %s", stsWebIdentityToken))
		return
	}

	// roleArn is ignored
	// roleArn := strings.TrimSpace(r.Form.Get(stsRoleArn))

	// VALIDATE JWT
	accessToken := r.Form.Get(stsWebIdentityToken)

	saAuthResult, err := c.ValidateServiceAccountJWT(&ctx, accessToken)
	if err != nil {
		writeSTSErrorResponse(ctx, w, true, ErrSTSInvalidIdentityToken, err)
		return
	}

	if !saAuthResult.Status.Authenticated {
		writeSTSErrorResponse(ctx, w, true, ErrSTSAccessDenied, fmt.Errorf("access denied: Invalid Token"))
		return
	}
	pbs, err := c.minioClientSet.StsV1beta1().PolicyBindings(tenantNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		writeSTSErrorResponse(ctx, w, true, ErrSTSInternalError, fmt.Errorf("error obtaining PolicyBindings: %s", err))
		return
	}

	chunks := strings.Split(strings.Replace(saAuthResult.Status.User.Username, "system:serviceaccount:", "", -1), ":")
	// saNamespace Service account Namespace
	saNamespace := chunks[0]
	// saName service account username
	saName := chunks[1]
	// Authorized PolicyBindings for the Service Account
	// Need to optimize it with a Cache (probably)
	policyBindings := []v1beta1.PolicyBinding{}
	for _, pb := range pbs.Items {
		if pb.Spec.Application.Namespace == saNamespace && pb.Spec.Application.ServiceAccount == saName {
			policyBindings = append(policyBindings, pb)
		}
	}

	if len(policyBindings) == 0 {
		writeSTSErrorResponse(ctx, w, true, ErrSTSAccessDenied, fmt.Errorf("service Account '%s' is not granted to AssumeRole in any Tenant", saAuthResult.Status.User.Username))
		return
	}

	tenants, err := c.minioClientSet.MinioV2().Tenants(tenantNamespace).List(ctx, metav1.ListOptions{})
	if err != nil || len(tenants.Items) == 0 {
		writeSTSErrorResponse(ctx, w, true, ErrSTSInvalidParameterValue, fmt.Errorf("no Tenants available in the namespace '%s'", tenantNamespace))
		return
	}

	// Only one tenant is allowed in a single namespace, gathering the first tenant in the list
	tenant := tenants.Items[0]

	// Session Policy
	sessionPolicyStr := r.Form.Get(stsPolicy)
	// The plain text that you use for both inline and managed session
	// policies shouldn't exceed 2048 characters.
	if len(sessionPolicyStr) > 2048 {
		writeSTSErrorResponse(ctx, w, true, ErrSTSPackedPolicyTooLarge, fmt.Errorf("session policy should not exceed 2048 characters"))
		return
	}

	if len(sessionPolicyStr) > 0 {
		sessionPolicy, err := iampolicy.ParseConfig(bytes.NewReader([]byte(sessionPolicyStr)))
		if err != nil {
			writeSTSErrorResponse(ctx, w, true, ErrSTSMalformedPolicyDocument, err)
			return
		}

		// Version in policy must not be empty
		if sessionPolicy.Version == "" {
			writeSTSErrorResponse(ctx, w, true, ErrSTSInvalidParameterValue, fmt.Errorf("invalid session policy version"))
			return
		}
	}

	durationStr := r.Form.Get(stsDurationSeconds)
	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		writeSTSErrorResponse(ctx, w, true, ErrSTSInvalidParameterValue, fmt.Errorf("invalid token expiry"))
	}

	if duration < 900 || duration > 31536000 {
		writeSTSErrorResponse(ctx, w, true, ErrSTSInvalidParameterValue, fmt.Errorf("invalid token expiry: min 900s, max 31536000s"))
	}

	stsCredentials, err := AssumeRole(ctx, c, &tenant, sessionPolicyStr, duration)
	if err != nil {
		writeSTSErrorResponse(ctx, w, true, ErrSTSInternalError, err)
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
