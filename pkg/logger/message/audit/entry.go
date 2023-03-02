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

package audit

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/minio/operator/pkg/utils"

	xhttp "github.com/minio/operator/pkg/http"
)

// Version - represents the current version of audit log structure.
const Version = "1"

// ObjectVersion object version key/versionId
type ObjectVersion struct {
	ObjectName string `json:"objectName"`
	VersionID  string `json:"versionId,omitempty"`
}

// Entry - audit entry logs.
type Entry struct {
	Version      string    `json:"version"`
	DeploymentID string    `json:"deploymentid,omitempty"`
	Time         time.Time `json:"time"`
	Trigger      string    `json:"trigger"`
	API          struct {
		Path            string `json:"path,omitempty"`
		Status          string `json:"status,omitempty"`
		Method          string `json:"method"`
		StatusCode      int    `json:"statusCode,omitempty"`
		InputBytes      int64  `json:"rx"`
		OutputBytes     int64  `json:"tx"`
		TimeToFirstByte string `json:"timeToFirstByte,omitempty"`
		TimeToResponse  string `json:"timeToResponse,omitempty"`
	} `json:"api"`
	RemoteHost string                 `json:"remotehost,omitempty"`
	RequestID  string                 `json:"requestID,omitempty"`
	SessionID  string                 `json:"sessionID,omitempty"`
	UserAgent  string                 `json:"userAgent,omitempty"`
	ReqClaims  map[string]interface{} `json:"requestClaims,omitempty"`
	ReqQuery   map[string]string      `json:"requestQuery,omitempty"`
	ReqHeader  map[string]string      `json:"requestHeader,omitempty"`
	RespHeader map[string]string      `json:"responseHeader,omitempty"`
	Tags       map[string]interface{} `json:"tags,omitempty"`
}

// NewEntry - constructs an audit entry object with some fields filled
func NewEntry(deploymentID string) Entry {
	return Entry{
		Version:      Version,
		DeploymentID: deploymentID,
		Time:         time.Now().UTC(),
	}
}

// ToEntry - constructs an audit entry from a http request
func ToEntry(w http.ResponseWriter, r *http.Request, reqClaims map[string]interface{}, deploymentID string) Entry {
	entry := NewEntry(deploymentID)

	entry.RemoteHost = r.RemoteAddr
	entry.UserAgent = r.UserAgent()
	entry.ReqClaims = reqClaims

	q := r.URL.Query()
	reqQuery := make(map[string]string, len(q))
	for k, v := range q {
		reqQuery[k] = strings.Join(v, ",")
	}
	entry.ReqQuery = reqQuery

	reqHeader := make(map[string]string, len(r.Header))
	for k, v := range r.Header {
		reqHeader[k] = strings.Join(v, ",")
	}
	entry.ReqHeader = reqHeader

	wh := w.Header()

	var requestID interface{}
	requestID = r.Context().Value(utils.ContextRequestID)
	if requestID == nil {
		requestID, _ = utils.NewUUID()
	}
	entry.RequestID = requestID.(string)

	if val := r.Context().Value(utils.ContextRequestUserID); val != nil {
		sessionID := val.(string)
		claims := jwt.MapClaims{}
		_, _ = jwt.ParseWithClaims(sessionID, claims, nil)
		if sub, ok := claims["sub"]; ok {
			sessionID = sub.(string)
		}
		entry.SessionID = sessionID
	}

	respHeader := make(map[string]string, len(wh))
	for k, v := range wh {
		respHeader[k] = strings.Join(v, ",")
	}
	entry.RespHeader = respHeader

	if etag := respHeader[xhttp.ETag]; etag != "" {
		respHeader[xhttp.ETag] = strings.Trim(etag, `"`)
	}

	return entry
}
