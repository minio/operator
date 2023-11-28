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
import {
  Box,
  breakPoints,
  Grid,
  SectionTitle,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Tabs,
  Tag,
  ValuePair,
} from "mds";
import { ErrorResponseHandler } from "../../../../../common/types";
import api from "../../../../../common/api";
import { useSelector } from "react-redux";
import { AppState, useAppDispatch } from "../../../../../store";
import { setErrorSnackMessage } from "../../../../../systemSlice";

interface IPodEventsProps {
  tenant: string;
  namespace: string;
  podName: string;
  propLoading: boolean;
}

interface Annotation {
  key: string;
  value: string;
}

interface Condition {
  status: string;
  type: string;
}

interface EnvVar {
  key: string;
  value: string;
}

interface Mount {
  mountPath: string;
  name: string;
}

interface State {
  started: string;
  state: string;
}

interface Container {
  args: string[];
  containerID: string;
  environmentVariables: EnvVar[];
  hostPorts: string[];
  image: string;
  imageID: string;
  lastState: any;
  mounts: Mount[];
  name: string;
  ports: string[];
  ready: boolean;
  state: State;
}

interface Label {
  key: string;
  value: string;
}

interface Toleration {
  effect: string;
  key: string;
  operator: string;
  tolerationSeconds: number;
}

interface VolumePVC {
  claimName: string;
}

interface Volume {
  name: string;
  pvc?: VolumePVC;
  projected?: any;
}

interface DescribeResponse {
  annotations: Annotation[];
  conditions: Condition[];
  containers: Container[];
  controllerRef: string;
  labels: Label[];
  name: string;
  namespace: string;
  nodeName: string;
  nodeSelector: string[];
  phase: string;
  podIP: string;
  qosClass: string;
  startTime: string;
  tolerations: Toleration[];
  volumes: Volume[];
}

interface IPodDescribeSummaryProps {
  describeInfo: DescribeResponse;
}

interface IPodDescribeAnnotationsProps {
  annotations: Annotation[];
}

interface IPodDescribeLabelsProps {
  labels: Label[];
}

interface IPodDescribeConditionsProps {
  conditions: Condition[];
}

interface IPodDescribeTolerationsProps {
  tolerations: Toleration[];
}

interface IPodDescribeVolumesProps {
  volumes: Volume[];
}

interface IPodDescribeContainersProps {
  containers: Container[];
}

interface IPodDescribeTableProps {
  title: string;
  columns: string[];
  columnsLabels: string[];
  items: any[];
}

const twoColCssGridLayoutConfig = {
  display: "grid",
  gridTemplateColumns: "2fr 1fr",
  gridAutoFlow: "row",
  gap: 2,
  padding: "15px",
  [`@media (max-width: ${breakPoints.sm}px)`]: {
    gridTemplateColumns: "1fr",
    gridAutoFlow: "dense",
  },
};

const HeaderSection = ({ title }: { title: string }) => {
  return (
    <SectionTitle separator sx={{ marginBottom: 5 }}>
      {title}
    </SectionTitle>
  );
};

const PodDescribeSummary = ({ describeInfo }: IPodDescribeSummaryProps) => {
  return (
    <React.Fragment>
      <div id="pod-describe-summary-content">
        <HeaderSection title={"Summary"} />
        <Box sx={{ ...twoColCssGridLayoutConfig }}>
          <ValuePair label={"Name"} value={describeInfo.name} />
          <ValuePair label={"Namespace"} value={describeInfo.namespace} />
          <ValuePair label={"Node"} value={describeInfo.nodeName} />
          <ValuePair label={"Start time"} value={describeInfo.startTime} />
          <ValuePair label={"Status"} value={describeInfo.phase} />
          <ValuePair label={"QoS Class"} value={describeInfo.qosClass} />
          <ValuePair label={"IP"} value={describeInfo.podIP} />
        </Box>
      </div>
    </React.Fragment>
  );
};

