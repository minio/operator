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
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/georgysavva/scany/sqlscan"
)

// QTemplate is used to represent queries that involve string substitution as
// well as SQL positional argument substitution.
type QTemplate string

func (t QTemplate) build(args ...interface{}) string {
	return fmt.Sprintf(string(t), args...)
}

// Table a database table
type Table struct {
	Name            string
	CreateStatement QTemplate
}

func (t *Table) getCreateStatement() string {
	return t.CreateStatement.build(t.Name)
}

var (
	auditLogEventsTable = Table{
		Name: "audit_log_events",
		CreateStatement: `CREATE TABLE %s (
                                    event_time TIMESTAMPTZ NOT NULL,
                                    log JSONB NOT NULL
                                  ) PARTITION BY RANGE (event_time);`,
	}
	requestInfoTable = Table{
		Name: "request_info",
		CreateStatement: `CREATE TABLE %s (
                                    time TIMESTAMPTZ NOT NULL,
                                    api_name TEXT NOT NULL,
                                    bucket TEXT,
                                    object TEXT,
                                    time_to_response_ns INT8,
                                    remote_host TEXT,
                                    request_id TEXT,
                                    user_agent TEXT,
                                    response_status TEXT,
                                    response_status_code INT8,
                                    request_content_length INT8,
                                    response_content_length INT8
                                  ) PARTITION BY RANGE (time);`,
	}

	// Allows iterating on all tables
	allTables = []Table{auditLogEventsTable, requestInfoTable}
)

// DBClient is a client object that makes requests to the DB.
type DBClient struct {
	*sql.DB
}

// NewDBClient creates a new DBClient.
func NewDBClient(ctx context.Context, connStr string) (*DBClient, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	log.Print("Connected to db.")

	return &DBClient{db}, nil
}

