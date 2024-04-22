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
  SectionTitle,
  Box,
  Table,
  TableHead,
  TableRow,
  TableCell,
  TableBody,
} from "mds";
import { useSelector } from "react-redux";
import { useParams } from "react-router-dom";
import { setErrorSnackMessage } from "../../../../systemSlice";
import { ErrorResponseHandler } from "../../../../common/types";
import { AppState, useAppDispatch } from "../../../../store";
import api from "../../../../common/api";

const TenantCSR = () => {
  const dispatch = useAppDispatch();
  const { tenantName, tenantNamespace } = useParams();

  const loadingTenant = useSelector(
    (state: AppState) => state.tenants.loadingTenant,
  );

  const [loading, setLoading] = useState<boolean>(true);
  const [csrStatus] = useState<string[]>([""]);
  const [csrName] = useState<string[]>([""]);
  const [csrAnnotations] = useState<string[]>([""]);

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
      <SectionTitle separator sx={{ marginBottom: 15 }}>
        Certificate Signing Requests
      </SectionTitle>
      <Box>
        <Table aria-label="collapsible table">
          <TableHead>
            <TableRow>
              <TableCell>Name</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Annotation</TableCell>
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
      </Box>
    </Fragment>
  );
};

export default TenantCSR;
