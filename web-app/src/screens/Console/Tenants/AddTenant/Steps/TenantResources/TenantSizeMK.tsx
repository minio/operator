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

import React, { Fragment, useCallback, useEffect, useState } from "react";
import { Box, Grid, InputBox, Select } from "mds";
import { useSelector } from "react-redux";
import get from "lodash/get";
import { AppState, useAppDispatch } from "../../../../../../store";
import { erasureCodeCalc, getBytes } from "../../../../../../common/utils";
import { clearValidationError } from "../../../utils";
import { ecListTransform } from "../../../ListTenants/utils";
import { IStorageDistribution } from "../../../../../../common/types";
import { commonFormValidation } from "../../../../../../utils/validationFunctions";
import {
  IMkEnvs,
  IntegrationConfiguration,
  mkPanelConfigurations,
} from "./utils";
import { isPageValid, updateAddField } from "../../createTenantSlice";
import api from "../../../../../../common/api";
import H3Section from "../../../../Common/H3Section";

interface ITenantSizeAWSProps {
  formToRender?: IMkEnvs;
}

const TenantSizeMK = ({ formToRender }: ITenantSizeAWSProps) => {
  const dispatch = useAppDispatch();

  const volumeSize = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.volumeSize,
  );
  const sizeFactor = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.sizeFactor,
  );
  const drivesPerServer = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.drivesPerServer,
  );
  const nodes = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.nodes,
  );
  const memoryNode = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.memoryNode,
  );
  const ecParity = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.ecParity,
  );
  const ecParityChoices = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.ecParityChoices,
  );
  const cleanECChoices = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.cleanECChoices,
  );
  const resourcesSize = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.resourcesSize,
  );
  const distribution = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.distribution,
  );
  const ecParityCalc = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.ecParityCalc,
  );
  const cpuToUse = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.cpuToUse,
  );
  const maxCPUsUse = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.maxCPUsUse,
  );
  const integrationSelection = useSelector(
    (state: AppState) =>
      state.createTenant.fields.tenantSize.integrationSelection,
  );
  const limitSize = useSelector(
    (state: AppState) => state.createTenant.limitSize,
  );
  const selectedStorageType = useSelector(
    (state: AppState) =>
      state.createTenant.fields.nameTenant.selectedStorageType,
  );

  const [validationErrors, setValidationErrors] = useState<any>({});

  // Common
  const updateField = useCallback(
    (field: string, value: any) => {
      dispatch(
        updateAddField({
          pageName: "tenantSize",
          field: field,
          value: value,
        }),
      );
    },
    [dispatch],
  );

  const updateMainField = useCallback(
    (field: string, value: string) => {
      dispatch(
        updateAddField({
          pageName: "nameTenant",
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

  /*Debounce functions*/

  // Storage Quotas
  useEffect(() => {
    if (ecParityChoices.length > 0 && distribution.error === "") {
      const ecCodeValidated = erasureCodeCalc(
        cleanECChoices,
        distribution.persistentVolumes,
        distribution.pvSize,
        distribution.nodes,
      );

      updateField("ecParityCalc", ecCodeValidated);

      if (!cleanECChoices.includes(ecParity) || ecParity === "") {
        updateField("ecParity", ecCodeValidated.defaultEC);
      }
    }
  }, [ecParity, ecParityChoices, distribution, cleanECChoices, updateField]);
  /*End debounce functions*/

  /*Set location Storage Types*/
  useEffect(() => {
    if (formToRender !== undefined && parseInt(nodes) >= 4) {
      const setConfigs = mkPanelConfigurations[formToRender];
      const keyCount = Object.keys(setConfigs).length;

      //Configuration is filled
      if (keyCount > 0) {
        const configs: IntegrationConfiguration[] = get(
          setConfigs,
          "configurations",
          [],
        );

        const mainSelection = configs.find(
          (item) => item.typeSelection === selectedStorageType,
        );

        if (mainSelection) {
          updateField("integrationSelection", mainSelection);
          updateMainField("selectedStorageClass", mainSelection.storageClass);

          let pvSize = parseInt(
            getBytes(
              mainSelection.driveSize.driveSize,
              mainSelection.driveSize.sizeUnit,
              true,
            ),
            10,
          );

          const distrCalculate: IStorageDistribution = {
            pvSize,
            nodes: parseInt(nodes),
            disks: mainSelection.drivesPerServer,
            persistentVolumes: mainSelection.drivesPerServer * parseInt(nodes),
            error: "",
          };

          updateField("distribution", distrCalculate);
          // apply requests, half of the available resources
          updateField(
            "resourcesCPURequest",
            Math.max(1, mainSelection.CPU / 2),
          );
          updateField(
            "resourcesMemoryRequest",
            Math.max(2, mainSelection.memory / 2),
          );
        }
      }
    }
  }, [nodes, selectedStorageType, formToRender, updateField, updateMainField]);

  /*Calculate Allocation End*/

  /* Validations of pages */

  useEffect(() => {
    const commonValidation = commonFormValidation([
      {
        fieldKey: "nodes",
        required: true,
        value: nodes,
        customValidation: parseInt(nodes) < 4,
        customValidationMessage: "Al least 4 servers must be selected",
      },
    ]);

    dispatch(
      isPageValid({
        pageName: "tenantSize",
        valid:
          !("nodes" in commonValidation) &&
          distribution.error === "" &&
          ecParityCalc.error === 0 &&
          resourcesSize.error === "" &&
          ecParity !== "" &&
          parseInt(nodes) >= 4,
      }),
    );

    setValidationErrors(commonValidation);
  }, [
    nodes,
    volumeSize,
    sizeFactor,
    memoryNode,
    distribution,
    ecParityCalc,
    resourcesSize,
    limitSize,
    selectedStorageType,
    cpuToUse,
    maxCPUsUse,
    dispatch,
    drivesPerServer,
    ecParity,
  ]);

  useEffect(() => {
    if (integrationSelection.drivesPerServer !== 0) {
      // Get EC Value
      if (nodes.trim() !== "") {
        api
          .invoke(
            "GET",
            `api/v1/get-parity/${nodes}/${integrationSelection.drivesPerServer}`,
          )
          .then((ecList: string[]) => {
            updateField("ecParityChoices", ecListTransform(ecList));
            updateField("cleanECChoices", ecList);
          })
          .catch((err: any) => {
            updateField("ecparityChoices", []);
            dispatch(
              isPageValid({
                pageName: "tenantSize",
                valid: false,
              }),
            );
            updateField("ecParity", "");
          });
      }
    }
  }, [integrationSelection, nodes, dispatch, updateField]);

  /* End Validation of pages */

  return (
    <Fragment>
      <Grid item xs={12}>
        <Box className={"inputItem"}>
          <H3Section>Tenant Size</H3Section>
          <span className={"muted"}>Please select the desired capacity</span>
        </Box>
      </Grid>
      {distribution.error !== "" && (
        <Grid item xs={12}>
          <div className={"error"}>{distribution.error}</div>
        </Grid>
      )}
      {resourcesSize.error !== "" && (
        <Grid item xs={12}>
          <div className={"error"}>{resourcesSize.error}</div>
        </Grid>
      )}
      <InputBox
        id="nodes"
        name="nodes"
        onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
          if (e.target.validity.valid) {
            updateField("nodes", e.target.value);
            cleanValidation("nodes");
          }
        }}
        label="Number of Servers"
        disabled={selectedStorageType === ""}
        value={nodes}
        min="4"
        required
        error={validationErrors["nodes"] || ""}
        pattern={"[0-9]*"}
      />
      <Select
        id="ec_parity"
        name="ec_parity"
        onChange={(value) => {
          updateField("ecParity", value);
        }}
        label="Erasure Code Parity"
        disabled={selectedStorageType === ""}
        value={ecParity}
        options={ecParityChoices}
      />
      <span className={"muted"}>
        Please select the desired parity. This setting will change the max
        usable capacity in the cluster
      </span>
    </Fragment>
  );
};

export default TenantSizeMK;
