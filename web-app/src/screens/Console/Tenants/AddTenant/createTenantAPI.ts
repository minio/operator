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

import api from "../../../../common/api";
import get from "lodash/get";
import { NewServiceAccount } from "../../Common/CredentialsPrompt/types";
import { ErrorResponseHandler, ITenantCreator } from "../../../../common/types";

export const createTenantCall = (dataSend: ITenantCreator) => {
  return new Promise<NewServiceAccount>((resolve, reject) => {
    api
      .invoke("POST", `/api/v1/tenants`, dataSend)
      .then((res) => {
        const consoleSAList = get(res, "console", []);

        let newSrvAcc: NewServiceAccount = {
          idp: get(res, "externalIDP", false),
          console: [],
        };

        if (consoleSAList) {
          if (Array.isArray(consoleSAList)) {
            newSrvAcc.console = consoleSAList.map((consoleKey) => {
              return {
                accessKey: consoleKey.access_key,
                secretKey: consoleKey.secret_key,
                api: "s3v4",
                path: "auto",
                url: consoleKey.url,
              };
            });
          } else {
            newSrvAcc = {
              console: {
                accessKey: res.console.access_key,
                secretKey: res.console.secret_key,
                url: res.console.url,
              },
            };
          }
        }
        resolve(newSrvAcc);
      })
      .catch((err: ErrorResponseHandler) => {
        reject(err);
      });
  });
};
