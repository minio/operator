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
import { useSelector } from "react-redux";
import { AppState } from "../../../store";
import { Box } from "@mui/material";
import { CircleIcon } from "mds";
import { getLicenseConsent } from "../License/utils";
import { registeredCluster } from "../../../config";

const LicenseBadge = () => {
  const licenseInfo = useSelector(
    (state: AppState) => state?.system?.licenseInfo,
  );

  const isAgplAckDone = getLicenseConsent();
  const clusterRegistered = registeredCluster();

  const { plan = "" } = licenseInfo || {};

  if (plan || isAgplAckDone || clusterRegistered) {
    return null;
  }

  return (
    <Box
      sx={{
        position: "absolute",
        top: 1,
        transform: "translateX(5px)",
        zIndex: 400,
        border: 0,
      }}
      style={{
        border: 0,
      }}
    >
      <CircleIcon
        style={{
          fill: "#FF3958",
          border: "1px solid #FF3958",
          borderRadius: "100%",
          width: 8,
          height: 8,
        }}
      />
    </Box>
  );
};

export default LicenseBadge;
