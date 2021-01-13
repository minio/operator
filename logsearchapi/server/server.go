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
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

// LogSearch represents the Log Search API server
type LogSearch struct {
	// Configuration
	PGConnStr                      string
	AuditAuthToken, QueryAuthToken string
	DiskCapacityGBs                int

	// Runtime
	DBClient *DBClient
	*http.ServeMux
}

// NewLogSearch creates a LogSearch
func NewLogSearch(pgConnStr, auditAuthToken string, queryAuthToken string, diskCapacity int) (ls *LogSearch, err error) {
	ls = &LogSearch{
		PGConnStr:       pgConnStr,
		AuditAuthToken:  auditAuthToken,
		QueryAuthToken:  queryAuthToken,
		DiskCapacityGBs: diskCapacity,
	}

	// Initialize DB Client
	ls.DBClient, err = NewDBClient(context.Background(), ls.PGConnStr)
	if err != nil {
		// FIXME(aditya): Remove connection string as it contains creds
		return nil, fmt.Errorf("Error connecting to db: %v, connection string: %s", err, ls.PGConnStr)
	}

	// Initialize tables in db
	err = ls.DBClient.InitDBTables(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Error initializing tables: %v", err)
	}

	// Initialize muxer
	ls.ServeMux = http.NewServeMux()
	ls.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {})
	ls.HandleFunc("/api/ingest", authorize(ls.ingestHandler, ls.AuditAuthToken))
	ls.HandleFunc("/api/query", authorize(ls.queryHandler, ls.QueryAuthToken))

	// Start vacuum thread
	go ls.DBClient.vacuumData(ls.DiskCapacityGBs)
	go ls.DBClient.partitionTables()

	return ls, nil
}

func (ls *LogSearch) writeErrorResponse(w http.ResponseWriter, status int, msg string, err error) {
	http.Error(w, fmt.Sprintf("%s: %v", msg, err), status)
	log.Printf("%s: %v (%d)", msg, err, status)
}

// ingestHandler handles:
//
//   POST /api/ingest?token=xxx
//
// The json body represents the Audit log data. If it is an empty object the
// request is ignored but returns success.
func (ls *LogSearch) ingestHandler(w http.ResponseWriter, r *http.Request) {
	// Request is assumed to be authenticated at this point.

	if r.Method != "POST" {
		ls.writeErrorResponse(w, 400, "Non post request", nil)
		return
	}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ls.writeErrorResponse(w, 500, "Error reading request body", err)
		return
	}

	err = ls.DBClient.InsertEvent(r.Context(), buf)
	if err != nil {
		ls.writeErrorResponse(w, 500, "Error writing to DB", err)
	}
}

// queryHandler handles:
//
//   GET /api/query?token=xxx&q=(raw|reqinfo)&pageNo=0&pageSize=50&timeAsc|timeDesc&timeStart=?
func (ls *LogSearch) queryHandler(w http.ResponseWriter, r *http.Request) {
	// Request is assumed to be authenticated at this point.

	sq, err := searchQueryFromRequest(r)
	if err != nil {
		ls.writeErrorResponse(w, 400, "Bad params:", err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	err = ls.DBClient.Search(r.Context(), sq, w)
	if err != nil {
		w.Header().Del("Content-Type")
		ls.writeErrorResponse(w, 500, "Unhandled error:", err)
	}
}

// LoadEnv loads environment variables and returns
// a new LogSearch.
func LoadEnv() (*LogSearch, error) {
	pgConnStr := os.Getenv(PgConnStrEnv)
	if pgConnStr == "" {
		return nil, errors.New(PgConnStrEnv + " env variable is required.")
	}
	auditAuthToken := os.Getenv(AuditAuthTokenEnv)
	if auditAuthToken == "" {
		return nil, errors.New(AuditAuthTokenEnv + " env variable is required.")
	}
	queryAuthToken := os.Getenv(QueryAuthTokenEnv)
	if queryAuthToken == "" {
		return nil, errors.New(QueryAuthTokenEnv + " env variable is required.")
	}
	diskCapacity, err := strconv.Atoi(os.Getenv(DiskCapacityEnv))
	if err != nil {
		return nil, errors.New(DiskCapacityEnv + " env variable is required and must be an integer.")
	}

	return NewLogSearch(pgConnStr, auditAuthToken, queryAuthToken, diskCapacity)
}
