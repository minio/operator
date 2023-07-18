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
import { IAffinityModel } from "../../../../common/types";
import { Tenant } from "../../../../api/operatorApi";

export const getDefaultAffinity = (tenantName: string, poolName: string) => {
  const defaultAffinity: IAffinityModel = {
    podAntiAffinity: {
      requiredDuringSchedulingIgnoredDuringExecution: [
        {
          labelSelector: {
            matchExpressions: [
              {
                key: "v1.min.io/tenant",
                operator: "In",
                values: [tenantName],
              },
              {
                key: "v1.min.io/pool",
                operator: "In",
                values: [poolName],
              },
            ],
          },
          topologyKey: "kubernetes.io/hostname",
        },
      ],
    },
  };
  return defaultAffinity;
};

export const getNodeSelector = (
  labels: string,
  withPodAntiAffinity: boolean,
  tenantName: string,
  poolName: string,
) => {
  // Labels in the form of key1=value1&key2=value2&key3=value3...
  const splittedLabels = labels.split("&");
  let matchExpressions: any = [];

  splittedLabels.forEach((label: string) => {
    const splitKeyValue = label.split("=");
    if (splitKeyValue.length === 2) {
      matchExpressions.push({
        key: splitKeyValue[0],
        operator: "In",
        values: [splitKeyValue[1]],
      });
    }
  });

  const nodeSelector: IAffinityModel = {
    nodeAffinity: {
      requiredDuringSchedulingIgnoredDuringExecution: {
        nodeSelectorTerms: [
          {
            matchExpressions: matchExpressions,
          },
        ],
      },
    },
  };
  if (withPodAntiAffinity) {
    const def = getDefaultAffinity(tenantName, poolName);
    nodeSelector.podAntiAffinity = def.podAntiAffinity;
  }
  return nodeSelector;
};

export const tenantIsOnline = (tenant: Tenant): boolean => {
  if (tenant.currentState === "") {
    return false;
  }
  return !(
    tenant.status &&
    tenant.status.health_status !== "green" &&
    tenant.status.health_status !== "yellow"
  );
};
