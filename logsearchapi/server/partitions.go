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
	"fmt"
	"log"
	"time"

	"github.com/georgysavva/scany/sqlscan"
)

const (
	createTablePartition QTemplate = `CREATE TABLE %s PARTITION OF %s
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

type partitionTimeRange struct {
	GivenTime          time.Time
	StartDate, EndDate time.Time
}

// newPartitionTimeRange computes the partitionTimeRange including the
// givenTime.
func newPartitionTimeRange(givenTime time.Time) partitionTimeRange {
	// Zero out the time and use UTC
	t := givenTime
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
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
	rows, _ := c.QueryContext(ctx, q)
	var childTables []childTableInfo
	if err := sqlscan.ScanAll(&childTables, rows); err != nil {
		return nil, fmt.Errorf("Error accessing db: %v", err)
	}
	for _, ct := range childTables {
		tableNames = append(tableNames, ct.Child)
	}

	return tableNames, nil
}

func (c *DBClient) getTablesDiskUsage(ctx context.Context) (m map[Table]map[string]uint64, _ error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	const (
		tableSize QTemplate = `SELECT pg_total_relation_size('%s');`
	)

	m = make(map[Table]map[string]uint64, len(allTables))
	for _, table := range allTables {
		parts, err := c.getExistingPartitions(ctx, table)
		if err != nil {
			return nil, err
		}
		cm := make(map[string]uint64, len(parts))
		for _, tableName := range parts {
			q := tableSize.build(tableName)
			row := c.QueryRowContext(ctx, q)
			var size uint64
			if err := row.Scan(&size); err != nil {
				return nil, fmt.Errorf("Unable to query relation size: %v", err)
			}
			cm[tableName] = size
		}
		m[table] = cm
	}
	return m, nil
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

func totalDiskUsage(m map[Table]map[string]uint64) (sum uint64) {
	for _, cm := range m {
		for _, v := range cm {
			sum += v
		}
	}
	return
}

func calculateHiLoWaterMarks(totalCap uint64) (hi, lo float64) {
	const (
		highWaterMarkPercent = 90
		lowWaterMarkPercent  = 70
	)
	return highWaterMarkPercent * float64(totalCap) / 100, lowWaterMarkPercent * float64(totalCap) / 100
}

func (c *DBClient) maintainLowWatermarkUsage(ctx context.Context, diskCapacityGBs int) (err error) {
	tables := make(map[Table][]string)
	for _, table := range allTables {
		tables[table], err = c.getExistingPartitions(ctx, table)
		if err != nil {
			return err
		}
	}
	du, err := c.getTablesDiskUsage(ctx)
	if err != nil {
		return err
	}

	totalUsage := totalDiskUsage(du)
	diskCap := uint64(diskCapacityGBs) * 1024 * 1024 * 1024
	hi, lo := calculateHiLoWaterMarks(diskCap)
	var index int
	if float64(totalUsage) <= hi {
		return nil
	}

	// Delete oldest child tables in each parent table, until usage is below
	// `lo`.
	for float64(totalUsage) >= lo {
		var recoveredSpace uint64
		for _, table := range allTables {
			if index >= len(tables[table]) {
				break
			}
			tableName := tables[table][index]
			err = c.deleteChildTable(ctx, tableName, "disk usage high-water mark reached")
			if err != nil {
				return err
			}
			recoveredSpace += du[table][tableName]
		}
		totalUsage -= recoveredSpace
		index++
	}
	log.Printf("Current tables disk usage: %.1f GB", float64(totalUsage)/float64(1024*1024*1024))
	return nil
}

// vacuumData should be called in a new go routine.
func (c *DBClient) vacuumData(diskCapacityGBs int) {
	var (
		normalInterval time.Duration = 1 * time.Hour
		retryInterval  time.Duration = 2 * time.Minute
	)
	timer := time.NewTimer(normalInterval)
	defer timer.Stop()

	for range timer.C {
		timer.Reset(retryInterval) // timer fired, reset it right here.

		err := c.maintainLowWatermarkUsage(context.Background(), diskCapacityGBs)
		if err != nil {
			log.Printf("Error maintaining high-water mark disk usage: %v (retrying in %s)", err, retryInterval)
			continue
		}
	}
}

func (c *DBClient) partitionTables() {
	checkInterval := 1 * time.Hour
	bgCtx := context.Background()
	tables := []Table{auditLogEventsTable, requestInfoTable}
	for {
		// Check if the partition table to store audit logs 24hrs from now exists
		aDayLater := time.Now().Add(24 * time.Hour)
		for _, table := range tables {
			partitionExists, err := c.checkPartitionTableExists(bgCtx, table.Name, aDayLater)
			if err != nil {
				log.Printf("Error while checking if partition for %s exists %s", table.Name, err)
			}

			if partitionExists {
				continue
			}

			if err := c.createTablePartition(bgCtx, table); err != nil {
				log.Printf("Error while creating partition for %s", table.Name)
			}
		}

		// Check again after `checkInterval` hrs
		time.Sleep(checkInterval)
	}
}
