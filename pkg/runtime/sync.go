// This file is part of MinIO Operator
// Copyright (c) 2023 MinIO, Inc.
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

package runtime

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-test/deep"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ObjectSyncer is a Syncer for sync Objects only by passing a MutateFn.
type ObjectSyncer struct {
	Client   client.Client
	Ctx      context.Context
	Obj      client.Object
	Owner    client.Object
	MutateFn controllerutil.MutateFn
	SyncType SyncType
	// runtime use that
	previousObject runtime.Object
}

var _ Syncer = &ObjectSyncer{}

// NewObjectSyncer creates a new kubernetes.Object syncer for object
// will set owner. And it can set SyncType. Mostly we should set CreateOrUpdate
func NewObjectSyncer(ctx context.Context, client client.Client, owner client.Object, MutateFn controllerutil.MutateFn, obj client.Object, syncType SyncType) Syncer {
	return &ObjectSyncer{
		Ctx:      ctx,
		Client:   client,
		Obj:      obj,
		Owner:    owner,
		MutateFn: MutateFn,
		SyncType: syncType,
	}
}

// mutateFn Wrap for controllerutil.MutateFn. Do something before create or update or patch.
func (s *ObjectSyncer) mutateFn() controllerutil.MutateFn {
	return func() error {
		s.previousObject = s.Obj.DeepCopyObject()
		err := s.MutateFn()
		if err != nil {
			return err
		}
		if s.Owner == nil {
			return nil
		}
		// set owner reference only if owner resource is not being deleted, otherwise the owner
		// reference will be reset in case of deleting.
		if s.Owner.GetDeletionTimestamp().IsZero() {
			if err := controllerutil.SetControllerReference(s.Owner, s.Obj, s.Client.Scheme()); err != nil {
				return err
			}
		} else if ctime := s.Obj.GetCreationTimestamp(); ctime.IsZero() {
			// the owner is deleted, don't recreate the resource if does not exist, because gc
			// will not delete it again because has no owner reference set
			return ErrOwnerDeleted
		}

		return nil
	}
}

// objectTypeName returns the type of Object's Name
func (s *ObjectSyncer) objectTypeName(obj runtime.Object) string {
	if obj != nil {
		gvk, err := apiutil.GVKForObject(obj, s.Client.Scheme())
		if err != nil {
			return fmt.Sprintf("%T", obj)
		}
		return gvk.String()
	}
	return "nil"
}

// Sync does the actual syncing and implements the Syncer Sync method.
func (s *ObjectSyncer) Sync(ctx context.Context) (SyncResult, error) {
	var err error
	result := SyncResult{}
	key := client.ObjectKeyFromObject(s.Obj)
	switch s.SyncType {
	case SyncTypeCreateOrUpdate:
		result.Operation, err = controllerutil.CreateOrUpdate(ctx, s.Client, s.Obj, s.mutateFn())
	case SyncTypeCreateOrPatch:
		result.Operation, err = controllerutil.CreateOrPatch(ctx, s.Client, s.Obj, s.mutateFn())
	case SyncTypeFoundToUpdate:
		// found first
		key := client.ObjectKeyFromObject(s.Obj)
		if err := s.Client.Get(ctx, key, s.Obj); err != nil {
			if apierrors.IsNotFound(err) {
				result.Operation = controllerutil.OperationResultNone
				break
			}
		}
		result.Operation, err = controllerutil.CreateOrUpdate(ctx, s.Client, s.Obj, s.mutateFn())
	case SyncTypeFoundToPatch:
		// found first
		key := client.ObjectKeyFromObject(s.Obj)
		if err := s.Client.Get(ctx, key, s.Obj); err != nil {
			if apierrors.IsNotFound(err) {
				result.Operation = controllerutil.OperationResultNone
				break
			}
		}
		result.Operation, err = controllerutil.CreateOrPatch(ctx, s.Client, s.Obj, s.mutateFn())
	}

	// get the diff info
	diff := deep.Equal(redact(s.previousObject), redact(s.Obj))
	switch {
	case errors.Is(err, ErrOwnerDeleted):
		klog.Infof("%s key %s kind %s error %s", string(result.Operation), key, s.objectTypeName(s.Obj), err)
		err = nil
	case errors.Is(err, ErrIgnore):
		klog.Infof("syncer skipped  key %s kind %s error %s", key, s.objectTypeName(s.Obj), err)
		err = nil
	case err != nil:
		result.SetEventData(eventWarning, EventReason(s.Obj, err),
			fmt.Sprintf("%s %s failed syncing: %s", s.objectTypeName(s.Obj), key, err))
		klog.Errorf("%s key %s kind %s diff %s", string(result.Operation), key, s.objectTypeName(s.Obj), diff)
	default:
		result.SetEventData(eventNormal, EventReason(s.Obj, err),
			fmt.Sprintf("%s %s %s successfully", s.objectTypeName(s.Obj), key, result.Operation))
		// Only print this log if there is a difference to show for the sync
		// otherwise keep it like this to avoid flooding the Operator log with unchanged difference.
		if diff != nil {
			klog.Infof("%s key %s kind %s diff %s", string(result.Operation), key, s.objectTypeName(s.Obj), diff)
		}
	}

	return result, err
}

// ObjectOwner returns the ObjectSyncer owner.
func (s *ObjectSyncer) ObjectOwner() runtime.Object {
	return s.Owner
}

// Masking of sensitive information
func redact(obj runtime.Object) runtime.Object {
	switch exposed := obj.(type) {
	case *corev1.Secret:
		redacted := exposed.DeepCopy()
		redacted.Data = nil
		redacted.StringData = nil
		exposed.ObjectMeta.DeepCopyInto(&redacted.ObjectMeta)
		return redacted
	case *corev1.ConfigMap:
		redacted := exposed.DeepCopy()
		redacted.Data = nil
		return redacted
	}
	return obj
}
