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

import React, { Fragment, useEffect, useState } from "react";
import { Theme } from "@mui/material/styles";
import { useParams } from "react-router-dom";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { containerForHeader } from "../../Common/FormComponents/common/styleLibrary";
import Grid from "@mui/material/Grid";
import Tabs from "@mui/material/Tabs";
import Tab from "@mui/material/Tab";
import { DialogContentText } from "@mui/material";
import ConfirmDialog from "../../Common/ModalWrapper/ConfirmDialog";
import api from "../../../../common/api";
import { useSelector } from "react-redux";
import { AppState, useAppDispatch } from "../../../../store";
import { ErrorResponseHandler } from "../../../../common/types";
import { setErrorSnackMessage } from "../../../../systemSlice";
import FormSwitchWrapper from "../../Common/FormComponents/FormSwitchWrapper/FormSwitchWrapper";
import { IKeyValue, ITenantLogsStruct } from "../ListTenants/types";

import LoggingDetails from "./LoggingDetails";
import LoggingDBDetails from "./LoggingDBDetails";
import {
  resetAuditLogForm,
  setAuditLoggingEnabled,
  setCPURequest,
  setDBCPURequest,
  setDBFSGroup,
  setDBImage,
  setDBInitImage,
  setDBMemRequest,
  setDBRunAsGroup,
  setDBRunAsNonRoot,
  setDBRunAsUser,
  setDBServiceAccountName,
  setDiskCapacityGB,
  setFSGroup,
  setImage,
  setMemRequest,
  setRunAsGroup,
  setRunAsNonRoot,
  setRunAsUser,
  setServiceAccountName,
} from "../TenantDetails/tenantAuditLogSlice";

import { HelpBox, WarnIcon } from "mds";
import FormHr from "../../Common/FormHr";

interface ILoggingScreenProps {
  classes: any;
}

const styles = (theme: Theme) =>
  createStyles({
    ...containerForHeader,
  });