func (c *DBClient) checkTableExists(ctx context.Context, table string) (bool, error) {
	const existsQuery QTemplate = `SELECT 1 FROM %s WHERE false;`
	_, err := c.QueryContext(ctx, existsQuery.build(table))
	if err != nil {
		// check for table does not exist error
		if strings.Contains(err.Error(), fmt.Sprintf(`relation "%s" does not exist`, table)) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *DBClient) checkPartitionTableExists(ctx context.Context, table string, givenTime time.Time) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	p := newPartitionTimeRange(givenTime)
	partitionTable := fmt.Sprintf("%s_%s", table, p.getPartnameSuffix())
	const existsQuery QTemplate = `SELECT 1 FROM %s WHERE false;`
	_, err := c.QueryContext(ctx, existsQuery.build(partitionTable))
	if err != nil {
		// check for table does not exist error
		if strings.Contains(err.Error(), fmt.Sprintf(`relation "%s" does not exist`, table)) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (c *DBClient) createTablePartition(ctx context.Context, table Table) error {
	partTimeRange := newPartitionTimeRange(time.Now())
	_, err := c.ExecContext(ctx, table.getCreatePartitionStatement(partTimeRange))
	return err
}

func (c *DBClient) createTableAndPartition(ctx context.Context, table Table) error {
	if exists, err := c.checkTableExists(ctx, table.Name); err != nil {
		return err
	} else if exists {
		return nil
	}

	if _, err := c.ExecContext(ctx, table.getCreateStatement()); err != nil {
		return err
	}

	return c.createTablePartition(ctx, table)
}

func (c *DBClient) createTables(ctx context.Context) error {
	if err := c.createTableAndPartition(ctx, auditLogEventsTable); err != nil {
		return err
	}

	if err := c.createTableAndPartition(ctx, requestInfoTable); err != nil {
		return err
	}

	return nil
}

// InitDBTables Creates tables in the DB.
func (c *DBClient) InitDBTables(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return c.createTables(ctx)
}

// InsertEvent inserts audit event in the DB.
func (c *DBClient) InsertEvent(ctx context.Context, eventBytes []byte) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if isEmptyEvent(eventBytes) {
		return nil
	}
	event, err := parseJSONEvent(eventBytes)
	if err != nil {
		return err
	}

	const (
		insertAuditLogEvent QTemplate = `INSERT INTO %s (event_time, log) VALUES ($1, $2);`
		insertRequestInfo   QTemplate = `INSERT INTO %s (time,
                                                                 api_name,
                                                                 bucket,
                                                                 object,
                                                                 time_to_response_ns,
                                                                 remote_host,
                                                                 request_id,
                                                                 user_agent,
                                                                 response_status,
                                                                 response_status_code,
                                                                 request_content_length,
                                                                 response_content_length)
                                                   VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);`
	)

	// Start a database transaction
	tx, err := c.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// NOTE: Timestamps are nanosecond resolution from MinIO, however we are
	// using storing it with only microsecond precision in PG for simplicity
	// as that is the maximum precision supported by it.
	eventJSON, errJSON := json.Marshal(event)
	if errJSON != nil {
		return errJSON
	}
	_, err = tx.ExecContext(ctx, insertAuditLogEvent.build(auditLogEventsTable.Name), event.Time, eventJSON)
	if err != nil {
		return err
	}

	var reqLen *uint64
	rqlen, err := event.getRequestContentLength()
	if err == nil {
		reqLen = &rqlen
	}
	var respLen *uint64
	rsplen, err := event.getResponseContentLength()
	if err == nil {
		respLen = &rsplen
	}

	_, err = tx.ExecContext(ctx, insertRequestInfo.build(requestInfoTable.Name),
		event.Time,
		event.API.Name,
		event.API.Bucket,
		event.API.Object,
		event.API.TimeToResponse,
		event.RemoteHost,
		event.RequestID,
		event.UserAgent,
		event.API.Status,
		event.API.StatusCode,
		reqLen,
		respLen)
	if err != nil {
		return err
	}

	return tx.Commit()
}

type logEventRawRow struct {
	EventTime time.Time
	Log       string
}

// LogEventRow holds a raw log record
type LogEventRow struct {
	EventTime time.Time              `json:"event_time"`
	Log       map[string]interface{} `json:"log"`
}

// ReqInfoRow holds a structured log record
type ReqInfoRow struct {
	Time                  time.Time `json:"time"`
	APIName               string    `json:"api_name"`
	Bucket                string    `json:"bucket"`
	Object                string    `json:"object"`
	TimeToResponseNs      uint64    `json:"time_to_response_ns"`
	RemoteHost            string    `json:"remote_host"`
	RequestID             string    `json:"request_id"`
	UserAgent             string    `json:"user_agent"`
	ResponseStatus        string    `json:"response_status"`
	ResponseStatusCode    int       `json:"response_status_code"`
	RequestContentLength  *uint64   `json:"request_content_length"`
	ResponseContentLength *uint64   `json:"response_content_length"`
}

// Search executes a search query on the db.
func (c *DBClient) Search(ctx context.Context, s *SearchQuery, w io.Writer) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	const (
		logEventSelect QTemplate = `SELECT event_time,
                                                   log
                                              FROM %s
                                             WHERE %s
                                          ORDER BY event_time %s
                                            OFFSET $1 LIMIT $2;`
		reqInfoSelect QTemplate = `SELECT time,
                                                  api_name,
                                                  bucket,
                                                  object,
                                                  time_to_response_ns,
                                                  remote_host,
                                                  request_id,
                                                  user_agent,
                                                  response_status,
                                                  response_status_code,
                                                  request_content_length,
                                                  response_content_length
                                             FROM %s
                                            %s
                                         	ORDER BY time %s
                                           	OFFSET $1 LIMIT $2;`
	)

	timeOrder := "DESC"
	if s.TimeAscending {
		timeOrder = "ASC"
	}

	jw := json.NewEncoder(w)
	switch s.Query {
	case rawQ:
		whereClauses := []string{}
		// only filter by time if provided
		if s.TimeStart != nil {
			timeRangeOp := ">="
			timeRangeClause := fmt.Sprintf("event_time %s '%s'", timeRangeOp, s.TimeStart.Format(time.RFC3339Nano))
			whereClauses = append(whereClauses, timeRangeClause)
		}
		if s.TimeEnd != nil {
			timeRangeOp := "<"
			timeRangeClause := fmt.Sprintf("event_time %s '%s'", timeRangeOp, s.TimeEnd.Format(time.RFC3339Nano))
			whereClauses = append(whereClauses, timeRangeClause)
		}
		whereClause := strings.Join(whereClauses, " AND ")

		q := logEventSelect.build(auditLogEventsTable.Name, whereClause, timeOrder)
		rows, _ := c.QueryContext(ctx, q, s.PageNumber*s.PageSize, s.PageSize)
		var logEventsRaw []logEventRawRow
		if err := sqlscan.ScanAll(&logEventsRaw, rows); err != nil {
			return fmt.Errorf("Error accessing db: %v", err)
		}
		// parse the encoded json string stored in the db into a json
		// object for output
		logEvents := make([]LogEventRow, len(logEventsRaw))
		for i, e := range logEventsRaw {
			logEvents[i].EventTime = e.EventTime
			logEvents[i].Log = make(map[string]interface{})
			if err := json.Unmarshal([]byte(e.Log), &logEvents[i].Log); err != nil {
				return fmt.Errorf("Error decoding json log: %v", err)
			}
		}
		if err := jw.Encode(logEvents); err != nil {
			return fmt.Errorf("Error writing to output stream: %v", err)
		}
	case reqInfoQ:
		// For this query, $1 and $2 are used for offset and limit.
		sqlArgs := []interface{}{s.PageNumber * s.PageSize, s.PageSize}

		dollarStart := 3
		whereClauses := []string{}
		// only filter by time if provided
		if s.TimeStart != nil {
			// $3 will be used for the time parameter
			timeRangeOp := ">="
			timeRangeClause := fmt.Sprintf("time %s $%d", timeRangeOp, dollarStart)
			sqlArgs = append(sqlArgs, s.TimeStart.Format(time.RFC3339Nano))
			whereClauses = append(whereClauses, timeRangeClause)
			dollarStart++
		}
		// only filter by time if provided
		if s.TimeEnd != nil {
			// $3 will be used for the time parameter
			timeRangeOp := "<"
			timeRangeClause := fmt.Sprintf("time %s $%d", timeRangeOp, dollarStart)
			sqlArgs = append(sqlArgs, s.TimeEnd.Format(time.RFC3339Nano))
			whereClauses = append(whereClauses, timeRangeClause)
			dollarStart++
		}

		// Remaining dollar params are added for filter where clauses
		filterClauses, filterArgs := generateFilterClauses(s.FParams, dollarStart)
		whereClauses = append(whereClauses, filterClauses...)
		sqlArgs = append(sqlArgs, filterArgs...)

		whereClause := strings.Join(whereClauses, " AND ")
		if len(whereClauses) > 0 {
			whereClause = fmt.Sprintf("WHERE %s", whereClause)
		}
		q := reqInfoSelect.build(requestInfoTable.Name, whereClause, timeOrder)
		rows, _ := c.QueryContext(ctx, q, sqlArgs...)
		var reqInfos []ReqInfoRow
		err := sqlscan.ScanAll(&reqInfos, rows)
		if err != nil {
			return fmt.Errorf("Error accessing db: %v", err)
		}
		if err := jw.Encode(reqInfos); err != nil {
			return fmt.Errorf("Error writing to output stream: %v", err)
		}
	default:
		return fmt.Errorf("Invalid query name: %v", s.Query)
	}
	return nil
}
