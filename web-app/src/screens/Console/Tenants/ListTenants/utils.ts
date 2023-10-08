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

import get from "lodash/get";

export interface Opts {
  label: string;
  value: string;
}

export interface IQuotaElement {
  hard: number;
  name: string;
}

export interface IQuotas {
  elements?: IQuotaElement[];
  name: string;
}

export interface KeyPair {
  id: string;
  encoded_cert: string;
  encoded_key: string;
  cert: string;
  key: string;
}

export const ecListTransform = (
  ecList: string[],
  defaultEC: string = "",
): Opts[] => {
  return ecList.map((value) => {
    let defLabel = value;
    if (defaultEC !== "" && value === defaultEC) {
      defLabel = `${value} (Default)`;
    }

    return {
      label: defLabel,
      value,
    };
  });
};

export const getLimitSizes = (resourceQuotas: IQuotas) => {
  const quotas: IQuotaElement[] = get(resourceQuotas, "elements", []);
  if (quotas === null) {
    return {};
  }

  const returnQuotas: any = {};

  quotas.forEach((rsQuota) => {
    const stCName = rsQuota.name.split(
      ".storageclass.storage.k8s.io/requests.storage",
    )[0];
    const hard = get(rsQuota, "hard", 0);
    const used = get(rsQuota, "used", 0);

    returnQuotas[stCName] = hard - used;
  });

  return returnQuotas;
};
