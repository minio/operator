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

import React, { Fragment } from "react";
import { breakPoints, DrivesIcon, Grid, Box } from "mds";
import { useNavigate } from "react-router-dom";
import { CapacityValues, ValueUnit } from "./types";
import { setTenantName } from "../tenantsSlice";
import { getTenantAsync } from "../thunks/tenantDetailsAsync";
import { niceBytes, niceBytesInt } from "../../../../common/utils";
import InformationItem from "./InformationItem";
import TenantCapacity from "./TenantCapacity";
import { useAppDispatch } from "../../../../store";
import { TenantList } from "../../../../api/operatorApi";
import styled from "styled-components";
import get from "lodash/get";

const TenantListItemMain = styled.div(({ theme }) => ({
  border: `${get(theme, "borderColor", "#eaeaea")} 1px solid`,
  borderRadius: 3,
  padding: 15,
  cursor: "pointer",
  "&.disabled": {
    backgroundColor: get(theme, "signalColors.danger", "red"),
  },
  "&:hover": {
    backgroundColor: get(theme, "boxBackground", "#FBFAFA"),
  },
  "& .tenantTitle": {
    display: "flex",
    alignItems: "center",
    justifyContent: "space-between",
    gap: 10,
    "& h1": {
      padding: 0,
      margin: 0,
      marginBottom: 5,
      fontSize: 22,
      color: get(theme, "screenTitle.iconColor", "#07193E"),
      [`@media (max-width: ${breakPoints.md}px)`]: {
        marginBottom: 0,
      },
    },
  },
  "& .tenantDetails": {
    display: "flex",
    gap: 40,
    "& span": {
      fontSize: 14,
    },
    [`@media (max-width: ${breakPoints.md}px)`]: {
      flexFlow: "column-reverse",
      gap: 5,
    },
  },
  "& .tenantMetrics": {
    display: "flex",
    alignItems: "center",
    marginTop: 20,
    gap: 25,
    borderTop: `${get(theme, "borderColor", "#E2E2E2")} 1px solid`,
    paddingTop: 20,
    "& svg.tenantIcon": {
      color: get(theme, "screenTitle.iconColor", "#07193E"),
      fill: get(theme, "screenTitle.iconColor", "#07193E"),
    },
    "& .metric": {
      "& .min-icon": {
        color: get(theme, "fontColor", "#000"),
        width: 13,
        marginRight: 5,
      },
    },
    "& .metricLabel": {
      fontSize: 14,
      fontWeight: "bold",
      color: get(theme, "fontColor", "#000"),
    },
    "& .metricText": {
      fontSize: 24,
      fontWeight: "bold",
    },
    "& .unit": {
      fontSize: 12,
      fontWeight: "normal",
    },
    "& .status": {
      fontSize: 12,
      color: get(theme, "mutedText", "#87888d"),
    },
    [`@media (max-width: ${breakPoints.md}px)`]: {
      marginTop: 8,
      paddingTop: 8,
    },
  },
  "& .namespaceLabel": {
    display: "inline-flex",
    color: get(theme, "signalColors.dark", "#000"),
    backgroundColor: get(theme, "borderColor", "#E2E2E2"),
    borderRadius: 2,
    padding: "4px 8px",
    fontSize: 10,
    marginRight: 20,
  },
  "& .redState": {
    color: get(theme, "signalColors.danger", "#C51B3F"),
    "& .min-icon": {
      width: 16,
      height: 16,
      float: "left",
      marginRight: 4,
    },
  },
  "& .yellowState": {
    color: get(theme, "signalColors.warning", "#FFBD62"),
    "& .min-icon": {
      width: 16,
      height: 16,
      float: "left",
      marginRight: 4,
    },
  },
  "& .greenState": {
    color: get(theme, "signalColors.good", "#4CCB92"),
    "& .min-icon": {
      width: 16,
      height: 16,
      float: "left",
      marginRight: 4,
    },
  },
  "& .greyState": {
    color: get(theme, "signalColors.disabled", "#E6EBEB"),
    "& .min-icon": {
      width: 16,
      height: 16,
      float: "left",
      marginRight: 4,
    },
  },
}));

