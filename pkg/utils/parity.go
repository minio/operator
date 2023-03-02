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
	"errors"
	"fmt"
	"sort"

	"github.com/minio/pkg/ellipses"
)

// This file implements and supports ellipses pattern for
// `minio server` command line arguments.

// Supported set sizes this is used to find the optimal
// single set size.
var setSizes = []uint64{4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

// getDivisibleSize - returns a greatest common divisor of
// all the ellipses sizes.
func getDivisibleSize(totalSizes []uint64) (result uint64) {
	gcd := func(x, y uint64) uint64 {
		for y != 0 {
			x, y = y, x%y
		}
		return x
	}
	result = totalSizes[0]
	for i := 1; i < len(totalSizes); i++ {
		result = gcd(result, totalSizes[i])
	}
	return result
}

// isValidSetSize - checks whether given count is a valid set size for erasure coding.
var isValidSetSize = func(count uint64) bool {
	return (count >= setSizes[0] && count <= setSizes[len(setSizes)-1])
}

// possibleSetCountsWithSymmetry returns symmetrical setCounts based on the
// input argument patterns, the symmetry calculation is to ensure that
// we also use uniform number of drives common across all ellipses patterns.
func possibleSetCountsWithSymmetry(setCounts []uint64, argPatterns []ellipses.ArgPattern) []uint64 {
	newSetCounts := make(map[uint64]struct{})
	for _, ss := range setCounts {
		var symmetry bool
		for _, argPattern := range argPatterns {
			for _, p := range argPattern {
				if uint64(len(p.Seq)) > ss {
					symmetry = uint64(len(p.Seq))%ss == 0
				} else {
					symmetry = ss%uint64(len(p.Seq)) == 0
				}
			}
		}
		// With no arg patterns, it is expected that user knows
		// the right symmetry, so either ellipses patterns are
		// provided (recommended) or no ellipses patterns.
		if _, ok := newSetCounts[ss]; !ok && (symmetry || argPatterns == nil) {
			newSetCounts[ss] = struct{}{}
		}
	}

	setCounts = []uint64{}
	for setCount := range newSetCounts {
		setCounts = append(setCounts, setCount)
	}

	// Not necessarily needed but it ensures to the readers
	// eyes that we prefer a sorted setCount slice for the
	// subsequent function to figure out the right common
	// divisor, it avoids loops.
	sort.Slice(setCounts, func(i, j int) bool {
		return setCounts[i] < setCounts[j]
	})

	return setCounts
}

func commonSetDriveCount(divisibleSize uint64, setCounts []uint64) (setSize uint64) {
	// prefers setCounts to be sorted for optimal behavior.
	if divisibleSize < setCounts[len(setCounts)-1] {
		return divisibleSize
	}

	// Figure out largest value of total_drives_in_erasure_set which results
	// in least number of total_drives/total_drives_erasure_set ratio.
	prevD := divisibleSize / setCounts[0]
	for _, cnt := range setCounts {
		if divisibleSize%cnt == 0 {
			d := divisibleSize / cnt
			if d <= prevD {
				prevD = d
				setSize = cnt
			}
		}
	}
	return setSize
}

// getSetIndexes returns list of indexes which provides the set size
// on each index, this function also determines the final set size
// The final set size has the affinity towards choosing smaller
// indexes (total sets)
func getSetIndexes(args []string, totalSizes []uint64, argPatterns []ellipses.ArgPattern) (setIndexes [][]uint64, err error) {
	if len(totalSizes) == 0 || len(args) == 0 {
		return nil, errors.New("invalid argument")
	}

	setIndexes = make([][]uint64, len(totalSizes))
	for _, totalSize := range totalSizes {
		// Check if totalSize has minimum range upto setSize
		if totalSize < setSizes[0] {
			return nil, fmt.Errorf("incorrect number of endpoints provided %s", args)
		}
	}

	commonSize := getDivisibleSize(totalSizes)
	possibleSetCounts := func(setSize uint64) (ss []uint64) {
		for _, s := range setSizes {
			if setSize%s == 0 {
				ss = append(ss, s)
			}
		}
		return ss
	}

	setCounts := possibleSetCounts(commonSize)
	if len(setCounts) == 0 {
		err = fmt.Errorf("incorrect number of endpoints provided %s, number of disks %d is not divisible by any supported erasure set sizes %d", args, commonSize, setSizes)
		return nil, err
	}

	// Returns possible set counts with symmetry.
	setCounts = possibleSetCountsWithSymmetry(setCounts, argPatterns)
	if len(setCounts) == 0 {
		err = fmt.Errorf("no symmetric distribution detected with input endpoints provided %s, disks %d cannot be spread symmetrically by any supported erasure set sizes %d", args, commonSize, setSizes)
		return nil, err
	}

	// Final set size with all the symmetry accounted for.
	setSize := commonSetDriveCount(commonSize, setCounts)

	// Check whether setSize is with the supported range.
	if !isValidSetSize(setSize) {
		err = fmt.Errorf("incorrect number of endpoints provided %s, number of disks %d is not divisible by any supported erasure set sizes %d", args, commonSize, setSizes)
		return nil, err
	}

	for i := range totalSizes {
		for j := uint64(0); j < totalSizes[i]/setSize; j++ {
			setIndexes[i] = append(setIndexes[i], setSize)
		}
	}

	return setIndexes, nil
}

// Return the total size for each argument patterns.
func getTotalSizes(argPatterns []ellipses.ArgPattern) []uint64 {
	var totalSizes []uint64
	for _, argPattern := range argPatterns {
		var totalSize uint64 = 1
		for _, p := range argPattern {
			totalSize *= uint64(len(p.Seq))
		}
		totalSizes = append(totalSizes, totalSize)
	}
	return totalSizes
}

// PossibleParityValues returns possible parities for input args,
// parties are calculated in uniform manner for one pool or
// multiple pools, ensuring that parities returned are common
// and applicable across all pools.
func PossibleParityValues(args ...string) ([]string, error) {
	setIndexes, err := parseEndpointSet(args...)
	if err != nil {
		return nil, err
	}
	maximumParity := setIndexes[0][0] / 2
	var parities []string
	for maximumParity >= 2 {
		parities = append(parities, fmt.Sprintf("EC:%d", maximumParity))
		maximumParity--
	}
	return parities, nil
}

// Parses all arguments and returns an endpointSet which is a collection
// of endpoints following the ellipses pattern, this is what is used
// by the object layer for initializing itself.
func parseEndpointSet(args ...string) (setIndexes [][]uint64, err error) {
	argPatterns := make([]ellipses.ArgPattern, len(args))
	for i, arg := range args {
		patterns, err := ellipses.FindEllipsesPatterns(arg)
		if err != nil {
			return nil, err
		}
		argPatterns[i] = patterns
	}

	return getSetIndexes(args, getTotalSizes(argPatterns), argPatterns)
}
