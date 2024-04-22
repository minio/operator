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

import React, { Fragment, useEffect, useState } from "react";
import { Box, FormLayout, InputBox, Select } from "mds";
import get from "lodash/get";
import { niceBytes } from "../../../../../../common/utils";
import { ErrorResponseHandler } from "../../../../../../common/types";
import { IQuotaElement, IQuotas } from "../../../ListTenants/utils";
import { AppState, useAppDispatch } from "../../../../../../store";
import { useSelector } from "react-redux";
import {
  commonFormValidation,
  IValidation,
} from "../../../../../../utils/validationFunctions";
import InputUnitMenu from "../../../../Common/FormComponents/InputUnitMenu/InputUnitMenu";
import {
  isPoolPageValid,
  setPoolField,
  setPoolStorageClasses,
} from "./addPoolSlice";
import api from "../../../../../../common/api";
import H3Section from "../../../../Common/H3Section";

const PoolResources = () => {
  const dispatch = useAppDispatch();

  const tenant = useSelector((state: AppState) => state.tenants.tenantInfo);
  const storageClasses = useSelector(
    (state: AppState) => state.addPool.storageClasses,
  );
  const numberOfNodes = useSelector((state: AppState) =>
    state.addPool.setup.numberOfNodes.toString(),
  );
  const storageClass = useSelector(
    (state: AppState) => state.addPool.setup.storageClass,
  );
  const volumeSize = useSelector((state: AppState) =>
    state.addPool.setup.volumeSize.toString(),
  );
  const volumesPerServer = useSelector((state: AppState) =>
    state.addPool.setup.volumesPerServer.toString(),
  );

  const [validationErrors, setValidationErrors] = useState<any>({});

  const instanceCapacity: number =
    parseInt(volumeSize) * 1073741824 * parseInt(volumesPerServer);
  const totalCapacity: number = instanceCapacity * parseInt(numberOfNodes);

  // Validation
  useEffect(() => {
    let customAccountValidation: IValidation[] = [
      {
        fieldKey: "number_of_nodes",
        required: true,
        value: numberOfNodes.toString(),
        customValidation:
          parseInt(numberOfNodes) < 1 || isNaN(parseInt(numberOfNodes)),
        customValidationMessage: "Number of servers must be at least 1",
      },
      {
        fieldKey: "pool_size",
        required: true,
        value: volumeSize.toString(),
        customValidation:
          parseInt(volumeSize) < 1 || isNaN(parseInt(volumeSize)),
        customValidationMessage: "Pool Size cannot be 0",
      },
      {
        fieldKey: "volumes_per_server",
        required: true,
        value: volumesPerServer.toString(),
        customValidation:
          parseInt(volumesPerServer) < 1 || isNaN(parseInt(volumesPerServer)),
        customValidationMessage: "1 volume or more are required",
      },
    ];

    const commonVal = commonFormValidation(customAccountValidation);

    dispatch(
      isPoolPageValid({
        page: "setup",
        status: Object.keys(commonVal).length === 0,
      }),
    );

    setValidationErrors(commonVal);
  }, [dispatch, numberOfNodes, volumeSize, volumesPerServer, storageClass]);

  useEffect(() => {
    if (storageClasses.length === 0 && tenant) {
      api
        .invoke(
          "GET",
          `/api/v1/namespaces/${tenant.namespace}/resourcequotas/${tenant.namespace}-storagequota`,
        )
        .then((res: IQuotas) => {
          const elements: IQuotaElement[] = get(res, "elements", []);

          const newStorage = elements.map((storageClass: any) => {
            const name = get(storageClass, "name", "").split(
              ".storageclass.storage.k8s.io/requests.storage",
            )[0];

            return { label: name, value: name };
          });

          dispatch(
            setPoolField({
              page: "setup",
              field: "storageClass",
              value: newStorage[0].value,
            }),
          );

          dispatch(setPoolStorageClasses(newStorage));
        })
        .catch((err: ErrorResponseHandler) => {
          console.error(err);
        });
    }
  }, [tenant, storageClasses, dispatch]);

  const setFieldInfo = (fieldName: string, value: any) => {
    dispatch(
      setPoolField({
        page: "setup",
        field: fieldName,
        value: value,
      }),
    );
  };

  return (
    <Fragment>
      <Box className={"inputItem"} sx={{ marginBottom: 12 }}>
        <H3Section>New Pool Configuration</H3Section>
        <span className={"muted"}>
          Configure a new Pool to expand MinIO storage
        </span>
      </Box>

      <FormLayout withBorders={false} containerPadding={false}>
        <InputBox
          id="number_of_nodes"
          name="number_of_nodes"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            const intValue = parseInt(e.target.value);

            if (e.target.validity.valid && !isNaN(intValue)) {
              setFieldInfo("numberOfNodes", intValue);
            } else if (isNaN(intValue)) {
              setFieldInfo("numberOfNodes", 0);
            }
          }}
          label="Number of Servers"
          value={numberOfNodes}
          error={validationErrors["number_of_nodes"] || ""}
          pattern={"[0-9]*"}
        />
        <InputBox
          id="pool_size"
          name="pool_size"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            const intValue = parseInt(e.target.value);

            if (e.target.validity.valid && !isNaN(intValue)) {
              setFieldInfo("volumeSize", intValue);
            } else if (isNaN(intValue)) {
              setFieldInfo("volumeSize", 0);
            }
          }}
          label="Volume Size"
          value={volumeSize}
          error={validationErrors["pool_size"] || ""}
          pattern={"[0-9]*"}
          overlayObject={
            <InputUnitMenu
              id={"quota_unit"}
              onUnitChange={() => {}}
              unitSelected={"Gi"}
              unitsList={[{ label: "Gi", value: "Gi" }]}
              disabled={true}
            />
          }
        />
        <InputBox
          id="volumes_per_sever"
          name="volumes_per_sever"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            const intValue = parseInt(e.target.value);

            if (e.target.validity.valid && !isNaN(intValue)) {
              setFieldInfo("volumesPerServer", intValue);
            } else if (isNaN(intValue)) {
              setFieldInfo("volumesPerServer", 0);
            }
          }}
          label="Volumes per Server"
          value={volumesPerServer}
          error={validationErrors["volumes_per_server"] || ""}
          pattern={"[0-9]*"}
        />
        <Select
          id="storage_class"
          name="storage_class"
          onChange={(value) => {
            setFieldInfo("storageClass", value);
          }}
          label="Storage Class"
          value={storageClass}
          options={storageClasses}
          disabled={storageClasses.length < 1}
        />
        <Box
          sx={{
            display: "flex",
            justifyContent: "center",
            marginLeft: 30,
            gap: 25,
            "& .sizeNumber": {
              fontSize: 35,
              fontWeight: 700,
              textAlign: "center",
            },
            "& .sizeDescription": {
              fontSize: 14,
              textAlign: "center",
            },
          }}
        >
          <Box>
            <Box className={"sizeNumber"}>
              {niceBytes(instanceCapacity.toString(10))}
            </Box>
            <Box className={"sizeDescription muted"}>Instance Capacity</Box>
          </Box>
          <Box>
            <Box className={"sizeNumber"}>
              {niceBytes(totalCapacity.toString(10))}
            </Box>
            <Box className={"sizeDescription muted"}>Total Capacity</Box>
          </Box>
        </Box>
      </FormLayout>
    </Fragment>
  );
};

export default PoolResources;
