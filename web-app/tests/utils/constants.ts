// This file is part of MinIO Operator
// Copyright (c) 2022 MinIO, Inc.
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
import { readFileSync } from "fs";

const data = readFileSync(__dirname + "/../constants/timestamp.txt", "utf-8");
const unixTimestamp = data.trim();

export const TEST_BUCKET_NAME = "testbucket-" + unixTimestamp;
export const TEST_GROUP_NAME = "testgroup-" + unixTimestamp;
export const TEST_USER_NAME = "testuser-" + unixTimestamp;
export const TEST_PASSWORD = "password";
export const TEST_IAM_POLICY_NAME = "testpolicy-" + unixTimestamp;
export const TEST_IAM_POLICY = JSON.stringify({
  Version: "2012-10-17",
  Statement: [
    {
      Action: ["admin:*"],
      Effect: "Allow",
      Sid: "",
    },
    {
      Action: ["s3:*"],
      Effect: "Allow",
      Resource: ["arn:aws:s3:::*"],
      Sid: "",
    },
  ],
});
export const TEST_ASSIGN_POLICY_NAME = "consoleAdmin";
