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

import React, { Fragment, useEffect } from "react";
import { useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import Grid from "@mui/material/Grid";
import PageLayout from "../../../../Common/Layout/PageLayout";
import GenericWizard from "../../../../Common/GenericWizard/GenericWizard";
import ScreenTitle from "../../../../Common/ScreenTitle/ScreenTitle";
import { BackLink, TenantsIcon } from "mds";
import EditPoolResources from "./EditPoolResources";
import EditPoolConfiguration from "./EditPoolConfiguration";
import EditPoolPlacement from "./EditPoolPlacement";
import { IWizardElement } from "../../../../Common/GenericWizard/types";
import { LinearProgress } from "@mui/material";
import { niceBytes } from "../../../../../../common/utils";
import {
  formFieldStyles,
  modalStyleUtils,
} from "../../../../Common/FormComponents/common/styleLibrary";

import { AppState, useAppDispatch } from "../../../../../../store";
import { resetEditPoolForm, setInitialPoolDetails } from "./editPoolSlice";
import EditPoolButton from "./EditPoolButton";
import makeStyles from "@mui/styles/makeStyles";
import PageHeaderWrapper from "../../../../Common/PageHeaderWrapper/PageHeaderWrapper";

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    bottomContainer: {
      display: "flex",
      flexGrow: 1,
      alignItems: "center",
      margin: "auto",
      justifyContent: "center",
      "& div": {
        width: 150,
        "@media (max-width: 900px)": {
          flexFlow: "column",
        },
      },
    },
    pageBox: {
      border: "1px solid #EAEAEA",
      borderTop: 0,
    },
    editPoolTitle: {
      border: "1px solid #EAEAEA",
      borderBottom: 0,
    },
    ...formFieldStyles,
    ...modalStyleUtils,
  }),
);

const EditPool = () => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();

  const classes = useStyles();

  const tenant = useSelector((state: AppState) => state.tenants.tenantInfo);
  const selectedPool = useSelector(
    (state: AppState) => state.tenants.selectedPool,
  );

  const editSending = useSelector(
    (state: AppState) => state.editPool.editSending,
  );
  const navigateTo = useSelector(
    (state: AppState) => state.editPool.navigateTo,
  );

  const poolsURL = `/namespaces/${tenant?.namespace || ""}/tenants/${
    tenant?.name || ""
  }/pools`;

  useEffect(() => {
    if (selectedPool) {
      const poolDetails = tenant?.pools?.find(
        (pool) => pool.name === selectedPool,
      );

      if (poolDetails) {
        dispatch(setInitialPoolDetails(poolDetails));
      } else {
        navigate("/tenants");
      }
    }
  }, [selectedPool, dispatch, tenant, navigate]);

  useEffect(() => {
    if (navigateTo !== "") {
      const goTo = `${navigateTo}`;
      dispatch(resetEditPoolForm());
      navigate(goTo);
    }
  }, [navigateTo, navigate, dispatch]);

  const cancelButton = {
    label: "Cancel",
    type: "other",
    enabled: true,
    action: () => {
      dispatch(resetEditPoolForm());
      navigate(poolsURL);
    },
  };

  const createButton = {
    componentRender: <EditPoolButton />,
  };

  const wizardSteps: IWizardElement[] = [
    {
      label: "Pool Resources",
      componentRender: <EditPoolResources />,
      buttons: [cancelButton, createButton],
    },
    {
      label: "Configuration",
      advancedOnly: true,
      componentRender: <EditPoolConfiguration />,
      buttons: [cancelButton, createButton],
    },
    {
      label: "Pod Placement",
      advancedOnly: true,
      componentRender: <EditPoolPlacement />,
      buttons: [cancelButton, createButton],
    },
  ];

  return (
    <Fragment>
      <Grid item xs={12}>
        <PageHeaderWrapper
          label={
            <Fragment>
              <BackLink
                label={`Pool Details`}
                onClick={() => navigate(poolsURL)}
              />
            </Fragment>
          }
        />
        <PageLayout>
          <Grid item xs={12} className={classes.editPoolTitle}>
            <ScreenTitle
              icon={<TenantsIcon />}
              title={`Edit Pool - ${selectedPool}`}
              subTitle={
                <Fragment>
                  Namespace: {tenant?.namespace || ""} / Current Capacity:{" "}
                  {niceBytes((tenant?.total_size || 0).toString(10))} / Tenant:{" "}
                  {tenant?.name || ""}
                </Fragment>
              }
            />
          </Grid>

          {editSending && (
            <Grid item xs={12}>
              <LinearProgress />
            </Grid>
          )}
          <Grid item xs={12} className={classes.pageBox}>
            <GenericWizard wizardSteps={wizardSteps} />
          </Grid>
        </PageLayout>
      </Grid>
    </Fragment>
  );
};

export default EditPool;
