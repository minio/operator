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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	miniov1 "github.com/minio/operator/pkg/apis/minio.min.io/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/minio/operator/pkg/resources/statefulsets"

	"github.com/gorilla/mux"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
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
	tenant, err := c.tenantsLister.Tenants(namespace).Get(name)
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
	tenant, err := c.tenantsLister.Tenants(namespace).Get(name)
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
	// correct all statefulset names by loading them, this will fix their name on the tenant pool namess
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
		target := fmt.Sprintf("%s://%s:%s%s/%s/%s",
			"https",
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
			//tenantV1 := obj.Object.(*miniov1.Tenant)
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
				//convert to v2
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
				//convert to v1
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
