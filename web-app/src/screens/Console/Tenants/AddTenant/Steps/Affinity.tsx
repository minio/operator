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
import {
  AddIcon,
  RemoveIcon,
  FormLayout,
  Box,
  InputLabel,
  RadioGroup,
  Switch,
  Select,
  InputBox,
  IconButton,
  Grid,
} from "mds";
import { useSelector } from "react-redux";
import styled from "styled-components";
import { AppState, useAppDispatch } from "../../../../../store";
import {
  commonFormValidation,
  IValidation,
} from "../../../../../utils/validationFunctions";
import { ErrorResponseHandler } from "../../../../../common/types";
import { LabelKeyPair } from "../../types";
import { setModalErrorSnackMessage } from "../../../../../systemSlice";
import {
  addNewToleration,
  isPageValid,
  removeToleration,
  setKeyValuePairs,
  setTolerationInfo,
  updateAddField,
} from "../createTenantSlice";
import api from "../../../../../common/api";
import TolerationSelector from "../../../Common/TolerationSelector/TolerationSelector";
import H3Section from "../../../Common/H3Section";

const AffinityContainer = styled.div(() => ({
  "& .overlayAction": {
    marginLeft: 10,
    display: "flex",
    alignItems: "center",
  },
  "& .affinityConfigField": {
    display: "flex",
  },
  "& .affinityFieldLabel": {
    display: "flex",
    flexFlow: "column",
    flex: 1,
  },
  "& .affinityLabelKey": {
    "& div:first-child": {
      marginBottom: 0,
    },
  },
  "& .affinityLabelValue": {
    marginLeft: 10,
    "& div:first-child": {
      marginBottom: 0,
    },
  },
  "& .rowActions": {
    display: "flex",
    alignItems: "center",
  },
  "& .affinityRow": {
    marginBottom: 10,
    display: "flex",
  },
}));

interface OptionPair {
  label: string;
  value: string;
}

