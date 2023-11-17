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
import { useSelector } from "react-redux";
import { Box, InputBox, Select } from "mds";
import { AppState, useAppDispatch } from "../../../../../../store";
import {
  calculateDistribution,
  erasureCodeCalc,
  getBytes,
  k8sScalarUnitsExcluding,
  niceBytes,
} from "../../../../../../common/utils";
import { clearValidationError } from "../../../utils";
import { ecListTransform } from "../../../ListTenants/utils";
import { ICapacity } from "../../../../../../common/types";
import { commonFormValidation } from "../../../../../../utils/validationFunctions";
import { IMkEnvs } from "./utils";
import { isPageValid, updateAddField } from "../../createTenantSlice";
import api from "../../../../../../common/api";
import TenantSizeResources from "./TenantSizeResources";
import InputUnitMenu from "../../../../Common/FormComponents/InputUnitMenu/InputUnitMenu";
import H3Section from "../../../../Common/H3Section";

interface ITenantSizeProps {
  formToRender?: IMkEnvs;
}

const TenantSize = ({ formToRender }: ITenantSizeProps) => {
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
  const untouchedECField = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.untouchedECField,
  );
  const limitSize = useSelector(
    (state: AppState) => state.createTenant.limitSize,
  );
  const selectedStorageClass = useSelector(
    (state: AppState) =>
      state.createTenant.fields.nameTenant.selectedStorageClass,
  );
  const selectedStorageType = useSelector(
    (state: AppState) =>
      state.createTenant.fields.nameTenant.selectedStorageType,
  );

  const [validationErrors, setValidationErrors] = useState<any>({});
  const [errorFlag, setErrorFlag] = useState<boolean>(false);
  const [nodeError, setNodeError] = useState<string>("");

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

  const cleanValidation = (fieldName: string) => {
    setValidationErrors(clearValidationError(validationErrors, fieldName));
  };

  /*Debounce functions*/

  // Storage Quotas
  useEffect(() => {
    if (cleanECChoices.length > 0 && ecParityCalc.defaultEC !== "") {
      updateField(
        "ecParityChoices",
        ecListTransform(cleanECChoices, ecParityCalc.defaultEC),
      );
    }
  }, [ecParityCalc, cleanECChoices, updateField]);

  useEffect(() => {
    if (ecParity !== "" && ecParityCalc.defaultEC !== ecParity) {
      updateField("untouchedECField", false);
      return;
    }

    updateField("untouchedECField", true);
  }, [ecParity, ecParityCalc, updateField]);

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
  }, [
    ecParity,
    ecParityChoices.length,
    distribution,
    cleanECChoices,
    updateField,
    untouchedECField,
  ]);
  /*End debounce functions*/

  /*Calculate Allocation*/
  useEffect(() => {
    //Validate Cluster Size
    const size = volumeSize;
    const factor = sizeFactor;
    const limitSize = getBytes("16", "Ti", true);

    const clusterCapacity: ICapacity = {
      unit: factor,
      value: size.toString(),
    };

    const distrCalculate = calculateDistribution(
      clusterCapacity,
      parseInt(nodes),
      parseInt(limitSize),
      parseInt(drivesPerServer),
      formToRender,
      selectedStorageType,
    );

    updateField("distribution", distrCalculate);
    setErrorFlag(false);
    setNodeError("");
  }, [
    nodes,
    volumeSize,
    sizeFactor,
    updateField,
    drivesPerServer,
    selectedStorageType,
    formToRender,
  ]);

  /*Calculate Allocation End*/

  /* Validations of pages */

  useEffect(() => {
    const parsedSize = getBytes(volumeSize, sizeFactor, true);

    const commonValidation = commonFormValidation([
      {
        fieldKey: "nodes",
        required: true,
        value: nodes,
        customValidation: errorFlag,
        customValidationMessage: nodeError,
      },
      {
        fieldKey: "volume_size",
        required: true,
        value: volumeSize,
        customValidation:
          parseInt(parsedSize) < 1073741824 ||
          parseInt(parsedSize) > limitSize[selectedStorageClass],
        customValidationMessage: `Volume size must be greater than 1Gi and less than ${niceBytes(
          limitSize[selectedStorageClass],
          true,
        )}`,
      },
      {
        fieldKey: "drivesps",
        required: true,
        value: drivesPerServer,
        customValidation: parseInt(drivesPerServer) < 1,
        customValidationMessage: "There must be at least one drive",
      },
    ]);

    dispatch(
      isPageValid({
        pageName: "tenantSize",
        valid:
          !("nodes" in commonValidation) &&
          !("volume_size" in commonValidation) &&
          !("drivesps" in commonValidation) &&
          distribution.error === "" &&
          ecParityCalc.error === 0 &&
          ecParity !== "",
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
    selectedStorageClass,
    dispatch,
    errorFlag,
    nodeError,
    drivesPerServer,
    ecParity,
  ]);

  useEffect(() => {
    if (distribution.error === "") {
      // Get EC Value
      if (nodes.trim() !== "" && distribution.disks !== 0) {
        api
          .invoke("GET", `api/v1/get-parity/${nodes}/${distribution.disks}`)
          .then((ecList: string[]) => {
            updateField("ecParityChoices", ecListTransform(ecList));
            updateField("cleanECChoices", ecList);
            if (untouchedECField) {
              updateField("ecParity", "");
            }
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
  }, [distribution, dispatch, updateField, nodes, untouchedECField]);

  /* End Validation of pages */

  return (
    <Fragment>
      <Box className={"inputItem"}>
        <H3Section>Capacity</H3Section>
        <span className={"muted"}>Please select the desired capacity</span>
      </Box>
      {distribution.error !== "" && (
        <Box className={"inputItem error"}>{distribution.error}</Box>
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
        disabled={selectedStorageClass === ""}
        value={nodes}
        min="4"
        required
        error={validationErrors["nodes"] || ""}
        pattern={"[0-9]*"}
      />
      <InputBox
        id="drivesps"
        name="drivesps"
        onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
          if (e.target.validity.valid) {
            updateField("drivesPerServer", e.target.value);
            cleanValidation("drivesps");
          }
        }}
        label="Drives per Server"
        value={drivesPerServer}
        disabled={selectedStorageClass === ""}
        min="1"
        required
        error={validationErrors["drivesps"] || ""}
        pattern={"[0-9]*"}
      />
      <InputBox
        type="number"
        id="volume_size"
        name="volume_size"
        onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
          updateField("volumeSize", e.target.value);
          cleanValidation("volume_size");
        }}
        label="Total Size"
        value={volumeSize}
        disabled={selectedStorageClass === ""}
        required
        error={validationErrors["volume_size"] || ""}
        min="0"
        overlayObject={
          <InputUnitMenu
            id={"size-unit"}
            onUnitChange={(newValue) => {
              updateField("sizeFactor", newValue);
            }}
            unitSelected={sizeFactor}
            unitsList={k8sScalarUnitsExcluding(["Ki", "Mi"])}
            disabled={selectedStorageClass === ""}
          />
        }
      />
      <Select
        id="ec_parity"
        name="ec_parity"
        onChange={(value) => {
          updateField("ecParity", value);
        }}
        label="Erasure Code Parity"
        disabled={selectedStorageClass === ""}
        value={ecParity}
        options={ecParityChoices}
      />
      <Box className={"muted inputItem"}>
        Please select the desired parity. This setting will change the max
        usable capacity in the cluster
      </Box>

      <TenantSizeResources />
    </Fragment>
  );
};

export default TenantSize;
