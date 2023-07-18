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

import { Grid, IconButton, Tooltip, Typography } from "@mui/material";
import InputBoxWrapper from "../../../../Common/FormComponents/InputBoxWrapper/InputBoxWrapper";
import React, { Fragment, useCallback, useEffect, useState } from "react";
import FormSwitchWrapper from "../../../../Common/FormComponents/FormSwitchWrapper/FormSwitchWrapper";
import makeStyles from "@mui/styles/makeStyles";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import {
  createTenantCommon,
  formFieldStyles,
  modalBasic,
  wizardCommon,
} from "../../../../Common/FormComponents/common/styleLibrary";
import {
  addIDPADGroupAtIndex,
  addIDPADUsrAtIndex,
  isPageValid,
  removeIDPADGroupAtIndex,
  removeIDPADUsrAtIndex,
  setIDPADGroupAtIndex,
  setIDPADUsrAtIndex,
  updateAddField,
} from "../../createTenantSlice";
import { useSelector } from "react-redux";
import { clearValidationError } from "../../../utils";
import { AppState, useAppDispatch } from "../../../../../../store";
import AddIcon from "@mui/icons-material/Add";
import DeleteIcon from "@mui/icons-material/Delete";
import {
  commonFormValidation,
  IValidation,
} from "../../../../../../utils/validationFunctions";

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    adUserDnRows: {
      display: "flex",
      marginBottom: 10,
    },
    buttonTray: {
      marginLeft: 10,
      display: "flex",
      height: 38,
      "& button": {
        background: "#EAEAEA",
      },
    },
    overlayAction: {
      marginLeft: 10,
      "& svg": {
        maxWidth: 15,
        maxHeight: 15,
      },
      "& button": {
        background: "#EAEAEA",
      },
    },
    ...createTenantCommon,
    ...formFieldStyles,
    ...modalBasic,
    ...wizardCommon,
  }),
);

