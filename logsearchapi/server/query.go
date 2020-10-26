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
	"time"
)

type qType string

const (
	rawQ     qType = "raw"
	reqInfoQ qType = "reqinfo"
)

// SearchQuery represents a search query.
type SearchQuery struct {
	Query         qType
	TimeStart     time.Time
	TimeAscending bool
	PageNumber    int
	PageSize      int
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
func searchQueryFromRequest(r *http.Request) (*SearchQuery, error) {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return nil, err
	}

	q := qType(values.Get("q"))
	if q != rawQ && q != reqInfoQ {
		return nil, fmt.Errorf("Invalid query name: %s", string(q))
	}

	var timeStart time.Time
	if timeParam := values.Get("timeStart"); timeParam != "" {
		timeStart, err = parseSQTimeString(timeParam)
		if err != nil {
			return nil, fmt.Errorf("Invalid start start (must be RFC3339 format): %s", timeParam)
		}
	} else {
		timeStart = time.Now()
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

	return &SearchQuery{
		Query:         q,
		TimeStart:     timeStart,
		TimeAscending: timeAscending,
		PageSize:      pageSize,
		PageNumber:    pageNumber,
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
