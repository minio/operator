// This file is part of MinIO Operator
// Copyright (c) 2023 MinIO, Inc.
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

import React, { Fragment } from "react";
import Grid from "@mui/material/Grid";
import { Box } from "mds";

const getPolicyData = (policies: Record<string, any> = {}) => {
  const policyNames = Object.keys(policies);
  return policyNames.map((polName: string) => {
    const policyConfig = policies[polName] || {};
    return {
      name: polName || "",
      identities: policyConfig.identities || [],
      // v1 specific
      paths: policyConfig.paths || [],
      // v2 specific
      allow: policyConfig.allow || [],
      deny: policyConfig.deny || [],
    };
  });
};

const PolicyItem = ({
  items = [],
  title = "",
}: {
  items: string[];
  title: string;
}) => {
  return items?.length ? (
    <Fragment>
      <div
        style={{
          fontSize: "0.83em",
          fontWeight: "bold",
        }}
      >
        {title}
      </div>
      <div
        style={{
          display: "flex",
          gap: "2px",
          flexFlow: "column",
          marginLeft: "8px",
        }}
      >
        {items.map((iTxt: string) => {
          return <span style={{ fontSize: "12px" }}>- {iTxt}</span>;
        })}
      </div>
    </Fragment>
  ) : null;
};

const KMSPolicyInfo = ({
  policies = {},
}: {
  policies: Record<string, any>;
}) => {
  const fmtPolicies = getPolicyData(policies);
  return fmtPolicies.length ? (
    <Grid xs={12} marginBottom={"5px"}>
      <h4>Policies</h4>
      <Box
        withBorders
        sx={{
          maxHeight: "200px",
          overflow: "auto",
          padding: 0,
        }}
      >
        {fmtPolicies.map((pConf: Record<string, any>) => {
          return (
            <Box
              withBorders
              sx={{
                display: "flex",
                flexFlow: "column",
                gap: "2px",
                borderLeft: 0,
                borderRight: 0,
                borderTop: 0,
              }}
            >
              <div>
                <b
                  style={{
                    fontSize: "0.83em",
                    fontWeight: "bold",
                  }}
                >
                  Policy Name:
                </b>{" "}
                {pConf.name}
              </div>
              <PolicyItem title={"Allow"} items={pConf?.allow} />
              <PolicyItem title={"Deny"} items={pConf?.deny} />
              <PolicyItem title={"Paths"} items={pConf?.paths} />
              <PolicyItem title={"Identities"} items={pConf?.identities} />
            </Box>
          );
        })}
      </Box>
    </Grid>
  ) : null;
};

export default KMSPolicyInfo;
