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
	"math/rand"
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

// Generate a random time in the interval (now - 10 years, now + 10 years)
func randomTime() time.Time {
	dur10years := int64((3650 * 24 * time.Hour).Seconds()) // roughly 10 years in seconds

	// r is a unix time stamp in the desired 20 year interval.
	r := time.Now().Unix() - dur10years + rand.Int63n(2*dur10years)
	return time.Unix(r, 0)
}

func TestPartitionTimeRangeNextPrev(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// For some randomly generated times, test some properties.
	for i := 0; i < 1000; i++ {
		r := randomTime()
		p1 := newPartitionTimeRange(r)
		p0, p2 := p1.previous(), p1.next()

		p0next := p0.next()
		if !p0next.isSame(&p1) {
			t.Errorf("Test %d: r=%v p1=%s p0=%s (p0.next() != p1)", i, r, p1.String(), p0.String())
		}

		p2prev := p2.previous()
		if !p2prev.isSame(&p1) {
			t.Errorf("Test %d: r=%v p1=%s p2=%s (p2.previous() != p1)", i, r, p1.String(), p2.String())
		}

		p0nextnext := p0next.next()
		if !p0nextnext.isSame(&p2) {
			t.Errorf("Test %d: r=%v p0=%s p2=%s (p0.next().next() != p2)", i, r, p0.String(), p2.String())
		}

		p2prevprev := p2prev.previous()
		if !p2prevprev.isSame(&p0) {
			t.Errorf("Test %d: r=%v p0=%s p2=%s (p2.previous().previous() != p0)", i, r, p0.String(), p2.String())
		}

		if p := newPartitionTimeRange(p1.StartDate); !p.isSame(&p1) {
			t.Errorf("Test %d: r=%v p1=%s p=%s (newPartitionTimeRange(p1.StartTime) != p1)", i, r, p1.String(), p.String())
		}

		if p := newPartitionTimeRange(p1.EndDate); !p.isSame(&p2) {
			t.Errorf("Test %d: r=%v p1=%s p2=%s p=%s (newPartitionTimeRange(p1.EndDate) != p2)", i, r, p1.String(), p2.String(), p.String())
		}
	}
}
