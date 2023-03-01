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
	"errors"
	"sort"

	"github.com/minio/minio-go/v7/pkg/set"

	"github.com/minio/operator/api/operations/operator_api"

	"github.com/go-openapi/runtime/middleware"
	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func registerNodesHandlers(api *operations.OperatorAPI) {
	api.OperatorAPIGetMaxAllocatableMemHandler = operator_api.GetMaxAllocatableMemHandlerFunc(func(params operator_api.GetMaxAllocatableMemParams, principal *models.Principal) middleware.Responder {
		resp, err := getMaxAllocatableMemoryResponse(params.HTTPRequest.Context(), principal, params.NumNodes)
		if err != nil {
			return operator_api.NewGetMaxAllocatableMemDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetMaxAllocatableMemOK().WithPayload(resp)
	})

	api.OperatorAPIListNodeLabelsHandler = operator_api.ListNodeLabelsHandlerFunc(func(params operator_api.ListNodeLabelsParams, principal *models.Principal) middleware.Responder {
		resp, err := getNodeLabelsResponse(params.HTTPRequest.Context(), principal)
		if err != nil {
			return operator_api.NewListNodeLabelsDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewListNodeLabelsOK().WithPayload(*resp)
	})

	api.OperatorAPIGetAllocatableResourcesHandler = operator_api.GetAllocatableResourcesHandlerFunc(func(params operator_api.GetAllocatableResourcesParams, session *models.Principal) middleware.Responder {
		resp, err := getAllocatableResourcesResponse(session, params)
		if err != nil {
			return operator_api.NewGetAllocatableResourcesDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetAllocatableResourcesOK().WithPayload(resp)
	})
}

// NodeResourceInfo node information
type NodeResourceInfo struct {
	Name              string
	AllocatableMemory int64
	AllocatableCPU    int64
}

// getMaxAllocatableMemory get max allocatable memory given a desired number of nodes
func getMaxAllocatableMemory(ctx context.Context, clientset v1.CoreV1Interface, numNodes int32) (*models.MaxAllocatableMemResponse, error) {
	// can't request less than 4 nodes
	if numNodes < 4 {
		return nil, ErrFewerThanFourNodes
	}

	// get all nodes from cluster
	nodes, err := clientset.Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// requesting more nodes than are schedulable in the cluster
	schedulableNodes := len(nodes.Items)
	nonMasterNodes := len(nodes.Items)
	for _, node := range nodes.Items {
		// check taints to check if node is schedulable
		for _, taint := range node.Spec.Taints {
			if taint.Effect == corev1.TaintEffectNoSchedule {
				schedulableNodes--
			}
			// check if the node is a master
			if taint.Key == "node-role.kubernetes.io/master" {
				nonMasterNodes--
			}
		}
	}
	// requesting more nodes than schedulable and less than total number of workers
	if int(numNodes) > schedulableNodes && int(numNodes) < nonMasterNodes {
		return nil, ErrTooManyNodes
	}
	if nonMasterNodes < int(numNodes) {
		return nil, ErrTooFewNodes
	}

	// not enough schedulable nodes
	if schedulableNodes < int(numNodes) {
		return nil, ErrTooFewAvailableNodes
	}

	availableMemSizes := []int64{}
OUTER:
	for _, n := range nodes.Items {
		// Don't consider node if it has a NoSchedule or NoExecute Taint
		for _, t := range n.Spec.Taints {
			switch t.Effect {
			case corev1.TaintEffectNoSchedule:
				continue OUTER
			case corev1.TaintEffectNoExecute:
				continue OUTER
			default:
				continue
			}
		}
		if quantity, ok := n.Status.Allocatable[corev1.ResourceMemory]; ok {
			availableMemSizes = append(availableMemSizes, quantity.Value())
		}
	}

	maxAllocatableMemory := getMaxClusterMemory(numNodes, availableMemSizes)

	res := &models.MaxAllocatableMemResponse{
		MaxMemory: maxAllocatableMemory,
	}

	return res, nil
}

// getMaxClusterMemory returns the maximum memory size that can be used
// across numNodes (number of nodes)
func getMaxClusterMemory(numNodes int32, nodesMemorySizes []int64) int64 {
	if int32(len(nodesMemorySizes)) < numNodes || numNodes == 0 {
		return 0
	}

	// sort nodesMemorySizes int64 array
	sort.Slice(nodesMemorySizes, func(i, j int) bool { return nodesMemorySizes[i] < nodesMemorySizes[j] })
	maxIndex := 0
	maxAllocatableMemory := nodesMemorySizes[maxIndex]

	for i, size := range nodesMemorySizes {
		// maxAllocatableMemory is the minimum value of nodesMemorySizes array
		// only within the size of numNodes, if more nodes are available
		// then the maxAllocatableMemory is equal to the next minimum value
		// on the sorted nodesMemorySizes array.
		// e.g. with numNodes = 4;
		//   			maxAllocatableMemory of [2,4,8,8] => 2
		//      		maxAllocatableMemory of [2,4,8,8,16] => 4
		if int32(i) < numNodes {
			maxAllocatableMemory = min(maxAllocatableMemory, size)
		} else {
			maxIndex++
			maxAllocatableMemory = nodesMemorySizes[maxIndex]
		}
	}
	return maxAllocatableMemory
}

// min returns the smaller of x or y.
func min(x, y int64) int64 {
	if x > y {
		return y
	}
	return x
}

func getMaxAllocatableMemoryResponse(ctx context.Context, session *models.Principal, numNodes int32) (*models.MaxAllocatableMemResponse, *models.Error) {
	client, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}

	clusterResources, err := getMaxAllocatableMemory(ctx, client.CoreV1(), numNodes)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return clusterResources, nil
}

func getNodeLabels(ctx context.Context, clientset v1.CoreV1Interface) (*models.NodeLabels, error) {
	// get all nodes from cluster
	nodes, err := clientset.Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	// make a map[string]set to avoid duplicate values
	keyValueSet := map[string]set.StringSet{}

	for _, node := range nodes.Items {
		for k, v := range node.Labels {
			if _, ok := keyValueSet[k]; !ok {
				keyValueSet[k] = set.NewStringSet()
			}
			keyValueSet[k].Add(v)
		}
	}

	// convert to output
	res := models.NodeLabels{}
	for k, valSet := range keyValueSet {
		res[k] = valSet.ToSlice()
	}

	return &res, nil
}

func getNodeLabelsResponse(ctx context.Context, session *models.Principal) (*models.NodeLabels, *models.Error) {
	client, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}

	clusterResources, err := getNodeLabels(ctx, client.CoreV1())
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return clusterResources, nil
}

func getClusterResourcesInfo(numNodes int32, inNodesResources []NodeResourceInfo) *models.AllocatableResourcesResponse {
	// purge any nodes with 0 cpu
	var nodesResources []NodeResourceInfo
	for _, n := range inNodesResources {
		if n.AllocatableCPU > 0 {
			nodesResources = append(nodesResources, n)
		}
	}

	if int32(len(nodesResources)) < numNodes || numNodes == 0 {
		return &models.AllocatableResourcesResponse{
			CPUPriority: &models.NodeMaxAllocatableResources{
				MaxAllocatableCPU: 0,
				MaxAllocatableMem: 0,
			},
			MemPriority: &models.NodeMaxAllocatableResources{
				MaxAllocatableCPU: 0,
				MaxAllocatableMem: 0,
			},
			MinAllocatableCPU: 0,
			MinAllocatableMem: 0,
		}
	}

	allocatableResources := &models.AllocatableResourcesResponse{}

	// sort nodesResources giving CPU priority
	sort.Slice(nodesResources, func(i, j int) bool { return nodesResources[i].AllocatableCPU < nodesResources[j].AllocatableCPU })
	maxCPUNodesNeeded := len(nodesResources) - int(numNodes)
	maxMemNodesNeeded := maxCPUNodesNeeded

	maxAllocatableCPU := nodesResources[maxCPUNodesNeeded].AllocatableCPU
	minAllocatableCPU := nodesResources[maxCPUNodesNeeded].AllocatableCPU
	minAllocatableMem := nodesResources[maxMemNodesNeeded].AllocatableMemory

	availableMemsForMaxCPU := []int64{}
	for _, info := range nodesResources {
		if info.AllocatableCPU >= maxAllocatableCPU {
			availableMemsForMaxCPU = append(availableMemsForMaxCPU, info.AllocatableMemory)
		}
		// min allocatable resources overall
		minAllocatableCPU = min(minAllocatableCPU, info.AllocatableCPU)
		minAllocatableMem = min(minAllocatableMem, info.AllocatableMemory)
	}

	sort.Slice(availableMemsForMaxCPU, func(i, j int) bool { return availableMemsForMaxCPU[i] < availableMemsForMaxCPU[j] })
	maxAllocatableMem := availableMemsForMaxCPU[len(availableMemsForMaxCPU)-int(numNodes)]

	allocatableResources.MinAllocatableCPU = minAllocatableCPU
	allocatableResources.MinAllocatableMem = minAllocatableMem
	allocatableResources.CPUPriority = &models.NodeMaxAllocatableResources{
		MaxAllocatableCPU: maxAllocatableCPU,
		MaxAllocatableMem: maxAllocatableMem,
	}

	// sort nodesResources giving Mem priority
	sort.Slice(nodesResources, func(i, j int) bool { return nodesResources[i].AllocatableMemory < nodesResources[j].AllocatableMemory })
	maxMemNodesNeeded = len(nodesResources) - int(numNodes)
	maxAllocatableMem = nodesResources[maxMemNodesNeeded].AllocatableMemory

	availableCPUsForMaxMem := []int64{}
	for _, info := range nodesResources {
		if info.AllocatableMemory >= maxAllocatableMem {
			availableCPUsForMaxMem = append(availableCPUsForMaxMem, info.AllocatableCPU)
		}
	}

	sort.Slice(availableCPUsForMaxMem, func(i, j int) bool { return availableCPUsForMaxMem[i] < availableCPUsForMaxMem[j] })
	maxAllocatableCPU = availableCPUsForMaxMem[len(availableCPUsForMaxMem)-int(numNodes)]

	allocatableResources.MemPriority = &models.NodeMaxAllocatableResources{
		MaxAllocatableCPU: maxAllocatableCPU,
		MaxAllocatableMem: maxAllocatableMem,
	}

	return allocatableResources
}

// getAllocatableResources get max allocatable memory given a desired number of nodes
func getAllocatableResources(ctx context.Context, clientset v1.CoreV1Interface, numNodes int32) (*models.AllocatableResourcesResponse, error) {
	if numNodes == 0 {
		return nil, errors.New("error NumNodes must be greated than 0")
	}

	// get all nodes from cluster
	nodes, err := clientset.Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	nodesInfo := []NodeResourceInfo{}
OUTER:
	for _, n := range nodes.Items {
		// Don't consider node if it has a NoSchedule or NoExecute Taint
		for _, t := range n.Spec.Taints {
			switch t.Effect {
			case corev1.TaintEffectNoSchedule:
				continue OUTER
			case corev1.TaintEffectNoExecute:
				continue OUTER
			default:
				continue
			}
		}

		var nodeMemory int64
		var nodeCPU int64
		if quantity, ok := n.Status.Allocatable[corev1.ResourceMemory]; ok {
			// availableMemSizes = append(availableMemSizes, quantity.Value())
			nodeMemory = quantity.Value()
		}
		// we assume all nodes have allocatable cpu resource
		if quantity, ok := n.Status.Allocatable[corev1.ResourceCPU]; ok {
			// availableCPU = append(availableCPU, quantity.Value())
			nodeCPU = quantity.Value()
		}
		nodeInfo := NodeResourceInfo{
			Name:              n.Name,
			AllocatableCPU:    nodeCPU,
			AllocatableMemory: nodeMemory,
		}
		nodesInfo = append(nodesInfo, nodeInfo)
	}
	res := getClusterResourcesInfo(numNodes, nodesInfo)

	return res, nil
}

// Get allocatable resources response

func getAllocatableResourcesResponse(session *models.Principal, params operator_api.GetAllocatableResourcesParams) (*models.AllocatableResourcesResponse, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	client, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	clusterResources, err := getAllocatableResources(ctx, client.CoreV1(), params.NumNodes)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return clusterResources, nil
}