const Affinity = () => {
  const dispatch = useAppDispatch();

  const podAffinity = useSelector(
    (state: AppState) => state.createTenant.fields.affinity.podAffinity,
  );
  const nodeSelectorLabels = useSelector(
    (state: AppState) => state.createTenant.fields.affinity.nodeSelectorLabels,
  );
  const withPodAntiAffinity = useSelector(
    (state: AppState) => state.createTenant.fields.affinity.withPodAntiAffinity,
  );
  const keyValuePairs = useSelector(
    (state: AppState) => state.createTenant.nodeSelectorPairs,
  );
  const tolerations = useSelector(
    (state: AppState) => state.createTenant.tolerations,
  );

  const [validationErrors, setValidationErrors] = useState<any>({});
  const [loading, setLoading] = useState<boolean>(true);
  const [keyValueMap, setKeyValueMap] = useState<{ [key: string]: string[] }>(
    {},
  );
  const [keyOptions, setKeyOptions] = useState<OptionPair[]>([]);

  // Common
  const updateField = useCallback(
    (field: string, value: any) => {
      dispatch(
        updateAddField({
          pageName: "affinity",
          field: field,
          value: value,
        }),
      );
    },
    [dispatch],
  );

  useEffect(() => {
    if (loading) {
      api
        .invoke("GET", `/api/v1/nodes/labels`)
        .then((res: { [key: string]: string[] }) => {
          setLoading(false);
          setKeyValueMap(res);
          let keys: OptionPair[] = [];
          for (let k in res) {
            keys.push({
              label: k,
              value: k,
            });
          }
          setKeyOptions(keys);
        })
        .catch((err: ErrorResponseHandler) => {
          setLoading(false);
          dispatch(setModalErrorSnackMessage(err));
          setKeyValueMap({});
        });
    }
  }, [dispatch, loading]);

  useEffect(() => {
    if (keyValuePairs) {
      const vlr = keyValuePairs
        .filter((kvp) => kvp.key !== "")
        .map((kvp) => `${kvp.key}=${kvp.value}`)
        .filter((kvs, i, a) => a.indexOf(kvs) === i);
      const vl = vlr.join("&");
      updateField("nodeSelectorLabels", vl);
    }
  }, [keyValuePairs, updateField]);

  // Validation
  useEffect(() => {
    let customAccountValidation: IValidation[] = [];

    if (podAffinity === "nodeSelector") {
      let valid = true;

      const splittedLabels = nodeSelectorLabels.split("&");

      if (splittedLabels.length === 1 && splittedLabels[0] === "") {
        valid = false;
      }

      splittedLabels.forEach((item: string, index: number) => {
        const splitItem = item.split("=");

        if (splitItem.length !== 2) {
          valid = false;
        }

        if (index + 1 !== splittedLabels.length) {
          if (splitItem[0] === "" || splitItem[1] === "") {
            valid = false;
          }
        }
      });

      customAccountValidation = [
        ...customAccountValidation,
        {
          fieldKey: "labels",
          required: true,
          value: nodeSelectorLabels,
          customValidation: !valid,
          customValidationMessage:
            "You need to add at least one label key-pair",
        },
      ];
    }

    const commonVal = commonFormValidation(customAccountValidation);

    dispatch(
      isPageValid({
        pageName: "affinity",
        valid: Object.keys(commonVal).length === 0,
      }),
    );

    setValidationErrors(commonVal);
  }, [dispatch, podAffinity, nodeSelectorLabels]);

  const updateToleration = (index: number, field: string, value: any) => {
    const alterToleration = { ...tolerations[index], [field]: value };

    dispatch(
      setTolerationInfo({
        index: index,
        tolerationValue: alterToleration,
      }),
    );
  };

  return (
    <AffinityContainer>
      <FormLayout withBorders={false} containerPadding={false}>
        <Box className={"inputItem"}>
          <H3Section>Pod Placement</H3Section>
          <span className={"muted"}>
            Configure how pods will be assigned to nodes
          </span>
        </Box>
        <Box>
          <InputLabel>Type</InputLabel>
        </Box>
        <Box className={`muted inputItem`}>
          MinIO supports multiple configurations for Pod Affinity
        </Box>
        <RadioGroup
          currentValue={podAffinity}
          id="affinity-options"
          name="affinity-options"
          label={" "}
          onChange={(e) => {
            updateField("podAffinity", e.target.value);
          }}
          selectorOptions={[
            { label: "None", value: "none" },
            { label: "Default (Pod Anti-Affinity)", value: "default" },
            { label: "Node Selector", value: "nodeSelector" },
          ]}
          displayInColumn
        />
        {podAffinity === "nodeSelector" && (
          <Fragment>
            <Switch
              value="with_pod_anti_affinity"
              id="with_pod_anti_affinity"
              name="with_pod_anti_affinity"
              checked={withPodAntiAffinity}
              onChange={(e) => {
                const targetD = e.target;
                const checked = targetD.checked;

                updateField("withPodAntiAffinity", checked);
              }}
              label={"With Pod Anti-Affinity"}
            />
            <Box className={"inputBox"}>
              <h3>Labels</h3>
              <span className={"error"}>{validationErrors["labels"]}</span>
              <Grid container>
                {keyValuePairs &&
                  keyValuePairs.map((kvp, i) => {
                    return (
                      <Grid
                        item
                        xs={12}
                        className={"affinityRow"}
                        key={`affinity-keyVal-${i.toString()}`}
                      >
                        <Grid item xs={5} className={"affinityLabelKey"}>
                          {keyOptions.length > 0 && (
                            <Select
                              onChange={(value) => {
                                const newKey = value;
                                const newLKP: LabelKeyPair = {
                                  key: newKey,
                                  value: keyValueMap[newKey][0],
                                };
                                const arrCp: LabelKeyPair[] = [
                                  ...keyValuePairs,
                                ];
                                arrCp[i] = newLKP;
                                dispatch(setKeyValuePairs(arrCp));
                              }}
                              id="select-access-policy"
                              name="select-access-policy"
                              label={""}
                              value={kvp.key}
                              options={keyOptions}
                            />
                          )}
                          {keyOptions.length === 0 && (
                            <InputBox
                              id={`nodeselector-key-${i.toString()}`}
                              label={""}
                              name={`nodeselector-${i.toString()}`}
                              value={kvp.key}
                              onChange={(e) => {
                                const arrCp: LabelKeyPair[] = [
                                  ...keyValuePairs,
                                ];
                                arrCp[i] = {
                                  key: arrCp[i].key,
                                  value: e.target.value as string,
                                };
                                dispatch(setKeyValuePairs(arrCp));
                              }}
                              index={i}
                              placeholder={"Key"}
                            />
                          )}
                        </Grid>
                        <Grid item xs={5} className={"affinityLabelValue"}>
                          {keyOptions.length > 0 && (
                            <Select
                              onChange={(value) => {
                                const arrCp: LabelKeyPair[] = [
                                  ...keyValuePairs,
                                ];
                                arrCp[i] = {
                                  key: arrCp[i].key,
                                  value: value,
                                };
                                dispatch(setKeyValuePairs(arrCp));
                              }}
                              id="select-access-policy"
                              name="select-access-policy"
                              label={""}
                              value={kvp.value}
                              options={
                                keyValueMap[kvp.key]
                                  ? keyValueMap[kvp.key].map((v) => {
                                      return { label: v, value: v };
                                    })
                                  : []
                              }
                            />
                          )}
                          {keyOptions.length === 0 && (
                            <InputBox
                              id={`nodeselector-value-${i.toString()}`}
                              label={""}
                              name={`nodeselector-${i.toString()}`}
                              value={kvp.value}
                              onChange={(e) => {
                                const arrCp: LabelKeyPair[] = [
                                  ...keyValuePairs,
                                ];
                                arrCp[i] = {
                                  key: arrCp[i].key,
                                  value: e.target.value as string,
                                };
                                dispatch(setKeyValuePairs(arrCp));
                              }}
                              index={i}
                              placeholder={"value"}
                            />
                          )}
                        </Grid>
                        <Grid item xs={2} className={"rowActions"}>
                          <Box className={"overlayAction"}>
                            <IconButton
                              size={"small"}
                              onClick={() => {
                                const arrCp = [...keyValuePairs];
                                if (keyOptions.length > 0) {
                                  arrCp.push({
                                    key: keyOptions[0].value,
                                    value: keyValueMap[keyOptions[0].value][0],
                                  });
                                } else {
                                  arrCp.push({ key: "", value: "" });
                                }

                                dispatch(setKeyValuePairs(arrCp));
                              }}
                              disabled={i !== keyValuePairs.length - 1}
                            >
                              <AddIcon />
                            </IconButton>
                          </Box>
                          <Box className={"overlayAction"}>
                            <IconButton
                              size={"small"}
                              onClick={() => {
                                const arrCp = keyValuePairs.filter(
                                  (item, index) => index !== i,
                                );
                                dispatch(setKeyValuePairs(arrCp));
                              }}
                              disabled={keyValuePairs.length <= 1}
                            >
                              <RemoveIcon />
                            </IconButton>
                          </Box>
                        </Grid>
                      </Grid>
                    );
                  })}
              </Grid>
            </Box>
          </Fragment>
        )}
        <Grid item xs={12} className={"affinityConfigField"}>
          <Grid item className={"affinityFieldLabel"}>
            <h3>Tolerations</h3>
            <span className={"error"}>{validationErrors["tolerations"]}</span>
            <Grid container>
              {tolerations &&
                tolerations.map((tol, i) => {
                  return (
                    <Grid
                      item
                      xs={12}
                      className={"affinityRow"}
                      key={`affinity-keyVal-${i.toString()}`}
                    >
                      <TolerationSelector
                        effect={tol.effect}
                        onEffectChange={(value) => {
                          updateToleration(i, "effect", value);
                        }}
                        tolerationKey={tol.key}
                        onTolerationKeyChange={(value) => {
                          updateToleration(i, "key", value);
                        }}
                        operator={tol.operator}
                        onOperatorChange={(value) => {
                          updateToleration(i, "operator", value);
                        }}
                        value={tol.value}
                        onValueChange={(value) => {
                          updateToleration(i, "value", value);
                        }}
                        tolerationSeconds={tol.tolerationSeconds?.seconds || 0}
                        onSecondsChange={(value) => {
                          updateToleration(i, "tolerationSeconds", {
                            seconds: value,
                          });
                        }}
                        index={i}
                      />
                      <Box className={"overlayAction"}>
                        <IconButton
                          size={"small"}
                          onClick={() => {
                            dispatch(addNewToleration());
                          }}
                          disabled={i !== tolerations.length - 1}
                        >
                          <AddIcon />
                        </IconButton>
                      </Box>

                      <Box className={"overlayAction"}>
                        <IconButton
                          size={"small"}
                          onClick={() => dispatch(removeToleration(i))}
                          disabled={tolerations.length <= 1}
                        >
                          <RemoveIcon />
                        </IconButton>
                      </Box>
                    </Grid>
                  );
                })}
            </Grid>
          </Grid>
        </Grid>
      </FormLayout>
    </AffinityContainer>
  );
};

export default Affinity;
