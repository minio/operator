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

import React, { Fragment } from "react";
import {
  CircleIcon,
  Loader,
  ValuePair,
  Grid,
  Box,
  breakPoints,
  InformativeMessage,
} from "mds";
import get from "lodash/get";
import styled from "styled-components";
import { CapacityValues, ValueUnit } from "../../Tenants/ListTenants/types";
import { niceBytes, niceBytesInt } from "../../../../common/utils";
import { Tenant } from "../../../../api/operatorApi";
import TenantCapacity from "../../Tenants/ListTenants/TenantCapacity";

interface ISummaryUsageBar {
  tenant: Tenant;
  label: string;
  error: string;
  loading: boolean;
  labels?: boolean;
  healthStatus?: string;
}

const TenantCapacityMain = styled.div(({ theme }) => ({
  width: "100%",
  "& .tenantStatus": {
    marginTop: 2,
    color: get(theme, "signalColors.disabled", "#E6EBEB"),
    "& .min-icon": {
      width: 16,
      height: 16,
      marginRight: 4,
    },
    "&.red": {
      color: get(theme, "signalColors.danger", "#C51B3F"),
    },
    "&.yellow": {
      color: get(theme, "signalColors.warning", "#FFBD62"),
    },
    "&.green": {
      color: get(theme, "signalColors.good", "#4CCB92"),
    },
  },
}));

const SummaryUsageBar = ({
  tenant,
  healthStatus,
  loading,
  error,
}: ISummaryUsageBar) => {
  let raw: ValueUnit = { value: "n/a", unit: "" };
  let capacity: ValueUnit = { value: "n/a", unit: "" };
  let used: ValueUnit = { value: "n/a", unit: "" };
  let localUse: ValueUnit = { value: "n/a", unit: "" };
  let tieredUse: ValueUnit = { value: "n/a", unit: "" };

  if (tenant.status?.usage?.raw) {
    const b = niceBytes(`${tenant.status.usage.raw}`, true);
    const parts = b.split(" ");
    raw.value = parts[0];
    raw.unit = parts[1];
  }
  if (tenant.status?.usage?.capacity) {
    const b = niceBytes(`${tenant.status.usage.capacity}`, true);
    const parts = b.split(" ");
    capacity.value = parts[0];
    capacity.unit = parts[1];
  }
  if (tenant.status?.usage?.capacity_usage) {
    const b = niceBytesInt(tenant.status.usage.capacity_usage, true);
    const parts = b.split(" ");
    used.value = parts[0];
    used.unit = parts[1];
  }

  let spaceVariants: CapacityValues[] = [];
  if (!tenant.tiers || tenant.tiers.length === 0) {
    spaceVariants = [
      { value: tenant.status?.usage?.capacity_usage || 0, variant: "STANDARD" },
    ];
  } else {
    spaceVariants = tenant.tiers.map((itemTenant) => {
      return { value: itemTenant.size!, variant: itemTenant.name! };
    });
    let internalUsage = tenant.tiers
      .filter((itemTenant) => {
        return itemTenant.type === "internal";
      })
      .reduce((sum, itemTenant) => sum + itemTenant.size!, 0);
    let tieredUsage = tenant.tiers
      .filter((itemTenant) => {
        return itemTenant.type !== "internal";
      })
      .reduce((sum, itemTenant) => sum + itemTenant.size!, 0);

    const t = niceBytesInt(tieredUsage, true);
    const parts = t.split(" ");
    tieredUse.value = parts[0];
    tieredUse.unit = parts[1];

    const is = niceBytesInt(internalUsage, true);
    const partsInternal = is.split(" ");
    localUse.value = partsInternal[0];
    localUse.unit = partsInternal[1];
  }

  const renderComponent = () => {
    if (!loading) {
      return error !== "" ? (
        <InformativeMessage title={"Error"} message={error} variant={"error"} />
      ) : (
        <TenantCapacityMain>
          <TenantCapacity
            totalCapacity={tenant.status?.usage?.raw || 0}
            usedSpaceVariants={spaceVariants}
            statusClass={""}
            render={"bar"}
          />
          <Box
            sx={{
              display: "flex",
              alignItems: "stretch",
              margin: "0 0 15px 0",
              flexDirection: "row",
              gap: 20,
              [`@media (max-width: ${breakPoints.sm}px)`]: {
                flexDirection: "column",
                gap: 5,
              },
              [`@media (max-width: ${breakPoints.md}px)`]: {
                gap: 15,
              },
            }}
          >
            {(!tenant.tiers || tenant.tiers.length === 0) && (
              <Fragment>
                <ValuePair
                  label={"Internal:"}
                  direction={"row"}
                  value={`${used.value} ${used.unit}`}
                />
              </Fragment>
            )}
            {tenant.tiers && tenant.tiers.length > 0 && (
              <Fragment>
                <ValuePair
                  label={"Internal:"}
                  direction={"row"}
                  value={`${localUse.value} ${localUse.unit}`}
                />
                <ValuePair
                  label={"Tiered:"}
                  direction={"row"}
                  value={`${tieredUse.value} ${tieredUse.unit}`}
                />
              </Fragment>
            )}
            {healthStatus && (
              <ValuePair
                direction={"row"}
                label={"Health:"}
                value={
                  <div className={`tenantStatus ${healthStatus}`}>
                    <CircleIcon />
                  </div>
                }
              />
            )}
          </Box>
        </TenantCapacityMain>
      );
    }

    return null;
  };

  return (
    <React.Fragment>
      {loading && (
        <div style={{ padding: 5 }}>
          <Grid
            item
            xs={12}
            style={{
              textAlign: "center",
            }}
          >
            <Loader style={{ width: 40, height: 40 }} />
          </Grid>
        </div>
      )}
      {renderComponent()}
    </React.Fragment>
  );
};

export default SummaryUsageBar;
