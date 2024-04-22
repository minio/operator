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

import React, { Fragment, useEffect, useState } from "react";
import { Loader, SectionTitle, Box, Grid } from "mds";
import { useSelector } from "react-redux";
import { SubnetInfo } from "../../License/types";
import { AppState, useAppDispatch } from "../../../../store";
import { ErrorResponseHandler } from "../../../../common/types";
import SubnetLicenseTenant from "./SubnetLicenseTenant";
import api from "../../../../common/api";

import { setErrorSnackMessage } from "../../../../systemSlice";
import { setTenantDetailsLoad } from "../tenantsSlice";

const TenantLicense = () => {
  const dispatch = useAppDispatch();

  const loadingTenant = useSelector(
    (state: AppState) => state.tenants.loadingTenant,
  );
  const tenant = useSelector((state: AppState) => state.tenants.tenantInfo);

  const [licenseInfo, setLicenseInfo] = useState<SubnetInfo>();
  const [loadingLicenseInfo, setLoadingLicenseInfo] = useState<boolean>(true);
  const [loadingActivateProduct, setLoadingActivateProduct] =
    useState<boolean>(false);

  const activateProduct = (namespace: string, tenant: string) => {
    if (loadingActivateProduct) {
      return;
    }
    setLoadingActivateProduct(true);
    api
      .invoke(
        "POST",
        `/api/v1/subscription/namespaces/${namespace}/tenants/${tenant}/activate`,
        {},
      )
      .then(() => {
        setLoadingActivateProduct(false);
        dispatch(setTenantDetailsLoad(true));
        setLoadingLicenseInfo(true);
      })
      .catch((err: ErrorResponseHandler) => {
        setLoadingActivateProduct(false);
        dispatch(setErrorSnackMessage(err));
      });
  };

  useEffect(() => {
    if (loadingLicenseInfo) {
      api
        .invoke("GET", `/api/v1/subscription/info`)
        .then((res: SubnetInfo) => {
          setLicenseInfo(res);
          setLoadingLicenseInfo(false);
        })
        .catch((err: ErrorResponseHandler) => {
          setLoadingLicenseInfo(false);
        });
    }
  }, [loadingLicenseInfo]);

  return (
    <Fragment>
      <SectionTitle separator sx={{ marginBottom: 15 }}>
        License
      </SectionTitle>
      {loadingTenant ? (
        <Box
          sx={{
            textAlign: "center",
          }}
        >
          <Loader />
        </Box>
      ) : (
        <Fragment>
          {tenant && (
            <Grid container>
              <Grid item xs={12}>
                <SubnetLicenseTenant
                  tenant={tenant}
                  loadingLicenseInfo={loadingLicenseInfo}
                  loadingActivateProduct={loadingActivateProduct}
                  licenseInfo={licenseInfo}
                  activateProduct={activateProduct}
                />
              </Grid>
            </Grid>
          )}
        </Fragment>
      )}
    </Fragment>
  );
};

export default TenantLicense;
