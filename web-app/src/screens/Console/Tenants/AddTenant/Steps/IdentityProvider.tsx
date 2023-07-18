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
import { useSelector } from "react-redux";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import { Grid, Paper } from "@mui/material";
import {
  createTenantCommon,
  modalBasic,
  wizardCommon,
} from "../../../Common/FormComponents/common/styleLibrary";
import { AppState, useAppDispatch } from "../../../../../store";
import RadioGroupSelector from "../../../Common/FormComponents/RadioGroupSelector/RadioGroupSelector";
import { setIDP } from "../createTenantSlice";
import IDPActiveDirectory from "./IdentityProvider/IDPActiveDirectory";
import IDPOpenID from "./IdentityProvider/IDPOpenID";
import makeStyles from "@mui/styles/makeStyles";
import IDPBuiltIn from "./IdentityProvider/IDPBuiltIn";
import {
  BuiltInLogoElement,
  LDAPLogoElement,
  OIDCLogoElement,
} from "../../LogoComponents";
import H3Section from "../../../Common/H3Section";

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    protocolRadioOptions: {
      display: "flex",
      flexFlow: "column",
      marginBottom: 10,

      "& label": {
        fontSize: 16,
        fontWeight: 600,
      },
      "& div": {
        display: "flex",
        flexFlow: "row",
        alignItems: "top",
      },
    },
    ...createTenantCommon,
    ...modalBasic,
    ...wizardCommon,
  }),
);

const IdentityProvider = () => {
  const dispatch = useAppDispatch();
  const classes = useStyles();

  const idpSelection = useSelector(
    (state: AppState) =>
      state.createTenant.fields.identityProvider.idpSelection,
  );

  return (
    <Paper className={classes.paperWrapper}>
      <div className={classes.headerElement}>
        <H3Section>Identity Provider</H3Section>
        <span className={classes.descriptionText}>
          Access to the tenant can be controlled via an external Identity
          Manager.
        </span>
      </div>
      <Grid item xs={12} padding="10px">
        <RadioGroupSelector
          currentSelection={idpSelection}
          id="idp-options"
          name="idp-options"
          label="Protocol"
          onChange={(e) => {
            dispatch(setIDP(e.target.value));
          }}
          selectorOptions={[
            { label: <BuiltInLogoElement />, value: "Built-in" },
            { label: <OIDCLogoElement />, value: "OpenID" },
            { label: <LDAPLogoElement />, value: "AD" },
          ]}
        />
      </Grid>
      {idpSelection === "Built-in" && <IDPBuiltIn />}
      {idpSelection === "OpenID" && <IDPOpenID />}
      {idpSelection === "AD" && <IDPActiveDirectory />}
    </Paper>
  );
};

export default IdentityProvider;