const IDPActiveDirectory = () => {
  const dispatch = useAppDispatch();
  const classes = useStyles();

  const idpSelection = useSelector(
    (state: AppState) =>
      state.createTenant.fields.identityProvider.idpSelection,
  );
  const ADURL = useSelector(
    (state: AppState) => state.createTenant.fields.identityProvider.ADURL,
  );
  const ADSkipTLS = useSelector(
    (state: AppState) => state.createTenant.fields.identityProvider.ADSkipTLS,
  );
  const ADServerInsecure = useSelector(
    (state: AppState) =>
      state.createTenant.fields.identityProvider.ADServerInsecure,
  );
  const ADGroupSearchBaseDN = useSelector(
    (state: AppState) =>
      state.createTenant.fields.identityProvider.ADGroupSearchBaseDN,
  );
  const ADGroupSearchFilter = useSelector(
    (state: AppState) =>
      state.createTenant.fields.identityProvider.ADGroupSearchFilter,
  );
  const ADUserDNs = useSelector(
    (state: AppState) => state.createTenant.fields.identityProvider.ADUserDNs,
  );
  const ADGroupDNs = useSelector(
    (state: AppState) => state.createTenant.fields.identityProvider.ADGroupDNs,
  );
  const ADLookupBindDN = useSelector(
    (state: AppState) =>
      state.createTenant.fields.identityProvider.ADLookupBindDN,
  );
  const ADLookupBindPassword = useSelector(
    (state: AppState) =>
      state.createTenant.fields.identityProvider.ADLookupBindPassword,
  );
  const ADUserDNSearchBaseDN = useSelector(
    (state: AppState) =>
      state.createTenant.fields.identityProvider.ADUserDNSearchBaseDN,
  );
  const ADUserDNSearchFilter = useSelector(
    (state: AppState) =>
      state.createTenant.fields.identityProvider.ADUserDNSearchFilter,
  );
  const ADServerStartTLS = useSelector(
    (state: AppState) =>
      state.createTenant.fields.identityProvider.ADServerStartTLS,
  );

  const [validationErrors, setValidationErrors] = useState<any>({});

  const updateField = useCallback(
    (field: string, value: any) => {
      dispatch(
        updateAddField({
          pageName: "identityProvider",
          field: field,
          value: value,
        }),
      );
    },
    [dispatch],
  );

  const cleanValidation = (fieldName: string) => {
    setValidationErrors(clearValidationError(validationErrors, fieldName));
  };

  // Validation
  useEffect(() => {
    let customIDPValidation: IValidation[] = [];

    if (idpSelection === "AD") {
      customIDPValidation = [
        ...customIDPValidation,
        {
          fieldKey: "AD_URL",
          required: true,
          value: ADURL,
        },
        {
          fieldKey: "ad_lookupBindDN",
          required: true,
          value: ADLookupBindDN,
        },
      ];
    }

    const commonVal = commonFormValidation(customIDPValidation);

    dispatch(
      isPageValid({
        pageName: "identityProvider",
        valid: Object.keys(commonVal).length === 0,
      }),
    );

    setValidationErrors(commonVal);
  }, [
    ADLookupBindDN,
    idpSelection,
    ADURL,
    ADGroupSearchBaseDN,
    ADGroupSearchFilter,
    ADUserDNs,
    ADGroupDNs,
    dispatch,
  ]);

  return (
    <Fragment>
      <Grid item xs={12} className={classes.formFieldRow}>
        <InputBoxWrapper
          id="AD_URL"
          name="AD_URL"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("ADURL", e.target.value);
            cleanValidation("AD_URL");
          }}
          label="LDAP Server Address"
          value={ADURL}
          placeholder="ldap-server:636"
          error={validationErrors["AD_URL"] || ""}
          required
        />
      </Grid>
      <Grid item xs={12} className={classes.formFieldRow}>
        <FormSwitchWrapper
          value="ad_skipTLS"
          id="ad_skipTLS"
          name="ad_skipTLS"
          checked={ADSkipTLS}
          onChange={(e) => {
            const targetD = e.target;
            const checked = targetD.checked;
            updateField("ADSkipTLS", checked);
          }}
          label={"Skip TLS Verification"}
        />
      </Grid>
      <Grid item xs={12} className={classes.formFieldRow}>
        <FormSwitchWrapper
          value="ad_serverInsecure"
          id="ad_serverInsecure"
          name="ad_serverInsecure"
          checked={ADServerInsecure}
          onChange={(e) => {
            const targetD = e.target;
            const checked = targetD.checked;
            updateField("ADServerInsecure", checked);
          }}
          label={"Server Insecure"}
        />
      </Grid>
      {ADServerInsecure ? (
        <Grid item xs={12}>
          <Typography
            className={classes.error}
            variant="caption"
            display="block"
            gutterBottom
          >
            Warning: All traffic with Active Directory will be unencrypted
          </Typography>
          <br />
        </Grid>
      ) : null}
      <Grid item xs={12} className={classes.formFieldRow}>
        <FormSwitchWrapper
          value="ad_serverStartTLS"
          id="ad_serverStartTLS"
          name="ad_serverStartTLS"
          checked={ADServerStartTLS}
          onChange={(e) => {
            const targetD = e.target;
            const checked = targetD.checked;
            updateField("ADServerStartTLS", checked);
          }}
          label={"Start TLS connection to AD/LDAP server"}
        />
      </Grid>
      <Grid item xs={12} className={classes.formFieldRow}>
        <InputBoxWrapper
          id="ad_lookupBindDN"
          name="ad_lookupBindDN"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("ADLookupBindDN", e.target.value);
            cleanValidation("ad_lookupBindDN");
          }}
          label="Lookup Bind DN"
          value={ADLookupBindDN}
          placeholder="cn=admin,dc=min,dc=io"
          error={validationErrors["ad_lookupBindDN"] || ""}
          required
        />
      </Grid>
      <Grid item xs={12} className={classes.formFieldRow}>
        <InputBoxWrapper
          id="ad_lookupBindPassword"
          name="ad_lookupBindPassword"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("ADLookupBindPassword", e.target.value);
          }}
          label="Lookup Bind Password"
          value={ADLookupBindPassword}
          placeholder="admin"
        />
      </Grid>
      <Grid item xs={12} className={classes.formFieldRow}>
        <InputBoxWrapper
          id="ad_userDNSearchBaseDN"
          name="ad_userDNSearchBaseDN"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("ADUserDNSearchBaseDN", e.target.value);
          }}
          label="User DN Search Base DN"
          value={ADUserDNSearchBaseDN}
          placeholder="dc=min,dc=io"
        />
      </Grid>
      <Grid item xs={12} className={classes.formFieldRow}>
        <InputBoxWrapper
          id="ad_userDNSearchFilter"
          name="ad_userDNSearchFilter"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("ADUserDNSearchFilter", e.target.value);
          }}
          label="User DN Search Filter"
          value={ADUserDNSearchFilter}
          placeholder="(sAMAcountName=%s)"
        />
      </Grid>
      <Grid item xs={12} className={classes.formFieldRow}>
        <InputBoxWrapper
          id="ad_groupSearchBaseDN"
          name="ad_groupSearchBaseDN"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("ADGroupSearchBaseDN", e.target.value);
          }}
          label="Group Search Base DN"
          value={ADGroupSearchBaseDN}
          placeholder="ou=hwengg,dc=min,dc=io;ou=swengg,dc=min,dc=io"
        />
      </Grid>
      <Grid item xs={12} className={classes.formFieldRow}>
        <InputBoxWrapper
          id="ad_groupSearchFilter"
          name="ad_groupSearchFilter"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("ADGroupSearchFilter", e.target.value);
          }}
          label="Group Search Filter"
          value={ADGroupSearchFilter}
          placeholder="(&(objectclass=groupOfNames)(member=%s))"
        />
      </Grid>
      <fieldset className={classes.fieldGroup}>
        <legend className={classes.descriptionText}>
          List of user DNs (Distinguished Names) to be Tenant Administrators
        </legend>
        <Grid item xs={12}>
          {ADUserDNs.map((_, index) => {
            return (
              <Fragment key={`identityField-${index.toString()}`}>
                <div className={classes.adUserDnRows}>
                  <InputBoxWrapper
                    id={`ad-userdn-${index.toString()}`}
                    label={""}
                    placeholder=""
                    name={`ad-userdn-${index.toString()}`}
                    value={ADUserDNs[index]}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      dispatch(
                        setIDPADUsrAtIndex({
                          index: index,
                          userDN: e.target.value,
                        }),
                      );
                      cleanValidation(`ad-userdn-${index.toString()}`);
                    }}
                    index={index}
                    key={`csv-ad-userdn-${index.toString()}`}
                    error={
                      validationErrors[`ad-userdn-${index.toString()}`] || ""
                    }
                  />
                  <div className={classes.buttonTray}>
                    <Tooltip title="Add User" aria-label="add">
                      <IconButton
                        size={"small"}
                        onClick={() => {
                          dispatch(addIDPADUsrAtIndex());
                        }}
                      >
                        <AddIcon />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="Remove" aria-label="add">
                      <IconButton
                        size={"small"}
                        style={{ marginLeft: 16 }}
                        onClick={() => {
                          if (ADUserDNs.length > 1) {
                            dispatch(removeIDPADUsrAtIndex(index));
                          }
                        }}
                      >
                        <DeleteIcon />
                      </IconButton>
                    </Tooltip>
                  </div>
                </div>
              </Fragment>
            );
          })}
        </Grid>
      </fieldset>
      <fieldset className={classes.fieldGroup}>
        <legend className={classes.descriptionText}>
          List of group DNs (Distinguished Names) to be Tenant Administrators
        </legend>
        <Grid item xs={12}>
          {ADGroupDNs.map((_, index) => {
            return (
              <Fragment key={`identityField-${index.toString()}`}>
                <div className={classes.adUserDnRows}>
                  <InputBoxWrapper
                    id={`ad-groupdn-${index.toString()}`}
                    label={""}
                    placeholder=""
                    name={`ad-groupdn-${index.toString()}`}
                    value={ADGroupDNs[index]}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      dispatch(
                        setIDPADGroupAtIndex({
                          index: index,
                          userDN: e.target.value,
                        }),
                      );
                      cleanValidation(`ad-groupdn-${index.toString()}`);
                    }}
                    index={index}
                    key={`csv-ad-groupdn-${index.toString()}`}
                    error={
                      validationErrors[`ad-groupdn-${index.toString()}`] || ""
                    }
                  />
                  <div className={classes.buttonTray}>
                    <Tooltip title="Add Group" aria-label="add">
                      <IconButton
                        size={"small"}
                        onClick={() => {
                          dispatch(addIDPADGroupAtIndex());
                        }}
                      >
                        <AddIcon />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="Remove" aria-label="add">
                      <IconButton
                        size={"small"}
                        style={{ marginLeft: 16 }}
                        onClick={() => {
                          if (ADGroupDNs.length > 1) {
                            dispatch(removeIDPADGroupAtIndex(index));
                          }
                        }}
                      >
                        <DeleteIcon />
                      </IconButton>
                    </Tooltip>
                  </div>
                </div>
              </Fragment>
            );
          })}
        </Grid>
      </fieldset>
    </Fragment>
  );
};

export default IDPActiveDirectory;
