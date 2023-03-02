// This file is part of MinIO Operator
// Copyright (c) 2021 MinIO, Inc.
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
import { Grid } from "@mui/material";
import { HelpBox } from "mds";

interface IMissingIntegration {
  iconComponent: any;
  entity: string;
  documentationLink: string;
}

const MissingIntegration = ({
  iconComponent,
  entity,
  documentationLink,
}: IMissingIntegration) => {
  return (
    <Grid
      container
      justifyContent={"center"}
      alignContent={"center"}
      alignItems={"center"}
    >
      <Grid item xs={8}>
        <HelpBox
          title={`${entity} not available`}
          iconComponent={iconComponent}
          help={
            <Fragment>
              This feature is not available.
              <br />
              Please configure{" "}
              <a href={documentationLink} target="_blank" rel="noopener">
                {entity}
              </a>{" "}
              first to use this feature.
            </Fragment>
          }
        />
      </Grid>
    </Grid>
  );
};

export default MissingIntegration;
