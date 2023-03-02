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
import { Theme } from "@mui/material/styles";
import { Button } from "mds";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import {
  containerForHeader,
  createTenantCommon,
  formFieldStyles,
  modalBasic,
  spacingUtils,
  tenantDetailsStyles,
  wizardCommon,
} from "../../Common/FormComponents/common/styleLibrary";
import { useSelector } from "react-redux";
import { AppState, useAppDispatch } from "../../../../store";
import api from "../../../../common/api";
import { ErrorResponseHandler } from "../../../../common/types";
import { useParams } from "react-router-dom";
import FormSwitchWrapper from "../../Common/FormComponents/FormSwitchWrapper/FormSwitchWrapper";
import Grid from "@mui/material/Grid";
import InputBoxWrapper from "../../Common/FormComponents/InputBoxWrapper/InputBoxWrapper";
import { DialogContentText } from "@mui/material";
import ConfirmDialog from "../../Common/ModalWrapper/ConfirmDialog";
import {
  setErrorSnackMessage,
  setSnackBarMessage,
} from "../../../../systemSlice";
import { IKeyValue, ITenantMonitoringStruct } from "../ListTenants/types";
import KeyPairEdit from "./KeyPairEdit";
import InputUnitMenu from "../../Common/FormComponents/InputUnitMenu/InputUnitMenu";
import {
  setCPURequest,
  setDiskCapacityGB,
  setFSGroup,
  setImage,
  setInitImage,
  setMemRequest,
  setPrometheusEnabled,
  setRunAsGroup,
  setRunAsNonRoot,
  setRunAsUser,
  setServiceAccountName,
  setSidecarImage,
  setStorageClassName,
} from "../TenantDetails/tenantMonitoringSlice";
import { clearValidationError, imagePattern, numericPattern } from "../utils";
import SecurityContextSelector from "../securityContextSelector";
import { setFSGroupChangePolicy } from "../tenantSecurityContextSlice";
import { fsGroupChangePolicyType } from "../types";
import FormHr from "../../Common/FormHr";

interface ITenantMonitoring {
  classes: any;
}

const styles = (theme: Theme) =>
  createStyles({
    ...tenantDetailsStyles,
    ...spacingUtils,
    ...containerForHeader,
    ...createTenantCommon,
    ...formFieldStyles,
    ...modalBasic,
    ...wizardCommon,
  });

