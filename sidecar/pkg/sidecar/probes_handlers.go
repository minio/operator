// This file is part of MinIO Operator
// Copyright (c) 2024 MinIO, Inc.
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

package sidecar

import (
	"crypto/tls"
	"fmt"
	"io"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

func configureProbesServer(c *Controller, tenantTLS bool) *http.Server {
	router := mux.NewRouter().SkipClean(true).UseEncodedPath()

	router.Methods(http.MethodGet).
		Path("/ready").
		HandlerFunc(readinessHandler(tenantTLS))

	router.NotFoundHandler = http.NotFoundHandler()

	s := &http.Server{
		Addr:           "0.0.0.0:4444",
		Handler:        router,
		ReadTimeout:    time.Minute,
		WriteTimeout:   time.Minute,
		MaxHeaderBytes: 1 << 20,
	}

	return s
}

// we do insecure skip verify because we are checking against the local instance and don't care for the certificate
var probeHTTPClient = &http.Client{
	Timeout: time.Millisecond * 500,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

func readinessHandler(tenantTLS bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		schema := "https"
		if !tenantTLS {
			schema = "http"
		}
		// we only check against the local instance of MinIO
		url := schema + "://localhost:9000"
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create request: %s", err), http.StatusInternalServerError)
			return
		}

		response, err := probeHTTPClient.Do(request)
		if err != nil {
			http.Error(w, fmt.Sprintf("HTTP request failed: %s", err), http.StatusInternalServerError)
			return
		}
		defer response.Body.Close()
		_, _ = io.Copy(io.Discard, response.Body) // Discard body to enable connection reuse

		if response.StatusCode == 403 {
			fmt.Fprintln(w, "Readiness probe succeeded.")
		} else {
			http.Error(w, fmt.Sprintf("Readiness probe failed: expected status 403, got %d", response.StatusCode), http.StatusServiceUnavailable)
		}
	}
}
