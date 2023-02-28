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

import { Grid } from "@mui/material";
import { LDAPIcon, OIDCIcon, UsersIcon } from "mds";

export const OIDCLogoElement = () => {
  return (
    <Grid container columnGap={1}>
      <Grid>
        <OIDCIcon width={"16px"} height={"16px"} />
      </Grid>
      <Grid item>Open ID</Grid>
    </Grid>
  );
};

export const LDAPLogoElement = () => {
  return (
    <Grid container columnGap={1}>
      <Grid>
        <LDAPIcon width={"16px"} height={"16px"} />
      </Grid>
      <Grid item>LDAP / Active Directory</Grid>
    </Grid>
  );
};

export const BuiltInLogoElement = () => {
  return (
    <Grid container columnGap={1}>
      <Grid>
        <UsersIcon width={"16px"} height={"16px"} />
      </Grid>
      <Grid item>Built-in</Grid>
    </Grid>
  );
};