const TenantMonitoring = ({ classes }: ITenantMonitoring) => {
  const dispatch = useAppDispatch();
  const { tenantName, tenantNamespace } = useParams();
  const prometheusEnabled = useSelector(
    (state: AppState) => state.editTenantMonitoring.prometheusEnabled
  );
  const image = useSelector(
    (state: AppState) => state.editTenantMonitoring.image
  );
  const sidecarImage = useSelector(
    (state: AppState) => state.editTenantMonitoring.sidecarImage
  );
  const initImage = useSelector(
    (state: AppState) => state.editTenantMonitoring.initImage
  );
  const diskCapacityGB = useSelector(
    (state: AppState) => state.editTenantMonitoring.diskCapacityGB
  );
  const cpuRequest = useSelector(
    (state: AppState) => state.editTenantMonitoring.monitoringCPURequest
  );
  const memRequest = useSelector(
    (state: AppState) => state.editTenantMonitoring.monitoringMemRequest
  );
  const serviceAccountName = useSelector(
    (state: AppState) => state.editTenantMonitoring.serviceAccountName
  );
  const storageClassName = useSelector(
    (state: AppState) => state.editTenantMonitoring.storageClassName
  );
  const [validationErrors, setValidationErrors] = useState<any>({});
  const [toggleConfirmOpen, setToggleConfirmOpen] = useState<boolean>(false);

  const [labels, setLabels] = useState<IKeyValue[]>([{ key: "", value: "" }]);
  const [annotations, setAnnotations] = useState<IKeyValue[]>([
    { key: "", value: "" },
  ]);
  const [nodeSelector, setNodeSelector] = useState<IKeyValue[]>([
    { key: "", value: "" },
  ]);

  const [refreshMonitoringInfo, setRefreshMonitoringInfo] =
    useState<boolean>(true);
  const [labelsError, setLabelsError] = useState<any>({});
  const [annotationsError, setAnnotationsError] = useState<any>({});
  const [nodeSelectorError, setNodeSelectorError] = useState<any>({});

  const runAsGroup = useSelector(
    (state: AppState) => state.editTenantMonitoring.runAsGroup
  );
  const runAsUser = useSelector(
    (state: AppState) => state.editTenantMonitoring.runAsUser
  );
  const fsGroup = useSelector(
    (state: AppState) => state.editTenantMonitoring.fsGroup
  );
  const runAsNonRoot = useSelector(
    (state: AppState) => state.editTenantMonitoring.runAsNonRoot
  );
  const fsGroupChangePolicy = useSelector(
    (state: AppState) => state.editTenantSecurityContext.fsGroupChangePolicy
  );

  const cleanValidation = (fieldName: string) => {
    setValidationErrors(clearValidationError(validationErrors, fieldName));
  };

  const setMonitoringInfo = (res: ITenantMonitoringStruct) => {
    dispatch(setImage(res.image));
    dispatch(setSidecarImage(res.sidecarImage));
    dispatch(setInitImage(res.initImage));
    dispatch(setStorageClassName(res.storageClassName));
    dispatch(setDiskCapacityGB(res.diskCapacityGB));
    dispatch(setServiceAccountName(res.serviceAccountName));
    dispatch(setCPURequest(res.monitoringCPURequest));
    if (res.monitoringMemRequest) {
      dispatch(
        setMemRequest(
          Math.floor(
            parseInt(res.monitoringMemRequest, 10) / 1000000000
          ).toString()
        )
      );
    } else {
      dispatch(setMemRequest("0"));
    }
    res.labels != null
      ? setLabels(res.labels)
      : setLabels([{ key: "", value: "" }]);
    res.annotations != null
      ? setAnnotations(res.annotations)
      : setAnnotations([{ key: "", value: "" }]);
    res.nodeSelector != null
      ? setNodeSelector(res.nodeSelector)
      : setNodeSelector([{ key: "", value: "" }]);
    dispatch(setRunAsGroup(res.securityContext.runAsGroup));
    dispatch(setRunAsUser(res.securityContext.runAsUser));
    dispatch(setRunAsNonRoot(res.securityContext.runAsNonRoot));
    dispatch(setFSGroup(res.securityContext.fsGroup));
  };

  const trim = (x: IKeyValue[]): IKeyValue[] => {
    let retval: IKeyValue[] = [];
    for (let i = 0; i < x.length; i++) {
      if (x[i].key !== "") {
        retval.push(x[i]);
      }
    }
    return retval;
  };

  const checkValid = (): boolean => {
    if (
      Object.keys(validationErrors).length !== 0 ||
      Object.keys(labelsError).length !== 0 ||
      Object.keys(annotationsError).length !== 0 ||
      Object.keys(nodeSelectorError).length !== 0
    ) {
      let err: ErrorResponseHandler = {
        errorMessage: "Invalid entry",
        detailedError: "",
      };
      dispatch(setErrorSnackMessage(err));
      return false;
    } else {
      return true;
    }
  };

  useEffect(() => {
    if (refreshMonitoringInfo) {
      api
        .invoke(
          "GET",
          `/api/v1/namespaces/${tenantNamespace || ""}/tenants/${
            tenantName || ""
          }/monitoring`
        )
        .then((res: ITenantMonitoringStruct) => {
          dispatch(setPrometheusEnabled(res.prometheusEnabled));
          setMonitoringInfo(res);
          setRefreshMonitoringInfo(false);
        })
        .catch((err: ErrorResponseHandler) => {
          dispatch(setErrorSnackMessage(err));
          setRefreshMonitoringInfo(false);
        });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [refreshMonitoringInfo]);

  const submitMonitoringInfo = () => {
    if (checkValid()) {
      const securityContext = {
        runAsGroup: runAsGroup != null ? runAsGroup : "0",
        runAsUser: runAsUser != null ? runAsUser : "0",
        fsGroup: fsGroup != null ? fsGroup : "0",
        runAsNonRoot: runAsNonRoot != null ? runAsNonRoot : true,
      };
      api
        .invoke(
          "PUT",
          `/api/v1/namespaces/${tenantNamespace}/tenants/${tenantName}/monitoring`,
          {
            labels: trim(labels),
            annotations: trim(annotations),
            nodeSelector: trim(nodeSelector),
            image: image,
            sidecarImage: sidecarImage,
            initImage: initImage,
            diskCapacityGB: diskCapacityGB,
            serviceAccountName: serviceAccountName,
            storageClassName: storageClassName,
            monitoringCPURequest: cpuRequest,
            monitoringMemRequest: memRequest + "Gi",
            securityContext: securityContext,
          }
        )
        .then(() => {
          setRefreshMonitoringInfo(true);
          dispatch(setSnackBarMessage(`Prometheus configuration updated.`));
        })
        .catch((err: ErrorResponseHandler) => {
          setErrorSnackMessage(err);
        });
    }
  };

  const togglePrometheus = () => {
    const configInfo = {
      prometheusEnabled: prometheusEnabled,
      toggle: true,
    };
    api
      .invoke(
        "PUT",
        `/api/v1/namespaces/${tenantNamespace}/tenants/${tenantName}/monitoring`,
        configInfo
      )
      .then(() => {
        dispatch(setPrometheusEnabled(!prometheusEnabled));
        setRefreshMonitoringInfo(true);
        setToggleConfirmOpen(false);
        setRefreshMonitoringInfo(true);
      })
      .catch((err: ErrorResponseHandler) => {
        dispatch(setErrorSnackMessage(err));
      });
  };

  return (
    <Fragment>
      {toggleConfirmOpen && (
        <ConfirmDialog
          isOpen={toggleConfirmOpen}
          title={
            !prometheusEnabled
              ? "Enable Prometheus monitoring for this tenant?"
              : "Disable Prometheus monitoring for this tenant?"
          }
          confirmText={!prometheusEnabled ? "Enable" : "Disable"}
          cancelText="Cancel"
          onClose={() => setToggleConfirmOpen(false)}
          onConfirm={togglePrometheus}
          confirmationContent={
            <DialogContentText>
              {!prometheusEnabled
                ? "A small Prometheus server will be started per the configuration provided, which will collect the Prometheus metrics for your tenant."
                : " Current configuration will be lost, and defaults reset if reenabled."}
            </DialogContentText>
          }
        />
      )}
      <Grid container spacing={1}>
        <Grid item xs>
          <h1 className={classes.sectionTitle}>Prometheus Monitoring </h1>
        </Grid>
        <Grid item xs={7} justifyContent={"end"} textAlign={"right"}>
          <FormSwitchWrapper
            label={""}
            indicatorLabels={["Enabled", "Disabled"]}
            checked={prometheusEnabled}
            value={"tenant_monitoring"}
            id="tenant-monitoring"
            name="tenant-monitoring"
            onChange={() => {
              setToggleConfirmOpen(true);
            }}
            description=""
          />
        </Grid>
        <Grid item xs={12}>
          <FormHr />
        </Grid>
      </Grid>

      {prometheusEnabled && (
        <Fragment>
          <Grid item xs={12} paddingBottom={2}>
            <InputBoxWrapper
              id={`prometheus_image`}
              label={"Image"}
              placeholder={"quay.io/prometheus/prometheus:latest"}
              name={`image`}
              value={image}
              onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                if (event.target.validity.valid) {
                  dispatch(setImage(event.target.value));
                }
                cleanValidation(`image`);
              }}
              key={`image`}
              pattern={imagePattern}
              error={validationErrors[`image`] || ""}
            />
          </Grid>
          <Grid item xs={12} paddingBottom={2}>
            <InputBoxWrapper
              id={`sidecarImage`}
              label={"Sidecar Image"}
              placeholder={"library/alpine:latest"}
              name={`sidecarImage`}
              value={sidecarImage}
              onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                if (event.target.validity.valid) {
                  dispatch(setSidecarImage(event.target.value));
                }
                cleanValidation(`sidecarImage`);
              }}
              key={`sidecarImage`}
              pattern={imagePattern}
              error={validationErrors[`sidecarImage`] || ""}
            />
          </Grid>
          <Grid item xs={12} paddingBottom={2}>
            <InputBoxWrapper
              id={`initImage`}
              label={"Init Image"}
              placeholder={"library/busybox:1.33.1"}
              name={`initImage`}
              value={initImage}
              onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                if (event.target.validity.valid) {
                  dispatch(setInitImage(event.target.value));
                }
                cleanValidation(`initImage`);
              }}
              key={`initImage`}
              pattern={imagePattern}
              error={validationErrors[`initImage`] || ""}
            />
          </Grid>
          <Grid item xs={12} paddingBottom={2}>
            <InputBoxWrapper
              id={`diskCapacityGB`}
              label={"Disk Capacity"}
              placeholder={"Disk Capacity"}
              name={`diskCapacityGB`}
              value={diskCapacityGB}
              onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                if (event.target.validity.valid) {
                  dispatch(setDiskCapacityGB(event.target.value));
                }
                cleanValidation(`diskCapacityGB`);
              }}
              key={`diskCapacityGB`}
              pattern={numericPattern}
              error={validationErrors[`diskCapacityGB`] || ""}
              overlayObject={
                <InputUnitMenu
                  id={"size-unit"}
                  onUnitChange={() => {}}
                  unitSelected={"Gi"}
                  unitsList={[{ label: "Gi", value: "Gi" }]}
                  disabled={true}
                />
              }
            />
          </Grid>
          <Grid item xs={12} paddingBottom={2}>
            <InputBoxWrapper
              id={`cpuRequest`}
              label={"CPU Request"}
              placeholder={"CPU Request"}
              name={`cpuRequest`}
              value={cpuRequest}
              pattern={numericPattern}
              onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                if (event.target.validity.valid) {
                  dispatch(setCPURequest(event.target.value));
                }
                cleanValidation(`cpuRequest`);
              }}
              key={`cpuRequest`}
              error={validationErrors[`cpuRequest`] || ""}
            />
          </Grid>
          <Grid item xs={12} paddingBottom={2}>
            <InputBoxWrapper
              id={`memRequest`}
              label={"Memory Request"}
              placeholder={"Memory request"}
              name={`memRequest`}
              value={memRequest}
              onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                if (event.target.validity.valid) {
                  dispatch(setMemRequest(event.target.value));
                }
                cleanValidation(`memRequest`);
              }}
              pattern={numericPattern}
              key={`memRequest`}
              error={validationErrors[`memRequest`] || ""}
              overlayObject={
                <InputUnitMenu
                  id={"size-unit"}
                  onUnitChange={() => {}}
                  unitSelected={"Gi"}
                  unitsList={[{ label: "Gi", value: "Gi" }]}
                  disabled={true}
                />
              }
            />
          </Grid>
          <Grid item xs={12} paddingBottom={2}>
            <InputBoxWrapper
              id={`serviceAccountName`}
              label={"Service Account"}
              placeholder={"Service Account Name"}
              name={`serviceAccountName`}
              value={serviceAccountName}
              onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                if (event.target.validity.valid) {
                  dispatch(setServiceAccountName(event.target.value));
                }
                cleanValidation(`serviceAccountName`);
              }}
              key={`serviceAccountName`}
              pattern={"^[a-zA-Z0-9-.]{1,253}$"}
              error={validationErrors[`serviceAccountName`] || ""}
            />
          </Grid>
          <Grid item xs={12} paddingBottom={2}>
            <InputBoxWrapper
              id={`storageClassName`}
              label={"Storage Class"}
              placeholder={"Storage Class Name"}
              name={`storageClassName`}
              value={storageClassName}
              onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                if (event.target.validity.valid) {
                  dispatch(setStorageClassName(event.target.value));
                }
                cleanValidation(`storageClassName`);
              }}
              key={`storageClassName`}
              pattern={"^[a-zA-Z0-9-.]{1,253}$"}
              error={validationErrors[`storageClassName`] || ""}
            />
          </Grid>
          {labels !== null && (
            <Grid item xs={12} className={classes.formFieldRow}>
              <span className={classes.inputLabel}>Labels</span>
              <KeyPairEdit
                newValues={labels}
                setNewValues={setLabels}
                paramName={"Labels"}
                error={labelsError}
                setError={setLabelsError}
              />
            </Grid>
          )}

          {annotations !== null && (
            <Grid item xs={12} className={classes.formFieldRow}>
              <span className={classes.inputLabel}>Annotations</span>
              <KeyPairEdit
                newValues={annotations}
                setNewValues={setAnnotations}
                paramName={"Annotations"}
                error={annotationsError}
                setError={setAnnotationsError}
              />
            </Grid>
          )}
          {nodeSelector !== null && (
            <Grid item xs={12} className={classes.formFieldRow}>
              <span className={classes.inputLabel}>Node Selector</span>
              <KeyPairEdit
                newValues={nodeSelector}
                setNewValues={setNodeSelector}
                paramName={"Node Selector"}
                error={nodeSelectorError}
                setError={setNodeSelectorError}
              />
            </Grid>
          )}
          <Grid item xs={12} className={classes.formFieldRow}>
            <SecurityContextSelector
              classes={classes}
              runAsGroup={runAsGroup}
              runAsUser={runAsUser}
              fsGroup={fsGroup}
              runAsNonRoot={runAsNonRoot}
              fsGroupChangePolicy={fsGroupChangePolicy}
              setFSGroup={(value: string) => dispatch(setFSGroup(value))}
              setRunAsUser={(value: string) => dispatch(setRunAsUser(value))}
              setRunAsGroup={(value: string) => dispatch(setRunAsGroup(value))}
              setRunAsNonRoot={(value: boolean) =>
                dispatch(setRunAsNonRoot(value))
              }
              setFSGroupChangePolicy={(value: fsGroupChangePolicyType) =>
                dispatch(setFSGroupChangePolicy(value))
              }
            />
          </Grid>
          <Grid
            item
            xs={12}
            sx={{ display: "flex", justifyContent: "flex-end" }}
          >
            <Button
              type="submit"
              id={"submit_button"}
              variant="callAction"
              disabled={!checkValid()}
              onClick={() => submitMonitoringInfo()}
              label={"Save"}
            />
          </Grid>
        </Fragment>
      )}
    </Fragment>
  );
};

export default withStyles(styles)(TenantMonitoring);
