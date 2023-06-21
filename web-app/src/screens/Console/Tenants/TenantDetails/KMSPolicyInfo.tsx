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
