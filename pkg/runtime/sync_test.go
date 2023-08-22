// Copyright (C) 2023, MinIO, Inc.
//
// This code is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License, version 3,
// as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License, version 3,
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package runtime

import (
	"context"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func Test_NewObjectSyncer_CreateOrPatch_Patch(t *testing.T) {
	// test patch
	patchReg := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "test-namespace",
			Annotations: map[string]string{
				"an1": "vn1",
				"an2": "vn2",
			},
		},
		Immutable: nil,
		Data: map[string][]byte{
			"before": []byte("before"),
		},
	}
	// save the obj
	cli := fake.NewClientBuilder().WithRuntimeObjects(patchReg).WithScheme(scheme.Scheme).Build()
	patch := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "test-namespace",
		},
	}
	sync, err := NewObjectSyncer(context.Background(), cli, nil, func() error {
		if patch.Data == nil {
			patch.Data = map[string][]byte{}
		}
		patch.Data["after"] = []byte("after")
		if patch.Annotations == nil {
			patch.Annotations = map[string]string{}
		}
		patch.Annotations["an3"] = "vn3"
		return nil
	}, patch, SyncTypeCreateOrPatch).Sync(context.Background())
	if err != nil {
		return
	}
	if sync.Operation != controllerutil.OperationResultUpdated || patch.ResourceVersion == "" {
		t.Errorf("should make a update call")
	}
	if !reflect.DeepEqual(patch.Data, map[string][]byte{
		"after":  []byte("after"),
		"before": []byte("before"),
	}) {
		t.Errorf("patch failed")
	}
	if !reflect.DeepEqual(patch.Annotations, map[string]string{
		"an1": "vn1",
		"an2": "vn2",
		"an3": "vn3",
	}) {
		t.Errorf("patch failed")
	}
	t.Log(sync.EventReason)
}

func Test_NewObjectSyncer_CreateOrUpdate_Update(t *testing.T) {
	// test update
	updateReg := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "test-namespace",
			Annotations: map[string]string{
				"an1": "vn1",
				"an2": "vn2",
			},
		},
		Data: map[string][]byte{
			"before": []byte("before"),
		},
	}
	// save the obj
	cli := fake.NewClientBuilder().WithRuntimeObjects(updateReg).WithScheme(scheme.Scheme).Build()
	update := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "test-namespace",
		},
	}
	sync, err := NewObjectSyncer(context.Background(), cli, nil, func() error {
		if update.Data == nil {
			update.Data = map[string][]byte{}
		}
		update.Data["after"] = []byte("after")
		if update.Annotations == nil {
			update.Annotations = map[string]string{}
		}
		update.Annotations["an3"] = "vn3"
		return nil
	}, update, SyncTypeCreateOrUpdate).Sync(context.Background())
	if err != nil {
		return
	}
	if sync.Operation != controllerutil.OperationResultUpdated || update.ResourceVersion == "" {
		t.Errorf("should make a update call")
	}
	if !reflect.DeepEqual(update.Data, map[string][]byte{
		"after":  []byte("after"),
		"before": []byte("before"),
	}) {
		t.Errorf("update failed")
	}
	if !reflect.DeepEqual(update.Annotations, map[string]string{
		"an1": "vn1",
		"an2": "vn2",
		"an3": "vn3",
	}) {
		t.Errorf("update failed")
	}
	t.Log(sync.EventReason)
}

func Test_NewObjectSyncer_CreateOrUpdate_Create(t *testing.T) {
	// test create
	create := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "test-namespace",
		},
	}
	data := map[string][]byte{
		"create": []byte("create_data"),
	}
	cli := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
	sync, err := NewObjectSyncer(context.Background(), cli, nil, func() error {
		create.Data = data
		return nil
	}, create, SyncTypeCreateOrUpdate).Sync(context.Background())
	if err != nil {
		return
	}
	if sync.Operation != controllerutil.OperationResultCreated || create.ResourceVersion == "" {
		t.Errorf("should make a create call")
	}
	if !reflect.DeepEqual(create.Data, data) {
		t.Errorf("create failed")
	}
	t.Log(sync.EventReason)
}

type wrapK8sClientCanceledTest struct {
	client.Client
}

func (w *wrapK8sClientCanceledTest) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return context.Canceled
}

func Test_NewObjectSyncer_CreateOrUpdate_CreateCanceled(t *testing.T) {
	// test create
	create := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "test-namespace",
		},
	}
	data := map[string][]byte{
		"create": []byte("create_data"),
	}
	cli := &wrapK8sClientCanceledTest{fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()}
	sync, err := NewObjectSyncer(context.Background(), cli, nil, func() error {
		create.Data = data
		return nil
	}, create, SyncTypeCreateOrUpdate).Sync(context.Background())
	if err == nil {
		return
	}
	if sync.Operation != controllerutil.OperationResultNone || create.ResourceVersion != "" {
		t.Errorf("shouldn't make a create call")
	}
	t.Log(sync.EventReason)
}
