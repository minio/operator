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

import React, { Fragment, useCallback, useEffect, useState } from "react";
import { useSelector } from "react-redux";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { Grid, IconButton, Paper, SelectChangeEvent } from "@mui/material";
import { AppState, useAppDispatch } from "../../../../../../store";

import {
  modalBasic,
  wizardCommon,
} from "../../../../Common/FormComponents/common/styleLibrary";
import {
  commonFormValidation,
  IValidation,
} from "../../../../../../utils/validationFunctions";
import { ErrorResponseHandler } from "../../../../../../common/types";
import { LabelKeyPair } from "../../../types";
import RadioGroupSelector from "../../../../Common/FormComponents/RadioGroupSelector/RadioGroupSelector";
import FormSwitchWrapper from "../../../../Common/FormComponents/FormSwitchWrapper/FormSwitchWrapper";
import api from "../../../../../../common/api";
import InputBoxWrapper from "../../../../Common/FormComponents/InputBoxWrapper/InputBoxWrapper";
import { AddIcon, RemoveIcon } from "mds";
import SelectWrapper from "../../../../Common/FormComponents/SelectWrapper/SelectWrapper";
import TolerationSelector from "../../../../Common/TolerationSelector/TolerationSelector";
import { setModalErrorSnackMessage } from "../../../../../../systemSlice";
import {
  addNewEditPoolToleration,
  isEditPoolPageValid,
  removeEditPoolToleration,
  setEditPoolField,
  setEditPoolKeyValuePairs,
  setEditPoolTolerationInfo,
} from "./editPoolSlice";
import H3Section from "../../../../Common/H3Section";

interface IAffinityProps {
  classes: any;
}

const styles = (theme: Theme) =>
  createStyles({
    overlayAction: {
      marginLeft: 10,
      display: "flex",
      alignItems: "center",
      "& svg": {
        maxWidth: 15,
        maxHeight: 15,
      },
      "& button": {
        background: "#EAEAEA",
      },
    },
    affinityConfigField: {
      display: "flex",
    },
    affinityFieldLabel: {
      display: "flex",
      flexFlow: "column",
      flex: 1,
    },
    radioField: {
      display: "flex",
      alignItems: "flex-start",
      marginTop: 10,
      "& div:first-child": {
        display: "flex",
        flexFlow: "column",
        alignItems: "baseline",
        textAlign: "left !important",
      },
    },
    affinityLabelKey: {
      "& div:first-child": {
        marginBottom: 0,
      },
    },
    affinityLabelValue: {
      marginLeft: 10,
      "& div:first-child": {
        marginBottom: 0,
      },
    },
    rowActions: {
      display: "flex",
      alignItems: "center",
    },
    affinityRow: {
      marginBottom: 10,
      display: "flex",
    },
    ...modalBasic,
    ...wizardCommon,
  });

interface OptionPair {
  label: string;
  value: string;
}

