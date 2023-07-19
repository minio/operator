//  This file is part of MinIO Console Server
//  Copyright (c) 2022 MinIO, Inc.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

import React, { Fragment, useEffect, useState } from "react";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { containerForHeader } from "../../../Common/FormComponents/common/styleLibrary";
import Grid from "@mui/material/Grid";
import Tabs from "@mui/material/Tabs";
import Tab from "@mui/material/Tab";
import { Link, useParams } from "react-router-dom";

import api from "../../../../../common/api";
import { IEvent } from "../../ListTenants/types";
import { niceDays } from "../../../../../common/utils";
import { ErrorResponseHandler } from "../../../../../common/types";
import EventsList from "../events/EventsList";
import PVCDescribe from "./PVCDescribe";

import { setErrorSnackMessage } from "../../../../../systemSlice";
import { useAppDispatch } from "../../../../../store";

interface IPVCDetailsProps {
  classes: any;
}

const styles = (theme: Theme) =>
  createStyles({
    breadcrumLink: {
      textDecoration: "none",
      color: "black",
    },
    ...containerForHeader,
  });

const TenantVolumes = ({ classes }: IPVCDetailsProps) => {
  const dispatch = useAppDispatch();
  const { tenantName, PVCName, tenantNamespace } = useParams();

  const [curTab, setCurTab] = useState<number>(0);
  const [loading, setLoading] = useState<boolean>(true);
  const [events, setEvents] = useState<IEvent[]>([]);

  useEffect(() => {
    if (loading) {
      api
        .invoke(
          "GET",
          `/api/v1/namespaces/${tenantNamespace}/tenants/${tenantName}/pvcs/${PVCName}/events`,
        )
        .then((res: IEvent[]) => {
          for (let i = 0; i < res.length; i++) {
            let currentTime = (Date.now() / 1000) | 0;

            res[i].seen = niceDays((currentTime - res[i].last_seen).toString());
          }
          setEvents(res);
          setLoading(false);
        })
        .catch((err: ErrorResponseHandler) => {
          dispatch(setErrorSnackMessage(err));
          setLoading(false);
        });
    }
  }, [loading, PVCName, tenantNamespace, tenantName, dispatch]);

  return (
    <Fragment>
      <Grid item xs={12}>
        <h1 className={classes.sectionTitle}>
          <Link
            to={`/namespaces/${tenantNamespace}/tenants/${tenantName}/volumes`}
            className={classes.breadcrumLink}
          >
            PVCs
          </Link>{" "}
          &gt; {PVCName}
        </h1>
      </Grid>
      <Grid container>
        <Grid item xs={12}>
          <Tabs
            value={curTab}
            onChange={(e: React.ChangeEvent<{}>, newValue: number) => {
              setCurTab(newValue);
            }}
            indicatorColor="primary"
            textColor="primary"
            aria-label="cluster-tabs"
            variant="scrollable"
            scrollButtons="auto"
          >
            <Tab label="Events" id="simple-tab-0" />
            <Tab label="Describe" id="simple-tab-1" />
          </Tabs>
        </Grid>
        {curTab === 0 && (
          <React.Fragment>
            <h1 className={classes.sectionTitle}>Events</h1>
            <EventsList events={events} loading={loading} />
          </React.Fragment>
        )}
        {curTab === 1 && (
          <PVCDescribe
            tenant={tenantName || ""}
            namespace={tenantNamespace || ""}
            pvcName={PVCName || ""}
            propLoading={loading}
          />
        )}
      </Grid>
    </Fragment>
  );
};

export default withStyles(styles)(TenantVolumes);
