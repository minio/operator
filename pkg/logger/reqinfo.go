// This file is part of MinIO Operator
// Copyright (c) 2022 MinIO, Inc.
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

package logger

import (
	"context"
	"fmt"
	"sync"

	"github.com/minio/operator/pkg/utils"
)

// KeyVal - appended to ReqInfo.Tags
type KeyVal struct {
	Key string
	Val interface{}
}

// ObjectVersion object version key/versionId
type ObjectVersion struct {
	ObjectName string
	VersionID  string `json:"VersionId,omitempty"`
}

// ReqInfo stores the request info.
type ReqInfo struct {
	RemoteHost   string          // Client Host/IP
	Host         string          // Node Host/IP
	UserAgent    string          // User Agent
	DeploymentID string          // x-minio-deployment-id
	RequestID    string          // x-amz-request-id
	SessionID    string          // custom session id
	API          string          // API name - GetObject PutObject NewMultipartUpload etc.
	BucketName   string          `json:",omitempty"` // Bucket name
	ObjectName   string          `json:",omitempty"` // Object name
	VersionID    string          `json:",omitempty"` // corresponding versionID for the object
	Objects      []ObjectVersion `json:",omitempty"` // Only set during MultiObject delete handler.
	AccessKey    string          // Access Key
	tags         []KeyVal        // Any additional info not accommodated by above fields
	sync.RWMutex
}

// GetTags - returns the user defined tags
func (r *ReqInfo) GetTags() []KeyVal {
	if r == nil {
		return nil
	}
	r.RLock()
	defer r.RUnlock()
	return append([]KeyVal(nil), r.tags...)
}

// GetTagsMap - returns the user defined tags in a map structure
func (r *ReqInfo) GetTagsMap() map[string]interface{} {
	if r == nil {
		return nil
	}
	r.RLock()
	defer r.RUnlock()
	m := make(map[string]interface{}, len(r.tags))
	for _, t := range r.tags {
		m[t.Key] = t.Val
	}
	return m
}

// SetReqInfo sets ReqInfo in the context.
func SetReqInfo(ctx context.Context, req *ReqInfo) context.Context {
	if ctx == nil {
		LogIf(context.Background(), fmt.Errorf("context is nil"))
		return nil
	}
	return context.WithValue(ctx, utils.ContextLogKey, req)
}

// GetReqInfo returns ReqInfo if set.
func GetReqInfo(ctx context.Context) *ReqInfo {
	if ctx != nil {
		r, ok := ctx.Value(utils.ContextLogKey).(*ReqInfo)
		if ok {
			return r
		}
		r = &ReqInfo{}
		if val, o := ctx.Value(utils.ContextRequestID).(string); o {
			r.RequestID = val
		}
		if val, o := ctx.Value(utils.ContextRequestUserID).(string); o {
			r.SessionID = val
		}
		if val, o := ctx.Value(utils.ContextRequestUserAgent).(string); o {
			r.UserAgent = val
		}
		if val, o := ctx.Value(utils.ContextRequestHost).(string); o {
			r.Host = val
		}
		if val, o := ctx.Value(utils.ContextRequestRemoteAddr).(string); o {
			r.RemoteHost = val
		}
		SetReqInfo(ctx, r)
		return r
	}
	return nil
}