const PodDescribeAnnotations = ({
  annotations,
}: IPodDescribeAnnotationsProps) => {
  return (
    <React.Fragment>
      <div id="pod-describe-annotations-content">
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
    </React.Fragment>
  );
};

const PodDescribeLabels = ({ labels }: IPodDescribeLabelsProps) => {
  return (
    <React.Fragment>
      <div id="pod-describe-labels-content">
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
    </React.Fragment>
  );
};

const PodDescribeConditions = ({ conditions }: IPodDescribeConditionsProps) => {
  return (
    <div id="pod-describe-conditions-content">
      <PodDescribeTable
        title="Conditions"
        columns={["type", "status"]}
        columnsLabels={["Type", "Status"]}
        items={conditions}
      />
    </div>
  );
};

const PodDescribeTolerations = ({
  tolerations,
}: IPodDescribeTolerationsProps) => {
  return (
    <div id="pod-describe-tolerations-content">
      <PodDescribeTable
        title="Tolerations"
        columns={["effect", "key", "operator", "tolerationSeconds"]}
        columnsLabels={["Effect", "Key", "Operator", "Seconds of toleration"]}
        items={tolerations}
      />
    </div>
  );
};

const PodDescribeVolumes = ({ volumes }: IPodDescribeVolumesProps) => {
  return (
    <React.Fragment>
      <div id="pod-describe-volumes-content">
        {volumes.map((volume, index) => (
          <React.Fragment key={index}>
            <HeaderSection title={`Volume ${volume.name}`} />
            <Box sx={{ ...twoColCssGridLayoutConfig }}>
              {volume.pvc && (
                <React.Fragment>
                  <ValuePair label={"Type"} value="Persistant Volume Claim" />
                  <ValuePair
                    label={"Claim Name"}
                    value={volume.pvc.claimName}
                  />
                </React.Fragment>
              )}
              {/* TODO Add component to display projected data (Maybe change API response) */}
              {volume.projected && (
                <ValuePair label={"Type"} value="Projected" />
              )}
            </Box>
          </React.Fragment>
        ))}
      </div>
    </React.Fragment>
  );
};

