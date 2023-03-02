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

import {
  erasureCodeCalc,
  getBytes,
  niceBytes,
  setMemoryResource,
} from "../utils";

test("A variety of formatting results", () => {
  expect(niceBytes("1024")).toBe("1.0 KiB");
  expect(niceBytes("1048576")).toBe("1.0 MiB");
  expect(niceBytes("1073741824")).toBe("1.0 GiB");
});

test("From value and unit to a number of bytes", () => {
  expect(getBytes("1", "KiB")).toBe("1024");
  expect(getBytes("1", "MiB")).toBe("1048576");
  expect(getBytes("1", "GiB")).toBe("1073741824");
});

test("From value and unit to a number of bytes for kubernetes", () => {
  expect(getBytes("1", "Ki", true)).toBe("1024");
  expect(getBytes("1", "Mi", true)).toBe("1048576");
  expect(getBytes("1", "Gi", true)).toBe("1073741824");
  expect(getBytes("7500", "Gi", true)).toBe("8053063680000");
});

test("Determine the amount of memory to use", () => {
  expect(setMemoryResource(1024, "1024", 1024)).toStrictEqual({
    error: "There are not enough memory resources available",
    limit: 0,
    request: 0,
  });
  expect(setMemoryResource(64, "1099511627776", 34359738368)).toStrictEqual({
    error:
      "The requested memory is greater than the max available memory for the selected number of nodes",
    limit: 0,
    request: 0,
  });
  expect(setMemoryResource(2, "17179869184", 34359738368)).toStrictEqual({
    error: "",
    limit: 34359738368,
    request: 2147483648,
  });
});

test("Determine the correct values for EC Parity calculation", () => {
  expect(erasureCodeCalc([], 50, 5000, 4)).toStrictEqual({
    error: 1,
    defaultEC: "",
    erasureCodeSet: 0,
    maxEC: "",
    rawCapacity: "0",
    storageFactors: [],
  });
  expect(erasureCodeCalc(["EC:2"], 4, 26843545600, 4)).toStrictEqual({
    error: 0,
    storageFactors: [
      {
        erasureCode: "EC:2",
        storageFactor: 2,
        maxCapacity: "53687091200",
        maxFailureTolerations: 2,
      },
    ],
    maxEC: "EC:2",
    rawCapacity: "107374182400",
    erasureCodeSet: 4,
    defaultEC: "EC:2",
  });
});
