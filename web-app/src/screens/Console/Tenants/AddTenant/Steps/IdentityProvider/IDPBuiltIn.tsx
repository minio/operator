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
import { IconButton, Tooltip, InputBox, AddIcon, RemoveIcon, Box } from "mds";
import { useSelector } from "react-redux";
import CasinoIcon from "@mui/icons-material/Casino"; // TODO: Implement this in mds
import {
  addIDPNewKeyPair,
  isPageValid,
  removeIDPKeyPairAtIndex,
  setIDPPwdAtIndex,
  setIDPUsrAtIndex,
} from "../../createTenantSlice";
import { clearValidationError, getRandomString } from "../../../utils";
import { AppState, useAppDispatch } from "../../../../../../store";
import {
  commonFormValidation,
  IValidation,
} from "../../../../../../utils/validationFunctions";

const IDPBuiltIn = () => {
  const dispatch = useAppDispatch();

  const idpSelection = useSelector(
    (state: AppState) =>
      state.createTenant.fields.identityProvider.idpSelection,
  );
  const accessKeys = useSelector(
    (state: AppState) => state.createTenant.fields.identityProvider.accessKeys,
  );
  const secretKeys = useSelector(
    (state: AppState) => state.createTenant.fields.identityProvider.secretKeys,
  );

  const [validationErrors, setValidationErrors] = useState<any>({});

  const cleanValidation = (fieldName: string) => {
    setValidationErrors(clearValidationError(validationErrors, fieldName));
  };

  // Validation
  useEffect(() => {
    let customIDPValidation: IValidation[] = [];

    if (idpSelection === "Built-in") {
      customIDPValidation = [...customIDPValidation];
      for (var i = 0; i < accessKeys.length; i++) {
        customIDPValidation.push({
          fieldKey: `accesskey-${i.toString()}`,
          required: true,
          value: accessKeys[i],
          pattern: /^[a-zA-Z0-9-]{8,63}$/,
          customPatternMessage: "Keys must be at least length 8",
        });
        customIDPValidation.push({
          fieldKey: `secretkey-${i.toString()}`,
          required: true,
          value: secretKeys[i],
          pattern: /^[a-zA-Z0-9-]{8,63}$/,
          customPatternMessage: "Keys must be at least length 8",
        });
      }
    }

    const commonVal = commonFormValidation(customIDPValidation);

    dispatch(
      isPageValid({
        pageName: "identityProvider",
        valid: Object.keys(commonVal).length === 0,
      }),
    );

    setValidationErrors(commonVal);
  }, [idpSelection, accessKeys, secretKeys, dispatch]);

  return (
    <Fragment>
      Add additional users
      {accessKeys.map((_, index) => {
        return (
          <Fragment key={`identityField-${index.toString()}`}>
            <Box
              sx={{
                gridTemplateColumns: "auto auto 50px 50px",
                display: "grid",
                gap: 10,
                marginBottom: 10,
              }}
            >
              <InputBox
                id={`accesskey-${index.toString()}`}
                label={""}
                placeholder={"Access Key"}
                name={`accesskey-${index.toString()}`}
                value={accessKeys[index]}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                  dispatch(
                    setIDPUsrAtIndex({
                      index,
                      accessKey: e.target.value,
                    }),
                  );
                  cleanValidation(`accesskey-${index.toString()}`);
                }}
                index={index}
                key={`csv-accesskey-${index.toString()}`}
                error={validationErrors[`accesskey-${index.toString()}`] || ""}
              />
              <InputBox
                id={`secretkey-${index.toString()}`}
                label={""}
                placeholder={"Secret Key"}
                name={`secretkey-${index.toString()}`}
                value={secretKeys[index]}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                  dispatch(
                    setIDPPwdAtIndex({
                      index,
                      secretKey: e.target.value,
                    }),
                  );
                  cleanValidation(`secretkey-${index.toString()}`);
                }}
                index={index}
                key={`csv-secretkey-${index.toString()}`}
                error={validationErrors[`secretkey-${index.toString()}`] || ""}
              />
              <Box
                sx={{
                  display: "flex",
                  alignItems: "center",
                  gap: 10,
                  height: 38,
                }}
              >
                <IconButton
                  size={"small"}
                  onClick={() => {
                    dispatch(addIDPNewKeyPair());
                  }}
                  disabled={index !== accessKeys.length - 1}
                >
                  <AddIcon />
                </IconButton>
                <IconButton
                  size={"small"}
                  onClick={() => {
                    dispatch(removeIDPKeyPairAtIndex(index));
                  }}
                  disabled={accessKeys.length <= 1}
                >
                  <RemoveIcon />
                </IconButton>
                <Tooltip tooltip="Randomize Credentials" aria-label="add">
                  <IconButton
                    onClick={() => {
                      dispatch(
                        setIDPUsrAtIndex({
                          index,
                          accessKey: getRandomString(16),
                        }),
                      );
                      dispatch(
                        setIDPPwdAtIndex({
                          index,
                          secretKey: getRandomString(16),
                        }),
                      );
                    }}
                    size={"small"}
                  >
                    <CasinoIcon />
                  </IconButton>
                </Tooltip>
              </Box>
            </Box>
          </Fragment>
        );
      })}
    </Fragment>
  );
};

export default IDPBuiltIn;
