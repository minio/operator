// +build go1.13

/*
 * Copyright (C) 2020, MinIO, Inc.
 *
 * This code is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, version 3,
 * as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License, version 3,
 * along with this program.  If not, see <http://www.gnu.org/licenses/>
 *
 */

package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Entry - audit entry logs.
type Entry struct {
	Version      string `json:"version"`
	DeploymentID string `json:"deploymentid,omitempty"`
	Time         string `json:"time"`
	Trigger      string `json:"trigger"`
	API          struct {
		Name            string `json:"name,omitempty"`
		Bucket          string `json:"bucket,omitempty"`
		Object          string `json:"object,omitempty"`
		Status          string `json:"status,omitempty"`
		StatusCode      int    `json:"statusCode,omitempty"`
		TimeToFirstByte string `json:"timeToFirstByte,omitempty"`
		TimeToResponse  string `json:"timeToResponse,omitempty"`
	} `json:"api"`
	RemoteHost string                 `json:"remotehost,omitempty"`
	RequestID  string                 `json:"requestID,omitempty"`
	UserAgent  string                 `json:"userAgent,omitempty"`
	ReqClaims  map[string]interface{} `json:"requestClaims,omitempty"`
	ReqQuery   map[string]string      `json:"requestQuery,omitempty"`
	ReqHeader  map[string]string      `json:"requestHeader,omitempty"`
	RespHeader map[string]string      `json:"responseHeader,omitempty"`
	Tags       map[string]interface{} `json:"tags,omitempty"`
}

// API is struct with same info an Entry.API, but with more strong types.
type API struct {
	Name            string         `json:"name,omitempty"`
	Bucket          string         `json:"bucket,omitempty"`
	Object          string         `json:"object,omitempty"`
	Status          string         `json:"status,omitempty"`
	StatusCode      int            `json:"statusCode,omitempty"`
	TimeToFirstByte *time.Duration `json:"timeToFirstByte,omitempty"`
	TimeToResponse  time.Duration  `json:"timeToResponse,omitempty"`
}

// Event is the same as Entry but with more typed values.
type Event struct {
	Version      string                 `json:"version"`
	DeploymentID string                 `json:"deploymentid,omitempty"`
	Time         time.Time              `json:"time"`
	API          API                    `json:"api"`
	RemoteHost   string                 `json:"remotehost,omitempty"`
	RequestID    string                 `json:"requestID,omitempty"`
	UserAgent    string                 `json:"userAgent,omitempty"`
	ReqClaims    map[string]interface{} `json:"requestClaims,omitempty"`
	ReqQuery     map[string]string      `json:"requestQuery,omitempty"`
	ReqHeader    map[string]string      `json:"requestHeader,omitempty"`
	RespHeader   map[string]string      `json:"responseHeader,omitempty"`
}

// EventFromEntry performs a type conversion
func EventFromEntry(e *Entry) (*Event, error) {
	ret := Event{
		Version:      e.Version,
		DeploymentID: e.DeploymentID,
		API: API{
			Name:       e.API.Name,
			Bucket:     e.API.Bucket,
			Object:     e.API.Object,
			Status:     e.API.Status,
			StatusCode: e.API.StatusCode,
		},
		RemoteHost: e.RemoteHost,
		RequestID:  e.RequestID,
		UserAgent:  e.UserAgent,
		ReqClaims:  e.ReqClaims,
		ReqQuery:   e.ReqQuery,
		ReqHeader:  e.ReqHeader,
		RespHeader: e.RespHeader,
	}

	// Parse time
	var err error
	ret.Time, err = time.Parse(time.RFC3339Nano, e.Time)
	if err != nil {
		return nil, err
	}

	parseNanosec := func(s string) (time.Duration, error) {
		if !strings.HasSuffix(s, "ns") {
			return 0, errors.New("'ns' suffix not present")
		}
		s = s[:len(s)-2]
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("nanosec duration parse error: %v", err)
		}
		return time.Duration(n) * time.Nanosecond, nil
	}

	//Parse durations

	// TTFB can be absent
	if e.API.TimeToFirstByte != "" {
		dur, err := parseNanosec(e.API.TimeToFirstByte)
		if err != nil {
			return nil, err
		}
		ret.API.TimeToFirstByte = &dur
	}
	ret.API.TimeToResponse, err = parseNanosec(e.API.TimeToResponse)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func parseJSONEvent(b []byte) (*Event, error) {
	var entry Entry
	if err := json.Unmarshal(b, &entry); err != nil {
		return nil, err
	}

	return EventFromEntry(&entry)
}

func isEmptyEvent(b []byte) bool {
	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		return false
	}
	return len(m) == 0
}

func (ev *Event) getRequestContentLength() (uint64, error) {
	val, ok := ev.ReqHeader["Content-Length"]
	if !ok {
		return 0, errors.New("Request Content-Length not present")
	}
	return strconv.ParseUint(val, 10, 64)
}

func (ev *Event) getResponseContentLength() (uint64, error) {
	val, ok := ev.RespHeader["Content-Length"]
	if !ok {
		return 0, errors.New("Response Content-Length not present")
	}
	return strconv.ParseUint(val, 10, 64)
}
