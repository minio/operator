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

//import {  ISecurityContext} from "../types";
import { Theme } from "@mui/material/styles";
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
import React, { Fragment, useState } from "react";
import { useSelector } from "react-redux";
import { AppState, useAppDispatch } from "../../../../store";
import api from "../../../../common/api";
import { ErrorResponseHandler } from "../../../../common/types";
import { useParams } from "react-router-dom";
import Grid from "@mui/material/Grid";
import InputBoxWrapper from "../../Common/FormComponents/InputBoxWrapper/InputBoxWrapper";
import { Button } from "mds";
import {
  setErrorSnackMessage,
  setSnackBarMessage,
} from "../../../../systemSlice";
import { IKeyValue, ITenantAuditLogs } from "../ListTenants/types";
import KeyPairEdit from "./KeyPairEdit";
import InputUnitMenu from "../../Common/FormComponents/InputUnitMenu/InputUnitMenu";
import SecurityContextSelector from "../securityContextSelector";
import { clearValidationError, imagePattern, numericPattern } from "../utils";
import {
  setCPURequest,
  setDiskCapacityGB,
  setFSGroup,
  setImage,
  setMemRequest,
  setRefreshLoggingInfo,
  setRunAsGroup,
  setRunAsNonRoot,
  setRunAsUser,
  setServiceAccountName,
} from "../TenantDetails/tenantAuditLogSlice";
import { setFSGroupChangePolicy } from "../tenantSecurityContextSlice";
import { fsGroupChangePolicyType } from "../types";

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

const TenantAuditLogging = ({
  classes,
  labels,
  annotations,
  nodeSelector,
}: ITenantAuditLogs) => {
  const dispatch = useAppDispatch();
  const { tenantName, tenantNamespace } = useParams();
  const auditLoggingEnabled = useSelector(
    (state: AppState) => state.editTenantLogging.auditLoggingEnabled
  );
  const image = useSelector((state: AppState) => state.editTenantLogging.image);
  const diskCapacityGB = useSelector(
    (state: AppState) => state.editTenantLogging.diskCapacityGB
  );
  const cpuRequest = useSelector(
    (state: AppState) => state.editTenantLogging.cpuRequest
  );
  const memRequest = useSelector(
    (state: AppState) => state.editTenantLogging.memRequest
  );
  const serviceAccountName = useSelector(
    (state: AppState) => state.editTenantLogging.serviceAccountName
  );
  const runAsGroup = useSelector(
    (state: AppState) => state.editTenantLogging.securityContext.runAsGroup
  );
  const runAsUser = useSelector(
    (state: AppState) => state.editTenantLogging.securityContext.runAsUser
  );
  const fsGroup = useSelector(
    (state: AppState) => state.editTenantLogging.securityContext.fsGroup
  );
  const runAsNonRoot = useSelector(
    (state: AppState) => state.editTenantLogging.securityContext.runAsNonRoot
  );
  const fsGroupChangePolicy = useSelector(
    (state: AppState) => state.editTenantSecurityContext.fsGroupChangePolicy
  );

  const [validationErrors, setValidationErrors] = useState<any>({});
  const [loading, setLoading] = useState<boolean>(false);

  const [logLabels, setLabels] = useState<IKeyValue[]>(
    labels != null && labels.length > 0 ? labels : [{ key: "", value: "" }]
  );
  const [logAnnotations, setAnnotations] = useState<IKeyValue[]>(
    annotations != null && annotations.length > 0
      ? annotations
      : [{ key: "", value: "" }]
  );
  const [logNodeSelector, setNodeSelector] = useState<IKeyValue[]>(
    nodeSelector != null && nodeSelector.length > 0
      ? nodeSelector
      : [{ key: "", value: "" }]
  );

  const [labelsError, setLabelsError] = useState<any>({});
  const [annotationsError, setAnnotationsError] = useState<any>({});
  const [nodeSelectorError, setNodeSelectorError] = useState<any>({});

  const cleanValidation = (fieldName: string) => {
    setValidationErrors(clearValidationError(validationErrors, fieldName));
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

  const submitLoggingInfo = () => {
    if (checkValid()) {
      setLoading(true);
      const securityContext = {
        runAsGroup: runAsGroup != null ? runAsGroup : "",
        runAsUser: runAsUser != null ? runAsUser : "",
        fsGroup: fsGroup != null ? fsGroup : "",
        runAsNonRoot: runAsNonRoot != null ? runAsNonRoot : true,
      };

      api
        .invoke(
          "PUT",
          `/api/v1/namespaces/${tenantNamespace}/tenants/${tenantName}/log`,
          {
            labels: trim(logLabels),
            annotations: trim(logAnnotations),
            nodeSelector: trim(logNodeSelector),
            image: image,
            diskCapacityGB: diskCapacityGB.toString(),
            serviceAccountName: serviceAccountName,
            logCPURequest: cpuRequest,
            logMemRequest: memRequest,
            securityContext: securityContext,
          }
        )
        .then(() => {
          setRefreshLoggingInfo(true);
          dispatch(setSnackBarMessage(`Audit Log configuration updated.`));
          setLoading(false);
        })
        .catch((err: ErrorResponseHandler) => {
          setErrorSnackMessage(err);
          setLoading(false);
        });
    }
  };

  return (
    <Fragment>
      {auditLoggingEnabled && (
        <Fragment>
          <Grid item xs={12} paddingBottom={2}>
            <InputBoxWrapper
              id={`image`}
              label={"Image"}
              placeholder={"minio/operator:v4.4.22"}
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
              id={`diskCapacityGB`}
              label={"Disk Capacity"}
              placeholder={"Disk Capacity"}
              name={`diskCapacityGB`}
              value={!isNaN(diskCapacityGB) ? diskCapacityGB.toString() : "0"}
              onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                if (event.target.validity.valid) {
                  dispatch(setDiskCapacityGB(parseInt(event.target.value)));
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

          <Grid item xs={12} className={classes.formFieldRow}>
            <span className={classes.inputLabel}>Labels</span>
            <KeyPairEdit
              newValues={logLabels}
              setNewValues={setLabels}
              paramName={"Labels"}
              error={labelsError}
              setError={setLabelsError}
            />
          </Grid>

          <Grid item xs={12} className={classes.formFieldRow}>
            <span className={classes.inputLabel}>Annotations</span>
            <KeyPairEdit
              newValues={logAnnotations}
              setNewValues={setAnnotations}
              paramName={"Annotations"}
              error={annotationsError}
              setError={setAnnotationsError}
            />
          </Grid>

          <Grid item xs={12} className={classes.formFieldRow}>
            <span className={classes.inputLabel}>Node Selector</span>
            <KeyPairEdit
              newValues={logNodeSelector}
              setNewValues={setNodeSelector}
              paramName={"Node Selector"}
              error={nodeSelectorError}
              setError={setNodeSelectorError}
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
              disabled={loading || !checkValid()}
              onClick={() => submitLoggingInfo()}
              label={"Save"}
            />
          </Grid>
        </Fragment>
      )}
    </Fragment>
  );
};

export default withStyles(styles)(TenantAuditLogging);
