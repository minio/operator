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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type qType string

const (
	rawQ     qType = "raw"
	reqInfoQ qType = "reqinfo"
)

type fParam string

func stringToFParam(s string) (f fParam, err error) {
	f = fParam(s)
	switch f {
	case "bucket", "object", "api_name", "request_id", "user_agent", "response_status":
	default:
		return "", fmt.Errorf("Unknown filter param: %s", s)
	}
	return
}

// SearchQuery represents a search query.
type SearchQuery struct {
	Query         qType
	TimeStart     *time.Time
	TimeEnd       *time.Time
	TimeAscending bool
	PageNumber    int
	PageSize      int
	FParams       map[fParam]string
}

// searchQueryFromRequest creates a SearchQuery from the search parameters of a
// HTTP request. The query parameters are:
//
// "q" - name of the query (a qType string constant). Required.
//
// "timeStart" - A timestamp bound for the first result to be returned.
// Optional, defaults to current server time. Format is time.RFC3339Nano
//
// "timeAsc" or "timeDesc" - A flag (value is IGNORED) that specifies the
// ordering of results as ASCENDING time or DESCENDING time. Optional, defaults
// to DESCENDING ordering. At most one of these must be specified.
//
// "pageSize" - Maximum number of result records to return in a request.
// Optional, defaults to 10. Allowed range is 10 to 1000.
//
// "pageNo" - 0-based page number of results. Optional, defaults to 0.
//
// "fp" - Repeatable parameter to specify key-value match filters. The format is
// `key:value-pattern`, where key is the name of a field to match on, and
// value-pattern is a glob expression using `.` to signify a single character
// match and a `*` to match any text. For example, `bucket:photos-*` matches any
// bucket with a "photos-" prefix. To match a literal '.' or '*' prefix with
// '\'. To match a literal '\', just double it: '\\'.
func searchQueryFromRequest(r *http.Request) (*SearchQuery, error) {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return nil, err
	}

	q := qType(values.Get("q"))
	if q != rawQ && q != reqInfoQ {
		return nil, fmt.Errorf("Invalid query name: %s", string(q))
	}

	var timeStart *time.Time
	if timeParam := values.Get("timeStart"); timeParam != "" {
		ts, err := parseSQTimeString(timeParam)
		if err != nil {
			return nil, fmt.Errorf("Invalid start date (must be RFC3339 format): %s", timeParam)
		}
		timeStart = &ts
	}

	var timeEnd *time.Time
	if timeParam := values.Get("timeEnd"); timeParam != "" {
		ts, err := parseSQTimeString(timeParam)
		if err != nil {
			return nil, fmt.Errorf("Invalid start date (must be RFC3339 format): %s", timeParam)
		}
		timeEnd = &ts
	}

	var pageSize int = 10
	if psParam := values.Get("pageSize"); psParam != "" {
		pageSize, err = strconv.Atoi(psParam)
		if err != nil {
			return nil, fmt.Errorf("Invalid pageSize parameter: %s", psParam)
		}
		if pageSize < 10 || pageSize > 10000 {
			return nil, fmt.Errorf("pageSize must be between 10 and 10000, got: %d", pageSize)
		}
	}

	var pageNumber int = 0
	if pnParam := values.Get("pageStart"); pnParam != "" {
		pageNumber, err = strconv.Atoi(pnParam)
		if err != nil {
			return nil, fmt.Errorf("Invalid pageStart parameter: %s", pnParam)
		}
	}

	m := map[string][]string(values)

	_, isTimeAsc := m["timeAsc"]
	_, isTimeDesc := m["timeDesc"]
	if isTimeDesc && isTimeAsc {
		return nil, errors.New("both timeasc and timedesc may not be specified")
	}
	var timeAscending bool
	if isTimeAsc {
		timeAscending = true
	}

	var fParams map[fParam]string
	if vs, ok := m["fp"]; ok {
		fParams = make(map[fParam]string)
		for _, v := range vs {
			ps := strings.SplitN(v, ":", 2)
			if len(ps) != 2 {
				return nil, fmt.Errorf("Invalid filter parameter: %s", v)
			}
			key, err := stringToFParam(ps[0])
			if err != nil {
				return nil, err
			}
			fParams[key] = ps[1]
		}
	}

	return &SearchQuery{
		Query:         q,
		TimeStart:     timeStart,
		TimeEnd:       timeEnd,
		TimeAscending: timeAscending,
		PageSize:      pageSize,
		PageNumber:    pageNumber,
		FParams:       fParams,
	}, nil
}

func parseSQTimeString(s string) (r time.Time, err error) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02",
	}
	for _, layout := range layouts {
		r, err = time.Parse(layout, s)
		if err == nil {
			return
		}
	}
	err = fmt.Errorf("Unknown time format %s - RFC3339 time or date is known", s)
	return
}

func generateFilterClauses(m map[fParam]string, dollarStart int) (clauses []string, args []interface{}) {
	for k, v := range m {
		arg, op := v, "="
		if strings.Contains(v, ".") || strings.Contains(v, "*") {
			arg = strings.Replace(arg, ".", "_", -1)
			arg = strings.Replace(arg, "*", "%", -1)
			op = "LIKE"
		}

		clause := fmt.Sprintf("%s %s $%d", k, op, dollarStart)
		clauses = append(clauses, clause)
		args = append(args, arg)
		dollarStart++
	}
	return
}
