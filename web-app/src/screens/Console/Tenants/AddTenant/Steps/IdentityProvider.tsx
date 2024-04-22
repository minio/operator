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

import React from "react";
import {
  Box,
  FormLayout,
  Grid,
  LDAPIcon,
  OIDCIcon,
  RadioGroup,
  UsersIcon,
} from "mds";
import { useSelector } from "react-redux";
import { AppState, useAppDispatch } from "../../../../../store";
import { setIDP } from "../createTenantSlice";
import IDPActiveDirectory from "./IdentityProvider/IDPActiveDirectory";
import IDPOpenID from "./IdentityProvider/IDPOpenID";
import IDPBuiltIn from "./IdentityProvider/IDPBuiltIn";
import H3Section from "../../../Common/H3Section";

const IdentityProvider = () => {
  const dispatch = useAppDispatch();

  const idpSelection = useSelector(
    (state: AppState) =>
      state.createTenant.fields.identityProvider.idpSelection,
  );

  return (
    <FormLayout withBorders={false} containerPadding={false}>
      <Box className={"inputItem"}>
        <H3Section>Identity Provider</H3Section>
        <span className={"muted"}>
          Access to the tenant can be controlled via an external Identity
          Manager.
        </span>
      </Box>
      <Grid item xs={12} sx={{ padding: 10 }}>
        <RadioGroup
          currentValue={idpSelection}
          id="idp-options"
          name="idp-options"
          label="Protocol"
          onChange={(e) => {
            dispatch(setIDP(e.target.value));
          }}
          selectorOptions={[
            { label: "Built-in", value: "Built-in", icon: <UsersIcon /> },
            { label: "Open ID", value: "OpenID", icon: <OIDCIcon /> },
            {
              label: "LDAP / Active Directory",
              value: "AD",
              icon: <LDAPIcon />,
            },
          ]}
        />
      </Grid>
      {idpSelection === "Built-in" && <IDPBuiltIn />}
      {idpSelection === "OpenID" && <IDPOpenID />}
      {idpSelection === "AD" && <IDPActiveDirectory />}
    </FormLayout>
  );
};

export default IdentityProvider;
