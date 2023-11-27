//  This file is part of MinIO Operator
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
import { Tabs, SectionTitle } from "mds";
import { Link, useParams } from "react-router-dom";
import { setErrorSnackMessage } from "../../../../../systemSlice";
import { useAppDispatch } from "../../../../../store";
import { IEvent } from "../../ListTenants/types";
import { niceDays } from "../../../../../common/utils";
import { ErrorResponseHandler } from "../../../../../common/types";
import EventsList from "../events/EventsList";
import PVCDescribe from "./PVCDescribe";
import api from "../../../../../common/api";

const TenantVolumes = () => {
  const dispatch = useAppDispatch();
  const { tenantName, PVCName, tenantNamespace } = useParams();

  const [curTab, setCurTab] = useState<string>("events-tab");
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
      <SectionTitle separator sx={{ marginBottom: 15 }}>
        <Link
          to={`/namespaces/${tenantNamespace}/tenants/${tenantName}/volumes`}
        >
          PVCs
        </Link>{" "}
        &gt; {PVCName}
      </SectionTitle>
      <Tabs
        options={[
          {
            tabConfig: { id: "events-tab", label: "Events" },
            content: (
              <Fragment>
                <SectionTitle separator sx={{ marginBottom: 15 }}>
                  Events
                </SectionTitle>
                <EventsList events={events} loading={loading} />
              </Fragment>
            ),
          },
          {
            tabConfig: { id: "describe-tab", label: "Describe" },
            content: (
              <PVCDescribe
                tenant={tenantName || ""}
                namespace={tenantNamespace || ""}
                pvcName={PVCName || ""}
                propLoading={loading}
              />
            ),
          },
        ]}
        currentTabOrPath={curTab}
        onTabClick={(tab) => {
          setCurTab(tab);
        }}
        horizontal
      />
    </Fragment>
  );
};

export default TenantVolumes;
