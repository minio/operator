// This file is part of MinIO Operator
// Copyright (c) 2021 MinIO, Inc.
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

package utils

import (
	"reflect"
	"testing"

	"github.com/minio/pkg/ellipses"
)

func TestGetDivisibleSize(t *testing.T) {
	testCases := []struct {
		totalSizes []uint64
		result     uint64
	}{
		{[]uint64{24, 32, 16}, 8},
		{[]uint64{32, 8, 4}, 4},
		{[]uint64{8, 8, 8}, 8},
		{[]uint64{24}, 24},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run("", func(t *testing.T) {
			gotGCD := getDivisibleSize(testCase.totalSizes)
			if testCase.result != gotGCD {
				t.Errorf("Expected %v, got %v", testCase.result, gotGCD)
			}
		})
	}
}

// Test tests calculating set indexes.
func TestGetSetIndexes(t *testing.T) {
	testCases := []struct {
		args       []string
		totalSizes []uint64
		indexes    [][]uint64
		success    bool
	}{
		// Invalid inputs.
		{
			[]string{"data{1...3}"},
			[]uint64{3},
			nil,
			false,
		},
		{
			[]string{"data/controller1/export{1...2}, data/controller2/export{1...4}, data/controller3/export{1...8}"},
			[]uint64{2, 4, 8},
			nil,
			false,
		},
		{
			[]string{"data{1...17}/export{1...52}"},
			[]uint64{14144},
			nil,
			false,
		},
		// Valid inputs.
		{
			[]string{"data{1...27}"},
			[]uint64{27},
			[][]uint64{{9, 9, 9}},
			true,
		},
		{
			[]string{"http://host{1...3}/data{1...180}"},
			[]uint64{540},
			[][]uint64{{15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15}},
			true,
		},
		{
			[]string{"http://host{1...2}.rack{1...4}/data{1...180}"},
			[]uint64{1440},
			[][]uint64{{16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16}},
			true,
		},
		{
			[]string{"http://host{1...2}/data{1...180}"},
			[]uint64{360},
			[][]uint64{{12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12}},
			true,
		},
		{
			[]string{"data/controller1/export{1...4}, data/controller2/export{1...8}, data/controller3/export{1...12}"},
			[]uint64{4, 8, 12},
			[][]uint64{{4}, {4, 4}, {4, 4, 4}},
			true,
		},
		{
			[]string{"data{1...64}"},
			[]uint64{64},
			[][]uint64{{16, 16, 16, 16}},
			true,
		},
		{
			[]string{"data{1...24}"},
			[]uint64{24},
			[][]uint64{{12, 12}},
			true,
		},
		{
			[]string{"data/controller{1...11}/export{1...8}"},
			[]uint64{88},
			[][]uint64{{11, 11, 11, 11, 11, 11, 11, 11}},
			true,
		},
		{
			[]string{"data{1...4}"},
			[]uint64{4},
			[][]uint64{{4}},
			true,
		},
		{
			[]string{"data/controller1/export{1...10}, data/controller2/export{1...10}, data/controller3/export{1...10}"},
			[]uint64{10, 10, 10},
			[][]uint64{{10}, {10}, {10}},
			true,
		},
		{
			[]string{"data{1...16}/export{1...52}"},
			[]uint64{832},
			[][]uint64{{16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16}},
			true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run("", func(t *testing.T) {
			argPatterns := make([]ellipses.ArgPattern, len(testCase.args))
			for i, arg := range testCase.args {
				patterns, err := ellipses.FindEllipsesPatterns(arg)
				if err != nil {
					t.Fatalf("Unexpected failure %s", err)
				}
				argPatterns[i] = patterns
			}
			gotIndexes, err := getSetIndexes(testCase.args, testCase.totalSizes, argPatterns)
			if err != nil && testCase.success {
				t.Errorf("Expected success but failed instead %s", err)
			}
			if err == nil && !testCase.success {
				t.Errorf("Expected failure but passed instead")
			}
			if !reflect.DeepEqual(testCase.indexes, gotIndexes) {
				t.Errorf("Expected %v, got %v", testCase.indexes, gotIndexes)
			}
		})
	}
}

// Test tests possible parities returned for any input args
func TestPossibleParities(t *testing.T) {
	testCases := []struct {
		arg      string
		parities []string
		success  bool
	}{
		// Tests invalid inputs.
		{
			"...",
			nil,
			false,
		},
		// No range specified.
		{
			"{...}",
			nil,
			false,
		},
		// Invalid range.
		{
			"http://minio{2...3}/export/set{1...0}",
			nil,
			false,
		},
		// Range cannot be smaller than 4 minimum.
		{
			"/export{1..2}",
			nil,
			false,
		},
		// Unsupported characters.
		{
			"/export/test{1...2O}",
			nil,
			false,
		},
		// Tests valid inputs.
		{
			"{1...27}",
			[]string{"EC:4", "EC:3", "EC:2"},
			true,
		},
		{
			"/export/set{1...64}",
			[]string{"EC:8", "EC:7", "EC:6", "EC:5", "EC:4", "EC:3", "EC:2"},
			true,
		},
		// Valid input for distributed setup.
		{
			"http://minio{2...3}/export/set{1...64}",
			[]string{"EC:8", "EC:7", "EC:6", "EC:5", "EC:4", "EC:3", "EC:2"},
			true,
		},
		// Supporting some advanced cases.
		{
			"http://minio{1...64}.mydomain.net/data",
			[]string{"EC:8", "EC:7", "EC:6", "EC:5", "EC:4", "EC:3", "EC:2"},
			true,
		},
		{
			"http://rack{1...4}.mydomain.minio{1...16}/data",
			[]string{"EC:8", "EC:7", "EC:6", "EC:5", "EC:4", "EC:3", "EC:2"},
			true,
		},
		// Supporting kubernetes cases.
		{
			"http://minio{0...15}.mydomain.net/data{0...1}",
			[]string{"EC:8", "EC:7", "EC:6", "EC:5", "EC:4", "EC:3", "EC:2"},
			true,
		},
		// No host regex, just disks.
		{
			"http://server1/data{1...32}",
			[]string{"EC:8", "EC:7", "EC:6", "EC:5", "EC:4", "EC:3", "EC:2"},
			true,
		},
		// No host regex, just disks with two position numerics.
		{
			"http://server1/data{01...32}",
			[]string{"EC:8", "EC:7", "EC:6", "EC:5", "EC:4", "EC:3", "EC:2"},
			true,
		},
		// More than 2 ellipses are supported as well.
		{
			"http://minio{2...3}/export/set{1...64}/test{1...2}",
			[]string{"EC:8", "EC:7", "EC:6", "EC:5", "EC:4", "EC:3", "EC:2"},
			true,
		},
		// More than 1 ellipses per argument for standalone setup.
		{
			"/export{1...10}/disk{1...10}",
			[]string{"EC:5", "EC:4", "EC:3", "EC:2"},
			true,
		},
		// IPv6 ellipses with hexadecimal expansion
		{
			"http://[2001:3984:3989::{1...a}]/disk{1...10}",
			[]string{"EC:5", "EC:4", "EC:3", "EC:2"},
			true,
		},
		// IPv6 ellipses with hexadecimal expansion with 3 position numerics.
		{
			"http://[2001:3984:3989::{001...00a}]/disk{1...10}",
			[]string{"EC:5", "EC:4", "EC:3", "EC:2"},
			true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run("", func(t *testing.T) {
			gotPs, err := PossibleParityValues(testCase.arg)
			if err != nil && testCase.success {
				t.Errorf("Expected success but failed instead %s", err)
			}
			if err == nil && !testCase.success {
				t.Errorf("Expected failure but passed instead")
			}
			if !reflect.DeepEqual(testCase.parities, gotPs) {
				t.Errorf("Expected %v, got %v", testCase.parities, gotPs)
			}
		})
	}
}
