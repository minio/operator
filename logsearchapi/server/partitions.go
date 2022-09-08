//
// This file is part of MinIO Operator
// Copyright (C) 2020-2022, MinIO, Inc.
//
// This code is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License, version 3,
// as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License, version 3,
// along with this program.  If not, see <http://www.gnu.org/licenses/>
//

package server

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/georgysavva/scany/sqlscan"
)

const (
	createTablePartition QTemplate = `CREATE TABLE IF NOT EXISTS %s PARTITION OF %s
                                            FOR VALUES FROM ('%s') TO ('%s');`
)

const (
	partitionsPerMonth = 4
)

func (t *Table) getCreatePartitionStatement(p partitionTimeRange) string {
	partitionName := fmt.Sprintf("%s_%s", t.Name, p.getPartnameSuffix())
	start, end := p.getRangeArgs()
	return createTablePartition.build(partitionName, t.Name, start, end)
}

// partitionTimeRange is created from a given time by `newPartitionTimeRange`.
// It represents an interval of dates (i.e whole days) within the same month
// including the given time.
type partitionTimeRange struct {
	GivenTime          time.Time
	StartDate, EndDate time.Time
}

// newPartitionTimeRange computes the partitionTimeRange including the
// givenTime. For a fixed value of partitionsPerMonth, the days in a month are
// always partitioned in the same way regardless of the given time.
//
// Using partitionsPerMonth = 4:
//
// - the partitions for a 28 day month have number of days: [7,7,7,7]
//
// - the partitions for a 30 day month have number of days: [8,8,7,7]
//
// - the partitions for a 31 day month have number of days: [8,8,8,7]
func newPartitionTimeRange(givenTime time.Time) partitionTimeRange {
	// Convert to UTC and zero out the time.
	t := givenTime.In(time.UTC)
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)

	// Find the number of days in the month.
	lastDateOfMonth := t.AddDate(0, 1, -t.Day())
	daysInMonth := lastDateOfMonth.Day()

	quot := daysInMonth / partitionsPerMonth
	remDays := daysInMonth % partitionsPerMonth
	rangeStart := t.AddDate(0, 0, 1-t.Day())
	var rangeEnd time.Time
	for {
		rangeDays := quot
		if remDays > 0 {
			rangeDays++
			remDays--
		}
		rangeEnd = rangeStart.AddDate(0, 0, rangeDays)
		if t.Before(rangeEnd) {
			break
		}
		rangeStart = rangeEnd
	}
	return partitionTimeRange{
		GivenTime: givenTime,
		StartDate: rangeStart,
		EndDate:   rangeEnd,
	}
}

func (p *partitionTimeRange) getPartnameSuffix() string {
	return p.StartDate.Format("2006_01_02")
}

func (p *partitionTimeRange) getRangeArgs() (string, string) {
	return p.StartDate.Format("2006_01_02"), p.EndDate.Format("2006_01_02")
}

func (p *partitionTimeRange) String() string {
	return fmt.Sprintf("%s -> %s", p.StartDate.Format(time.RFC3339), p.EndDate.Format(time.RFC3339))
}

// isSame checks if the partitions represented by the arguments are the same.
func (p *partitionTimeRange) isSame(q *partitionTimeRange) bool {
	return p.StartDate == q.StartDate && p.EndDate == q.EndDate
}

func (p *partitionTimeRange) previous() partitionTimeRange {
	return newPartitionTimeRange(p.StartDate.Add(-time.Second))
}

func (p *partitionTimeRange) next() partitionTimeRange {
	return newPartitionTimeRange(p.EndDate)
}

func getPartitionTimeRangeForTable(name string) (partitionTimeRange, error) {
	fmtStr := []rune("2006_01_02")
	runes := []rune(name)

	errFn := func(msg string) error {
		title := "invalid partition name"
		s := ": " + msg
		if msg == "" {
			s = ""
		}
		return fmt.Errorf("%s%s", title, s)
	}

	if len(runes) <= len(fmtStr) {
		return partitionTimeRange{}, errFn("too short")
	}

	// Split out the date part of the table name
	partSuffix := string(runes[len(runes)-len(fmtStr):])
	startTime, err := time.Parse(string(fmtStr), partSuffix)
	if err != nil {
		return partitionTimeRange{}, errFn("bad time value: " + partSuffix)
	}
	return newPartitionTimeRange(startTime), nil
}

