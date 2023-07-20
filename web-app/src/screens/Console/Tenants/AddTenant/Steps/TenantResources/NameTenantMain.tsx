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

import React, { Fragment, useCallback, useEffect } from "react";
import { useSelector } from "react-redux";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import get from "lodash/get";
import Grid from "@mui/material/Grid";
import {
  formFieldStyles,
  modalBasic,
  wizardCommon,
} from "../../../../Common/FormComponents/common/styleLibrary";
import { AppState, useAppDispatch } from "../../../../../../store";
import InputBoxWrapper from "../../../../Common/FormComponents/InputBoxWrapper/InputBoxWrapper";
import SelectWrapper from "../../../../Common/FormComponents/SelectWrapper/SelectWrapper";
import SizePreview from "../SizePreview";
import TenantSize from "./TenantSize";
import { Paper, SelectChangeEvent } from "@mui/material";
import { IMkEnvs, mkPanelConfigurations } from "./utils";
import {
  isPageValid,
  setStorageType,
  setTenantName,
  updateAddField,
} from "../../createTenantSlice";
import { selFeatures } from "../../../../consoleSlice";
import NamespaceSelector from "./NamespaceSelector";
import H3Section from "../../../../Common/H3Section";

const styles = (theme: Theme) =>
  createStyles({
    sizePreview: {
      marginLeft: 10,
      background: "#FFFFFF",
      border: "1px solid #EAEAEA",
      padding: 2,
      marginTop: 20,
    },
    ...formFieldStyles,
    ...modalBasic,
    ...wizardCommon,
  });

const NameTenantField = () => {
  const dispatch = useAppDispatch();
  const tenantName = useSelector(
    (state: AppState) => state.createTenant.fields.nameTenant.tenantName,
  );

  const tenantNameError = useSelector(
    (state: AppState) => state.createTenant.validationErrors["tenant-name"],
  );

  return (
    <InputBoxWrapper
      id="tenant-name"
      name="tenant-name"
      onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
        dispatch(setTenantName(e.target.value));
      }}
      label="Name"
      value={tenantName}
      required
      error={tenantNameError || ""}
    />
  );
};

interface INameTenantMainScreen {
  classes: any;
  formToRender?: IMkEnvs;
}

const NameTenantMain = ({ classes, formToRender }: INameTenantMainScreen) => {
  const dispatch = useAppDispatch();

  const selectedStorageClass = useSelector(
    (state: AppState) =>
      state.createTenant.fields.nameTenant.selectedStorageClass,
  );
  const selectedStorageType = useSelector(
    (state: AppState) =>
      state.createTenant.fields.nameTenant.selectedStorageType,
  );
  const storageClasses = useSelector(
    (state: AppState) => state.createTenant.storageClasses,
  );
  const features = useSelector(selFeatures);

  // Common
  const updateField = useCallback(
    (field: string, value: string) => {
      dispatch(
        updateAddField({ pageName: "nameTenant", field: field, value: value }),
      );
    },
    [dispatch],
  );

  // Validation
  useEffect(() => {
    const isValid =
      (formToRender === IMkEnvs.default && storageClasses.length > 0) ||
      (formToRender !== IMkEnvs.default && selectedStorageType !== "");

    dispatch(isPageValid({ pageName: "nameTenant", valid: isValid }));
  }, [storageClasses, dispatch, selectedStorageType, formToRender]);

  return (
    <Fragment>
      <Grid container>
        <Grid item sx={{ width: "calc(100% - 320px)" }}>
          <Paper className={classes.paperWrapper} sx={{ minHeight: 550 }}>
            <Grid container>
              <Grid item xs={12}>
                <div className={classes.headerElement}>
                  <H3Section>Name</H3Section>
                  <span className={classes.descriptionText}>
                    How would you like to name this new tenant?
                  </span>
                </div>
                <div className={classes.formFieldRow}>
                  <NameTenantField />
                </div>
              </Grid>
              <Grid item xs={12} className={classes.formFieldRow}>
                <NamespaceSelector formToRender={formToRender} />
              </Grid>
              {formToRender === IMkEnvs.default ? (
                <Grid item xs={12} className={classes.formFieldRow}>
                  <SelectWrapper
                    id="storage_class"
                    name="storage_class"
                    onChange={(e: SelectChangeEvent<string>) => {
                      updateField(
                        "selectedStorageClass",
                        e.target.value as string,
                      );
                    }}
                    label="Storage Class"
                    value={selectedStorageClass}
                    options={storageClasses}
                    disabled={storageClasses.length < 1}
                  />
                </Grid>
              ) : (
                <Grid item xs={12} className={classes.formFieldRow}>
                  <SelectWrapper
                    id="storage_type"
                    name="storage_type"
                    onChange={(e: SelectChangeEvent<string>) => {
                      dispatch(
                        setStorageType({
                          storageType: e.target.value as string,
                          features: features,
                        }),
                      );
                    }}
                    label={get(
                      mkPanelConfigurations,
                      `${formToRender}.variantSelectorLabel`,
                      "Storage Type",
                    )}
                    value={selectedStorageType}
                    options={get(
                      mkPanelConfigurations,
                      `${formToRender}.variantSelectorValues`,
                      [],
                    )}
                  />
                </Grid>
              )}
              {formToRender === IMkEnvs.default ? (
                <TenantSize />
              ) : (
                get(
                  mkPanelConfigurations,
                  `${formToRender}.sizingComponent`,
                  null,
                )
              )}
            </Grid>
          </Paper>
        </Grid>
        <Grid item>
          <div className={classes.sizePreview}>
            <SizePreview />
          </div>
        </Grid>
      </Grid>
    </Fragment>
  );
};

export default withStyles(styles)(NameTenantMain);
