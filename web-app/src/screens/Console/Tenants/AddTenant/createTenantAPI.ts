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

import get from "lodash/get";
import { NewServiceAccount } from "../../Common/CredentialsPrompt/types";
import { ErrorResponseHandler } from "../../../../common/types";
import { api } from "../../../../api";
import {
  CreateTenantRequest,
  CreateTenantResponse,
  Error,
  HttpResponse,
} from "../../../../api/operatorApi";

export const createTenantCall = (dataSend: CreateTenantRequest) => {
  return new Promise<NewServiceAccount>((resolve, reject) => {
    api.tenants
      .createTenant(dataSend)
      .then((res: HttpResponse<CreateTenantResponse, Error>) => {
        const consoleSAList = res.data.console ?? [];

        let newSrvAcc: NewServiceAccount = {
          idp: get(res, "externalIDP", false),
          console: [],
        };

        newSrvAcc.console = consoleSAList.map((consoleKey) => {
          return {
            accessKey: consoleKey.access_key!,
            secretKey: consoleKey.secret_key!,
            api: "s3v4",
            path: "auto",
            url: consoleKey.url!,
          };
        });

        resolve(newSrvAcc);
      })
      .catch((err: ErrorResponseHandler) => {
        reject(err);
      });
  });
};
