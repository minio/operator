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

const (
	// QueryAuthTokenEnv environment variable
	QueryAuthTokenEnv = "LOGSEARCH_QUERY_AUTH_TOKEN"
	// PgConnStrEnv environment variable
	PgConnStrEnv = "LOGSEARCH_PG_CONN_STR"
	// AuditAuthTokenEnv environment variable
	AuditAuthTokenEnv = "LOGSEARCH_AUDIT_AUTH_TOKEN"
	// DiskCapacityEnv environment variable
	DiskCapacityEnv = "LOGSEARCH_DISK_CAPACITY_GB"
)
