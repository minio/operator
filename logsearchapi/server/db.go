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
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	createTablePartition QTemplate = `CREATE TABLE %s PARTITION OF %s
                                            FOR VALUES FROM ('%s') TO ('%s');`
)

const (
	partitionsPerMonth = 4
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

func (t *Table) getCreatePartitionStatement(partitionNameSuffix, rangeStart, rangeEnd string) string {
	partitionName := fmt.Sprintf("%s_%s", t.Name, partitionNameSuffix)
	return createTablePartition.build(partitionName, t.Name, rangeStart, rangeEnd)
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
)

func getPartitionRange(t time.Time) (time.Time, time.Time) {
	// Zero out the time and use UTC
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	daysInMonth := t.AddDate(0, 1, -t.Day()).Day()
	quot := daysInMonth / partitionsPerMonth
	remDays := daysInMonth % partitionsPerMonth
	rangeStart := t.AddDate(0, 0, 1-t.Day())
	for {
		rangeDays := quot
		if remDays > 0 {
			rangeDays++
			remDays--
		}
		rangeEnd := rangeStart.AddDate(0, 0, rangeDays)
		if t.Before(rangeEnd) {
			return rangeStart, rangeEnd
		}
		rangeStart = rangeEnd
	}
}

// DBClient is a client object that makes requests to the DB.
type DBClient struct {
	*pgxpool.Pool
}

// NewDBClient creates a new DBClient.
func NewDBClient(ctx context.Context, connStr string) (*DBClient, error) {
	pool, err := pgxpool.Connect(ctx, connStr)
	if err != nil {
		return nil, err
	}
	return &DBClient{pool}, nil
}

func (c *DBClient) checkTableExists(ctx context.Context, table string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	const existsQuery QTemplate = `SELECT 1 FROM %s WHERE false;`
	res, _ := c.Query(ctx, existsQuery.build(table))
	if res.Err() != nil {
		// check for table does not exist error
		if strings.Contains(res.Err().Error(), "(SQLSTATE 42P01)") {
			return false, nil
		}
		return false, res.Err()
	}
	return true, nil
}

func (c *DBClient) createTableAndPartition(ctx context.Context, table Table) error {
	if exists, err := c.checkTableExists(ctx, table.Name); err != nil {
		return err
	} else if exists {
		return nil
	}

	if _, err := c.Exec(ctx, table.getCreateStatement()); err != nil {
		return err
	}

	start, end := getPartitionRange(time.Now())
	partSuffix := start.Format("2006_01_02")
	rangeStart, rangeEnd := start.Format("2006-01-02"), end.Format("2006-01-02")
	_, err := c.Exec(ctx, table.getCreatePartitionStatement(partSuffix, rangeStart, rangeEnd))
	return err
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
	tx, err := c.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// NOTE: Timestamps are nanosecond resolution from MinIO, however we are
	// using storing it with only microsecond precision in PG for simplicity
	// as that is the maximum precision supported by it.
	_, err = tx.Exec(ctx, insertAuditLogEvent.build(auditLogEventsTable.Name), event.Time, event)
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

	_, err = tx.Exec(ctx, insertRequestInfo.build(requestInfoTable.Name),
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

	return tx.Commit(ctx)
}

type logEventRawRow struct {
	EventTime time.Time
	Log       string
}

type logEventRow struct {
	EventTime time.Time
	Log       map[string]interface{}
}

type reqInfoRow struct {
	Time                  time.Time
	APIName               string
	Bucket                string
	Object                string
	TimeToResponseNs      uint64
	RemoteHost            string
	RequestID             string
	UserAgent             string
	ResponseStatus        string
	ResponseStatusCode    int
	RequestContentLength  *uint64
	ResponseContentLength *uint64
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
                                            WHERE %s
                                         ORDER BY time %s
                                           OFFSET $1 LIMIT $2;`
	)

	timeRangeOp := "<="
	timeOrder := "DESC"
	if s.TimeAscending {
		timeRangeOp = ">="
		timeOrder = "ASC"
	}

	jw := json.NewEncoder(w)
	switch s.Query {
	case rawQ:
		timeRangeClause := fmt.Sprintf("event_time %s '%s'", timeRangeOp, s.TimeStart.Format(time.RFC3339Nano))
		q := logEventSelect.build(auditLogEventsTable.Name, timeRangeClause, timeOrder)
		rows, _ := c.Query(ctx, q, s.PageNumber*s.PageSize, s.PageSize)
		var logEventsRaw []logEventRawRow
		if err := pgxscan.ScanAll(&logEventsRaw, rows); err != nil {
			return fmt.Errorf("Error accessing db: %v", err)
		}
		// parse the encoded json string stored in the db into a json
		// object for output
		logEvents := make([]logEventRow, len(logEventsRaw))
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
		timeRangeClause := fmt.Sprintf("time %s '%s'", timeRangeOp, s.TimeStart.Format(time.RFC3339Nano))
		q := reqInfoSelect.build(requestInfoTable.Name, timeRangeClause, timeOrder)
		rows, _ := c.Query(ctx, q, s.PageNumber*s.PageSize, s.PageSize)
		var reqInfos []reqInfoRow
		err := pgxscan.ScanAll(&reqInfos, rows)
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
