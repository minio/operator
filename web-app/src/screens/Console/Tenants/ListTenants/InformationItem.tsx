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

import React, { Fragment } from "react";
import { Box } from "mds";
import styled from "styled-components";
import get from "lodash/get";

const InformationItemMain = styled.div(({ theme }) => ({
  margin: "0px 20px",
  "& .value": {
    fontSize: 18,
    color: get(theme, "mutedText", "#87888d"),
    fontWeight: 400,
    "&.normal": {
      color: get(theme, "fontColor", "#000"),
    },
  },
  "& .unit": {
    fontSize: 12,
    color: get(theme, "secondaryText", "#5B5C5C"),
    fontWeight: "bold",
  },
  "& .label": {
    textAlign: "center",
    color: get(theme, "mutedText", "#87888d"),
    fontSize: 12,
    whiteSpace: "nowrap",
    "&.normal": {
      color: get(theme, "secondaryText", "#5B5C5C"),
    },
  },
}));

interface IInformationItemProps {
  label: string;
  value: string;
  unit?: string;
  variant?: "normal" | "faded";
}

const InformationItem = ({
  label,
  value,
  unit,
  variant = "normal",
}: IInformationItemProps) => {
  return (
    <InformationItemMain>
      <Box style={{ textAlign: "center" }}>
        <span className={`value ${variant}`}>{value}</span>
        {unit && (
          <Fragment>
            {" "}
            <span className={"unit"}>{unit}</span>
          </Fragment>
        )}
      </Box>
      <Box className={"label"}>{label}</Box>
    </InformationItemMain>
  );
};

export default InformationItem;