const TenantListItem = ({ tenant }: { tenant: TenantList }) => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();

  const healthStatusToClass = (health_status: string) => {
    switch (health_status) {
      case "red":
        return "redState";
      case "yellow":
        return "yellowState";
      case "green":
        return "greenState";
      default:
        return "greyState";
    }
  };

  let raw: ValueUnit = { value: "n/a", unit: "" };
  let capacity: ValueUnit = { value: "n/a", unit: "" };
  let used: ValueUnit = { value: "n/a", unit: "" };
  let localUse: ValueUnit = { value: "n/a", unit: "" };
  let tieredUse: ValueUnit = { value: "n/a", unit: "" };

  if (tenant.capacity_raw) {
    const b = niceBytes(`${tenant.capacity_raw}`, true);
    const parts = b.split(" ");
    raw.value = parts[0];
    raw.unit = parts[1];
  }
  if (tenant.capacity) {
    const b = niceBytes(`${tenant.capacity}`, true);
    const parts = b.split(" ");
    capacity.value = parts[0];
    capacity.unit = parts[1];
  }
  if (tenant.capacity_usage) {
    const b = niceBytesInt(tenant.capacity_usage, true);
    const parts = b.split(" ");
    used.value = parts[0];
    used.unit = parts[1];
  }

  let spaceVariants: CapacityValues[] = [];
  if (!tenant.tiers || tenant.tiers.length === 0) {
    spaceVariants = [
      { value: tenant.capacity_usage || 0, variant: "STANDARD" },
    ];
  } else {
    spaceVariants = tenant.tiers?.map((itemTenant) => {
      return { value: itemTenant.size!, variant: itemTenant.name! };
    });
    let internalUsage = tenant.tiers
      ?.filter((itemTenant) => {
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

  const openTenantDetails = () => {
    dispatch(
      setTenantName({
        name: tenant.name!,
        namespace: tenant.namespace!,
      }),
    );
    dispatch(getTenantAsync());
    navigate(`/namespaces/${tenant.namespace}/tenants/${tenant.name}/summary`);
  };

  return (
    <Fragment>
      <TenantListItemMain
        id={`list-tenant-${tenant.name}`}
        onClick={openTenantDetails}
      >
        <Grid container>
          <Grid item xs={12} className={"tenantTitle"}>
            <Box>
              <h1>{tenant.name}</h1>
            </Box>
            <Box>
              <span className={"namespaceLabel"}>
                Namespace:&nbsp;{tenant.namespace}
              </span>
            </Box>
          </Grid>
          <Grid item xs={12} sx={{ marginTop: 2 }}>
            <Grid container>
              <Grid item xs={2}>
                <TenantCapacity
                  totalCapacity={tenant.capacity || 0}
                  usedSpaceVariants={spaceVariants}
                  statusClass={healthStatusToClass(tenant.health_status!)}
                />
              </Grid>
              <Grid item xs={7}>
                <Grid
                  item
                  xs
                  sx={{
                    display: "flex",
                    justifyContent: "flex-start",
                    alignItems: "center",
                    marginTop: "10px",
                  }}
                >
                  <InformationItem
                    label={"Raw Capacity"}
                    value={raw.value}
                    unit={raw.unit}
                  />
                  <InformationItem
                    label={"Usable Capacity"}
                    value={capacity.value}
                    unit={capacity.unit}
                  />
                  <InformationItem
                    label={"Pools"}
                    value={`${tenant.pool_count}`}
                  />
                </Grid>
                <Grid
                  item
                  xs={12}
                  sx={{ paddingLeft: "20px", marginTop: "15px" }}
                >
                  <span className={"status"}>
                    <strong>State:</strong> {tenant.currentState}
                  </span>
                </Grid>
              </Grid>
              <Grid item xs={3}>
                <Fragment>
                  <Grid container sx={{ gap: 20 }}>
                    <Grid
                      item
                      xs={2}
                      sx={{
                        display: "flex",
                        flexDirection: "column",
                        alignItems: "center",
                      }}
                    >
                      <DrivesIcon className={"muted"} style={{ width: 25 }} />
                      <Box
                        className={"muted"}
                        sx={{
                          fontSize: 12,
                          fontWeight: "400",
                        }}
                      >
                        Usage
                      </Box>
                    </Grid>
                    <Grid item xs={9} sx={{ paddingTop: 8 }}>
                      {(!tenant.tiers || tenant.tiers.length === 0) && (
                        <Box
                          sx={{
                            fontSize: 14,
                            fontWeight: 400,
                          }}
                        >
                          <span className={"muted"}>Internal: </span>{" "}
                          {`${used.value} ${used.unit}`}
                        </Box>
                      )}

                      {tenant.tiers && tenant.tiers.length > 0 && (
                        <Fragment>
                          <Box
                            sx={{
                              fontSize: 14,
                              fontWeight: 400,
                            }}
                          >
                            <span className={"muted"}>Internal: </span>{" "}
                            {`${localUse.value} ${localUse.unit}`}
                          </Box>
                          <Box
                            sx={{
                              fontSize: 14,
                              fontWeight: 400,
                            }}
                          >
                            <span className={"muted"}>Tiered: </span>{" "}
                            {`${tieredUse.value} ${tieredUse.unit}`}
                          </Box>
                        </Fragment>
                      )}
                    </Grid>
                  </Grid>
                </Fragment>
              </Grid>
            </Grid>
          </Grid>
        </Grid>
      </TenantListItemMain>
    </Fragment>
  );
};

export default TenantListItem;
