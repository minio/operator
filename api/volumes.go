// This file is part of MinIO Operator
// Copyright (c) 2021 MinIO, Inc.
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
	"sort"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"

	"github.com/go-openapi/runtime/middleware"
	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	v1 "k8s.io/api/certificates/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func registerVolumesHandlers(api *operations.OperatorAPI) {
	api.OperatorAPIListPVCsHandler = operator_api.ListPVCsHandlerFunc(func(params operator_api.ListPVCsParams, session *models.Principal) middleware.Responder {
		payload, err := getPVCsResponse(session, params)
		if err != nil {
			return operator_api.NewListPVCsDefault(int(err.Code)).WithPayload(err)
		}

		return operator_api.NewListPVCsOK().WithPayload(payload)
	})

	api.OperatorAPIListPVCsForTenantHandler = operator_api.ListPVCsForTenantHandlerFunc(func(params operator_api.ListPVCsForTenantParams, session *models.Principal) middleware.Responder {
		payload, err := getPVCsForTenantResponse(session, params)
		if err != nil {
			return operator_api.NewListPVCsForTenantDefault(int(err.Code)).WithPayload(err)
		}

		return operator_api.NewListPVCsForTenantOK().WithPayload(payload)
	})

	api.OperatorAPIListTenantCertificateSigningRequestHandler = operator_api.ListTenantCertificateSigningRequestHandlerFunc(func(params operator_api.ListTenantCertificateSigningRequestParams, session *models.Principal) middleware.Responder {
		payload, err := getTenantCSResponse(session, params)
		if err != nil {
			return operator_api.NewListTenantCertificateSigningRequestDefault(int(err.Code)).WithPayload(err)
		}

		return operator_api.NewListTenantCertificateSigningRequestOK().WithPayload(payload)
	})

	api.OperatorAPIDeletePVCHandler = operator_api.DeletePVCHandlerFunc(func(params operator_api.DeletePVCParams, session *models.Principal) middleware.Responder {
		err := getDeletePVCResponse(session, params)
		if err != nil {
			return operator_api.NewDeletePVCDefault(int(err.Code)).WithPayload(err)
		}
		return nil
	})

	api.OperatorAPIGetPVCEventsHandler = operator_api.GetPVCEventsHandlerFunc(func(params operator_api.GetPVCEventsParams, session *models.Principal) middleware.Responder {
		payload, err := getPVCEventsResponse(session, params)
		if err != nil {
			return operator_api.NewGetPVCEventsDefault(int(err.Code)).WithPayload(err)
		}

		return operator_api.NewGetPVCEventsOK().WithPayload(payload)
	})

	api.OperatorAPIGetPVCDescribeHandler = operator_api.GetPVCDescribeHandlerFunc(func(params operator_api.GetPVCDescribeParams, session *models.Principal) middleware.Responder {
		payload, err := getPVCDescribeResponse(session, params)
		if err != nil {
			return operator_api.NewGetPVCDescribeDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetPVCDescribeOK().WithPayload(payload)
	})
}

func getPVCsResponse(session *models.Principal, params operator_api.ListPVCsParams) (*models.ListPVCsResponse, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	clientset, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}

	// Filter Tenant PVCs. They keep their v1 tenant annotation
	listOpts := metav1.ListOptions{
		LabelSelector: miniov2.TenantLabel,
	}

	// List all PVCs
	listAllPvcs, err2 := clientset.CoreV1().PersistentVolumeClaims("").List(ctx, listOpts)

	if err2 != nil {
		return nil, ErrorWithContext(ctx, err2)
	}

	var ListPVCs []*models.PvcsListResponse

	for _, pvc := range listAllPvcs.Items {
		status := string(pvc.Status.Phase)
		if pvc.DeletionTimestamp != nil {
			status = "Terminating"
		}
		pvcResponse := models.PvcsListResponse{
			Name:         pvc.Name,
			Age:          pvc.CreationTimestamp.String(),
			Capacity:     pvc.Status.Capacity.Storage().String(),
			Namespace:    pvc.Namespace,
			Status:       status,
			StorageClass: *pvc.Spec.StorageClassName,
			Volume:       pvc.Spec.VolumeName,
			Tenant:       pvc.Labels["v1.min.io/tenant"],
		}
		ListPVCs = append(ListPVCs, &pvcResponse)
	}

	PVCsResponse := models.ListPVCsResponse{
		Pvcs: ListPVCs,
	}

	return &PVCsResponse, nil
}

