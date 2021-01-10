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

package cluster

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
)

// Used for registering with rest handlers (have a look at registerStorageRESTHandlers for usage example)
// If it is passed ["aaaa", "bbbb"], it returns ["aaaa", "{aaaa:.*}", "bbbb", "{bbbb:.*}"]
func restQueries(keys ...string) []string {
	var accumulator []string
	for _, key := range keys {
		accumulator = append(accumulator, key, "{"+key+":.*}")
	}
	return accumulator
}

func configureWebhookServer(c *Controller) *http.Server {
	router := mux.NewRouter().SkipClean(true).UseEncodedPath()

	router.Methods(http.MethodGet).
		Path(miniov2.WebhookAPIGetenv + "/{namespace}/{name:.+}").
		HandlerFunc(c.GetenvHandler).
		Queries(restQueries("key")...)
	router.Methods(http.MethodPost).
		Path(miniov2.WebhookAPIBucketService + "/{namespace}/{name:.+}").
		HandlerFunc(c.BucketSrvHandler).
		Queries(restQueries("bucket")...)
	router.Methods(http.MethodGet).
		PathPrefix(miniov2.WebhookAPIUpdate).
		Handler(http.StripPrefix(miniov2.WebhookAPIUpdate, http.FileServer(http.Dir(updatePath))))
	// CRD Conversion
	router.Methods(http.MethodPost).
		Path(miniov2.WebhookCRDConversaion).
		HandlerFunc(c.CRDConversionHandler)
	//.
	//		Queries(restQueries("bucket")...)

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r)
	})

	s := &http.Server{
		Addr:           ":" + miniov2.WebhookDefaultPort,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return s
}
