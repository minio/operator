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

package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/minio/operator/logsearchapi/server"
)

const (
	pgConnStrEnv      = "LOGSEARCH_PG_CONN_STR"
	auditAuthTokenEnv = "LOGSEARCH_AUDIT_AUTH_TOKEN"
	queryAuthTokenEnv = "LOGSEARCH_QUERY_AUTH_TOKEN"
	diskCapacityEnv   = "LOGSEARCH_DISK_CAPACITY_GB"
)

func loadEnv() (*server.LogSearch, error) {
	pgConnStr := os.Getenv(pgConnStrEnv)
	if pgConnStr == "" {
		return nil, errors.New(pgConnStrEnv + " env variable is required.")
	}
	auditAuthToken := os.Getenv(auditAuthTokenEnv)
	if auditAuthToken == "" {
		return nil, errors.New(auditAuthTokenEnv + " env variable is required.")
	}
	queryAuthToken := os.Getenv(queryAuthTokenEnv)
	if queryAuthToken == "" {
		return nil, errors.New(queryAuthTokenEnv + " env variable is required.")
	}
	diskCapacity, err := strconv.Atoi(os.Getenv(diskCapacityEnv))
	if err != nil {
		return nil, errors.New(diskCapacityEnv + " env variable is required and must be an integer.")
	}

	return server.NewLogSearch(pgConnStr, auditAuthToken, queryAuthToken, diskCapacity)
}

func main() {
	ls, err := loadEnv()
	if err != nil {
		log.Fatal(err)
	}
	s := &http.Server{
		Addr:    ":8080",
		Handler: ls,
	}
	log.Fatal(s.ListenAndServe())
}
