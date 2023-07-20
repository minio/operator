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

import { redirectRule } from "../screens/LoginPage/types";

interface userInterface {
  accessKey: string;
}

interface policyInterface {
  name: string;
}

interface policyDetailsInterface {
  policy: string;
}

export const usersSort = (a: userInterface, b: userInterface) => {
  if (a.accessKey > b.accessKey) {
    return 1;
  }
  if (a.accessKey < b.accessKey) {
    return -1;
  }
  // a must be equal to b
  return 0;
};

export const policySort = (a: policyInterface, b: policyInterface) => {
  if (a.name > b.name) {
    return 1;
  }
  if (a.name < b.name) {
    return -1;
  }
  // a must be equal to b
  return 0;
};

export const stringSort = (a: string, b: string) => {
  if (a > b) {
    return 1;
  }
  if (a < b) {
    return -1;
  }
  // a must be equal to b
  return 0;
};

export const policyDetailsSort = (
  a: policyDetailsInterface,
  b: policyDetailsInterface,
) => {
  if (a.policy > b.policy) {
    return 1;
  }
  if (a.policy < b.policy) {
    return -1;
  }
  // a must be equal to b
  return 0;
};

export const redirectRules = (a: redirectRule, b: redirectRule) => {
  if (a.displayName > b.displayName) {
    return 1;
  }
  if (a.displayName < b.displayName) {
    return -1;
  }
  return 0;
};
