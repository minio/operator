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
import { useSelector } from "react-redux";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { setErrorSnackMessage } from "../../../../systemSlice";
import {
  actionsTray,
  containerForHeader,
  searchField,
  tableStyles,
} from "../../Common/FormComponents/common/styleLibrary";
import { ErrorResponseHandler } from "../../../../common/types";
import api from "../../../../common/api";
import TableContainer from "@mui/material/TableContainer";
import Paper from "@mui/material/Paper";
import Table from "@mui/material/Table";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import TableCell from "@mui/material/TableCell";
import TableBody from "@mui/material/TableBody";
import { useParams } from "react-router-dom";
import { AppState, useAppDispatch } from "../../../../store";

interface ITenantCSRProps {
  classes: any;
}

const styles = (theme: Theme) =>
  createStyles({
    ...actionsTray,
    ...searchField,
    ...tableStyles,
    ...containerForHeader,
  });

const TenantCSR = ({ classes }: ITenantCSRProps) => {
  const dispatch = useAppDispatch();
  const { tenantName, tenantNamespace } = useParams();

  const loadingTenant = useSelector(
    (state: AppState) => state.tenants.loadingTenant,
  );

  const [loading, setLoading] = useState<boolean>(true);
  const [csrStatus] = useState([""]);
  const [csrName] = useState([""]);
  const [csrAnnotations] = useState([""]);

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
          `/api/v1/namespaces/${tenantNamespace || ""}/tenants/${
            tenantName || ""
          }/csr`,
        )
        .then((res) => {
          for (var _i = 0; _i < res.csrElement.length; _i++) {
            var entry = res.csrElement[_i];
            csrStatus.push(entry.status);
            csrName.push(entry.name);
            csrAnnotations.push(entry.annotations);
          }
          setLoading(false);
        })
        .catch((err: ErrorResponseHandler) => {
          dispatch(setErrorSnackMessage(err));
        });
    }
  }, [
    loading,
    tenantNamespace,
    tenantName,
    csrAnnotations,
    csrName,
    csrStatus,
    dispatch,
  ]);

  return (
    <Fragment>
      <h1 className={classes.sectionTitle}>Certificate Signing Requests</h1>
      <TableContainer component={Paper}>
        <Table aria-label="collapsible table">
          <TableHead>
            <TableRow>
              <TableCell>Name</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Annotation</TableCell>
              <TableCell />
            </TableRow>
          </TableHead>
          <TableBody>
            <TableRow>
              <TableCell>
                {csrName.map((csrName) => (
                  <p>{csrName}</p>
                ))}
              </TableCell>
              <TableCell>
                {csrStatus.map((csrStatus) => (
                  <p>{csrStatus}</p>
                ))}
              </TableCell>
              <TableCell>
                {csrAnnotations.map((csrAnnotations) => (
                  <p>{csrAnnotations}</p>
                ))}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </TableContainer>
    </Fragment>
  );
};

export default withStyles(styles)(TenantCSR);
