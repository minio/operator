//
// This file is part of MinIO Operator
// Copyright (C) 2022, MinIO, Inc.
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
	"testing"
	"time"
)

func TestNewPartitionTimeRange(t *testing.T) {
	pst, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatalf("Could not load loc: %v", err)
	}
	ist, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		t.Fatalf("Could not load loc: %v", err)
	}

	utc11hr := time.Date(2022, 1, 24, 11, 0, 0, 0, time.UTC)
	pst15hr := time.Date(2022, 1, 24, 15, 48, 0, 0, pst)
	pst16hr := time.Date(2022, 1, 24, 16, 48, 0, 0, pst)
	ist4hr := time.Date(2022, 1, 25, 4, 30, 0, 0, ist)

	testCases := []struct {
		givenTime                  time.Time
		expectedPartitionTimeRange partitionTimeRange
	}{
		{
			givenTime: utc11hr,
			expectedPartitionTimeRange: partitionTimeRange{
				GivenTime: utc11hr,
				StartDate: time.Date(2022, 1, 17, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2022, 1, 25, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			givenTime: pst15hr,
			expectedPartitionTimeRange: partitionTimeRange{
				GivenTime: pst15hr,
				StartDate: time.Date(2022, time.January, 17, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2022, time.January, 25, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			givenTime: pst16hr,
			expectedPartitionTimeRange: partitionTimeRange{
				GivenTime: pst16hr,
				StartDate: time.Date(2022, time.January, 25, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2022, time.February, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			givenTime: ist4hr,
			expectedPartitionTimeRange: partitionTimeRange{
				GivenTime: ist4hr,
				StartDate: time.Date(2022, time.January, 17, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2022, time.January, 25, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for i, testCase := range testCases {
		got := newPartitionTimeRange(testCase.givenTime)
		if got != testCase.expectedPartitionTimeRange {
			t.Errorf("%v:\ngot: %#v\nexpected: %#v", i+1, got, testCase.expectedPartitionTimeRange)
		}
	}
}