const Affinity = ({ classes }: IAffinityProps) => {
  const dispatch = useAppDispatch();

  const podAffinity = useSelector(
    (state: AppState) => state.editPool.fields.affinity.podAffinity,
  );
  const nodeSelectorLabels = useSelector(
    (state: AppState) => state.editPool.fields.affinity.nodeSelectorLabels,
  );
  const withPodAntiAffinity = useSelector(
    (state: AppState) => state.editPool.fields.affinity.withPodAntiAffinity,
  );
  const keyValuePairs = useSelector(
    (state: AppState) => state.editPool.fields.nodeSelectorPairs,
  );
  const tolerations = useSelector(
    (state: AppState) => state.editPool.fields.tolerations,
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
        setEditPoolField({
          page: "affinity",
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
      isEditPoolPageValid({
        page: "affinity",
        status: Object.keys(commonVal).length === 0,
      }),
    );

    setValidationErrors(commonVal);
  }, [dispatch, podAffinity, nodeSelectorLabels]);

  const updateToleration = (index: number, field: string, value: any) => {
    const alterToleration = { ...tolerations[index], [field]: value };

    dispatch(
      setEditPoolTolerationInfo({
        index: index,
        tolerationValue: alterToleration,
      }),
    );
  };

  return (
    <Paper className={classes.paperWrapper}>
      <div className={classes.headerElement}>
        <H3Section>Pod Placement</H3Section>
      </div>
      <Grid item xs={12} className={classes.affinityConfigField}>
        <Grid item className={classes.affinityFieldLabel}>
          <div className={classes.label}>Type</div>
          <div
            className={`${classes.descriptionText} ${classes.affinityHelpText}`}
          >
            MinIO supports multiple configurations for Pod Affinity
          </div>
          <Grid item className={classes.radioField}>
            <RadioGroupSelector
              currentSelection={podAffinity}
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
            />
          </Grid>
        </Grid>
      </Grid>
      {podAffinity === "nodeSelector" && (
        <Fragment>
          <br />
          <Grid item xs={12}>
            <FormSwitchWrapper
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
          </Grid>
          <Grid item xs={12}>
            <h3>Labels</h3>
            <span className={classes.error}>{validationErrors["labels"]}</span>
            <Grid container>
              {keyValuePairs &&
                keyValuePairs.map((kvp, i) => {
                  return (
                    <Grid
                      item
                      xs={12}
                      className={classes.affinityRow}
                      key={`affinity-keyVal-${i.toString()}`}
                    >
                      <Grid item xs={5} className={classes.affinityLabelKey}>
                        {keyOptions.length > 0 && (
                          <SelectWrapper
                            onChange={(e: SelectChangeEvent<string>) => {
                              const newKey = e.target.value as string;
                              const newLKP: LabelKeyPair = {
                                key: newKey,
                                value: keyValueMap[newKey][0],
                              };
                              const arrCp: LabelKeyPair[] = [...keyValuePairs];
                              arrCp[i] = newLKP;
                              dispatch(setEditPoolKeyValuePairs(arrCp));
                            }}
                            id="select-access-policy"
                            name="select-access-policy"
                            label={""}
                            value={kvp.key}
                            options={keyOptions}
                          />
                        )}
                        {keyOptions.length === 0 && (
                          <InputBoxWrapper
                            id={`nodeselector-key-${i.toString()}`}
                            label={""}
                            name={`nodeselector-${i.toString()}`}
                            value={kvp.key}
                            onChange={(e) => {
                              const arrCp: LabelKeyPair[] = [...keyValuePairs];
                              arrCp[i] = {
                                key: arrCp[i].key,
                                value: e.target.value as string,
                              };
                              dispatch(setEditPoolKeyValuePairs(arrCp));
                            }}
                            index={i}
                            placeholder={"Key"}
                          />
                        )}
                      </Grid>
                      <Grid item xs={5} className={classes.affinityLabelValue}>
                        {keyOptions.length > 0 && (
                          <SelectWrapper
                            onChange={(e: SelectChangeEvent<string>) => {
                              const arrCp: LabelKeyPair[] = [...keyValuePairs];
                              arrCp[i] = {
                                key: arrCp[i].key,
                                value: e.target.value as string,
                              };
                              dispatch(setEditPoolKeyValuePairs(arrCp));
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
                          <InputBoxWrapper
                            id={`nodeselector-value-${i.toString()}`}
                            label={""}
                            name={`nodeselector-${i.toString()}`}
                            value={kvp.value}
                            onChange={(e) => {
                              const arrCp: LabelKeyPair[] = [...keyValuePairs];
                              arrCp[i] = {
                                key: arrCp[i].key,
                                value: e.target.value as string,
                              };
                              dispatch(setEditPoolKeyValuePairs(arrCp));
                            }}
                            index={i}
                            placeholder={"value"}
                          />
                        )}
                      </Grid>
                      <Grid item xs={2} className={classes.rowActions}>
                        <div className={classes.overlayAction}>
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

                              dispatch(setEditPoolKeyValuePairs(arrCp));
                            }}
                          >
                            <AddIcon />
                          </IconButton>
                        </div>
                        {keyValuePairs.length > 1 && (
                          <div className={classes.overlayAction}>
                            <IconButton
                              size={"small"}
                              onClick={() => {
                                const arrCp = keyValuePairs.filter(
                                  (item, index) => index !== i,
                                );
                                dispatch(setEditPoolKeyValuePairs(arrCp));
                              }}
                            >
                              <RemoveIcon />
                            </IconButton>
                          </div>
                        )}
                      </Grid>
                    </Grid>
                  );
                })}
            </Grid>
          </Grid>
        </Fragment>
      )}
      <Grid item xs={12} className={classes.affinityConfigField}>
        <Grid item className={classes.affinityFieldLabel}>
          <h3>Tolerations</h3>
          <span className={classes.error}>
            {validationErrors["tolerations"]}
          </span>
          <Grid container>
            {tolerations &&
              tolerations.map((tol, i) => {
                return (
                  <Grid
                    item
                    xs={12}
                    className={classes.affinityRow}
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
                    <div className={classes.overlayAction}>
                      <IconButton
                        size={"small"}
                        onClick={() => {
                          dispatch(addNewEditPoolToleration());
                        }}
                        disabled={i !== tolerations.length - 1}
                      >
                        <AddIcon />
                      </IconButton>
                    </div>

                    <div className={classes.overlayAction}>
                      <IconButton
                        size={"small"}
                        onClick={() => {
                          dispatch(removeEditPoolToleration(i));
                        }}
                        disabled={tolerations.length <= 1}
                      >
                        <RemoveIcon />
                      </IconButton>
                    </div>
                  </Grid>
                );
              })}
          </Grid>
        </Grid>
      </Grid>
    </Paper>
  );
};

export default withStyles(styles)(Affinity);
