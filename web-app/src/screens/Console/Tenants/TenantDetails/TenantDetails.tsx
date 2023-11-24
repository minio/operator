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
import { useSelector } from "react-redux";
import {
  BackLink,
  Box,
  Button,
  CircleIcon,
  EditIcon,
  Grid,
  MinIOTierIconXs,
  PageLayout,
  ProgressBar,
  RefreshIcon,
  ScreenTitle,
  Tabs,
  TenantsIcon,
  TrashIcon,
} from "mds";
import {
  Navigate,
  Route,
  Routes,
  useLocation,
  useNavigate,
  useParams,
} from "react-router-dom";
import styled from "styled-components";
import get from "lodash/get";
import { AppState, useAppDispatch } from "../../../../store";
import { niceBytes } from "../../../../common/utils";
import { IAM_PAGES } from "../../../../common/SecureComponent/permissions";
import { setSnackBarMessage } from "../../../../systemSlice";
import { setTenantName } from "../tenantsSlice";
import { getTenantAsync } from "../thunks/tenantDetailsAsync";
import { tenantIsOnline } from "./utils";
import withSuspense from "../../Common/Components/withSuspense";
import TooltipWrapper from "../../Common/TooltipWrapper/TooltipWrapper";
import PageHeaderWrapper from "../../Common/PageHeaderWrapper/PageHeaderWrapper";

const TenantYAML = withSuspense(React.lazy(() => import("./TenantYAML")));
const TenantSummary = withSuspense(React.lazy(() => import("./TenantSummary")));
const TenantLicense = withSuspense(React.lazy(() => import("./TenantLicense")));
const PoolsSummary = withSuspense(React.lazy(() => import("./PoolsSummary")));
const PodsSummary = withSuspense(React.lazy(() => import("./PodsSummary")));
const TenantEvents = withSuspense(React.lazy(() => import("./TenantEvents")));
const TenantCSR = withSuspense(React.lazy(() => import("./TenantCSR")));
const VolumesSummary = withSuspense(
  React.lazy(() => import("./VolumesSummary")),
);
const TenantMetrics = withSuspense(React.lazy(() => import("./TenantMetrics")));
const TenantTrace = withSuspense(React.lazy(() => import("./TenantTrace")));
const TenantVolumes = withSuspense(
  React.lazy(() => import("./pvcs/TenantVolumes")),
);
const TenantIdentityProvider = withSuspense(
  React.lazy(() => import("./TenantIdentityProvider")),
);
const TenantSecurity = withSuspense(
  React.lazy(() => import("./TenantSecurity")),
);
const TenantEncryption = withSuspense(
  React.lazy(() => import("./TenantEncryption")),
);
const DeleteTenant = withSuspense(
  React.lazy(() => import("../ListTenants/DeleteTenant")),
);
const PodDetails = withSuspense(React.lazy(() => import("./pods/PodDetails")));

const TenantConfiguration = withSuspense(
  React.lazy(() => import("./TenantConfiguration")),
);

