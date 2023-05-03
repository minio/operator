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
import { connect } from "react-redux";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import {
  actionsTray,
  searchField,
} from "../../../Common/FormComponents/common/styleLibrary";
import { Box } from "@mui/material";
import Grid from "@mui/material/Grid";
import Chip from "@mui/material/Chip";
import Tabs from "@mui/material/Tabs";
import Tab from "@mui/material/Tab";

import { setErrorSnackMessage } from "../../../../../systemSlice";
import { ErrorResponseHandler } from "../../../../../common/types";
import api from "../../../../../common/api";
import { AppState, useAppDispatch } from "../../../../../store";
import LabelValuePair from "../../../Common/UsageBarWrapper/LabelValuePair";
import {
  DescribeResponse,
  IPVCDescribeAnnotationsProps,
  IPVCDescribeLabelsProps,
  IPVCDescribeProps,
  IPVCDescribeSummaryProps,
} from "./pvcTypes";

const styles = (theme: Theme) =>
  createStyles({
    ...actionsTray,

    ...searchField,
  });

const twoColCssGridLayoutConfig = {
  display: "grid",
  gridTemplateColumns: { xs: "1fr", sm: "2fr 1fr" },
  gridAutoFlow: { xs: "dense", sm: "row" },
  gap: 2,
  padding: "15px",
};

const HeaderSection = ({ title }: { title: string }) => {
  return (
    <Box
      sx={{
        borderBottom: "1px solid #eaeaea",
        margin: 0,
        marginBottom: "20px",
      }}
    >
      <h3>{title}</h3>
    </Box>
  );
};

const PVCDescribeSummary = ({ describeInfo }: IPVCDescribeSummaryProps) => {
  return (
    <Fragment>
      <div id="pvc-describe-summary-content">
        <HeaderSection title={"Summary"} />
        <Box sx={{ ...twoColCssGridLayoutConfig }}>
          <LabelValuePair label={"Name"} value={describeInfo.name} />
          <LabelValuePair label={"Namespace"} value={describeInfo.namespace} />
          <LabelValuePair label={"Capacity"} value={describeInfo.capacity} />
          <LabelValuePair label={"Status"} value={describeInfo.status} />
          <LabelValuePair
            label={"Storage Class"}
            value={describeInfo.storageClass}
          />
          <LabelValuePair
            label={"Access Modes"}
            value={describeInfo.accessModes.join(", ")}
          />
          <LabelValuePair
            label={"Finalizers"}
            value={describeInfo.finalizers.join(", ")}
          />
          <LabelValuePair label={"Volume"} value={describeInfo.volume} />
          <LabelValuePair
            label={"Volume Mode"}
            value={describeInfo.volumeMode}
          />
        </Box>
      </div>
    </Fragment>
  );
};

const PVCDescribeAnnotations = ({
  annotations,
}: IPVCDescribeAnnotationsProps) => {
  return (
    <Fragment>
      <div id="pvc-describe-annotations-content">
        <HeaderSection title={"Annotations"} />
        <Box>
          {annotations.map((annotation, index) => (
            <Chip
              style={{ margin: "0.5%" }}
              label={`${annotation.key}: ${annotation.value}`}
              key={index}
            />
          ))}
        </Box>
      </div>
    </Fragment>
  );
};

const PVCDescribeLabels = ({ labels }: IPVCDescribeLabelsProps) => {
  return (
    <Fragment>
      <div id="pvc-describe-labels-content">
        <HeaderSection title={"Labels"} />
        <Box>
          {labels.map((label, index) => (
            <Chip
              style={{ margin: "0.5%" }}
              label={`${label.key}: ${label.value}`}
              key={index}
            />
          ))}
        </Box>
      </div>
    </Fragment>
  );
};

const PVCDescribe = ({
  tenant,
  namespace,
  pvcName,
  propLoading,
}: IPVCDescribeProps) => {
  const [describeInfo, setDescribeInfo] = useState<DescribeResponse>();
  const [loading, setLoading] = useState<boolean>(true);
  const [curTab, setCurTab] = useState<number>(0);
  const dispatch = useAppDispatch();

  useEffect(() => {
    if (propLoading) {
      setLoading(true);
    }
  }, [propLoading]);

  useEffect(() => {
    if (loading) {
      api
        .invoke(
          "GET",
          `/api/v1/namespaces/${namespace}/tenants/${tenant}/pvcs/${pvcName}/describe`
        )
        .then((res: DescribeResponse) => {
          setDescribeInfo(res);
          setLoading(false);
        })
        .catch((err: ErrorResponseHandler) => {
          dispatch(setErrorSnackMessage(err));
          setLoading(false);
        });
    }
  }, [loading, pvcName, namespace, tenant, dispatch]);

  const renderTabComponent = (index: number, info: DescribeResponse) => {
    switch (index) {
      case 0:
        return <PVCDescribeSummary describeInfo={info} />;
      case 1:
        return <PVCDescribeAnnotations annotations={info.annotations} />;
      case 2:
        return <PVCDescribeLabels labels={info.labels} />;
      default:
        break;
    }
  };
  return (
    <Fragment>
      {describeInfo && (
        <Grid item xs={12}>
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
            <Tab id="pvc-describe-summary" label="Summary" />
            <Tab id="pvc-describe-annotations" label="Annotations" />
            <Tab id="pvc-describe-labels" label="Labels" />
          </Tabs>
          {renderTabComponent(curTab, describeInfo)}
        </Grid>
      )}
    </Fragment>
  );
};
const mapState = (state: AppState) => ({
  loadingTenant: state.tenants.loadingTenant,
});
const connector = connect(mapState, {
  setErrorSnackMessage,
});

export default withStyles(styles)(connector(PVCDescribe));
