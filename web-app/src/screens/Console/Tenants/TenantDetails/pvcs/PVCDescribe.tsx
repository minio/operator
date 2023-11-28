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
import { Box, breakPoints, Tabs, Tag, ValuePair } from "mds";
import { setErrorSnackMessage } from "../../../../../systemSlice";
import { ErrorResponseHandler } from "../../../../../common/types";
import { useAppDispatch } from "../../../../../store";
import {
  DescribeResponse,
  IPVCDescribeAnnotationsProps,
  IPVCDescribeLabelsProps,
  IPVCDescribeProps,
  IPVCDescribeSummaryProps,
} from "./pvcTypes";
import api from "../../../../../common/api";

const twoColCssGridLayoutConfig = {
  display: "grid",
  gridTemplateColumns: "2fr 1fr",
  gridAutoFlow: "row",
  gap: 2,
  padding: 15,
  [`@media (max-width: ${breakPoints.sm}px)`]: {
    gridTemplateColumns: "1fr",
    gridAutoFlow: "dense",
  },
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
          <ValuePair label={"Name"} value={describeInfo.name} />
          <ValuePair label={"Namespace"} value={describeInfo.namespace} />
          <ValuePair label={"Capacity"} value={describeInfo.capacity} />
          <ValuePair label={"Status"} value={describeInfo.status} />
          <ValuePair
            label={"Storage Class"}
            value={describeInfo.storageClass}
          />
          <ValuePair
            label={"Access Modes"}
            value={describeInfo.accessModes.join(", ")}
          />
          <ValuePair
            label={"Finalizers"}
            value={describeInfo.finalizers.join(", ")}
          />
          <ValuePair label={"Volume"} value={describeInfo.volume} />
          <ValuePair label={"Volume Mode"} value={describeInfo.volumeMode} />
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
            <Tag
              id={`${annotation.key}-${annotation.value}`}
              sx={{ margin: "0.5%" }}
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
            <Tag
              id={`${label.key}-${label.value}`}
              sx={{ margin: "0.5%" }}
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
  const [curTab, setCurTab] = useState<string>("pvc-describe-summary");
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
          `/api/v1/namespaces/${namespace}/tenants/${tenant}/pvcs/${pvcName}/describe`,
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

  return (
    <Fragment>
      {describeInfo && (
        <Tabs
          currentTabOrPath={curTab}
          onTabClick={(newValue) => {
            setCurTab(newValue);
          }}
          options={[
            {
              tabConfig: { id: "pvc-describe-summary", label: "Summary" },
              content: <PVCDescribeSummary describeInfo={describeInfo} />,
            },
            {
              tabConfig: {
                id: "pvc-describe-annotations",
                label: "Annotations",
              },
              content: (
                <PVCDescribeAnnotations
                  annotations={describeInfo.annotations}
                />
              ),
            },
            {
              tabConfig: { id: "pvc-describe-labels", label: "Labels" },
              content: <PVCDescribeLabels labels={describeInfo.labels} />,
            },
          ]}
          horizontal
        />
      )}
    </Fragment>
  );
};

export default PVCDescribe;