type childTableInfo struct {
	ParentSchema string
	Parent       string
	ChildSchema  string
	Child        string
}

// getExistingPartitions returns child tables of the given table in
// lexicographical order.
func (c *DBClient) getExistingPartitions(ctx context.Context, t Table) (tableNames []string, _ error) {
	const (
		listPartitions QTemplate = `SELECT nmsp_parent.nspname AS parent_schema,
                                                   parent.relname      AS parent,
                                                   nmsp_child.nspname  AS child_schema,
                                                   child.relname       AS child
                                              FROM pg_inherits
                                                   JOIN pg_class parent            ON pg_inherits.inhparent = parent.oid
                                                   JOIN pg_class child             ON pg_inherits.inhrelid   = child.oid
                                                   JOIN pg_namespace nmsp_parent   ON nmsp_parent.oid  = parent.relnamespace
                                                   JOIN pg_namespace nmsp_child    ON nmsp_child.oid   = child.relnamespace
                                             WHERE parent.relname='%s'
                                          ORDER BY child.relname ASC;`
	)

	q := listPartitions.build(t.Name)
	rows, err := c.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("Error listing partitions for %s: %v", t.Name, err)
	}

	var childTables []childTableInfo
	if err := sqlscan.ScanAll(&childTables, rows); err != nil {
		return nil, fmt.Errorf("Error accessing db: %v", err)
	}
	for _, ct := range childTables {
		tableNames = append(tableNames, ct.Child)
	}

	return tableNames, nil
}

func (c *DBClient) getTableDiskUsage(ctx context.Context, tableName string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	const (
		tableSize QTemplate = `SELECT pg_total_relation_size('%s');`
	)

	q := tableSize.build(tableName)
	row := c.QueryRowContext(ctx, q)
	var size int64
	err := row.Scan(&size)
	return size, err
}

func (c *DBClient) deleteChildTable(ctx context.Context, table, reason string) error {
	q := fmt.Sprintf("DROP TABLE %s;", table)
	_, err := c.ExecContext(ctx, q)
	if err != nil {
		return fmt.Errorf("Table deletion error for %s: %v (attempted for reason: %s)", table, err, reason)
	}
	log.Printf("Deleted table `%s` (%s)", table, reason)
	return nil
}

func calculateHiLoWaterMarks(totalCap uint64) (hi, lo float64) {
	const (
		highWaterMarkPercent = 90
		lowWaterMarkPercent  = 70
	)
	return highWaterMarkPercent * float64(totalCap) / 100, lowWaterMarkPercent * float64(totalCap) / 100
}

// getEarliestPartitionStartTime - finds the earliest start time of all existing
// table partitions - this is the minimum start time over the first existing
// partitions for each parent table.
func getEarliestPartitionStartTime(tables map[Table][]string, indices []int) (time.Time, error) {
	var earliestStartTime time.Time
	isSet := false
	for i, table := range allTables {
		pt, err := getPartitionTimeRangeForTable(tables[table][indices[i]])
		if err != nil {
			return time.Time{}, err
		}
		if !isSet {
			earliestStartTime = pt.StartDate
			isSet = true
		}
		if earliestStartTime.After(pt.StartDate) {
			earliestStartTime = pt.StartDate
		}
	}
	return earliestStartTime, nil
}

