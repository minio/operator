// Copyright (C) 2022, MinIO, Inc.
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

package controller

import (
	"context"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
)

// RegisterEvent creates an event for a given tenant
func (c *Controller) RegisterEvent(ctx context.Context, tenant *miniov2.Tenant, eventType, reason, message string) {
	c.recorder.Event(tenant, eventType, reason, message)
}
