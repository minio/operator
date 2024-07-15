package controller

import (
	"context"
	"testing"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestExpandPVCs(t *testing.T) {
	ctx := context.Background()

	tenant := &miniov2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-tenant",
		},
		Spec: miniov2.TenantSpec{
			Pools: []miniov2.Pool{
				{
					Name: "pool1",
					VolumeClaimTemplate: &corev1.PersistentVolumeClaim{
						Spec: corev1.PersistentVolumeClaimSpec{
							Resources: corev1.VolumeResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceStorage: resource.MustParse("2Gi"),
								},
							},
						},
					},
				},
			},
		},
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pvc",
			Namespace: "test-namespace",
			Labels: map[string]string{
				miniov2.TenantLabel: tenant.Name,
				miniov2.PoolLabel:   "pool1",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		},
	}

	kubeClient := fake.NewSimpleClientset(pvc)

	err := ExpandPVCs(ctx, kubeClient, tenant, "test-namespace")
	if err != nil {
		t.Fatalf("ExpandPVCs failed: %v", err)
	}

	updatedPVC, err := kubeClient.CoreV1().PersistentVolumeClaims("test-namespace").Get(ctx, "test-pvc", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get PVC: %v", err)
	}

	if updatedPVC.Spec.Resources.Requests[corev1.ResourceStorage] != resource.MustParse("2Gi") {
		t.Errorf("Expected PVC storage to be 2Gi, but got %v", updatedPVC.Spec.Resources.Requests[corev1.ResourceStorage])
	}
}
