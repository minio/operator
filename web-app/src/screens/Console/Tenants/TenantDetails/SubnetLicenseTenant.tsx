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
import { Box, Button, Grid } from "mds";
import styled from "styled-components";
import { DateTime } from "luxon";
import { Link } from "react-router-dom";
import get from "lodash/get";
import { niceBytes } from "../../../../common/utils";
import { SubnetInfo } from "../../License/types";
import { Tenant } from "../../../../api/operatorApi";
import TooltipWrapper from "../../Common/TooltipWrapper/TooltipWrapper";

interface ISubnetLicenseTenant {
  tenant: Tenant | null;
  loadingActivateProduct: any;
  loadingLicenseInfo: boolean;
  licenseInfo: SubnetInfo | undefined;
  activateProduct: any;
}

const LicenseContainer = styled.div(({ theme }) => ({
  "& .licenseInfoValue": {
    textTransform: "none",
    fontSize: 14,
    fontWeight: "bold",
  },
  "&.licenseContainer": {
    position: "relative",
    padding: "20px 52px 0px 28px",
    background: get(theme, "signalColors.info", "#2781B0"),
    boxShadow: "0px 3px 7px #00000014",
    "& h2": {
      color: get(theme, "bgColor", "#fff"),
      marginBottom: 67,
    },
    "& a": {
      textDecoration: "none",
    },
    "& h3": {
      color: get(theme, "bgColor", "#fff"),
      marginBottom: "30px",
      fontWeight: "bold",
    },
    "& h6": {
      color: "#FFFFFF !important",
    },
  },
  "& .licenseInfo": {
    color: get(theme, "bgColor", "#fff"),
    position: "relative",
  },
  "& .licenseInfoTitle": {
    textTransform: "none",
    color: get(theme, "mutedText", "#87888d"),
    fontSize: 11,
  },
  "& .verifiedIcon": {
    width: 96,
    position: "absolute",
    right: 0,
    bottom: 29,
  },
  "& .noUnderLine": {
    textDecoration: "none",
  },
}));

const SubnetLicenseTenant = ({
  tenant,
  loadingActivateProduct,
  loadingLicenseInfo,
  licenseInfo,
  activateProduct,
}: ISubnetLicenseTenant) => {
  const expiryTime = tenant?.subnet_license
    ? DateTime.fromISO(tenant.subnet_license?.expires_at!)
    : DateTime.now();

  return (
    <LicenseContainer
      className={tenant && tenant.subnet_license ? "licenseContainer" : ""}
    >
      {tenant && tenant.subnet_license ? (
        <Fragment>
          <Grid container className={"licenseInfo"}>
            <Grid item xs={6}>
              <Box sx={{ marginBottom: 12 }} className={"licenseInfoTitle"}>
                License
              </Box>
              <Box sx={{ marginBottom: 12 }} className={"licenseInfoValue"}>
                Commercial License
              </Box>
              <Box sx={{ marginBottom: 12 }} className={"licenseInfoTitle"}>
                Organization
              </Box>
              <Box sx={{ marginBottom: 12 }} className={"licenseInfoValue"}>
                {tenant.subnet_license.organization}
              </Box>
              <Box sx={{ marginBottom: 12 }} className={"licenseInfoTitle"}>
                Registered Capacity
              </Box>
              <Box sx={{ marginBottom: 12 }} className={"licenseInfoValue"}>
                {niceBytes(
                  (
                    (tenant.subnet_license?.storage_capacity || 0) *
                    1099511627776
                  ) // 1 Terabyte = 1099511627776 Bytes
                    .toString(10),
                )}
              </Box>
              <Box sx={{ marginBottom: 12 }} className={"licenseInfoTitle"}>
                Expiry Date
              </Box>
              <Box sx={{ marginBottom: 12 }} className={"licenseInfoValue"}>
                {expiryTime.toFormat("yyyy-MM-dd")}
              </Box>
            </Grid>
            <Grid item xs={6}>
              <Box sx={{ marginBottom: 12 }} className={"licenseInfoTitle"}>
                Subscription Plan
              </Box>
              <Box sx={{ marginBottom: 12 }} className={"licenseInfoValue"}>
                {tenant.subnet_license.plan}
              </Box>
              <Box sx={{ marginBottom: 12 }} className={"licenseInfoTitle"}>
                Requestor
              </Box>
              <Box sx={{ marginBottom: 12 }} className={"licenseInfoValue"}>
                {tenant.subnet_license.email}
              </Box>
            </Grid>
            <img
              className={"verifiedIcon"}
              src={"/verified.svg"}
              alt="verified"
            />
          </Grid>
        </Fragment>
      ) : (
        !loadingLicenseInfo && (
          <Box
            withBorders
            sx={{
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
            }}
          >
            {!licenseInfo && (
              <Link
                to={"/license"}
                onClick={(e) => {
                  e.stopPropagation();
                }}
                className={"noUnderLine"}
              >
                <TooltipWrapper tooltip={"Activate Product"}>
                  <Button
                    id={"activate-product"}
                    label={"Activate Product"}
                    onClick={() => false}
                    variant={"callAction"}
                  />
                </TooltipWrapper>
              </Link>
            )}
            {licenseInfo && tenant && (
              <TooltipWrapper tooltip={"Attach License"}>
                <Button
                  id={"attach-license"}
                  disabled={loadingActivateProduct}
                  label={"Attach License"}
                  onClick={() => activateProduct(tenant.namespace, tenant.name)}
                  variant={"callAction"}
                />
              </TooltipWrapper>
            )}
          </Box>
        )
      )}
    </LicenseContainer>
  );
};

export default SubnetLicenseTenant;
