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
import get from "lodash/get";
import { useSelector } from "react-redux";
import Grid from "@mui/material/Grid";
import { LinearProgress } from "@mui/material";

import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import {
  modalBasic,
  settingsCommon,
  wizardCommon,
} from "../../Common/FormComponents/common/styleLibrary";
import GenericWizard from "../../Common/GenericWizard/GenericWizard";
import {
  IWizardButton,
  IWizardElement,
} from "../../Common/GenericWizard/types";
import { AppState, useAppDispatch } from "../../../../store";
import Configure from "./Steps/Configure";
import IdentityProvider from "./Steps/IdentityProvider";
import Security from "./Steps/Security";
import Encryption from "./Steps/Encryption";
import Affinity from "./Steps/Affinity";
import Images from "./Steps/Images";
import PageLayout from "../../Common/Layout/PageLayout";

import TenantResources from "./Steps/TenantResources/TenantResources";
import {
  IMkEnvs,
  resourcesConfigurations,
} from "./Steps/TenantResources/utils";
import { BackLink, HelpBox, StorageIcon } from "mds";
import { selFeatures } from "../../consoleSlice";
import makeStyles from "@mui/styles/makeStyles";
import { resetAddTenantForm } from "./createTenantSlice";
import CreateTenantButton from "./CreateTenantButton";
import NewTenantCredentials from "./NewTenantCredentials";
import { useNavigate } from "react-router-dom";
import PageHeaderWrapper from "../../Common/PageHeaderWrapper/PageHeaderWrapper";

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    pageBox: {
      border: "1px solid #EAEAEA",
    },
    ...modalBasic,
    ...wizardCommon,
    ...settingsCommon,
  }),
);

const AddTenant = () => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const classes = useStyles();

  const features = useSelector(selFeatures);

  // Fields
  const addSending = useSelector(
    (state: AppState) => state.createTenant.addingTenant,
  );
  const [formRender, setFormRender] = useState<IMkEnvs | null>(null);

  useEffect(() => {
    let setConfiguration = IMkEnvs.default;

    if (features && features.length !== 0) {
      const possibleVariables = Object.keys(resourcesConfigurations);

      possibleVariables.forEach((element) => {
        if (features.includes(element)) {
          setConfiguration = get(
            resourcesConfigurations,
            element,
            IMkEnvs.default,
          );
        }
      });
    }

    setFormRender(setConfiguration);
  }, [features]);

  const cancelButton = {
    label: "Cancel",
    type: "other",
    enabled: true,
    action: () => {
      dispatch(resetAddTenantForm());
      navigate("/tenants");
    },
  };

  const createButton: IWizardButton = {
    componentRender: <CreateTenantButton key={"create-tenant"} />,
  };

  const wizardSteps: IWizardElement[] = [
    {
      label: "Setup",
      componentRender: <TenantResources />,
      buttons: [cancelButton, createButton],
    },
    {
      label: "Configure",
      advancedOnly: true,
      componentRender: <Configure />,
      buttons: [cancelButton, createButton],
    },
    {
      label: "Images",
      advancedOnly: true,
      componentRender: <Images />,
      buttons: [cancelButton, createButton],
    },
    {
      label: "Pod Placement",
      advancedOnly: true,
      componentRender: <Affinity />,
      buttons: [cancelButton, createButton],
    },
    {
      label: "Identity Provider",
      advancedOnly: true,
      componentRender: <IdentityProvider />,
      buttons: [cancelButton, createButton],
    },
    {
      label: "Security",
      advancedOnly: true,
      componentRender: <Security />,
      buttons: [cancelButton, createButton],
    },
    {
      label: "Encryption",
      advancedOnly: true,
      componentRender: <Encryption />,
      buttons: [cancelButton, createButton],
    },
  ];

  let filteredWizardSteps = wizardSteps;

  return (
    <Fragment>
      <NewTenantCredentials />
      <PageHeaderWrapper
        label={
          <BackLink
            onClick={() => {
              dispatch(resetAddTenantForm());
              navigate("/tenants");
            }}
            label={"Tenants"}
          />
        }
      />

      <PageLayout>
        {addSending && (
          <Grid item xs={12}>
            <LinearProgress />
          </Grid>
        )}
        <Grid item xs={12} className={classes.pageBox}>
          <GenericWizard wizardSteps={filteredWizardSteps} />
        </Grid>
        {formRender === IMkEnvs.aws && (
          <Grid item xs={12} style={{ marginTop: 16 }}>
            <HelpBox
              title={"EBS Volume Configuration."}
              iconComponent={<StorageIcon />}
              help={
                <Fragment>
                  <b>Performance Optimized</b>: Uses the <i>gp3</i> EBS storage
                  class class configured at 1,000Mi/s throughput and 16,000
                  IOPS, however the minimum volume size for this type of EBS
                  volume is <b>32Gi</b>.
                  <br />
                  <br />
                  <b>Storage Optimized</b>: Uses the <i>sc1</i> EBS storage
                  class, however the minimum volume size for this type of EBS
                  volume is &nbsp;
                  <b>16Ti</b> to unlock their maximum throughput speed of
                  250Mi/s.
                </Fragment>
              }
            />
          </Grid>
        )}
      </PageLayout>
    </Fragment>
  );
};

export default AddTenant;