const LoggingScreen = ({ classes }: ILoggingScreenProps) => {
  const deprecated = true; // Use a flag to hide UI for the moment, all related code will be removed once deprecation actually happens
  const { tenantNamespace, tenantName } = useParams();
  const [curTab, setCurTab] = useState<number>(0);
  const [loading, setLoading] = useState<boolean>(true);
  const [toggleConfirmOpen, setToggleConfirmOpen] = useState<boolean>(false);
  const [refreshLoggingInfo, setRefreshLoggingInfo] = useState<boolean>(true);
  const [dbLabels, setDBLabels] = useState<IKeyValue[]>([
    { key: "", value: "" },
  ]);
  const [dbAnnotations, setDBAnnotations] = useState<IKeyValue[]>([
    { key: "", value: "" },
  ]);
  const [dbNodeSelector, setDBNodeSelector] = useState<IKeyValue[]>([
    { key: "", value: "" },
  ]);
  const [labels, setLabels] = useState<IKeyValue[]>([{ key: "", value: "" }]);
  const [annotations, setAnnotations] = useState<IKeyValue[]>([
    { key: "", value: "" },
  ]);
  const [nodeSelector, setNodeSelector] = useState<IKeyValue[]>([
    { key: "", value: "" },
  ]);
  const dispatch = useAppDispatch();
  const auditLoggingEnabled = useSelector(
    (state: AppState) => state.editTenantLogging.auditLoggingEnabled
  );

  function a11yProps(index: any) {
    return {
      id: `simple-tab-${index}`,
      "aria-controls": `simple-tabpanel-${index}`,
    };
  }

  const setLoggingInfo = (res: ITenantLogsStruct) => {
    if (res !== null) {
      dispatch(setAuditLoggingEnabled(res !== null && !res.disabled));
      res.dbServiceAccountName != null &&
        dispatch(setDBServiceAccountName(res.dbServiceAccountName));
      res.dbImage != null && dispatch(setDBImage(res.dbImage));
      res.dbInitImage != null && dispatch(setDBInitImage(res.dbInitImage));
      res.logDBCPURequest != null &&
        dispatch(setDBCPURequest(res.logDBCPURequest));
      if (res.logDBMemRequest) {
        dispatch(
          setDBMemRequest(
            Math.floor(parseInt(res.logDBMemRequest, 10)).toString()
          )
        );
      } else {
        dispatch(setDBMemRequest("0"));
      }
      if (res.dbSecurityContext) {
        dispatch(setDBRunAsGroup(res.dbSecurityContext.runAsGroup));
        dispatch(setDBRunAsUser(res.dbSecurityContext.runAsUser));
        dispatch(setDBFSGroup(res.dbSecurityContext.fsGroup));
        dispatch(setDBRunAsNonRoot(res.dbSecurityContext.runAsNonRoot));
      }
      res.image != null && dispatch(setImage(res.image));
      res.serviceAccountName != null &&
        dispatch(setServiceAccountName(res.serviceAccountName));
      res.logCPURequest != null && dispatch(setCPURequest(res.logCPURequest));
      if (res.logMemRequest) {
        dispatch(
          setMemRequest(Math.floor(parseInt(res.logMemRequest, 10)).toString())
        );
      } else {
        dispatch(setMemRequest("0"));
      }
      if (res.securityContext) {
        dispatch(setRunAsGroup(res.securityContext.runAsGroup));
        dispatch(setRunAsUser(res.securityContext.runAsUser));
        dispatch(setFSGroup(res.securityContext.fsGroup));
        dispatch(setRunAsNonRoot(res.securityContext.runAsNonRoot));
      }

      res.diskCapacityGB != null &&
        dispatch(setDiskCapacityGB(res.diskCapacityGB));
      res.labels != null
        ? setLabels(res.labels)
        : setLabels([{ key: "", value: "" }]);
      res.annotations != null
        ? setAnnotations(res.annotations)
        : setAnnotations([{ key: "", value: "" }]);
      res.nodeSelector != null
        ? setNodeSelector(res.nodeSelector)
        : setNodeSelector([{ key: "", value: "" }]);
      res.dbLabels != null
        ? setDBLabels(res.dbLabels)
        : setDBLabels([{ key: "", value: "" }]);
      res.dbAnnotations != null
        ? setDBAnnotations(res.dbAnnotations)
        : setDBAnnotations([{ key: "", value: "" }]);
      res.dbNodeSelector != null
        ? setDBNodeSelector(res.dbNodeSelector)
        : setDBNodeSelector([{ key: "", value: "" }]);
      setRefreshLoggingInfo(false);
    }
  };

  useEffect(() => {
    if (refreshLoggingInfo) {
      api
        .invoke(
          "GET",
          `/api/v1/namespaces/${tenantNamespace}/tenants/${tenantName}/log`
        )
        .then((res: ITenantLogsStruct) => {
          if (res !== null) {
            dispatch(setAuditLoggingEnabled(res.auditLoggingEnabled));
            setLoggingInfo(res);
            setRefreshLoggingInfo(false);
          }
        })
        .catch((err: ErrorResponseHandler) => {
          dispatch(setErrorSnackMessage(err));
          setRefreshLoggingInfo(false);
        });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [refreshLoggingInfo]);

  useEffect(() => {
    if (loading) {
      setLoading(false);
    }
  }, [loading, refreshLoggingInfo]);

  const toggleLogging = () => {
    dispatch(resetAuditLogForm());
    if (!auditLoggingEnabled) {
      api
        .invoke(
          "POST",
          `/api/v1/namespaces/${tenantNamespace}/tenants/${tenantName}/enable-logging`
        )
        .then(() => {
          setRefreshLoggingInfo(true);
          setToggleConfirmOpen(false);
          setAuditLoggingEnabled(true);
        })
        .catch((err: ErrorResponseHandler) => {
          dispatch(
            setErrorSnackMessage({
              errorMessage: "Error enabling logging",
              detailedError: err.detailedError,
            })
          );
        });
    } else {
      api
        .invoke(
          "POST",
          `/api/v1/namespaces/${tenantNamespace}/tenants/${tenantName}/disable-logging`
        )
        .then(() => {
          setAuditLoggingEnabled(false);
          setRefreshLoggingInfo(true);
          setToggleConfirmOpen(false);
          dispatch(resetAuditLogForm());
        })
        .catch((err: ErrorResponseHandler) => {
          dispatch(
            setErrorSnackMessage({
              errorMessage: "Error disabling logging",
              detailedError: err.detailedError,
            })
          );
        });
    }
  };

  return deprecated ? (
    <Fragment>
      <HelpBox
        title={
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              flexGrow: 1,
            }}
          >
            <span>
              Current Audit logs functionality will be deprecated soon, please
              refer to the
              <a
                href="https://min.io/docs/minio/kubernetes/upstream/operations/monitoring/minio-logging.html"
                target="_blank"
                rel="noopener"
              >
                {" documentation "}
              </a>
              in order to setup an external service for logs
            </span>
          </div>
        }
        iconComponent={<WarnIcon />}
        help={<Fragment />}
      />
    </Fragment>
  ) : (
    <Fragment>
      <Grid item xs>
        {toggleConfirmOpen && (
          <ConfirmDialog
            isOpen={toggleConfirmOpen}
            title={
              !auditLoggingEnabled
                ? "Enable Audit Logging for this tenant?"
                : "Disable Audit Logging for this tenant?"
            }
            confirmText={!auditLoggingEnabled ? "Enable" : "Disable"}
            cancelText="Cancel"
            onClose={() => setToggleConfirmOpen(false)}
            onConfirm={toggleLogging}
            confirmationContent={
              <DialogContentText>
                {!auditLoggingEnabled
                  ? "A small Postgres server will be started per the configuration provided, which will collect the audit logs for your tenant."
                  : " Current configuration will be lost, and defaults reset if reenabled."}
              </DialogContentText>
            }
          />
        )}
      </Grid>
      <Grid container>
        <Grid item xs>
          <h1 className={classes.sectionTitle}>Audit Logs</h1>
        </Grid>
        <Grid>
          <FormSwitchWrapper
            label={""}
            indicatorLabels={["Enabled", "Disabled"]}
            checked={auditLoggingEnabled}
            value={"tenant_logging"}
            id="tenant_logging"
            name="tenant_logging"
            onChange={() => {
              setToggleConfirmOpen(true);
            }}
            description=""
          />
        </Grid>
      </Grid>
      <Grid container>
        {auditLoggingEnabled && (
          <Fragment>
            <Grid item xs={9}>
              <Tabs
                value={curTab}
                onChange={(e: React.ChangeEvent<{}>, newValue: number) => {
                  setCurTab(newValue);
                }}
                indicatorColor="primary"
                textColor="primary"
                aria-label="cluster-tabs"
                variant="scrollable"
                scrollButtons="auto"
              >
                <Tab label="Configuration" {...a11yProps(0)} />
                <Tab label="DB Configuration" {...a11yProps(1)} />
              </Tabs>
            </Grid>
            <Grid item xs={12}>
              <FormHr />
            </Grid>
            {curTab === 0 && (
              <LoggingDetails
                classes={classes}
                labels={labels}
                annotations={annotations}
                nodeSelector={nodeSelector}
              />
            )}
            {curTab === 1 && (
              <LoggingDBDetails
                classes={classes}
                labels={dbLabels}
                annotations={dbAnnotations}
                nodeSelector={dbNodeSelector}
              />
            )}
          </Fragment>
        )}
      </Grid>
    </Fragment>
  );
};

export default withStyles(styles)(LoggingScreen);
