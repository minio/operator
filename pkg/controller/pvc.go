package controller

import (
	"context"
	"fmt"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/resources/statefulsets"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TryToDeletePVCS - try to delete pvc if set ReclaimStorageLabel:true
func (c *Controller) TryToDeletePVCS(ctx context.Context, namespace string, tenantName string) {
	pvcList := corev1.PersistentVolumeClaimList{}
	listOpt := client.ListOptions{
		Namespace: namespace,
	}
	client.MatchingLabels{
		"v1.min.io/tenant": tenantName,
	}.ApplyToList(&listOpt)
	err := c.k8sClient.List(ctx, &pvcList, &listOpt)
	if err != nil {
		runtime.HandleError(fmt.Errorf("PersistentVolumeClaimList  '%s/%s' error:%s", namespace, tenantName, err.Error()))
	}
	for _, pvc := range pvcList.Items {
		if pvc.Labels[statefulsets.ReclaimStorageLabel] == "true" {
			err := c.k8sClient.Delete(ctx, &pvc)
			if err != nil {
				runtime.HandleError(fmt.Errorf("Delete PersistentVolumeClaim '%s/%s/%s' error:%s", namespace, tenantName, pvc.Name, err.Error()))
			}
		}
	}
}

// ResizePVCS - try to resize pvc to Request+AdditionalStorage if set AdditionalStorage to pool
func (c *Controller) ResizePVCS(ctx context.Context, tenant *miniov2.Tenant) {
	fmt.Println("try to resizePVC now")
	for _, pool := range tenant.Spec.Pools {
		if pool.AdditionalStorage != nil {
			q, err := resource.ParseQuantity(*pool.AdditionalStorage)
			if err != nil {
				// if parse error. Continue
				fmt.Printf("ParseQuantity %s error: %s \n", *pool.AdditionalStorage, err.Error())
				continue
			}
			storageRequest := pool.VolumeClaimTemplate.Spec.Resources.Requests.Storage()
			if storageRequest != nil {
				q.Add(*storageRequest)
			}
			pvcList := corev1.PersistentVolumeClaimList{}
			listOpt := client.ListOptions{
				Namespace: tenant.Namespace,
			}
			client.MatchingLabels{
				"v1.min.io/tenant": tenant.Name,
				"v1.min.io/pool":   pool.Name,
			}.ApplyToList(&listOpt)
			err = c.k8sClient.List(ctx, &pvcList, &listOpt)
			if err != nil {
				runtime.HandleError(fmt.Errorf("PersistentVolumeClaimList '%s/%s/%s' error:%s", tenant.Namespace, tenant.Name, pool.Name, err.Error()))
			}
			fmt.Println("Get pvc number:", len(pvcList.Items))
			for _, pvc := range pvcList.Items {
				// if already resized with a bigger or equal size, do nothing
				if pvc.Spec.Resources.Requests.Storage().Cmp(q) != -1 {
					fmt.Println("ignore to resize:")
					continue
				}
				pvc.Spec.Resources.Requests[corev1.ResourceStorage] = q
				err := c.k8sClient.Update(ctx, &pvc)
				if err != nil {
					runtime.HandleError(fmt.Errorf("Update PersistentVolumeClaim '%s/%s' to %s error:%s", tenant.Namespace, pvc.Name, q.String(), err.Error()))
				} else {
					fmt.Println("resize success")
				}
			}
		} else {
			fmt.Println("nothing to resize for ", pool.Name, " empty here")
		}
	}
}
