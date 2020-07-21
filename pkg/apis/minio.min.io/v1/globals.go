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

package v1

import (
	"os"
)

// ClusterDomain is used to store the Kubernetes cluster domain
var ClusterDomain string

// Scheme indicates communication over http or https
var Scheme string

// Identity is the public identity generated for MinIO Server based on
// Used only during KES Deployments
var Identity string

func getEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "cluster.local"
	}
	return value
}

func identifyScheme(t *Tenant) string {
	scheme := "http"
	if t.AutoCert() || t.ExternalCert() {
		scheme = "https"
	}
	return scheme
}

// InitGlobals initiates the global variables while Operator starts
func InitGlobals(t *Tenant) {
	ClusterDomain = getEnv("CLUSTER_DOMAIN")
	Scheme = identifyScheme(t)
}
