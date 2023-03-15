import React, { Fragment } from "react";
import { Stack } from "@mui/material";
import Grid from "@mui/material/Grid";
import { CapacityValues, ValueUnit } from "../../Tenants/ListTenants/types";
import { CircleIcon, Loader } from "mds";
import { niceBytes, niceBytesInt } from "../../../../common/utils";
import TenantCapacity from "../../Tenants/ListTenants/TenantCapacity";
import ErrorBlock from "../../../shared/ErrorBlock";
import LabelValuePair from "./LabelValuePair";
import { Tenant } from "../../../../api/operatorApi";

interface ISummaryUsageBar {
  tenant: Tenant;
  label: string;
  error: string;
  loading: boolean;
  labels?: boolean;
  healthStatus?: string;
}

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
        <ErrorBlock errorMessage={error} withBreak={false} />
      ) : (
        <Grid item xs={12}>
          <TenantCapacity
            totalCapacity={tenant.status?.usage?.raw || 0}
            usedSpaceVariants={spaceVariants}
            statusClass={""}
            render={"bar"}
          />
          <Stack
            direction={{ xs: "column", sm: "row" }}
            spacing={{ xs: 1, sm: 2, md: 4 }}
            alignItems={"stretch"}
            margin={"0 0 15px 0"}
          >
            {(!tenant.tiers || tenant.tiers.length === 0) && (
              <Fragment>
                <LabelValuePair
                  label={"Internal:"}
                  orientation={"row"}
                  value={`${used.value} ${used.unit}`}
                />
              </Fragment>
            )}
            {tenant.tiers && tenant.tiers.length > 0 && (
              <Fragment>
                <LabelValuePair
                  label={"Internal:"}
                  orientation={"row"}
                  value={`${localUse.value} ${localUse.unit}`}
                />
                <LabelValuePair
                  label={"Tiered:"}
                  orientation={"row"}
                  value={`${tieredUse.value} ${tieredUse.unit}`}
                />
              </Fragment>
            )}
            {healthStatus && (
              <LabelValuePair
                orientation={"row"}
                label={"Health:"}
                value={
                  <span className={healthStatus}>
                    <CircleIcon />
                  </span>
                }
              />
            )}
          </Stack>
        </Grid>
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
