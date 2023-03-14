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
import {
  setDBCPURequest,
  setDBFSGroup,
  setDBFSGroupChangePolicy,
  setDBImage,
  setDBInitImage,
  setDBMemRequest,
  setDBRunAsGroup,
  setDBRunAsNonRoot,
  setDBRunAsUser,
  setRefreshLoggingInfo,
} from "./tenantAuditLogSlice";

import SecurityContextSelector from "../securityContextSelector";

import { clearValidationError, imagePattern, numericPattern } from "../utils";
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

const LoggingDBDetails = ({
  classes,
  labels,
  annotations,
  nodeSelector,
}: ITenantAuditLogs) => {
  const dispatch = useAppDispatch();
  const { tenantName, tenantNamespace } = useParams();
  const dbImage = useSelector(
    (state: AppState) => state.editTenantLogging.dbImage
  );
  const dbInitImage = useSelector(
    (state: AppState) => state.editTenantLogging.dbInitImage
  );
  const dbCpuRequest = useSelector(
    (state: AppState) => state.editTenantLogging.dbCPURequest
  );
  const dbMemRequest = useSelector(
    (state: AppState) => state.editTenantLogging.dbMemRequest
  );
  const dbServiceAccountName = useSelector(
    (state: AppState) => state.editTenantLogging.dbServiceAccountName
  );

  const dbRunAsGroup = useSelector(
    (state: AppState) => state.editTenantLogging.dbSecurityContext.runAsGroup
  );
  const dbRunAsUser = useSelector(
    (state: AppState) => state.editTenantLogging.dbSecurityContext.runAsUser
  );
  const dbFSGroup = useSelector(
    (state: AppState) => state.editTenantLogging.dbSecurityContext.fsGroup
  );
  const dbFSGroupChangePolicy = useSelector(
    (state: AppState) =>
      state.editTenantLogging.dbSecurityContext.fsGroupChangePolicy
  );
  const dbRunAsNonRoot = useSelector(
    (state: AppState) => state.editTenantLogging.dbSecurityContext.runAsNonRoot
  );
  const [validationErrors, setValidationErrors] = useState<any>({});

  const [dbLabels, setDBLabels] = useState<IKeyValue[]>(
    labels != null && labels.length > 0 ? labels : [{ key: "", value: "" }]
  );
  const [dbAnnotations, setDBAnnotations] = useState<IKeyValue[]>(
    annotations != null && annotations.length > 0
      ? annotations
      : [{ key: "", value: "" }]
  );
  const [dbNodeSelector, setDBNodeSelector] = useState<IKeyValue[]>(
    nodeSelector != null && nodeSelector.length > 0
      ? nodeSelector
      : [{ key: "", value: "" }]
  );

  const [dbLabelsError, setDBLabelsError] = useState<any>({});
  const [dbAnnotationsError, setDBAnnotationsError] = useState<any>({});
  const [dbNodeSelectorError, setDBNodeSelectorError] = useState<any>({});

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
      Object.keys(dbNodeSelectorError).length !== 0 ||
      Object.keys(dbAnnotationsError).length !== 0 ||
      Object.keys(dbLabelsError).length !== 0
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
      const dbSecurityContext = {
        runAsGroup: dbRunAsGroup != null ? dbRunAsGroup : "",
        runAsUser: dbRunAsUser != null ? dbRunAsUser : "",
        fsGroup: dbFSGroup != null ? dbFSGroup : "",
        runAsNonRoot: dbRunAsNonRoot != null ? dbRunAsNonRoot : true,
        fsGroupChangePolicy:
          dbFSGroupChangePolicy != null ? dbFSGroupChangePolicy : "Always",
      };
      api
        .invoke(
          "PUT",
          `/api/v1/namespaces/${tenantNamespace}/tenants/${tenantName}/log`,
          {
            dbLabels: trim(dbLabels),
            dbAnnotations: trim(dbAnnotations),
            dbNodeSelector: trim(dbNodeSelector),
            dbImage: dbImage,
            dbInitImage: dbInitImage,
            dbServiceAccountName: dbServiceAccountName,
            logDBCPURequest: dbCpuRequest,
            logDBMemRequest: dbMemRequest,
            dbSecurityContext: dbSecurityContext,
          }
        )
        .then(() => {
          setRefreshLoggingInfo(true);
          dispatch(setSnackBarMessage(`Audit Log DB configuration updated.`));
        })
        .catch((err: ErrorResponseHandler) => {
          setErrorSnackMessage(err);
        });
    }
  };

  return (
    <Fragment>
      <Fragment>
        <Grid item xs={12} paddingBottom={2}>
          <InputBoxWrapper
            id={`dbImage`}
            label={"DB Postgres Image"}
            placeholder={"library/postgres:13"}
            name={`dbImage`}
            value={dbImage}
            onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
              if (event.target.validity.valid) {
                dispatch(setDBImage(event.target.value));
              }
              cleanValidation(`dbImage`);
            }}
            key={`dbImage`}
            pattern={imagePattern}
            error={validationErrors[`dbImage`] || ""}
          />
        </Grid>
        <Grid item xs={12} paddingBottom={2}>
          <InputBoxWrapper
            id={`dbInitImage`}
            label={"DB Init Image"}
            placeholder={"library/busybox:1.33.1"}
            name={`dbInitImage`}
            value={dbInitImage}
            onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
              if (event.target.validity.valid) {
                dispatch(setDBInitImage(event.target.value));
              }
              cleanValidation(`dbInitImage`);
            }}
            key={`dbInitImage`}
            pattern={imagePattern}
            error={validationErrors[`dbInitImage`] || ""}
          />
        </Grid>
        <Grid item xs={12} paddingBottom={2}>
          <InputBoxWrapper
            id={`dbCPURequest`}
            label={"DB CPU Request"}
            placeholder={"DB CPU Request"}
            name={`dbCPURequest`}
            value={dbCpuRequest}
            pattern={numericPattern}
            onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
              if (event.target.validity.valid) {
                dispatch(setDBCPURequest(event.target.value));
              }
              cleanValidation(`dbCPURequest`);
            }}
            key={`dbCPURequest`}
            error={validationErrors[`dbCPURequest`] || ""}
          />
        </Grid>
        <Grid item xs={12} paddingBottom={2}>
          <InputBoxWrapper
            id={`dbMemRequest`}
            label={"DB Memory Request"}
            placeholder={"DB Memory request"}
            name={`dbMemRequest`}
            value={dbMemRequest}
            onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
              if (event.target.validity.valid) {
                dispatch(setDBMemRequest(event.target.value));
              }
              cleanValidation(`dbMemRequest`);
            }}
            pattern={numericPattern}
            key={`dbMemRequest`}
            error={validationErrors[`dbMemRequest`] || ""}
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

        <Grid item xs={12} className={classes.formFieldRow}>
          <span className={classes.inputLabel}>DB Labels</span>
          <KeyPairEdit
            newValues={dbLabels}
            setNewValues={setDBLabels}
            paramName={"dbLabels"}
            error={dbLabelsError}
            setError={setDBLabelsError}
          />
        </Grid>
        <Grid item xs={12} className={classes.formFieldRow}>
          <span className={classes.inputLabel}>DB Annotations</span>
          <KeyPairEdit
            newValues={dbAnnotations}
            setNewValues={setDBAnnotations}
            paramName={"dbAnnotations"}
            error={dbAnnotationsError}
            setError={setDBAnnotationsError}
          />
        </Grid>

        <Grid item xs={12} className={classes.formFieldRow}>
          <span className={classes.inputLabel}>DB Node Selector</span>
          <KeyPairEdit
            newValues={dbNodeSelector}
            setNewValues={setDBNodeSelector}
            paramName={"DB Node Selector"}
            error={dbNodeSelectorError}
            setError={setDBNodeSelectorError}
          />
        </Grid>

        <Grid item xs={12} className={classes.formFieldRow}>
          <SecurityContextSelector
            classes={classes}
            runAsGroup={dbRunAsGroup}
            runAsUser={dbRunAsUser}
            fsGroup={dbFSGroup!}
            fsGroupChangePolicy={
              dbFSGroupChangePolicy as fsGroupChangePolicyType
            }
            runAsNonRoot={dbRunAsNonRoot}
            setFSGroup={(value: string) => dispatch(setDBFSGroup(value))}
            setRunAsUser={(value: string) => dispatch(setDBRunAsUser(value))}
            setRunAsGroup={(value: string) => dispatch(setDBRunAsGroup(value))}
            setRunAsNonRoot={(value: boolean) =>
              dispatch(setDBRunAsNonRoot(value))
            }
            setFSGroupChangePolicy={(value: fsGroupChangePolicyType) =>
              dispatch(setDBFSGroupChangePolicy(value))
            }
          />
        </Grid>
        <Grid item xs={12} sx={{ display: "flex", justifyContent: "flex-end" }}>
          <Button
            type="submit"
            id={"submit_button"}
            variant="callAction"
            disabled={!checkValid()}
            onClick={() => submitLoggingInfo()}
            label={"Save"}
          />
        </Grid>
      </Fragment>
    </Fragment>
  );
};

export default withStyles(styles)(LoggingDBDetails);
