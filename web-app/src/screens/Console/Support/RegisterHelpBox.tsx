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

import React from "react";
import {
  Box,
  breakPoints,
  CallHomeFeatureIcon,
  DiagnosticsFeatureIcon,
  ExtraFeaturesIcon,
  HelpIconFilled,
  PerformanceFeatureIcon,
} from "mds";
import styled from "styled-components";
import get from "lodash/get";

const LinkElement = styled.a(({ theme }) => ({
  color: get(theme, "linkColor", "#2781B0"),
  fontWeight: 600,
}));

const FeatureItem = ({
  icon,
  description,
}: {
  icon: any;
  description: string | React.ReactNode;
}) => {
  return (
    <Box
      sx={{
        display: "flex",
        "& .min-icon": {
          marginRight: "10px",
          height: "23px",
          width: "23px",
          marginBottom: "10px",
        },
      }}
    >
      {icon}{" "}
      <Box className="muted" sx={{ fontSize: "14px", fontStyle: "italic" }}>
        {description}
      </Box>
    </Box>
  );
};
const RegisterHelpBox = ({ hasMargin = true }: { hasMargin?: boolean }) => {
  return (
    <Box
      withBorders
      sx={{
        flex: 1,
        display: "flex",
        flexFlow: "column",
        padding: "20px",
        marginLeft: hasMargin ? "30px" : "",
        [`@media (max-width: ${breakPoints.sm}px)`]: {
          marginLeft: 0,
          marginTop: 0,
        },
        [`@media (max-width: ${breakPoints.md}px)`]: {
          marginLeft: 0,
          marginTop: hasMargin ? "30px" : "",
        },
      }}
    >
      <Box
        sx={{
          fontSize: "16px",
          fontWeight: 600,
          display: "flex",
          alignItems: "center",
          marginBottom: "16px",

          "& .min-icon": {
            height: "21px",
            width: "21px",
            marginRight: "15px",
          },
        }}
      >
        <HelpIconFilled />
        <div>Why should I register?</div>
      </Box>
      <Box sx={{ fontSize: "14px", marginBottom: "15px" }}>
        Registering this cluster with the MinIO Subscription Network (SUBNET)
        provides the following benefits in addition to the commercial license
        and SLA backed support.
      </Box>

      <Box
        sx={{
          display: "flex",
          flexFlow: "column",
        }}
      >
        <FeatureItem
          icon={<CallHomeFeatureIcon />}
          description={`Call Home Monitoring`}
        />
        <FeatureItem
          icon={<DiagnosticsFeatureIcon />}
          description={`Health Diagnostics`}
        />
        <FeatureItem
          icon={<PerformanceFeatureIcon />}
          description={`Performance Analysis`}
        />
        <FeatureItem
          icon={<ExtraFeaturesIcon />}
          description={
            <LinkElement href="https://min.io/signup?ref=op" target="_blank">
              More Features
            </LinkElement>
          }
        />
      </Box>
    </Box>
  );
};

export default RegisterHelpBox;
