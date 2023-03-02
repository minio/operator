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
import { Grid, ThemedLogo } from "mds";
import { useSelector } from "react-redux";
import { AppState } from "../../../../store";

interface IAutoColorIcon {
  marginRight: number;
  marginTop: number;
}

const AutoColorIcon = ({ marginRight, marginTop }: IAutoColorIcon) => {
  let tinycolor = require("tinycolor2");

  const colorVariants = useSelector(
    (state: AppState) => state.system.overrideStyles
  );

  const isDark =
    tinycolor(colorVariants?.backgroundColor || "#fff").getBrightness() <= 128;

  return (
    <Grid
      sx={{
        "& svg": {
          width: 105,
          marginRight,
          marginTop,
          fill: isDark ? "#fff" : "#081C42",
        },
      }}
    >
      <ThemedLogo />
    </Grid>
  );
};

export default AutoColorIcon;
