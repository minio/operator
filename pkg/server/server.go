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
	"net/http"

	"github.com/gorilla/mux"
)

func StartServer() {
	m := mux.NewRouter()
	m.HandleFunc("/bucket/{id}", createBucketHandler).Methods("POST")
	m.HandleFunc("/buckets", listBucketHandler).Methods("GET")
	m.HandleFunc("/env", fetchEnvHandler).Methods("GET")
}

func createBucketHandler(http.ResponseWriter, *http.Request) {
	// 1. Extract bucket name from request
	// 2. Extract tenant name, service endpoint, namespace and jwt token from cert
	// 3. Create the DNS service name
}

func listBucketHandler(http.ResponseWriter, *http.Request) {
	// 1. Extract the namespace
	// 2. List all services running in that namespace which is a minio bucket.
}

func fetchEnvHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Extract the tenant information
	// 2. Calculate the environment based on the CRD for the tenant
	// 3. Return string
}
