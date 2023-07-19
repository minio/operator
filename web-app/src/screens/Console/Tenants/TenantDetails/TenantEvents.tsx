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

import React, { useEffect, useState } from "react";
import { connect, useSelector } from "react-redux";
import { Theme } from "@mui/material/styles";
import { useParams } from "react-router-dom";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import {
  actionsTray,
  containerForHeader,
  searchField,
  tableStyles,
} from "../../Common/FormComponents/common/styleLibrary";
import Grid from "@mui/material/Grid";
import { IEvent } from "../ListTenants/types";
import { niceDays } from "../../../../common/utils";
import { ErrorResponseHandler } from "../../../../common/types";
import api from "../../../../common/api";
import { AppState, useAppDispatch } from "../../../../store";
import EventsList from "./events/EventsList";
import { setErrorSnackMessage } from "../../../../systemSlice";

interface ITenantEventsProps {
  classes: any;
}

const styles = (theme: Theme) =>
  createStyles({
    ...actionsTray,
    ...searchField,
    ...tableStyles,
    ...containerForHeader,
  });

const TenantEvents = ({ classes }: ITenantEventsProps) => {
  const dispatch = useAppDispatch();
  const params = useParams();

  const loadingTenant = useSelector(
    (state: AppState) => state.tenants.loadingTenant,
  );

  const [events, setEvents] = useState<IEvent[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const tenantName = params.tenantName || "";
  const tenantNamespace = params.tenantNamespace || "";

  useEffect(() => {
    if (loadingTenant) {
      setLoading(true);
    }
  }, [loadingTenant]);

  useEffect(() => {
    if (loading) {
      api
        .invoke(
          "GET",
          `/api/v1/namespaces/${tenantNamespace}/tenants/${tenantName}/events`,
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
  }, [loading, tenantNamespace, tenantName, dispatch]);

  return (
    <React.Fragment>
      <h1 className={classes.sectionTitle}>Events</h1>
      <Grid item xs={12}>
        <EventsList events={events} loading={loading} />
      </Grid>
    </React.Fragment>
  );
};
const mapState = (state: AppState) => ({
  loadingTenant: state.tenants.loadingTenant,
});
const connector = connect(mapState, null);

export default withStyles(styles)(connector(TenantEvents));
