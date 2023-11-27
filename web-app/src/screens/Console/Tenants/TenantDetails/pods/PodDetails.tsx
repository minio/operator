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
import { SectionTitle, Tabs } from "mds";
import { Link, useParams } from "react-router-dom";

import PodLogs from "./PodLogs";
import PodEvents from "./PodEvents";
import PodDescribe from "./PodDescribe";

const PodDetails = () => {
  const { tenantNamespace, tenantName, podName } = useParams();

  const [curTab, setCurTab] = useState<string>("events-tab");
  const [loading, setLoading] = useState<boolean>(true);

  useEffect(() => {
    if (loading) {
      setLoading(false);
    }
  }, [loading]);

  return (
    <Fragment>
      <SectionTitle separator sx={{ marginBottom: 15 }}>
        <Link
          to={`/namespaces/${tenantNamespace || ""}/tenants/${
            tenantName || ""
          }/pods`}
        >
          Pods
        </Link>{" "}
        &gt; {podName}
      </SectionTitle>
      <Tabs
        options={[
          {
            tabConfig: { id: "events-tab", label: "Events" },
            content: (
              <PodEvents
                tenant={tenantName || ""}
                namespace={tenantNamespace || ""}
                podName={podName || ""}
                propLoading={loading}
              />
            ),
          },
          {
            tabConfig: { id: "describe-tab", label: "Describe" },
            content: (
              <PodDescribe
                tenant={tenantName || ""}
                namespace={tenantNamespace || ""}
                podName={podName || ""}
                propLoading={loading}
              />
            ),
          },
          {
            tabConfig: { id: "logs-tab", label: "Logs" },
            content: (
              <PodLogs
                tenant={tenantName || ""}
                namespace={tenantNamespace || ""}
                podName={podName || ""}
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

export default PodDetails;
