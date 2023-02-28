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

import React, { Fragment, useEffect, useState } from "react";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import { spacingUtils } from "../Common/FormComponents/common/styleLibrary";
import withStyles from "@mui/styles/withStyles";
import { Box } from "@mui/material";
import PageLayout from "../Common/Layout/PageLayout";
import api from "../../../common/api";

import { SubnetRegTokenResponse } from "../License/types";
import { ErrorResponseHandler } from "../../../common/types";
import { useSelector } from "react-redux";
import { setErrorSnackMessage } from "../../../systemSlice";
import { AppState, useAppDispatch } from "../../../store";
import Tabs from "@mui/material/Tabs";
import Tab from "@mui/material/Tab";
import { TabPanel } from "../../shared/tabs";
import { ClusterRegistered, ProxyConfiguration } from "./utils";
import ApiKeyRegister from "./ApiKeyRegister";
import { fetchLicenseInfo } from "./registerThunks";
import {
  resetRegisterForm,
  setCurTab,
  setLoading,
  setSubnetRegToken,
} from "./registerSlice";
import OfflineRegistration from "./OfflineRegistration";
import SubnetMFAToken from "./SubnetMFAToken";
import ClusterRegistrationForm from "./ClusterRegistrationForm";
import OnlineRegistration from "./OnlineRegistration";
import PageHeaderWrapper from "../Common/PageHeaderWrapper/PageHeaderWrapper";

interface IRegister {
  classes: any;
}

const styles = (theme: Theme) =>
  createStyles({
    sizedLabel: {
      minWidth: "75px",
    },
    ...spacingUtils,
  });

const Register = ({ classes }: IRegister) => {
  const dispatch = useAppDispatch();

  const subnetMFAToken = useSelector(
    (state: AppState) => state.register.subnetMFAToken
  );
  const subnetAccessToken = useSelector(
    (state: AppState) => state.register.subnetAccessToken
  );

  const subnetRegToken = useSelector(
    (state: AppState) => state.register.subnetRegToken
  );
  const subnetOrganizations = useSelector(
    (state: AppState) => state.register.subnetOrganizations
  );

  const loading = useSelector((state: AppState) => state.register.loading);
  const loadingLicenseInfo = useSelector(
    (state: AppState) => state.register.loadingLicenseInfo
  );
  const clusterRegistered = useSelector(
    (state: AppState) => state.register.clusterRegistered
  );
  const licenseInfo = useSelector(
    (state: AppState) => state.register.licenseInfo
  );
  const curTab = useSelector((state: AppState) => state.register.curTab);

  const [initialLicenseLoading, setInitialLicenseLoading] =
    useState<boolean>(true);

  useEffect(() => {
    // when unmounted, reset
    return () => {
      dispatch(resetRegisterForm());
    };
  }, [dispatch]);

  const fetchSubnetRegToken = () => {
    if (loading || subnetRegToken) {
      return;
    }
    dispatch(setLoading(true));
    api
      .invoke("GET", "/api/v1/subnet/registration-token")
      .then((resp: SubnetRegTokenResponse) => {
        dispatch(setLoading(false));
        if (resp && resp.regToken) {
          dispatch(setSubnetRegToken(resp.regToken));
        }
      })
      .catch((err: ErrorResponseHandler) => {
        console.error(err);
        dispatch(setErrorSnackMessage(err));
        dispatch(setLoading(false));
      });
  };

  useEffect(() => {
    if (initialLicenseLoading) {
      dispatch(fetchLicenseInfo());
      setInitialLicenseLoading(false);
    }
  }, [initialLicenseLoading, setInitialLicenseLoading, dispatch]);

  let clusterRegistrationForm: JSX.Element = <Fragment />;

  if (subnetAccessToken && subnetOrganizations.length > 0) {
    clusterRegistrationForm = <ClusterRegistrationForm />;
  } else if (subnetMFAToken) {
    clusterRegistrationForm = <SubnetMFAToken />;
  } else {
    clusterRegistrationForm = <OnlineRegistration />;
  }

  const apiKeyRegistration = (
    <Fragment>
      <Box
        sx={{
          border: "1px solid #eaeaea",
          borderRadius: "2px",
          display: "flex",
          flexFlow: "column",
          padding: "43px",
        }}
      >
        {clusterRegistered && licenseInfo ? (
          <ClusterRegistered email={licenseInfo.email} />
        ) : (
          <ApiKeyRegister registerEndpoint={"/api/v1/subnet/login"} />
        )}
      </Box>
      <ProxyConfiguration />
    </Fragment>
  );

  const offlineRegistration = <OfflineRegistration />;

  const regUi = (
    <Fragment>
      <Box
        sx={{
          border: "1px solid #eaeaea",
          borderRadius: "2px",
          display: "flex",
          flexFlow: "column",
          padding: "43px",
        }}
      >
        {clusterRegistered && licenseInfo ? (
          <ClusterRegistered email={licenseInfo.email} />
        ) : (
          clusterRegistrationForm
        )}
      </Box>

      {!clusterRegistered && <ProxyConfiguration />}
    </Fragment>
  );

  const loadingUi = <div>Loading..</div>;
  const uiToShow = loadingLicenseInfo ? loadingUi : regUi;

  return (
    <Fragment>
      <PageHeaderWrapper
        label="Register to MinIO Subscription Network"
        actions={<React.Fragment />}
      />

      <PageLayout>
        <Tabs
          value={curTab}
          onChange={(e: React.ChangeEvent<{}>, newValue: number) => {
            dispatch(setCurTab(newValue));
          }}
          indicatorColor="primary"
          textColor="primary"
          aria-label="cluster-tabs"
          variant="scrollable"
          scrollButtons="auto"
        >
          <Tab
            label="Credentials"
            id="simple-tab-0"
            aria-controls="simple-tab-panel-0"
          />
          <Tab
            label="API Key"
            id="simple-tab-1"
            aria-controls="simple-tab-panel-1"
          />
          <Tab
            label="Air-Gap"
            id="simple-tab-2"
            aria-controls="simple-tab-panel-2"
            onClick={() => fetchSubnetRegToken()}
          />
        </Tabs>

        <TabPanel index={0} value={curTab}>
          {uiToShow}
        </TabPanel>
        <TabPanel index={1} value={curTab}>
          {apiKeyRegistration}
        </TabPanel>
        <TabPanel index={2} value={curTab}>
          {offlineRegistration}
        </TabPanel>
      </PageLayout>
    </Fragment>
  );
};

export default withStyles(styles)(Register);