const PodDescribeTable = ({
  title,
  items,
  columns,
  columnsLabels,
}: IPodDescribeTableProps) => {
  return (
    <Fragment>
      <HeaderSection title={title} />
      <Box>
        <Table aria-label="collapsible table">
          <TableHead>
            <TableRow>
              {columnsLabels.map((label, index) => (
                <TableCell key={index}>{label}</TableCell>
              ))}
            </TableRow>
          </TableHead>
          <TableBody>
            {items.map((item, i) => {
              return (
                <TableRow key={i}>
                  {columns.map((column, j) => (
                    <TableCell key={j}>{item[column]}</TableCell>
                  ))}
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </Box>
    </Fragment>
  );
};

const PodDescribeContainers = ({ containers }: IPodDescribeContainersProps) => {
  return (
    <React.Fragment>
      <div id="pod-describe-containers-content">
        {containers.map((container, index) => (
          <React.Fragment key={index}>
            <HeaderSection title={`Container ${container.name}`} />
            <Box
              style={{ wordBreak: "break-all" }}
              sx={{ ...twoColCssGridLayoutConfig }}
            >
              <ValuePair label={"Image"} value={container.image} />
              <ValuePair label={"Ready"} value={`${container.ready}`} />
              <ValuePair label={"Ports"} value={container.ports.join(", ")} />
              <ValuePair
                label={"Host Ports"}
                value={container.hostPorts.join(", ")}
              />
              <ValuePair
                label={"Arguments"}
                value={container.args.join(", ")}
              />
              <ValuePair label={"Started"} value={container.state?.started} />
              <ValuePair label={"State"} value={container.state?.state} />
            </Box>
            <Box
              style={{ wordBreak: "break-all" }}
              sx={{ ...twoColCssGridLayoutConfig }}
            >
              <ValuePair label={"Image ID"} value={container.imageID} />
              <ValuePair label={"Container ID"} value={container.containerID} />
            </Box>
            <PodDescribeTable
              title="Mounts"
              columns={["name", "mountPath"]}
              columnsLabels={["Name", "Mount Path"]}
              items={container.mounts}
            />
            <PodDescribeTable
              title="Environment Variables"
              columns={["key", "value"]}
              columnsLabels={["Key", "Value"]}
              items={container.environmentVariables}
            />
          </React.Fragment>
        ))}
      </div>
    </React.Fragment>
  );
};

const PodDescribe = ({
  tenant,
  namespace,
  podName,
  propLoading,
}: IPodEventsProps) => {
  const dispatch = useAppDispatch();
  const loadingTenant = useSelector(
    (state: AppState) => state.tenants.loadingTenant,
  );

  const [describeInfo, setDescribeInfo] = useState<DescribeResponse>();
  const [loading, setLoading] = useState<boolean>(true);
  const [curTab, setCurTab] = useState<string>("pod-describe-summary");

  useEffect(() => {
    if (propLoading) {
      setLoading(true);
    }
  }, [propLoading]);

  useEffect(() => {
    if (loadingTenant) {
      setLoading(true);
    }
  }, [loadingTenant]);

  useEffect(() => {
    if (loading) {
      api
        .invoke(
          "GET",
          `/api/v1/namespaces/${namespace}/tenants/${tenant}/pods/${podName}/describe`,
        )
        .then((res: DescribeResponse) => {
          const cleanRes = cleanDescribeResponseEnvVariables(res);
          setDescribeInfo(cleanRes);
          setLoading(false);
        })
        .catch((err: ErrorResponseHandler) => {
          dispatch(setErrorSnackMessage(err));
          setLoading(false);
        });
    }
  }, [loading, podName, namespace, tenant, dispatch]);

  const cleanDescribeResponseEnvVariables = (
    res: DescribeResponse,
  ): DescribeResponse => {
    res.containers = res.containers.map((c) => {
      c.environmentVariables = c.environmentVariables.filter(
        (item) => item !== null,
      );
      return c;
    });
    return res;
  };

  return (
    <React.Fragment>
      {describeInfo && (
        <Grid item xs={12}>
          <Tabs
            currentTabOrPath={curTab}
            onTabClick={(newValue) => {
              setCurTab(newValue);
            }}
            horizontal
            options={[
              {
                tabConfig: { id: "pod-describe-summary", label: "Summary" },
                content: <PodDescribeSummary describeInfo={describeInfo} />,
              },
              {
                tabConfig: {
                  id: "pod-describe-annotations",
                  label: "Annotations",
                },
                content: (
                  <PodDescribeAnnotations
                    annotations={describeInfo.annotations}
                  />
                ),
              },
              {
                tabConfig: { id: "pod-describe-labels", label: "Labels" },
                content: <PodDescribeLabels labels={describeInfo.labels} />,
              },
              {
                tabConfig: {
                  id: "pod-describe-conditions",
                  label: "Conditions",
                },
                content: (
                  <PodDescribeConditions conditions={describeInfo.conditions} />
                ),
              },
              {
                tabConfig: {
                  id: "pod-describe-tolerations",
                  label: "Tolerations",
                },
                content: (
                  <PodDescribeTolerations
                    tolerations={describeInfo.tolerations}
                  />
                ),
              },
              {
                tabConfig: { id: "pod-describe-volumes", label: "Volumes" },
                content: <PodDescribeVolumes volumes={describeInfo.volumes} />,
              },
              {
                tabConfig: {
                  id: "pod-describe-containers",
                  label: "Containers",
                },
                content: (
                  <PodDescribeContainers containers={describeInfo.containers} />
                ),
              },
            ]}
          />
        </Grid>
      )}
    </React.Fragment>
  );
};

export default PodDescribe;
