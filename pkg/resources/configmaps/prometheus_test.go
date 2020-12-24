package configmaps

import (
	"reflect"
	"testing"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestRoundTripPrometheusConfig verifies helper functions serializing/deserializing prometheus config
func TestRoundTripPrometheusConfig(t *testing.T) {
	diskCap := 5 // 5GiB
	tenant := &miniov2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tenant-a",
		},
		Spec: miniov2.TenantSpec{
			Pools: []miniov2.Pool{
				{
					Name:             "pool-0",
					Servers:          4,
					VolumesPerServer: 4,
				},
			},
			Prometheus: &miniov2.PrometheusConfig{
				DiskCapacityDB: &diskCap,
			},
		},
	}

	// Builds prometheus config from the given tenant object and minio credentials
	ak, sk := "minio", "minio123"
	pCfg := getPrometheusConfig(tenant, ak, sk)
	cfgMap := pCfg.getConfigMap(tenant)

	// Unpacks the prometheus config from the config map created above.
	newPCfg, err := fromPrometheusConfigMap(cfgMap)
	if err != nil {
		t.Fatalf("failed to get prometheus config from configmap %v", err)
	}

	if !reflect.DeepEqual(pCfg, newPCfg) {
		t.Fatalf("expected %v to be equal to %v", pCfg, newPCfg)
	}
}
