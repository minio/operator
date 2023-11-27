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

import React, { Fragment } from "react";
import { Box, Grid } from "mds";
import styled from "styled-components";
import get from "lodash/get";
import RegistrationStatusBanner from "./RegistrationStatusBanner";

const LinkElement = styled.a(({ theme }) => ({
  color: get(theme, "linkColor", "#2781B0"),
  fontWeight: 600,
}));

export const FormTitle = ({
  icon = null,
  title,
}: {
  icon?: any;
  title: any;
}) => {
  return (
    <Box
      sx={{
        display: "flex",
        alignItems: "center",
        justifyContent: "flex-start",
      }}
    >
      {icon}
      <div className="title-text">{title}</div>
    </Box>
  );
};

export const ClusterRegistered = ({ email }: { email: string }) => {
  return (
    <Fragment>
      <RegistrationStatusBanner email={email} />
      <Grid item xs={12} sx={{ marginTop: 25 }}>
        <Box
          sx={{
            padding: "20px",
            "& a": {
              color: "#2781B0",
              cursor: "pointer",
            },
          }}
        >
          Login to{" "}
          <LinkElement href="https://subnet.min.io" target="_blank">
            SUBNET
          </LinkElement>{" "}
          to avail support for this MinIO cluster
        </Box>
      </Grid>
    </Fragment>
  );
};
