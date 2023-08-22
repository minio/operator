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
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	eventNormal  = "Normal"
	eventWarning = "Warning"
)

var (
	// ErrOwnerDeleted is returned when the object owner is marked for deletion.
	ErrOwnerDeleted = fmt.Errorf("owner is deleted")

	// ErrIgnore is returned for ignored errors.
	// Ignored errors are treated by the syncer as successful syncs.
	ErrIgnore = fmt.Errorf("ignored error")
)

// SyncType is for controlling syncer performance
type SyncType string

const (
	// SyncTypeCreateOrUpdate - if not found will create, if existing will update the object
	SyncTypeCreateOrUpdate = SyncType("CreateOrUpdate")
	// SyncTypeCreateOrPatch - if not found will create, if existing will patch the object
	SyncTypeCreateOrPatch = SyncType("CreateOrPatch")
	// SyncTypeFoundToUpdate - if not found will do nothing, if existing will update the object
	SyncTypeFoundToUpdate = SyncType("FoundToUpdate")
	// SyncTypeFoundToPatch - if not found will do nothing, if existing will patch the object
	SyncTypeFoundToPatch = SyncType("FoundToUpPatch")
)

// Syncer is for sync Object's action.
type Syncer interface {
	Sync(context.Context) (SyncResult, error)
	ObjectOwner() runtime.Object
}

// SyncResult is a result of an Sync.
type SyncResult struct {
	Operation    controllerutil.OperationResult
	EventType    string
	EventReason  string
	EventMessage string
}

// SetEventData sets event data on an SyncResult.
func (r *SyncResult) SetEventData(eventType, reason, message string) {
	r.EventType = eventType
	r.EventReason = reason
	r.EventMessage = message
}

// EventReason sets the syncer result reason for kind
func EventReason(obj client.Object, err error) string {
	objKindName := reflect.TypeOf(obj).String()
	if err != nil {
		return fmt.Sprintf("%sSyncFailed", objKindName)
	}
	return fmt.Sprintf("%sSyncSuccessfull", objKindName)
}
