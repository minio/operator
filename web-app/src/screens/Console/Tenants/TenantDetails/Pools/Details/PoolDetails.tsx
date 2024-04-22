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

import React, { Fragment } from "react";
import { useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";
import { AppState } from "../../../../../../store";
import { niceBytesInt } from "../../../../../../common/utils";
import {
  Box,
  breakPoints,
  Button,
  EditTenantIcon,
  SectionTitle,
  ValuePair,
} from "mds";
import { NodeSelectorTerm } from "../../../../../../api/operatorApi";

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

const PoolDetails = () => {
  const navigate = useNavigate();

  const tenant = useSelector((state: AppState) => state.tenants.tenantInfo);
  const selectedPool = useSelector(
    (state: AppState) => state.tenants.selectedPool,
  );
  if (tenant === null) {
    return <Fragment />;
  }

  const poolInformation =
    tenant.pools?.find((pool) => pool.name === selectedPool) || null;

  if (poolInformation === null) {
    return null;
  }

  let affinityType = "None";

  if (poolInformation.affinity) {
    if (poolInformation.affinity.nodeAffinity) {
      affinityType = "Node Selector";
    } else {
      affinityType = "Default (Pod Anti-Affinity)";
    }
  }

  return (
    <Fragment>
      <Box
        withBorders={true}
        customBorderPadding={"0px"}
        sx={{ width: "100%" }}
      >
        <SectionTitle
          separator
          actions={
            <Button
              icon={<EditTenantIcon />}
              onClick={() => {
                navigate(
                  `/namespaces/${tenant?.namespace || ""}/tenants/${
                    tenant?.name || ""
                  }/edit-pool`,
                );
              }}
              label={"Edit Pool"}
              id={"editPool"}
            />
          }
        >
          Pool Configuration
        </SectionTitle>
        <Box sx={{ ...twoColCssGridLayoutConfig }}>
          <ValuePair label={"Pool Name"} value={poolInformation.name} />
          <ValuePair
            label={"Total Volumes"}
            value={poolInformation.volumes_per_server}
          />
          <ValuePair
            label={"Volumes per server"}
            value={poolInformation.volumes_per_server}
          />
          <ValuePair
            label={"Capacity"}
            value={niceBytesInt(
              poolInformation.volumes_per_server *
                poolInformation.servers *
                poolInformation.volume_configuration.size,
            )}
          />
          <ValuePair
            label={"Runtime Class Name"}
            value={poolInformation.runtimeClassName}
          />
        </Box>
        <SectionTitle separator>Resources</SectionTitle>
        <Box sx={{ ...twoColCssGridLayoutConfig }}>
          {poolInformation.resources && (
            <Fragment>
              <ValuePair
                label={"CPU"}
                value={poolInformation.resources?.requests?.cpu}
              />
              <ValuePair
                label={"Memory"}
                value={niceBytesInt(
                  poolInformation.resources?.requests?.memory!,
                )}
              />
            </Fragment>
          )}
          <ValuePair
            label={"Volume Size"}
            value={niceBytesInt(poolInformation.volume_configuration.size)}
          />
          <ValuePair
            label={"Storage Class Name"}
            value={poolInformation.volume_configuration.storage_class_name}
          />
        </Box>
        {poolInformation.securityContext &&
          (poolInformation.securityContext.runAsNonRoot ||
            poolInformation.securityContext.runAsUser ||
            poolInformation.securityContext.runAsGroup ||
            poolInformation.securityContext.fsGroup) && (
            <Fragment>
              <SectionTitle separator>Security Context</SectionTitle>
              <Box>
                {poolInformation.securityContext.runAsNonRoot !== null && (
                  <Box sx={{ ...twoColCssGridLayoutConfig }}>
                    <ValuePair
                      label={"Run as Non Root"}
                      value={
                        poolInformation.securityContext.runAsNonRoot
                          ? "Yes"
                          : "No"
                      }
                    />
                  </Box>
                )}
                <Box
                  sx={{
                    display: "grid",
                    gridTemplateColumns: "2fr 1fr",
                    gridAutoFlow: "row",
                    gap: 2,
                    padding: "15px",
                    [`@media (max-width: ${breakPoints.sm}px)`]: {
                      gridTemplateColumns: "1fr",
                      gridAutoFlow: "dense",
                    },
                    [`@media (max-width: ${breakPoints.md}px)`]: {
                      gridTemplateColumns: "2fr 1fr",
                    },
                    [`@media (max-width: ${breakPoints.lg}px)`]: {
                      gridTemplateColumns: "1fr 1fr 1fr",
                    },
                  }}
                >
                  {poolInformation.securityContext.runAsUser && (
                    <ValuePair
                      label={"Run as User"}
                      value={poolInformation.securityContext.runAsUser}
                    />
                  )}
                  {poolInformation.securityContext.runAsGroup && (
                    <ValuePair
                      label={"Run as Group"}
                      value={poolInformation.securityContext.runAsGroup}
                    />
                  )}
                  {poolInformation.securityContext.fsGroup && (
                    <ValuePair
                      label={"FsGroup"}
                      value={poolInformation.securityContext.fsGroup}
                    />
                  )}
                </Box>
              </Box>
            </Fragment>
          )}
        <SectionTitle separator>Affinity</SectionTitle>
        <Box>
          <Box sx={{ ...twoColCssGridLayoutConfig }}>
            <ValuePair label={"Type"} value={affinityType} />
            {poolInformation.affinity?.nodeAffinity &&
            poolInformation.affinity?.podAntiAffinity ? (
              <ValuePair label={"With Pod Anti affinity"} value={"Yes"} />
            ) : (
              <span />
            )}
          </Box>
          {poolInformation.affinity?.nodeAffinity && (
            <Fragment>
              <SectionTitle separator>Labels</SectionTitle>
              <ul>
                {poolInformation.affinity?.nodeAffinity?.requiredDuringSchedulingIgnoredDuringExecution?.nodeSelectorTerms.map(
                  (term: NodeSelectorTerm) => {
                    return term.matchExpressions?.map((trm) => {
                      return (
                        <li>
                          {trm.key} - {trm.values?.join(", ")}
                        </li>
                      );
                    });
                  },
                )}
              </ul>
            </Fragment>
          )}
        </Box>
        {poolInformation.tolerations &&
          poolInformation.tolerations.length > 0 && (
            <Fragment>
              <SectionTitle separator>Tolerations</SectionTitle>
              <Box>
                <ul>
                  {poolInformation.tolerations.map((tolItem) => {
                    return (
                      <li>
                        {tolItem.operator === "Equal" ? (
                          <Fragment>
                            If <strong>{tolItem.key}</strong> is equal to{" "}
                            <strong>{tolItem.value}</strong> then{" "}
                            <strong>{tolItem.effect}</strong> after{" "}
                            <strong>
                              {tolItem.tolerationSeconds?.seconds || 0}
                            </strong>{" "}
                            seconds
                          </Fragment>
                        ) : (
                          <Fragment>
                            If <strong>{tolItem.key}</strong> exists then{" "}
                            <strong>{tolItem.effect}</strong> after{" "}
                            <strong>
                              {tolItem.tolerationSeconds?.seconds || 0}
                            </strong>{" "}
                            seconds
                          </Fragment>
                        )}
                      </li>
                    );
                  })}
                </ul>
              </Box>
            </Fragment>
          )}
      </Box>
    </Fragment>
  );
};

export default PoolDetails;
