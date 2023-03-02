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
import Grid from "@mui/material/Grid";

type Props = {
  separator?: boolean;
  actions?: React.ReactNode;
  icon?: React.ReactNode;
  children: string;
};

const SectionTitle: React.FC<Props> = ({
  children,
  separator = true,
  actions,
  icon,
}) => {
  return (
    <Grid
      container
      alignItems={"center"}
      justifyContent="space-between"
      sx={{
        borderBottom: separator ? "1px solid #eaeaea" : "",
      }}
    >
      <Grid item xs>
        <h3
          style={{
            margin: 0,
            marginBottom: 10,
          }}
        >
          {icon && (
            <Grid
              container
              alignItems={"center"}
              justifyContent={"flex-start"}
              spacing={1}
            >
              <Grid item>{icon}</Grid>
              <Grid item>{children}</Grid>
            </Grid>
          )}
          {!icon && children}
        </h3>
      </Grid>
      {actions && <Grid item> {actions}</Grid>}
    </Grid>
  );
};

export default SectionTitle;
