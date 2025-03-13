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
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/minio/operator/pkg/common"
)

func configureSidecarServer(c *Controller) *http.Server {
	router := mux.NewRouter().SkipClean(true).UseEncodedPath()

	router.Methods(http.MethodPost).
		Path(common.SidecarAPIConfigEndpoint).
		HandlerFunc(c.CheckConfigHandler).
		Queries(restQueries("c")...)

	router.NotFoundHandler = http.NotFoundHandler()

	s := &http.Server{
		Addr:           "0.0.0.0:" + common.SidecarHTTPPort,
		Handler:        router,
		ReadTimeout:    time.Minute,
		WriteTimeout:   time.Minute,
		MaxHeaderBytes: 1 << 20,
	}

	return s
}

// CheckConfigHandler - POST /sidecar/v1/config?c={hash}
func (c *Controller) CheckConfigHandler(_ http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	hash := vars["c"]

	log.Println("Checking config hash: ", hash)
}
