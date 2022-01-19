/*
 * Copyright (C) 2022, MinIO, Inc.
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
	"log"

	"github.com/lib/pq"
)

// dbMigration represents a DB migration using db-client c.
// Note: a migration func should be idempotent.
type dbMigration func(ctx context.Context, c *DBClient) error

var allMigrations = []dbMigration{
	addAccessKeyColAndIndex,
	addAuditLogIndices,
	addReqInfoIndices,

	// Add new migrations here below
}

func (c *DBClient) runMigrations(ctx context.Context) error {
	for _, migration := range allMigrations {
		if err := migration(ctx, c); err != nil {
			return err
		}
	}
	return nil
}

func duplicateColErr(err error) bool {
	if pqerr, ok := err.(*pq.Error); ok &&
		pqerr.Code == "42701" {
		return true
	}
	return false
}

func duplicateTblErr(err error) bool {
	if pqerr, ok := err.(*pq.Error); ok &&
		pqerr.Code == "42P07" {
		return true
	}
	return false
}

func (c *DBClient) runQueries(ctx context.Context, queries []string, ignoreErr func(error) bool) error {
	for _, query := range queries {
		if _, err := c.ExecContext(ctx, query); err != nil {
			if ignoreErr(err) {
				continue
			}
			return err
		}
	}
	return nil
}

func updateAccessKeyCol(ctx context.Context, c *DBClient) {
	updQ := `WITH req AS (
                             SELECT log->>'requestID' AS request_id,
                                    COALESCE(
                                       substring(
                                           log->'requestHeader'->>'Authorization',
                                           e'^AWS4-HMAC-SHA256\\s+Credential\\s*=\\s*([^/]+)'
                                       ),
                                       substring(log->'requestHeader'->>'Authorization', e'^AWS\\s+([^:]+)')
                                    ) AS access_key
                               FROM audit_log_events AS a JOIN request_info AS b ON (a.event_time = b.time)
                              WHERE b.access_key IS NULL
                           ORDER BY event_time
                              LIMIT $1
                             OFFSET $2
                          )
               UPDATE request_info
                  SET access_key = req.access_key
                 FROM req
                WHERE request_info.request_id = req.request_id`

	for off, lim := 0, 1000; ; off += lim {
		res, err := c.ExecContext(ctx, updQ, lim, off)
		if err != nil {
			log.Printf("Failed to update access_key column in request_info: %v", err)
			return
		}

		if rows, err := res.RowsAffected(); err != nil {
			log.Printf("Failed to get rows affected: %v", err)
			return
		} else if rows < 1000 {
			break
		}
	}
}

func addAccessKeyColAndIndex(ctx context.Context, c *DBClient) error {
	queries := []string{
		`alter table request_info add access_key text`,
		`create index request_info_access_key_index on request_info (access_key)`,
	}

	return c.runQueries(ctx, queries, func(err error) bool {
		if duplicateColErr(err) {
			go updateAccessKeyCol(ctx, c)
			return true
		}
		if duplicateTblErr(err) {
			return true
		}
		return false
	})
}

func addAuditLogIndices(ctx context.Context, c *DBClient) error {
	queries := []string{
		`create index audit_log_events_log_index on audit_log_events USING btree ((log->>'requestID'))`,
		`create index audit_log_events_event_time_index on audit_log_events (event_time desc)`,
	}

	return c.runQueries(ctx, queries, duplicateTblErr)
}

func addReqInfoIndices(ctx context.Context, c *DBClient) error {
	queries := []string{
		`create index request_info_api_name_index on request_info (api_name)`,
		`create index request_info_bucket_index on request_info (bucket)`,
		`create index request_info_object_index on request_info (object)`,
		`create index request_info_request_id_index on request_info (request_id)`,
		`create index request_info_response_status_index on request_info (response_status)`,
		`create index request_info_time_index on request_info (time)`,
	}

	return c.runQueries(ctx, queries, duplicateTblErr)
}
