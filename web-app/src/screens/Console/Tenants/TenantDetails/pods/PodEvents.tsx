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
import { Grid } from "mds";
import { useSelector } from "react-redux";
import { IEvent } from "../../ListTenants/types";
import { niceDays } from "../../../../../common/utils";
import { ErrorResponseHandler } from "../../../../../common/types";
import { AppState, useAppDispatch } from "../../../../../store";
import { setErrorSnackMessage } from "../../../../../systemSlice";
import api from "../../../../../common/api";
import EventsList from "../events/EventsList";

interface IPodEventsProps {
  tenant: string;
  namespace: string;
  podName: string;
  propLoading: boolean;
}

const PodEvents = ({
  tenant,
  namespace,
  podName,
  propLoading,
}: IPodEventsProps) => {
  const dispatch = useAppDispatch();
  const loadingTenant = useSelector(
    (state: AppState) => state.tenants.loadingTenant,
  );
  const [events, setEvents] = useState<IEvent[]>([]);
  const [loading, setLoading] = useState<boolean>(true);

  useEffect(() => {
    if (propLoading) {
      setLoading(true);
    }
  }, [propLoading]);

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
          `/api/v1/namespaces/${namespace}/tenants/${tenant}/pods/${podName}/events`,
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
  }, [loading, podName, namespace, tenant, dispatch]);

  return (
    <React.Fragment>
      <Grid item xs={12}>
        <EventsList events={events} loading={loading} />
      </Grid>
    </React.Fragment>
  );
};

export default PodEvents;