func (c *DBClient) maintainLowWatermarkUsage(ctx context.Context, diskCapacityGBs int) (err error) {
	tables := make(map[Table][]string, len(allTables))
	du := make(map[Table]map[string]int64, len(allTables))
	var totalUsage int64
	for _, table := range allTables {
		// Find partitions for the parent table `table`.
		tables[table], err = c.getExistingPartitions(ctx, table)
		if err != nil {
			return err
		}

		// Query disk usage of each partition in pg
		m := make(map[string]int64, len(tables[table]))
		for _, partition := range tables[table] {
			size, err := c.getTableDiskUsage(ctx, partition)
			if err != nil {
				return err
			}
			m[partition] = size
			totalUsage += size
		}
		du[table] = m

	}

	diskCap := uint64(diskCapacityGBs) * 1024 * 1024 * 1024
	hi, lo := calculateHiLoWaterMarks(diskCap)

	if float64(totalUsage) <= hi {
		return nil
	}

	// Print out disk usage after deletes - we defer this call because we could
	// exit this func due to an error.
	defer func() {
		log.Printf("Current tables disk usage: %.1f GB", float64(totalUsage)/float64(1024*1024*1024))
	}()

	// Delete oldest child tables in each parent table, until usage is below
	// `lo`.
	//
	// NOTE: Existing partitions for each parent table may not be in sync wrt
	// the time periods they correspond to, due to previous errors in deleting
	// from the db. So we keep track of the indices of the child tables for each
	// parent table to ensure we only delete the oldest tables.
	indices := make([]int, len(allTables))
	for float64(totalUsage) >= lo {
		earliestStartTime, err := getEarliestPartitionStartTime(tables, indices)
		if err != nil {
			return err
		}

		// Quit without deleting the current partition even if we are over the
		// highwater mark!
		currentPartStartTime := newPartitionTimeRange(time.Now()).StartDate
		if earliestStartTime.Equal(currentPartStartTime) {
			type ctinfo struct {
				Name string
				Size int64
			}
			var ct []ctinfo
			var total int64
			for i, table := range allTables {
				name := tables[table][indices[i]]
				ct = append(ct, ctinfo{name, du[table][name]})
				total += du[table][name]
			}

			log.Printf("WARNING: highwater mark reached: no non-current tables exist to delete!"+
				" Please increase the value of "+DiskCapacityEnv+" and ensure disk capacity for PostgreSQL!"+
				" Candidate tables and sizes: %v (total usage: %d)", ct, total)
			break
		}

		// Delete all child tables with the same StartTime = earliestStartTime
		for i, table := range allTables {
			pt, err := getPartitionTimeRangeForTable(tables[table][indices[i]])
			if err != nil {
				return err
			}

			if pt.StartDate.Equal(earliestStartTime) {
				tableName := tables[table][indices[i]]
				err := c.deleteChildTable(ctx, tableName, "disk usage high-water mark reached")
				if err != nil {
					return err
				}
				indices[i]++
				totalUsage -= du[table][tableName]
			}
		}
	}
	return nil
}

// vacuumData should be called in a new go routine.
func (c *DBClient) vacuumData(ctx context.Context, diskCapacityGBs int) {
	normalInterval := 1 * time.Hour
	retryInterval := 2 * time.Minute
	timer := time.NewTimer(normalInterval)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:

			err := c.maintainLowWatermarkUsage(ctx, diskCapacityGBs)
			if err != nil {
				log.Printf("Error maintaining high-water mark disk usage: %v (retrying in %s)", err, retryInterval)
				timer.Reset(retryInterval)
				continue
			}
			timer.Reset(normalInterval)

		case <-ctx.Done():
			log.Println("Vacuum thread exiting.")
			return
		}
	}
}

func (c *DBClient) partitionTables(ctx context.Context) {
	checkInterval := 1 * time.Hour

	timer := time.NewTimer(checkInterval)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			timer.Reset(checkInterval)
			// Check if the partition table to store audit logs 48hrs from now exists
			later := time.Now().Add(48 * time.Hour)
			for _, table := range allTables {
				partitionExists, err := c.checkPartitionTableExists(ctx, table.Name, later)
				if err != nil {
					log.Printf("Error while checking if partition for %s exists %s", table.Name, err)
				}

				if partitionExists {
					continue
				}

				if err := c.createTablePartition(ctx, table, later); err != nil {
					log.Printf("Error while creating partition for %s", table.Name)
				}
			}

		case <-ctx.Done():
			log.Println("Table partitioner thread exiting.")
			return
		}
	}
}
