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

export const months = [
  { value: "01", label: "January" },
  { value: "02", label: "February" },
  { value: "03", label: "March" },
  { value: "04", label: "April" },
  { value: "05", label: "May" },
  { value: "06", label: "June" },
  { value: "07", label: "July" },
  { value: "08", label: "August" },
  { value: "09", label: "September" },
  { value: "10", label: "October" },
  { value: "11", label: "November" },
  { value: "12", label: "December" },
];

export const days = Array.from(Array(31), (_, num) => num + 1);

const currentYear = new Date().getFullYear();

export const years = Array.from(
  Array(25),
  (_, numYear) => numYear + currentYear,
);

export const validDate = (year: string, month: string, day: string): any[] => {
  const currentDate = Date.parse(`${year}-${month}-${day}`);

  if (isNaN(currentDate)) {
    return [false, ""];
  }

  const parsedMonth = parseInt(month);
  const parsedDay = parseInt(day);

  const monthForString = parsedMonth < 10 ? `0${parsedMonth}` : parsedMonth;
  const dayForString = parsedDay < 10 ? `0${parsedDay}` : parsedDay;

  const parsedDate = new Date(currentDate).toISOString().split("T")[0];
  const dateString = `${year}-${monthForString}-${dayForString}`;

  return [parsedDate === dateString, dateString];
};

// twoDigitDate gets a two digit string number used for months or days
// returns "NaN" if number is NaN
export const twoDigitDate = (num: number): string => {
  return num < 10 ? `0${num}` : `${num}`;
};
