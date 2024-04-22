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
import {
  BackLink,
  TenantsIcon,
  PageLayout,
  ScreenTitle,
  Wizard,
  WizardElement,
  Box,
  ProgressBar,
  Grid,
} from "mds";
import { useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";
import { AppState, useAppDispatch } from "../../../../../../store";
import { niceBytes } from "../../../../../../common/utils";
import { resetPoolForm } from "./addPoolSlice";
import PoolResources from "./PoolResources";
import PoolConfiguration from "./PoolConfiguration";
import PoolPodPlacement from "./PoolPodPlacement";
import AddPoolCreateButton from "./AddPoolCreateButton";
import PageHeaderWrapper from "../../../../Common/PageHeaderWrapper/PageHeaderWrapper";

const AddPool = () => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();

  const tenant = useSelector((state: AppState) => state.tenants.tenantInfo);
  const sending = useSelector((state: AppState) => state.addPool.sending);
  const navigateTo = useSelector((state: AppState) => state.addPool.navigateTo);

  const poolsURL = `/namespaces/${tenant?.namespace || ""}/tenants/${
    tenant?.name || ""
  }/pools`;

  useEffect(() => {
    if (navigateTo !== "") {
      const goTo = `${navigateTo}`;
      dispatch(resetPoolForm());
      navigate(goTo);
    }
  }, [navigateTo, navigate, dispatch]);

  const cancelButton = {
    label: "Cancel",
    type: "custom" as "to" | "custom" | "next" | "back",
    enabled: true,
    action: () => {
      dispatch(resetPoolForm());
      navigate(poolsURL);
    },
  };

  const createButton = {
    componentRender: <AddPoolCreateButton key={"add-pool-crate"} />,
  };

  const wizardSteps: WizardElement[] = [
    {
      label: "Setup",
      componentRender: <PoolResources />,
      buttons: [cancelButton, createButton],
    },
    {
      label: "Configuration",
      advancedOnly: true,
      componentRender: <PoolConfiguration />,
      buttons: [cancelButton, createButton],
    },
    {
      label: "Pod Placement",
      advancedOnly: true,
      componentRender: <PoolPodPlacement />,
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
                label={`Tenant Pools`}
                onClick={() => navigate(poolsURL)}
              />
            </Fragment>
          }
        />
        <PageLayout variant={"constrained"}>
          <Box withBorders sx={{ padding: 0, borderBottom: 0 }}>
            <ScreenTitle
              icon={<TenantsIcon />}
              title={`Add New Pool to ${tenant?.name || ""}`}
              subTitle={
                <Fragment>
                  Namespace: {tenant?.namespace || ""} / Current Capacity:{" "}
                  {niceBytes((tenant?.total_size || 0).toString(10))}
                </Fragment>
              }
              actions={null}
            />
          </Box>
          {sending && (
            <Grid item xs={12}>
              <ProgressBar />
            </Grid>
          )}
          <Box
            withBorders
            sx={{ padding: 0, borderTop: 0, "& .muted": { fontSize: 13 } }}
          >
            <Wizard wizardSteps={wizardSteps} linearMode={false} />
          </Box>
        </PageLayout>
      </Grid>
    </Fragment>
  );
};

export default AddPool;