func getPVCsForTenantResponse(session *models.Principal, params operator_api.ListPVCsForTenantParams) (*models.ListPVCsResponse, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	clientset, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}

	// Filter Tenant PVCs. They keep their v1 tenant annotation
	listOpts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("v1.min.io/tenant=%s", params.Tenant),
	}

	// List all PVCs
	listAllPvcs, err2 := clientset.CoreV1().PersistentVolumeClaims(params.Namespace).List(ctx, listOpts)

	if err2 != nil {
		return nil, ErrorWithContext(ctx, err2)
	}

	var ListPVCs []*models.PvcsListResponse

	for _, pvc := range listAllPvcs.Items {
		status := string(pvc.Status.Phase)
		if pvc.DeletionTimestamp != nil {
			status = "Terminating"
		}
		pvcResponse := models.PvcsListResponse{
			Name:         pvc.Name,
			Age:          pvc.CreationTimestamp.String(),
			Capacity:     pvc.Status.Capacity.Storage().String(),
			Namespace:    pvc.Namespace,
			Status:       status,
			StorageClass: *pvc.Spec.StorageClassName,
			Volume:       pvc.Spec.VolumeName,
			Tenant:       pvc.Labels["v1.min.io/tenant"],
		}
		ListPVCs = append(ListPVCs, &pvcResponse)
	}

	PVCsResponse := models.ListPVCsResponse{
		Pvcs: ListPVCs,
	}

	return &PVCsResponse, nil
}

func getDeletePVCResponse(session *models.Principal, params operator_api.DeletePVCParams) *models.Error {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	// get Kubernetes Client
	clientset, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return ErrorWithContext(ctx, err)
	}
	listOpts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("v1.min.io/tenant=%s", params.Tenant),
		FieldSelector: fmt.Sprintf("metadata.name=%s", params.PVCName),
	}
	if err = clientset.CoreV1().PersistentVolumeClaims(params.Namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, listOpts); err != nil {
		return ErrorWithContext(ctx, err)
	}
	return nil
}

func getPVCEventsResponse(session *models.Principal, params operator_api.GetPVCEventsParams) (models.EventListWrapper, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	clientset, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	PVC, err := clientset.CoreV1().PersistentVolumeClaims(params.Namespace).Get(ctx, params.PVCName, metav1.GetOptions{})
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	events, err := clientset.CoreV1().Events(params.Namespace).List(ctx, metav1.ListOptions{FieldSelector: fmt.Sprintf("involvedObject.uid=%s", PVC.UID)})
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	retval := models.EventListWrapper{}
	for i := 0; i < len(events.Items); i++ {
		retval = append(retval, &models.EventListElement{
			Namespace: events.Items[i].Namespace,
			LastSeen:  events.Items[i].LastTimestamp.Unix(),
			Message:   events.Items[i].Message,
			EventType: events.Items[i].Type,
			Reason:    events.Items[i].Reason,
		})
	}
	sort.SliceStable(retval, func(i int, j int) bool {
		return retval[i].LastSeen < retval[j].LastSeen
	})
	return retval, nil
}

