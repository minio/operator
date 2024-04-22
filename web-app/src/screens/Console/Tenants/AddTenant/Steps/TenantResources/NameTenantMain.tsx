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
import { Box, FormLayout, Grid, InputBox, Select } from "mds";
import get from "lodash/get";
import { AppState, useAppDispatch } from "../../../../../../store";
import { IMkEnvs, mkPanelConfigurations } from "./utils";
import {
  isPageValid,
  setStorageType,
  setTenantName,
  updateAddField,
} from "../../createTenantSlice";
import { selFeatures } from "../../../../consoleSlice";
import SizePreview from "../SizePreview";
import TenantSize from "./TenantSize";
import NamespaceSelector from "./NamespaceSelector";
import H3Section from "../../../../Common/H3Section";

const NameTenantField = () => {
  const dispatch = useAppDispatch();
  const tenantName = useSelector(
    (state: AppState) => state.createTenant.fields.nameTenant.tenantName,
  );

  const tenantNameError = useSelector(
    (state: AppState) => state.createTenant.validationErrors["tenant-name"],
  );

  return (
    <InputBox
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
  formToRender?: IMkEnvs;
}

const NameTenantMain = ({ formToRender }: INameTenantMainScreen) => {
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
      <Grid container sx={{ justifyContent: "space-between" }}>
        <Grid item sx={{ width: "calc(100% - 320px)" }}>
          <Box sx={{ minHeight: 550 }}>
            <FormLayout withBorders={false} containerPadding={false}>
              <Box className={"inputItem"}>
                <H3Section>Name</H3Section>
                <span className={"muted"}>
                  How would you like to name this new tenant?
                </span>
              </Box>
              <NameTenantField />
              <NamespaceSelector formToRender={formToRender} />
              {formToRender === IMkEnvs.default ? (
                <Select
                  id="storage_class"
                  name="storage_class"
                  onChange={(value) => {
                    updateField("selectedStorageClass", value);
                  }}
                  label="Storage Class"
                  value={selectedStorageClass}
                  options={storageClasses}
                  disabled={storageClasses.length < 1}
                />
              ) : (
                <Select
                  id="storage_type"
                  name="storage_type"
                  onChange={(value) => {
                    dispatch(
                      setStorageType({
                        storageType: value,
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
            </FormLayout>
          </Box>
        </Grid>
        <Grid item xs={"hidden"} sm={"hidden"}>
          <Box
            sx={{
              marginLeft: 10,
              padding: 2,
              marginTop: 20,
            }}
            withBorders
            useBackground
          >
            <SizePreview />
          </Box>
        </Grid>
      </Grid>
    </Fragment>
  );
};

export default NameTenantMain;
