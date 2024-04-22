// This file is part of MinIO Operator
// Copyright (c) 2022 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public APIKey as published by
// the Free Software Foundation, either version 3 of the APIKey, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public APIKey for more details.
//
// You should have received a copy of the GNU Affero General Public APIKey
// along with this program.  If not, see <http://www.gnu.org/APIKeys/>.

import React, { Fragment, useCallback, useEffect, useState } from "react";
import { Box, PageLayout, Tabs } from "mds";
import { ErrorResponseHandler } from "../../../common/types";
import { ClusterRegistered } from "./utils";
import api from "../../../common/api";
import ApiKeyRegister from "./ApiKeyRegister";
import PageHeaderWrapper from "../Common/PageHeaderWrapper/PageHeaderWrapper";

const RegisterOperator = () => {
  const [apiKeyRegistered, setAPIKeyRegistered] = useState<boolean>(false);
  const [curTab, setCurTab] = useState<string>("simple-tab-0");

  const fetchAPIKeyInfo = useCallback(() => {
    api
      .invoke("GET", `/api/v1/subnet/apikey/info`)
      .then((res: any) => {
        setAPIKeyRegistered(true);
      })
      .catch((err: ErrorResponseHandler) => {
        setAPIKeyRegistered(false);
      });
  }, []);

  useEffect(() => {
    fetchAPIKeyInfo();
  }, [fetchAPIKeyInfo]);

  const apiKeyRegistration = (
    <Fragment>
      <Box
        withBorders
        sx={{
          display: "flex",
          flexFlow: "column",
          padding: "43px",
        }}
      >
        {apiKeyRegistered ? (
          <ClusterRegistered email={"Operator"} />
        ) : (
          <ApiKeyRegister registerEndpoint={"/api/v1/subnet/apikey/register"} />
        )}
      </Box>
    </Fragment>
  );

  return (
    <Fragment>
      <PageHeaderWrapper label="Register to MinIO Subscription Network" />

      <PageLayout>
        <Tabs
          currentTabOrPath={curTab}
          onTabClick={(nvTab) => {
            setCurTab(nvTab);
          }}
          options={[
            {
              tabConfig: { id: "simple-tab-0", label: "API Key" },
              content: apiKeyRegistration,
            },
          ]}
          horizontal
        />
      </PageLayout>
    </Fragment>
  );
};

export default RegisterOperator;
