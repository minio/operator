// This file is part of MinIO Console Server
// Copyright (c) 2023 MinIO, Inc.
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

import { ErrorResponseHandler } from "../common/types";
import { Error } from "./operatorApi";

// errorToHandler translates a swagger error to a ErrorResponseHandler which
// is legacy, when all API calls are using the swagger API, we can remove this.
export const errorToHandler = (e: Error): ErrorResponseHandler => {
  if (!e) {
    return {
      statusCode: 0,
      errorMessage: "",
      detailedError: "",
    };
  }
  return {
    statusCode: e.code,
    errorMessage: e.message,
    detailedError: e.detailedMessage,
  };
};
