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
import { useSelector } from "react-redux";
import { useNavigate, useParams } from "react-router-dom";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import {
  containerForHeader,
  tableStyles,
  tenantDetailsStyles,
} from "../../Common/FormComponents/common/styleLibrary";
import { niceDays } from "../../../../common/utils";
import { IPodListElement } from "../ListTenants/types";

import api from "../../../../common/api";
import TableWrapper from "../../Common/TableWrapper/TableWrapper";
import { AppState, useAppDispatch } from "../../../../store";
import { ErrorResponseHandler } from "../../../../common/types";
import DeletePod from "./DeletePod";
import { Grid, InputAdornment, TextField } from "@mui/material";
import { SearchIcon, Button } from "mds";
import { setErrorSnackMessage } from "../../../../systemSlice";
import TooltipWrapper from "../../Common/TooltipWrapper/TooltipWrapper";

interface IPodsSummary {
  classes: any;
}

const styles = (theme: Theme) =>
  createStyles({
    ...tenantDetailsStyles,
    ...tableStyles,
    ...containerForHeader,
  });

const PodsSummary = ({ classes }: IPodsSummary) => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const { tenantName, tenantNamespace } = useParams();

  const loadingTenant = useSelector(
    (state: AppState) => state.tenants.loadingTenant,
  );

  const [pods, setPods] = useState<IPodListElement[]>([]);
  const [loadingPods, setLoadingPods] = useState<boolean>(true);
  const [deleteOpen, setDeleteOpen] = useState<boolean>(false);
  const [selectedPod, setSelectedPod] = useState<any>(null);
  const [filter, setFilter] = useState("");
  const [logReportFileContent, setLogReportFileContent] = useState<string>("");
  const [startLogReport, setStartLogReport] = useState<boolean>(false);
  const [downloadReport, setDownloadReport] = useState<boolean>(false);
  const [downloadSuccess, setDownloadSuccess] = useState<boolean>(false);
  const [reportError, setReportError] = useState<boolean>(false);
  const [filename, setFilename] = useState<string>("");

  const podViewAction = (pod: IPodListElement) => {
    navigate(
      `/namespaces/${tenantNamespace || ""}/tenants/${tenantName || ""}/pods/${
        pod.name
      }`,
    );
    return;
  };

  useEffect(() => {
    if (downloadReport) {
      let element = document.createElement("a");
      element.setAttribute(
        "href",
        `data:application/gzip;base64,${logReportFileContent}`,
      );
      element.setAttribute("download", filename);

      element.style.display = "none";
      document.body.appendChild(element);

      element.click();

      document.body.removeChild(element);
      setDownloadReport(false);
      setDownloadSuccess(true);
    }
  }, [downloadReport, filename, logReportFileContent]);

  const closeDeleteModalAndRefresh = (reloadData: boolean) => {
    setDeleteOpen(false);
    setLoadingPods(true);
  };

  const confirmDeletePod = (pod: IPodListElement) => {
    pod.tenant = tenantName;
    pod.namespace = tenantNamespace;
    setSelectedPod(pod);
    setDeleteOpen(true);
  };

  const filteredRecords: IPodListElement[] = pods.filter((elementItem) =>
    elementItem.name.toLowerCase().includes(filter.toLowerCase()),
  );

  const podTableActions = [
    { type: "view", onClick: podViewAction },
    { type: "delete", onClick: confirmDeletePod },
  ];

  useEffect(() => {
    if (loadingTenant) {
      setLoadingPods(true);
    }
  }, [loadingTenant]);

  useEffect(() => {
    if (loadingPods) {
      api
        .invoke(
          "GET",
          `/api/v1/namespaces/${tenantNamespace || ""}/tenants/${
            tenantName || ""
          }/pods`,
        )
        .then((result: IPodListElement[]) => {
          for (let i = 0; i < result.length; i++) {
            let currentTime = (Date.now() / 1000) | 0;
            result[i].time = niceDays(
              (currentTime - parseInt(result[i].timeCreated)).toString(),
            );
          }
          setPods(result);
          setLoadingPods(false);
        })
        .catch((err: ErrorResponseHandler) => {
          dispatch(
            setErrorSnackMessage({
              errorMessage: "Error loading pods",
              detailedError: err.detailedError,
            }),
          );
        });
    }
  }, [loadingPods, tenantName, tenantNamespace, dispatch]);

  useEffect(() => {
    if (startLogReport) {
      setLogReportFileContent("");

      api
        .invoke(
          "GET",
          `/api/v1/namespaces/${tenantNamespace}/tenants/${tenantName}/log-report`,
        )
        .then(async (res: any) => {
          setLogReportFileContent(decodeURIComponent(res.blob));
          //@ts-ignore
          setFilename(res.filename || "tenant-log-report.zip");
          setStartLogReport(false);
          setDownloadReport(true);
          if (res.filename.length === 0 || res.blob.length === 0) {
            setReportError(true);
          } else {
            setReportError(false);
          }
        })
        .catch((err: ErrorResponseHandler) => {
          dispatch(setErrorSnackMessage(err));
          setStartLogReport(false);
          setReportError(true);
        });
    } else {
      // reset start status
      setStartLogReport(false);
    }
  }, [tenantName, tenantNamespace, startLogReport, dispatch]);

  const generateTenantLogReport = () => {
    setStartLogReport(true);
  };

  return (
    <Fragment>
      {deleteOpen && (
        <DeletePod
          deleteOpen={deleteOpen}
          selectedPod={selectedPod}
          closeDeleteModalAndRefresh={closeDeleteModalAndRefresh}
        />
      )}
      <Grid container>
        <Grid item xs={4}>
          <h1 className={classes.sectionTitle}>Pods</h1>
        </Grid>
        <Grid item xs={4}>
          {downloadSuccess &&
            !reportError &&
            "Tenant report downloaded to " + filename}
          {reportError && "There was a problem generating the report"}
        </Grid>
        <Grid item xs={4} display={"flex"} justifyContent={"flex-end"}>
          <TooltipWrapper tooltip="A report of all tenant logs will be generated as a .zip file and downloaded for analysis. This report can be uploaded to SUBNET to enable our team to best assist you in troubleshooting.">
            <Button
              id="log_report"
              onClick={generateTenantLogReport}
              disabled={pods.length === 0}
            >
              Download Log Report
            </Button>
          </TooltipWrapper>
        </Grid>
      </Grid>
      <Grid item xs={12} className={classes.actionsTray}>
        <TextField
          placeholder="Search Pods"
          className={classes.searchField}
          id="search-resource"
          label=""
          InputProps={{
            disableUnderline: true,
            startAdornment: (
              <InputAdornment position="start">
                <SearchIcon />
              </InputAdornment>
            ),
          }}
          onChange={(e) => {
            setFilter(e.target.value);
          }}
          variant="standard"
        />
      </Grid>
      <Grid item xs={12} className={classes.tableBlock}>
        <TableWrapper
          columns={[
            { label: "Name", elementKey: "name", width: 200 },
            { label: "Status", elementKey: "status" },
            { label: "Age", elementKey: "time" },
            { label: "Pod IP", elementKey: "podIP" },
            {
              label: "Restarts",
              elementKey: "restarts",
              renderFunction: (input) => {
                return input !== null ? input : 0;
              },
            },
            { label: "Node", elementKey: "node" },
          ]}
          isLoading={loadingPods}
          records={filteredRecords}
          itemActions={podTableActions}
          entityName="Pods"
          idField="name"
        />
      </Grid>
    </Fragment>
  );
};

export default withStyles(styles)(PodsSummary);
