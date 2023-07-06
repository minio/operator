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

import request from "superagent";
import get from "lodash/get";
import { clearSession } from "../utils";
import { ErrorResponseHandler } from "../types";
import { baseUrl } from "../../history";

type RequestHeaders = { [name: string]: string };

export class API {
  invoke(method: string, url: string, data?: object, headers?: RequestHeaders) {
    let targetURL = url;
    if (targetURL[0] === "/") {
      targetURL = targetURL.slice(1);
    }
    let req = request(method, targetURL);

    if (headers) {
      for (let k in headers) {
        req.set(k, headers[k]);
      }
    }

    return req
      .send(data)
      .then((res) => res.body)
      .catch((err) => {
        // if we get unauthorized and we are not doing login, kick out the user
        if (
          err.status === 401 &&
          localStorage.getItem("userLoggedIn") &&
          !targetURL.includes("api/v1/login")
        ) {
          if (window.location.pathname !== "/") {
            localStorage.setItem("redirect-path", window.location.pathname);
          }
          clearSession();
          // Refresh the whole page to ensure cache is clear
          // and we dont end on an infinite loop
          window.location.href = `${baseUrl}login`;
          return;
        }

        return this.onError(err);
      });
  }

  onError(err: any) {
    if (err.status) {
      const errMessage =
        get(err.response, "body.message", `Error ${err.status.toString()}`) ||
        "";

      let detailedMessage = get(err.response, "body.detailedMessage", "") || "";

      if (errMessage === detailedMessage) {
        detailedMessage = "";
      }

      const capMessage =
        errMessage.charAt(0).toUpperCase() + errMessage.slice(1);
      const capDetailed =
        detailedMessage.charAt(0).toUpperCase() + detailedMessage.slice(1);

      const throwMessage: ErrorResponseHandler = {
        errorMessage: capMessage,
        detailedError: capDetailed,
        statusCode: err.status,
      };

      return Promise.reject(throwMessage);
    } else {
      clearSession();
      window.location.href = `${baseUrl}login`;
    }
  }
}

const api = new API();
export default api;