const HealthsStatusIcon = styled.div(({ theme }) => ({
  position: "relative",
  fontSize: 10,
  left: 26,
  height: 10,
  top: 4,
  "& .statusIcon": {
    color: get(theme, "signalColors.disabled", "#E6EBEB"),
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

const TenantDetails = () => {
  const dispatch = useAppDispatch();
  const params = useParams();
  const navigate = useNavigate();
  const { pathname = "" } = useLocation();

  const loadingTenant = useSelector(
    (state: AppState) => state.tenants.loadingTenant,
  );
  const selectedTenant = useSelector(
    (state: AppState) => state.tenants.currentTenant,
  );
  const selectedNamespace = useSelector(
    (state: AppState) => state.tenants.currentNamespace,
  );
  const tenantInfo = useSelector((state: AppState) => state.tenants.tenantInfo);

  const tenantName = params.tenantName || "";
  const tenantNamespace = params.tenantNamespace || "";
  const [deleteOpen, setDeleteOpen] = useState<boolean>(false);

  // if the current tenant selected is not the one in the redux, reload it
  useEffect(() => {
    if (
      selectedNamespace !== tenantNamespace ||
      selectedTenant !== tenantName
    ) {
      dispatch(
        setTenantName({
          name: tenantName,
          namespace: tenantNamespace,
        }),
      );
      dispatch(getTenantAsync());
    }
  }, [
    selectedTenant,
    selectedNamespace,
    dispatch,
    tenantName,
    tenantNamespace,
  ]);

  const editYaml = () => {
    navigate(getRoutePath("summary/yaml"));
  };

  const getRoutePath = (newValue: string) => {
    return `/namespaces/${tenantNamespace}/tenants/${tenantName}/${newValue}`;
  };

  const confirmDeleteTenant = () => {
    setDeleteOpen(true);
  };

  const closeDeleteModalAndRefresh = (reloadData: boolean) => {
    setDeleteOpen(false);

    if (reloadData) {
      dispatch(setSnackBarMessage("Tenant Deleted"));
      navigate(`/tenants`);
    }
  };

  return (
    <Fragment>
      {deleteOpen && tenantInfo !== null && (
        <DeleteTenant
          deleteOpen={deleteOpen}
          selectedTenant={tenantInfo}
          closeDeleteModalAndRefresh={closeDeleteModalAndRefresh}
        />
      )}

      <PageHeaderWrapper
        label={
          <Fragment>
            <BackLink
              label="Tenants"
              onClick={() => navigate(IAM_PAGES.TENANTS)}
            />
          </Fragment>
        }
        actions={<Fragment />}
      />

      <PageLayout variant={"constrained"}>
        <Box withBorders={true} customBorderPadding={"0px"}>
          {loadingTenant && (
            <Grid item xs={12}>
              <ProgressBar />
            </Grid>
          )}
          <Grid item xs={12}>
            <ScreenTitle
              icon={
                <Fragment>
                  <HealthsStatusIcon>
                    {tenantInfo && tenantInfo.status && (
                      <span
                        className={`statusIcon ${tenantInfo.status
                          ?.health_status!}`}
                      >
                        <CircleIcon style={{ width: 15, height: 15 }} />
                      </span>
                    )}
                  </HealthsStatusIcon>
                  <TenantsIcon />
                </Fragment>
              }
              title={tenantName}
              subTitle={
                <Fragment>
                  Namespace: {tenantNamespace} / Capacity:{" "}
                  {niceBytes((tenantInfo?.total_size || 0).toString(10))}
                </Fragment>
              }
              actions={
                <Box
                  sx={{ display: "flex", justifyContent: "flex-end", gap: 10 }}
                >
                  <TooltipWrapper tooltip={"Delete"}>
                    <Button
                      id={"delete-tenant"}
                      variant="secondary"
                      onClick={() => {
                        confirmDeleteTenant();
                      }}
                      color="secondary"
                      icon={<TrashIcon />}
                    />
                  </TooltipWrapper>
                  <TooltipWrapper tooltip={"Edit YAML"}>
                    <Button
                      icon={<EditIcon />}
                      id={"yaml_button"}
                      variant="regular"
                      aria-label="Edit YAML"
                      onClick={() => {
                        editYaml();
                      }}
                    />
                  </TooltipWrapper>
                  <TooltipWrapper tooltip={"Management Console"}>
                    <Button
                      id={"tenant-hop"}
                      onClick={() => {
                        navigate(
                          `/namespaces/${tenantNamespace}/tenants/${tenantName}/hop`,
                        );
                      }}
                      disabled={!tenantInfo || !tenantIsOnline(tenantInfo)}
                      variant={"regular"}
                      icon={<MinIOTierIconXs style={{ height: 16 }} />}
                    />
                  </TooltipWrapper>
                  <TooltipWrapper tooltip={"Refresh"}>
                    <Button
                      id={"tenant-refresh"}
                      variant="regular"
                      aria-label="Refresh List"
                      onClick={() => {
                        dispatch(getTenantAsync());
                      }}
                      icon={<RefreshIcon />}
                    />
                  </TooltipWrapper>
                </Box>
              }
            />
          </Grid>

          <Tabs
            currentTabOrPath={pathname}
            useRouteTabs
            onTabClick={(route) => navigate(route)}
            routes={
              <Routes>
                <Route path={"summary"} element={<TenantSummary />} />
                <Route
                  path={"configuration"}
                  element={<TenantConfiguration />}
                />
                <Route path={`summary/yaml`} element={<TenantYAML />} />
                <Route path={"metrics"} element={<TenantMetrics />} />
                <Route path={"trace"} element={<TenantTrace />} />
                <Route
                  path={"identity-provider"}
                  element={<TenantIdentityProvider />}
                />
                <Route path={"security"} element={<TenantSecurity />} />
                <Route path={"encryption"} element={<TenantEncryption />} />
                <Route path={"pools"} element={<PoolsSummary />} />
                <Route path={"pods/:podName"} element={<PodDetails />} />
                <Route path={"pods"} element={<PodsSummary />} />
                <Route path={"pvcs/:PVCName"} element={<TenantVolumes />} />
                <Route path={"volumes"} element={<VolumesSummary />} />
                <Route path={"license"} element={<TenantLicense />} />
                <Route path={"events"} element={<TenantEvents />} />
                <Route path={"csr"} element={<TenantCSR />} />
                <Route
                  path={"/"}
                  element={
                    <Navigate
                      to={`/namespaces/${tenantNamespace}/tenants/${tenantName}/summary`}
                    />
                  }
                />
              </Routes>
            }
            options={[
              {
                tabConfig: {
                  label: "Summary",
                  id: `details-summary`,
                  to: getRoutePath("summary"),
                },
              },
              {
                tabConfig: {
                  label: "Configuration",
                  id: `details-configuration`,
                  to: getRoutePath("configuration"),
                },
              },
              {
                tabConfig: {
                  label: "Metrics",
                  id: `details-metrics`,
                  to: getRoutePath("metrics"),
                },
              },
              {
                tabConfig: {
                  label: "Identity Provider",
                  id: `details-idp`,
                  to: getRoutePath("identity-provider"),
                },
              },
              {
                tabConfig: {
                  label: "Security",
                  id: `details-security`,
                  to: getRoutePath("security"),
                },
              },
              {
                tabConfig: {
                  label: "Encryption",
                  id: `details-encryption`,
                  to: getRoutePath("encryption"),
                },
              },
              {
                tabConfig: {
                  label: "Pools",
                  id: `details-pools`,
                  to: getRoutePath("pools"),
                },
              },
              {
                tabConfig: {
                  label: "Pods",
                  id: "tenant-pod-tab",
                  to: getRoutePath("pods"),
                },
              },

              {
                tabConfig: {
                  label: "Volumes",
                  id: `details-volumes`,
                  to: getRoutePath("volumes"),
                },
              },
              {
                tabConfig: {
                  label: "Events",
                  id: `details-events`,
                  to: getRoutePath("events"),
                },
              },
              {
                tabConfig: {
                  label: "Certificate Requests",
                  id: `details-csr`,
                  to: getRoutePath("csr"),
                },
              },
              {
                tabConfig: {
                  label: "License",
                  id: `details-license`,
                  to: getRoutePath("license"),
                },
              },
            ]}
          />
        </Box>
      </PageLayout>
    </Fragment>
  );
};

export default TenantDetails;
