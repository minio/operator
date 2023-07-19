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
  Button,
  CircleIcon,
  EditIcon,
  MinIOTierIconXs,
  RefreshIcon,
  TenantsIcon,
  TrashIcon,
} from "mds";
import {
  Link,
  Navigate,
  Route,
  Routes,
  useLocation,
  useNavigate,
  useParams,
} from "react-router-dom";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import Grid from "@mui/material/Grid";
import {
  containerForHeader,
  pageContentStyles,
  tenantDetailsStyles,
} from "../../Common/FormComponents/common/styleLibrary";
import { AppState, useAppDispatch } from "../../../../store";
import { niceBytes } from "../../../../common/utils";
import ScreenTitle from "../../Common/ScreenTitle/ScreenTitle";
import PageLayout from "../../Common/Layout/PageLayout";

import VerticalTabs from "../../Common/VerticalTabs/VerticalTabs";
import withSuspense from "../../Common/Components/withSuspense";
import { IAM_PAGES } from "../../../../common/SecureComponent/permissions";
import { setSnackBarMessage } from "../../../../systemSlice";
import { setTenantName } from "../tenantsSlice";
import { getTenantAsync } from "../thunks/tenantDetailsAsync";
import { LinearProgress } from "@mui/material";
import TooltipWrapper from "../../Common/TooltipWrapper/TooltipWrapper";
import PageHeaderWrapper from "../../Common/PageHeaderWrapper/PageHeaderWrapper";
import { tenantIsOnline } from "./utils";

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

interface ITenantDetailsProps {
  classes: any;
}

const styles = (theme: Theme) =>
  createStyles({
    ...tenantDetailsStyles,
    pageContainer: {
      border: "1px solid #EAEAEA",
      width: "100%",
      height: "100%",
    },
    contentSpacer: {
      ...pageContentStyles.contentSpacer,
      minHeight: 400,
    },
    redState: {
      color: theme.palette.error.main,
      "& .min-icon": {
        width: 16,
        height: 16,
      },
    },
    yellowState: {
      color: theme.palette.warning.main,
      "& .min-icon": {
        width: 16,
        height: 16,
      },
    },
    greenState: {
      color: theme.palette.success.main,
      "& .min-icon": {
        width: 16,
        height: 16,
      },
    },
    greyState: {
      color: "grey",
      "& .min-icon": {
        width: 16,
        height: 16,
      },
    },
    healthStatusIcon: {
      position: "relative",
      fontSize: 10,
      left: 26,
      height: 10,
      top: 4,
    },
    ...containerForHeader,
  });

const TenantDetails = ({ classes }: ITenantDetailsProps) => {
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

  const splitSections = pathname.split("/");

  let highlightedTab = splitSections[splitSections.length - 1] || "summary";
  if (highlightedTab === ":podName" || highlightedTab === "pods") {
    // It has SUB Route
    highlightedTab = "pods";
  }
  const [activeTab, setActiveTab] = useState(highlightedTab);

  useEffect(() => {
    setActiveTab(highlightedTab);
  }, [highlightedTab]);

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

  const healthStatusToClass = (health_status: string) => {
    return health_status === "red"
      ? classes.redState
      : health_status === "yellow"
      ? classes.yellowState
      : health_status === "green"
      ? classes.greenState
      : classes.greyState;
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
        actions={<React.Fragment />}
      />

      <PageLayout className={classes.pageContainer}>
        {loadingTenant && (
          <Grid item xs={12}>
            <LinearProgress />
          </Grid>
        )}
        <Grid item xs={12}>
          <ScreenTitle
            icon={
              <Fragment>
                <div className={classes.healthStatusIcon}>
                  {tenantInfo && tenantInfo.status && (
                    <span
                      className={healthStatusToClass(
                        tenantInfo.status?.health_status!,
                      )}
                    >
                      <CircleIcon />
                    </span>
                  )}
                </div>
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
              <div style={{ display: "flex", justifyContent: "flex-end" }}>
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
              </div>
            }
          />
        </Grid>

        <VerticalTabs
          selectedTab={activeTab}
          isRouteTabs
          routes={
            <div className={classes.contentSpacer}>
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
            </div>
          }
        >
          {{
            tabConfig: {
              label: "Summary",
              value: "summary",
              component: Link,
              to: getRoutePath("summary"),
            },
          }}
          {{
            tabConfig: {
              label: "Configuration",
              value: "configuration",
              component: Link,
              to: getRoutePath("configuration"),
            },
          }}
          {{
            tabConfig: {
              label: "Metrics",
              value: "metrics",
              component: Link,
              to: getRoutePath("metrics"),
            },
          }}
          {{
            tabConfig: {
              label: "Identity Provider",
              value: "identity-provider",
              component: Link,
              to: getRoutePath("identity-provider"),
            },
          }}
          {{
            tabConfig: {
              label: "Security",
              value: "security",
              component: Link,
              to: getRoutePath("security"),
            },
          }}
          {{
            tabConfig: {
              label: "Encryption",
              value: "encryption",
              component: Link,
              to: getRoutePath("encryption"),
            },
          }}
          {{
            tabConfig: {
              label: "Pools",
              value: "pools",
              component: Link,
              to: getRoutePath("pools"),
            },
          }}
          {{
            tabConfig: {
              label: "Pods",
              value: "pods",
              component: Link,
              id: "tenant-pod-tab",
              to: getRoutePath("pods"),
            },
          }}

          {{
            tabConfig: {
              label: "Volumes",
              value: "volumes",
              component: Link,
              to: getRoutePath("volumes"),
            },
          }}
          {{
            tabConfig: {
              label: "Events",
              value: "events",
              component: Link,
              to: getRoutePath("events"),
            },
          }}
          {{
            tabConfig: {
              label: "Certificate Requests",
              value: "csr",
              component: Link,
              to: getRoutePath("csr"),
            },
          }}
          {{
            tabConfig: {
              label: "License",
              value: "license",
              component: Link,
              to: getRoutePath("license"),
            },
          }}
        </VerticalTabs>
      </PageLayout>
    </Fragment>
  );
};

export default withStyles(styles)(TenantDetails);
