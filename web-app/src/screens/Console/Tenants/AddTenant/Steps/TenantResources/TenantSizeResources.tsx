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
import { Box, InputBox, Switch } from "mds";
import { useSelector } from "react-redux";
import floor from "lodash/floor";
import get from "lodash/get";
import { AppState, useAppDispatch } from "../../../../../../store";
import { AllocableResourcesResponse } from "../../../types";
import { isPageValid, updateAddField } from "../../createTenantSlice";
import api from "../../../../../../common/api";
import InputUnitMenu from "../../../../Common/FormComponents/InputUnitMenu/InputUnitMenu";
import H3Section from "../../../../Common/H3Section";

const TenantSizeResources = () => {
  const dispatch = useAppDispatch();

  const nodes = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.nodes,
  );

  const resourcesSize = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.resourcesSize,
  );
  const selectedStorageClass = useSelector(
    (state: AppState) =>
      state.createTenant.fields.nameTenant.selectedStorageClass,
  );
  const maxCPUsUse = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.maxCPUsUse,
  );
  const maxMemorySize = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.maxMemorySize,
  );

  const resourcesSpecifyLimit = useSelector(
    (state: AppState) =>
      state.createTenant.fields.tenantSize.resourcesSpecifyLimit,
  );

  const resourcesCPURequestError = useSelector(
    (state: AppState) =>
      state.createTenant.fields.tenantSize.resourcesCPURequestError,
  );
  const resourcesCPURequest = useSelector(
    (state: AppState) =>
      state.createTenant.fields.tenantSize.resourcesCPURequest,
  );
  const resourcesCPULimitError = useSelector(
    (state: AppState) =>
      state.createTenant.fields.tenantSize.resourcesCPULimitError,
  );
  const resourcesCPULimit = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.resourcesCPULimit,
  );

  const resourcesMemoryRequestError = useSelector(
    (state: AppState) =>
      state.createTenant.fields.tenantSize.resourcesMemoryRequestError,
  );
  const resourcesMemoryRequest = useSelector(
    (state: AppState) =>
      state.createTenant.fields.tenantSize.resourcesMemoryRequest,
  );
  const resourcesMemoryLimitError = useSelector(
    (state: AppState) =>
      state.createTenant.fields.tenantSize.resourcesMemoryLimitError,
  );
  const resourcesMemoryLimit = useSelector(
    (state: AppState) =>
      state.createTenant.fields.tenantSize.resourcesMemoryLimit,
  );

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

  /*Debounce functions*/

  useEffect(() => {
    dispatch(
      isPageValid({
        pageName: "tenantSize",
        valid:
          resourcesMemoryRequestError === "" &&
          resourcesMemoryLimitError === "" &&
          resourcesCPURequestError === "" &&
          resourcesCPULimitError === "",
      }),
    );
  }, [
    dispatch,
    resourcesMemoryRequestError,
    resourcesMemoryLimitError,
    resourcesCPURequestError,
    resourcesCPULimitError,
  ]);

  /*End debounce functions*/

  /*Calculate Allocation*/
  useEffect(() => {
    // Get allocatable Resources
    api
      .invoke("GET", `api/v1/cluster/allocatable-resources?num_nodes=${nodes}`)
      .then((res: AllocableResourcesResponse) => {
        updateField("maxAllocatableResources", res);

        const maxAllocatableResources = res;

        const memoryExists = get(
          maxAllocatableResources,
          "min_allocatable_mem",
          false,
        );

        const cpuExists = get(
          maxAllocatableResources,
          "min_allocatable_cpu",
          false,
        );

        if (memoryExists === false || cpuExists === false) {
          updateField("cpuToUse", 0);

          updateField("maxMemorySize", "");
          updateField("maxCPUsUse", "");

          return;
        }

        const maxMemory = floor(
          res.mem_priority.max_allocatable_mem / 1024 / 1024 / 1024,
        );
        // We default to Best CPU Configuration
        updateField("maxMemorySize", maxMemory.toString());
        updateField(
          "maxCPUsUse",
          res.cpu_priority.max_allocatable_cpu.toString(),
        );

        const maxAllocatableCPU = get(
          maxAllocatableResources,
          "cpu_priority.max_allocatable_cpu",
          0,
        );

        const baseCpuUse = Math.max(1, floor(maxAllocatableCPU / 2));
        if (resourcesCPURequest === "") {
          updateField("resourcesCPURequest", baseCpuUse);
        }

        const baseMemoryUse = Math.max(2, floor(maxMemory / 2));
        if (resourcesMemoryRequest === "") {
          updateField("resourcesMemoryRequest", baseMemoryUse);
        }
      })
      .catch((err: any) => {
        updateField("maxMemorySize", 0);
        updateField("resourcesCPURequest", "");
        updateField("resourcesMemoryRequest", "");

        console.error(err);
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [nodes, updateField]);

  /*Calculate Allocation End*/

  return (
    <Fragment>
      <Box className={"inputItem"}>
        <H3Section>Resources</H3Section>
        <span className={"muted"}>
          You may specify the amount of CPU and Memory that MinIO servers should
          reserve on each node.
        </span>
      </Box>
      {resourcesSize.error !== "" && (
        <Box className={"inputItem error"}>{resourcesSize.error}</Box>
      )}
      <InputBox
        label={"CPU Request"}
        id={"resourcesCPURequest"}
        name={"resourcesCPURequest"}
        onChange={(e) => {
          let value = parseInt(e.target.value);
          if (e.target.value === "") {
            updateField("resourcesCPURequestError", "");
          } else if (isNaN(value)) {
            updateField("resourcesCPURequestError", "Invalid number");
          } else if (value > parseInt(maxCPUsUse)) {
            updateField(
              "resourcesCPURequestError",
              `Request exceeds available cores (${maxCPUsUse})`,
            );
          } else if (e.target.validity.valid) {
            updateField("resourcesCPURequestError", "");
          } else {
            updateField("resourcesCPURequestError", "Invalid configuration");
          }
          updateField("resourcesCPURequest", e.target.value);
        }}
        value={resourcesCPURequest}
        disabled={selectedStorageClass === ""}
        max={maxCPUsUse}
        error={resourcesCPURequestError}
        pattern={"[0-9]*"}
      />
      <InputBox
        id="resourcesMemoryRequest"
        name="resourcesMemoryRequest"
        onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
          let value = parseInt(e.target.value);
          if (e.target.value === "") {
            updateField("resourcesMemoryRequestError", "");
          } else if (isNaN(value)) {
            updateField("resourcesMemoryRequestError", "Invalid number");
          } else if (value > parseInt(maxMemorySize)) {
            updateField(
              "resourcesMemoryRequestError",
              `Request exceeds available memory across ${nodes} nodes (${maxMemorySize}Gi)`,
            );
          } else if (value < 2) {
            updateField(
              "resourcesMemoryRequestError",
              "At least 2Gi must be requested",
            );
          } else if (e.target.validity.valid) {
            updateField("resourcesMemoryRequestError", "");
          } else {
            updateField("resourcesMemoryRequestError", "Invalid configuration");
          }
          updateField("resourcesMemoryRequest", e.target.value);
        }}
        label="Memory Request"
        overlayObject={
          <InputUnitMenu
            id={"size-unit"}
            onUnitChange={() => {}}
            unitSelected={"Gi"}
            unitsList={[{ label: "Gi", value: "Gi" }]}
            disabled={true}
          />
        }
        value={resourcesMemoryRequest}
        disabled={selectedStorageClass === ""}
        error={resourcesMemoryRequestError}
        pattern={"[0-9]*"}
      />
      <Switch
        value="resourcesSpecifyLimit"
        id="resourcesSpecifyLimit"
        name="resourcesSpecifyLimit"
        checked={resourcesSpecifyLimit}
        onChange={(e) => {
          const targetD = e.target;
          const checked = targetD.checked;

          updateField("resourcesSpecifyLimit", checked);
        }}
        label={"Specify Limit"}
      />

      {resourcesSpecifyLimit && (
        <Fragment>
          <InputBox
            label={"CPU Limit"}
            id={"resourcesCPULimit"}
            name={"resourcesCPULimit"}
            onChange={(e) => {
              let value = parseInt(e.target.value);
              if (e.target.value === "") {
                updateField("resourcesCPULimitError", "");
              } else if (isNaN(value)) {
                updateField("resourcesCPULimitError", "Invalid number");
              } else if (e.target.validity.valid) {
                updateField("resourcesCPULimitError", "");
              } else {
                updateField("resourcesCPULimitError", "Invalid configuration");
              }
              updateField("resourcesCPULimit", e.target.value);
            }}
            value={resourcesCPULimit}
            disabled={selectedStorageClass === ""}
            max={maxCPUsUse}
            error={resourcesCPULimitError}
            pattern={"[0-9]*"}
          />
          <InputBox
            id="resourcesMemoryLimit"
            name="resourcesMemoryLimit"
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
              let value = parseInt(e.target.value);
              if (e.target.value === "") {
                updateField("resourcesMemoryLimitError", "");
              } else if (isNaN(value)) {
                updateField("resourcesMemoryLimitError", "Invalid number");
              } else if (e.target.validity.valid) {
                updateField("resourcesMemoryLimitError", "");
              } else {
                updateField(
                  "resourcesMemoryLimitError",
                  "Invalid configuration",
                );
              }
              updateField("resourcesMemoryLimit", e.target.value);
            }}
            label="Memory Limit"
            overlayObject={
              <InputUnitMenu
                id={"size-unit"}
                onUnitChange={() => {}}
                unitSelected={"Gi"}
                unitsList={[{ label: "Gi", value: "Gi" }]}
                disabled={true}
              />
            }
            value={resourcesMemoryLimit}
            disabled={selectedStorageClass === ""}
            error={resourcesMemoryLimitError}
            pattern={"[0-9]*"}
          />
        </Fragment>
      )}
    </Fragment>
  );
};

export default TenantSizeResources;
