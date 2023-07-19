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

import { useNavigate } from "react-router-dom";
import { Theme } from "@mui/material/styles";
import { CapacityValues, ValueUnit } from "./types";
import { setTenantName } from "../tenantsSlice";
import { getTenantAsync } from "../thunks/tenantDetailsAsync";
import { DrivesIcon } from "mds";
import { niceBytes, niceBytesInt } from "../../../../common/utils";
import Grid from "@mui/material/Grid";
import InformationItem from "./InformationItem";
import TenantCapacity from "./TenantCapacity";
import { useAppDispatch } from "../../../../store";
import makeStyles from "@mui/styles/makeStyles";
import { TenantList } from "../../../../api/operatorApi";

const useStyles = makeStyles((theme: Theme) => ({
  redState: {
    color: theme.palette.error.main,
    "& .min-icon": {
      width: 16,
      height: 16,
      float: "left",
      marginRight: 4,
    },
  },
  yellowState: {
    color: theme.palette.warning.main,
    "& .min-icon": {
      width: 16,
      height: 16,
      float: "left",
      marginRight: 4,
    },
  },
  greenState: {
    color: theme.palette.success.main,
    "& .min-icon": {
      width: 16,
      height: 16,
      float: "left",
      marginRight: 4,
    },
  },
  greyState: {
    color: "grey",
    "& .min-icon": {
      width: 16,
      height: 16,
      float: "left",
      marginRight: 4,
    },
  },
  tenantItem: {
    border: "1px solid #EAEAEA",
    marginBottom: 16,
    padding: "15px 30px",
    "&:hover": {
      backgroundColor: "#FAFAFA",
      cursor: "pointer",
    },
  },
  titleContainer: {
    display: "flex",
    justifyContent: "space-between",
    width: "100%",
  },
  title: {
    fontSize: 18,
    fontWeight: "bold",
  },
  namespaceLabel: {
    display: "inline-flex",
    backgroundColor: "#EAEDEF",
    borderRadius: 2,
    padding: "4px 8px",
    fontSize: 10,
    marginRight: 20,
  },
  status: {
    fontSize: 12,
    color: "#8F9090",
  },
}));

const TenantListItem = ({ tenant }: { tenant: TenantList }) => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const classes = useStyles();

  const healthStatusToClass = (health_status: string) => {
    switch (health_status) {
      case "red":
        return classes.redState;
      case "yellow":
        return classes.yellowState;
      case "green":
        return classes.greenState;
      default:
        return classes.greyState;
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
      <div
        className={classes.tenantItem}
        id={`list-tenant-${tenant.name}`}
        onClick={openTenantDetails}
      >
        <Grid container>
          <Grid item xs={12} className={classes.titleContainer}>
            <div className={classes.title}>
              <span>{tenant.name}</span>
            </div>
            <div>
              <span className={classes.namespaceLabel}>
                Namespace:&nbsp;{tenant.namespace}
              </span>
            </div>
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
              <Grid item xs>
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
                    variant={"faded"}
                  />
                </Grid>
                <Grid
                  item
                  xs={12}
                  sx={{ paddingLeft: "20px", marginTop: "15px" }}
                >
                  <span className={classes.status}>
                    <strong>State:</strong> {tenant.currentState}
                  </span>
                </Grid>
              </Grid>
              <Grid item xs={3}>
                <Fragment>
                  <Grid container>
                    <Grid
                      item
                      xs={2}
                      textAlign={"center"}
                      justifyContent={"center"}
                      justifyItems={"center"}
                    >
                      <DrivesIcon
                        style={{ width: 25, color: "rgb(91,91,91)" }}
                      />
                      <div
                        style={{
                          color: "rgb(118, 118, 118)",
                          fontSize: 12,
                          fontWeight: "400",
                        }}
                      >
                        Usage
                      </div>
                    </Grid>
                    <Grid item xs={1} />
                    <Grid item style={{ paddingTop: 8 }}>
                      {(!tenant.tiers || tenant.tiers.length === 0) && (
                        <div
                          style={{
                            fontSize: 14,
                            fontWeight: 400,
                          }}
                        >
                          <span
                            style={{
                              color: "rgb(62,62,62)",
                            }}
                          >
                            Internal:{" "}
                          </span>{" "}
                          {`${used.value} ${used.unit}`}
                        </div>
                      )}

                      {tenant.tiers && tenant.tiers.length > 0 && (
                        <Fragment>
                          <div
                            style={{
                              fontSize: 14,
                              fontWeight: 400,
                            }}
                          >
                            <span
                              style={{
                                color: "rgb(62,62,62)",
                              }}
                            >
                              Internal:{" "}
                            </span>{" "}
                            {`${localUse.value} ${localUse.unit}`}
                          </div>
                          <div
                            style={{
                              fontSize: 14,
                              fontWeight: 400,
                            }}
                          >
                            <span
                              style={{
                                color: "rgb(62,62,62)",
                              }}
                            >
                              Tiered:{" "}
                            </span>{" "}
                            {`${tieredUse.value} ${tieredUse.unit}`}
                          </div>
                        </Fragment>
                      )}
                    </Grid>
                  </Grid>
                </Fragment>
              </Grid>
            </Grid>
          </Grid>
        </Grid>
      </div>
    </Fragment>
  );
};

export default TenantListItem;
