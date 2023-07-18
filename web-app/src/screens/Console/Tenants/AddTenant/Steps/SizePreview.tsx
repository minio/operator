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

import React, { Fragment } from "react";
import { useSelector } from "react-redux";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { AppState } from "../../../../../store";
import {
  modalBasic,
  wizardCommon,
} from "../../../Common/FormComponents/common/styleLibrary";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableRow from "@mui/material/TableRow";
import { niceBytes } from "../../../../../common/utils";

import { Divider } from "@mui/material";

interface ISizePreviewProps {
  classes: any;
}

const styles = (theme: Theme) =>
  createStyles({
    root: {
      margin: 4,
    },
    table: {
      "& .MuiTableCell-root": {
        fontSize: 13,
      },
    },
    ...modalBasic,
    ...wizardCommon,
  });

const SizePreview = ({ classes }: ISizePreviewProps) => {
  const nodes = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.nodes,
  );
  const memoryNode = useSelector(
    (state: AppState) =>
      state.createTenant.fields.tenantSize.resourcesMemoryRequest,
  );
  const ecParity = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.ecParity,
  );

  const distribution = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.distribution,
  );
  const ecParityCalc = useSelector(
    (state: AppState) => state.createTenant.fields.tenantSize.ecParityCalc,
  );

  const cpuToUse = useSelector(
    (state: AppState) =>
      state.createTenant.fields.tenantSize.resourcesCPURequest,
  );
  const integrationSelection = useSelector(
    (state: AppState) =>
      state.createTenant.fields.tenantSize.integrationSelection,
  );

  const usableInformation = ecParityCalc.storageFactors.find(
    (element) => element.erasureCode === ecParity,
  );

  return (
    <div className={classes.root}>
      <h4>Resource Allocation</h4>
      <Divider />
      <Table className={classes.table} aria-label="simple table" size={"small"}>
        <TableBody>
          <TableRow>
            <TableCell scope="row">Number of Servers</TableCell>
            <TableCell align="right">
              {parseInt(nodes) > 0 ? nodes : "-"}
            </TableCell>
          </TableRow>
          {integrationSelection.typeSelection === "" &&
            integrationSelection.storageClass === "" && (
              <Fragment>
                <TableRow>
                  <TableCell scope="row">Drives per Server</TableCell>
                  <TableCell align="right">
                    {distribution ? distribution.disks : "-"}
                  </TableCell>
                </TableRow>
                <TableRow>
                  <TableCell scope="row">Drive Capacity</TableCell>
                  <TableCell align="right">
                    {distribution ? niceBytes(distribution.pvSize) : "-"}
                  </TableCell>
                </TableRow>
              </Fragment>
            )}

          <TableRow>
            <TableCell scope="row">Total Volumes</TableCell>
            <TableCell align="right">
              {distribution ? distribution.persistentVolumes : "-"}
            </TableCell>
          </TableRow>
          {integrationSelection.typeSelection === "" &&
            integrationSelection.storageClass === "" && (
              <Fragment>
                <TableRow>
                  <TableCell scope="row">Memory per Node</TableCell>
                  <TableCell align="right">{memoryNode} Gi</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell style={{ borderBottom: 0 }} scope="row">
                    CPU Selection
                  </TableCell>
                  <TableCell style={{ borderBottom: 0 }} align="right">
                    {cpuToUse}
                  </TableCell>
                </TableRow>
              </Fragment>
            )}
        </TableBody>
      </Table>
      {ecParityCalc.error === 0 && usableInformation && (
        <Fragment>
          <h4>Erasure Code Configuration</h4>
          <Divider />
          <Table
            className={classes.table}
            aria-label="simple table"
            size={"small"}
          >
            <TableBody>
              <TableRow>
                <TableCell scope="row">EC Parity</TableCell>
                <TableCell align="right">
                  {ecParity !== "" ? ecParity : "-"}
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell scope="row">Raw Capacity</TableCell>
                <TableCell align="right">
                  {niceBytes(ecParityCalc.rawCapacity)}
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell scope="row">Usable Capacity</TableCell>
                <TableCell align="right">
                  {niceBytes(usableInformation.maxCapacity)}
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell style={{ borderBottom: 0 }} scope="row">
                  Server Failures Tolerated
                </TableCell>
                <TableCell style={{ borderBottom: 0 }} align="right">
                  {distribution
                    ? Math.floor(
                        usableInformation.maxFailureTolerations /
                          distribution.disks,
                      )
                    : "-"}
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </Fragment>
      )}
      {integrationSelection.typeSelection !== "" &&
        integrationSelection.storageClass !== "" && (
          <Fragment>
            <h4>Single Instance Configuration</h4>
            <Divider />
            <Table
              className={classes.table}
              aria-label="simple table"
              size={"small"}
            >
              <TableBody>
                <TableRow>
                  <TableCell scope="row">CPU</TableCell>
                  <TableCell align="right">
                    {integrationSelection.CPU !== 0
                      ? integrationSelection.CPU
                      : "-"}
                  </TableCell>
                </TableRow>
                <TableRow>
                  <TableCell scope="row">Memory</TableCell>
                  <TableCell align="right">
                    {integrationSelection.memory !== 0
                      ? `${integrationSelection.memory} Gi`
                      : "-"}
                  </TableCell>
                </TableRow>
                <TableRow>
                  <TableCell scope="row">Drives per Server</TableCell>
                  <TableCell align="right">
                    {integrationSelection.drivesPerServer !== 0
                      ? `${integrationSelection.drivesPerServer}`
                      : "-"}
                  </TableCell>
                </TableRow>
                <TableRow>
                  <TableCell style={{ borderBottom: 0 }} scope="row">
                    Drive Size
                  </TableCell>
                  <TableCell style={{ borderBottom: 0 }} align="right">
                    {integrationSelection.driveSize.driveSize}
                    {integrationSelection.driveSize.sizeUnit}
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </Fragment>
        )}
    </div>
  );
};

export default withStyles(styles)(SizePreview);