func getTenantCSResponse(session *models.Principal, params operator_api.ListTenantCertificateSigningRequestParams) (*models.CsrElements, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	clientset, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}

	// Get CSRs by Label "v1.min.io/tenant=" + params.Tenant
	listByTenantLabel := metav1.ListOptions{LabelSelector: "v1.min.io/tenant=" + params.Tenant}
	listResult, listError := clientset.CertificatesV1().CertificateSigningRequests().List(ctx, listByTenantLabel)
	if listError != nil {
		return nil, ErrorWithContext(ctx, listError)
	}

	// Get CSR by label "v1.min.io/kes=" + params.Tenant + "-kes"
	listByKESLabel := metav1.ListOptions{LabelSelector: "v1.min.io/kes=" + params.Tenant + "-kes"}
	listKESResult, listKESError := clientset.CertificatesV1().CertificateSigningRequests().List(ctx, listByKESLabel)
	if listKESError != nil {
		return nil, ErrorWithContext(ctx, listKESError)
	}

	var listOfCSRs []v1.CertificateSigningRequest
	for index := 0; index < len(listResult.Items); index++ {
		listOfCSRs = append(listOfCSRs, listResult.Items[index])
	}
	for index := 0; index < len(listKESResult.Items); index++ {
		listOfCSRs = append(listOfCSRs, listKESResult.Items[index])
	}

	var arrayElements []*models.CsrElement
	for index := 0; index < len(listOfCSRs); index++ {
		csrResult := listOfCSRs[index]
		annotations := []*models.Annotation{}
		for k, v := range csrResult.ObjectMeta.Annotations {
			annotations = append(annotations, &models.Annotation{Key: k, Value: v})
		}
		var DeletionGracePeriodSeconds int64
		DeletionGracePeriodSeconds = 0
		if csrResult.ObjectMeta.DeletionGracePeriodSeconds != nil {
			DeletionGracePeriodSeconds = *csrResult.ObjectMeta.DeletionGracePeriodSeconds
		}
		messages := ""
		// A CSR.Status can contain multiple Conditions
		for i := 0; i < len(csrResult.Status.Conditions); i++ {
			messages = messages + " " + csrResult.Status.Conditions[i].Message
		}
		retval := &models.CsrElement{
			Name:                       csrResult.ObjectMeta.Name,
			Annotations:                annotations,
			DeletionGracePeriodSeconds: DeletionGracePeriodSeconds,
			GenerateName:               csrResult.ObjectMeta.GenerateName,
			Generation:                 csrResult.ObjectMeta.Generation,
			Namespace:                  csrResult.ObjectMeta.Namespace,
			ResourceVersion:            csrResult.ObjectMeta.ResourceVersion,
			Status:                     messages,
		}
		arrayElements = append(arrayElements, retval)
	}
	result := &models.CsrElements{CsrElement: arrayElements}
	return result, nil
}

func getPVCDescribeResponse(session *models.Principal, params operator_api.GetPVCDescribeParams) (*models.DescribePVCWrapper, *models.Error) {
	clientSet, err := K8sClient(session.STSSessionToken)
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	k8sClient := k8sClient{client: clientSet}
	return getPVCDescribe(ctx, params.Namespace, params.PVCName, &k8sClient)
}

func getPVCDescribe(ctx context.Context, namespace string, pvcName string, clientSet K8sClientI) (*models.DescribePVCWrapper, *models.Error) {
	pvc, err := clientSet.getPVC(ctx, namespace, pvcName, metav1.GetOptions{})
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	accessModes := []string{}
	for _, a := range pvc.Status.AccessModes {
		accessModes = append(accessModes, string(a))
	}
	return &models.DescribePVCWrapper{
		Name:         pvc.Name,
		Namespace:    pvc.Namespace,
		StorageClass: *pvc.Spec.StorageClassName,
		Status:       string(pvc.Status.Phase),
		Volume:       pvc.Spec.VolumeName,
		Labels:       castLabels(pvc.Labels),
		Annotations:  castAnnotations(pvc.Annotations),
		Finalizers:   pvc.Finalizers,
		Capacity:     pvc.Status.Capacity.Storage().String(),
		AccessModes:  accessModes,
		VolumeMode:   string(*pvc.Spec.VolumeMode),
	}, nil
}

func castLabels(labelsToCast map[string]string) (labels []*models.Label) {
	for k, v := range labelsToCast {
		labels = append(labels, &models.Label{Key: k, Value: v})
	}
	return labels
}

func castAnnotations(annotationsToCast map[string]string) (annotations []*models.Annotation) {
	for k, v := range annotationsToCast {
		annotations = append(annotations, &models.Annotation{Key: k, Value: v})
	}
	return annotations
}
