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

// Close codes for websockets defined in RFC 6455
export const WSCloseNormalClosure = 1000;
export const WSCloseCloseGoingAway = 1001;
export const WSCloseAbnormalClosure = 1006;
export const WSClosePolicyViolation = 1008;
export const WSCloseInternalServerErr = 1011;

export const wsProtocol = (protocol: string): string => {
  let wsProtocol = "ws";
  if (protocol === "https:") {
    wsProtocol = "wss";
  }
  return wsProtocol;
};
