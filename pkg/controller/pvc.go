package controller

import (
	"context"
	"fmt"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// ExpandPVCs expands the PVCs for a given tenant
func ExpandPVCs(ctx context.Context, kubeClientSet kubernetes.Interface, tenant *miniov2.Tenant, namespace string) error {
	uOpts := metav1.UpdateOptions{}

	for _, pool := range tenant.Spec.Pools {
		opts := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s,%s=%s", miniov2.TenantLabel, tenant.Name, miniov2.PoolLabel, pool.Name),
		}
		pvcList, err := kubeClientSet.CoreV1().PersistentVolumeClaims(namespace).List(ctx, opts)
		if err != nil {
			return err
		}

		for _, pvc := range pvcList.Items {
			if pool.VolumeClaimTemplate != nil {
				requestedStorage := pool.VolumeClaimTemplate.Spec.Resources.Requests[corev1.ResourceStorage]
				currentStorage := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
				if requestedStorage.Cmp(currentStorage) > 0 {
					pvc.Spec.Resources.Requests[corev1.ResourceStorage] = requestedStorage
					_, err = kubeClientSet.CoreV1().PersistentVolumeClaims(namespace).Update(ctx, &pvc, uOpts)
					if err != nil {
						return err
					}
					klog.Infof("Expanded PVC %s from %s to %s", pvc.Name, currentStorage.String(), requestedStorage.String())
				}
			}
		}
	}

	return nil
}
