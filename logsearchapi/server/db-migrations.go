// This file is part of MinIO Operator
// Copyright (c) 2022 MinIO, Inc.
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

package server

import (
	"context"
	"fmt"
	"log"

	"github.com/lib/pq"
)

// dbMigration represents a DB migration using db-client c.
// Note: a migration func should be idempotent.
type dbMigration func(ctx context.Context, c *DBClient) error

var allMigrations = []dbMigration{
	addAccessKeyCol,

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

// updateAccessKeyCol updates request_info records which where created before
// the introduction of access_key column.
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
                          )
               UPDATE request_info
                  SET access_key = req.access_key
                 FROM req
                WHERE request_info.request_id = req.request_id`

	for lim := 1000; ; {
		select {
		case <-ctx.Done():
			return
		default:
		}

		res, err := c.ExecContext(ctx, updQ, lim)
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

// addAccessKeyCol adds a new column access_key, to request_info table to store
// API requests access key/user information wherever applicable.
func addAccessKeyCol(ctx context.Context, c *DBClient) error {
	queries := []string{
		`ALTER table request_info ADD access_key text`,
	}
	err := c.runQueries(ctx, queries, func(err error) bool {
		if duplicateColErr(err) {
			return true
		}
		if duplicateTblErr(err) {
			return true
		}
		return false
	})
	go updateAccessKeyCol(ctx, c)
	return err
}

// CreateIndices creates table indexes for audit_log_events and request_info tables.
// See auditLogIndices, reqInfoIndices functions for actual indices details.
func (c *DBClient) CreateIndices(ctx context.Context) error {
	tables := []struct {
		t       Table
		indices []indexOpts
	}{
		{
			t:       auditLogEventsTable,
			indices: auditLogIndices(),
		},
		{
			t:       requestInfoTable,
			indices: reqInfoIndices(),
		},
	}

	for _, table := range tables {
		// The following procedure creates indices on all partitions of
		// this table. If an index was created on any of its partitions,
		// it checks if newer partitions were created meanwhile, so as
		// to create indices on those partitions too.
		for {
			partitions, err := c.getExistingPartitions(ctx, table.t)
			if err != nil {
				return err
			}

			var indexCreated bool
			for _, partition := range partitions {
				indexed, err := c.CreatePartitionIndices(ctx, table.indices, partition)
				if err != nil {
					return err
				}
				indexCreated = indexCreated || indexed
			}
			if !indexCreated {
				break
			}
		}

		// No more new non-indexed table partitions, creating
		// parent table indices.
		err := c.CreateParentIndices(ctx, table.indices)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreatePartitionIndices creates all indices described by optses on partition.
// It returns true if a new index was created on this partition. Note: this
// function ignores the index already exists error.
func (c *DBClient) CreatePartitionIndices(ctx context.Context, optses []indexOpts, partition string) (indexed bool, err error) {
	for _, opts := range optses {
		q := opts.createPartitionQuery(partition)
		_, err := c.ExecContext(ctx, q)
		if err == nil {
			indexed = true
		}
		if err != nil && !duplicateTblErr(err) {
			return indexed, err
		}
	}
	return indexed, nil
}

// CreateParentIndices creates all indices specified by optses on the parent table.
func (c *DBClient) CreateParentIndices(ctx context.Context, optses []indexOpts) error {
	for _, opts := range optses {
		q := opts.createParentQuery()
		_, err := c.ExecContext(ctx, q)
		if err != nil && !duplicateTblErr(err) {
			return err
		}
	}
	return nil
}

// auditLogIndices is a slice of audit_log_events' table indices specified as
// indexOpt values.
func auditLogIndices() []indexOpts {
	return []indexOpts{
		{
			tableName:   "audit_log_events",
			indexSuffix: "log",
			col:         idxCol{name: `(log->>'requestID')`},
			idxType:     "btree",
		},
		{
			tableName: "audit_log_events",
			col: idxCol{
				name:  "event_time",
				order: colDesc,
			},
		},
	}
}

// reqInfoIndices is a slice of request_info's table indices specified as indexOpt values.
func reqInfoIndices() []indexOpts {
	var idxOpts []indexOpts
	cols := []string{"access_key", "api_name", "bucket", "object", "request_id", "response_status", "time"}
	for _, col := range cols {
		idxOpts = append(idxOpts, indexOpts{
			tableName: "request_info",
			col:       idxCol{name: col},
		})
	}
	return idxOpts
}

type colOrder bool

const (
	colDesc colOrder = true
)

type idxCol struct {
	name  string
	order colOrder
}

func (col idxCol) colWithOrder() string {
	if col.order == colDesc {
		return fmt.Sprintf("(%s DESC)", col.name)
	}
	return fmt.Sprintf("(%s)", col.name)
}

// indexOpts type is used to specify a table index
type indexOpts struct {
	tableName   string
	indexSuffix string
	col         idxCol
	idxType     string
}

func (opts indexOpts) colWithOrder() string {
	return opts.col.colWithOrder()
}

func (opts indexOpts) createParentQuery() string {
	var idxName string
	if opts.indexSuffix != "" {
		idxName = fmt.Sprintf("%s_%s_index", opts.tableName, opts.indexSuffix)
	} else {
		idxName = fmt.Sprintf("%s_%s_index", opts.tableName, opts.col.name)
	}

	var q string
	if opts.idxType != "" {
		q = fmt.Sprintf("CREATE INDEX %s ON %s USING %s %s", idxName, opts.tableName, opts.idxType, opts.colWithOrder())
	} else {
		q = fmt.Sprintf("CREATE INDEX %s ON %s %s", idxName, opts.tableName, opts.colWithOrder())
	}
	return q
}

func (opts indexOpts) createPartitionQuery(partition string) string {
	var idxName string
	if opts.indexSuffix != "" {
		idxName = fmt.Sprintf("%s_%s_index", partition, opts.indexSuffix)
	} else {
		idxName = fmt.Sprintf("%s_%s_index", partition, opts.col.name)
	}

	var q string
	if opts.idxType != "" {
		q = fmt.Sprintf("CREATE INDEX CONCURRENTLY %s ON %s USING %s %s", idxName, partition, opts.idxType, opts.colWithOrder())
	} else {
		q = fmt.Sprintf("CREATE INDEX CONCURRENTLY %s ON %s %s", idxName, partition, opts.colWithOrder())
	}
	return q
}
