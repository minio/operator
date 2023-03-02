//  This file is part of MinIO Console Server
//  Copyright (c) 2021 MinIO, Inc.
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

import React from "react";
import { Grid } from "@mui/material";
import { Loader } from "mds";

const LoadingComponent = () => {
  return (
    <Grid
      container
      spacing={0}
      direction="column"
      alignItems="center"
      justifyContent="center"
      style={{ minHeight: "100vh" }}
    >
      <Grid item xs={3} style={{ textAlign: "center" }}>
        <Loader style={{ width: 35, height: 35 }} />
      </Grid>
    </Grid>
  );
};

export default LoadingComponent;
