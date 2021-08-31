//go:build go1.13
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
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"github.com/minio/operator/logsearchapi/server"
)

func main() {
	ls, err := server.LoadEnv()
	if err != nil {
		log.Fatal(err)
	}
	s := &http.Server{
		Addr:    ":8080",
		Handler: ls,
	}
	log.Print("Log Search API Listening on Port :8080")
	log.Fatal(s.ListenAndServe())
}
