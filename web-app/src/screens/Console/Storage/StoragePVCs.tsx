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
import get from "lodash/get";
import { Theme } from "@mui/material/styles";
import { Grid } from "@mui/material";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import {
  actionsTray,
  containerForHeader,
  searchField,
} from "../Common/FormComponents/common/styleLibrary";
import { ErrorResponseHandler } from "../../../common/types";
import { useAppDispatch } from "../../../store";
import { setErrorSnackMessage } from "../../../systemSlice";
import { IPVCsResponse, IStoragePVCs } from "./types";
import api from "../../../common/api";
import TableWrapper from "../Common/TableWrapper/TableWrapper";
import DeletePVC from "../Tenants/TenantDetails/DeletePVC";
import PageLayout from "../Common/Layout/PageLayout";
import SearchBox from "../Common/SearchBox";
import PageHeaderWrapper from "../Common/PageHeaderWrapper/PageHeaderWrapper";

interface IStorageVolumesProps {
  classes: any;
}

const styles = (theme: Theme) =>
  createStyles({
    tableWrapper: {
      height: "calc(100vh - 150px)",
    },
    ...actionsTray,
    ...searchField,
    ...containerForHeader,
  });

const StorageVolumes = ({ classes }: IStorageVolumesProps) => {
  const dispatch = useAppDispatch();

  const [records, setRecords] = useState<IStoragePVCs[]>([]);
  const [filter, setFilter] = useState("");
  const [loading, setLoading] = useState<boolean>(true);
  const [selectedPVC, setSelectedPVC] = useState<any>(null);
  const [deleteOpen, setDeleteOpen] = useState<boolean>(false);

  useEffect(() => {
    if (loading) {
      api
        .invoke("GET", `/api/v1/list-pvcs`)
        .then((res: IPVCsResponse) => {
          let volumes = get(res, "pvcs", []);
          setRecords(volumes ? volumes : []);
          setLoading(false);
        })
        .catch((err: ErrorResponseHandler) => {
          setLoading(false);
          dispatch(setErrorSnackMessage(err));
        });
    }
  }, [loading, dispatch]);

  const filteredRecords: IStoragePVCs[] = records.filter((elementItem) =>
    elementItem.name.toLowerCase().includes(filter.toLowerCase()),
  );

  const confirmDeletePVC = (pvcItem: IStoragePVCs) => {
    const delPvc = {
      ...pvcItem,
      tenant: pvcItem.tenant,
      namespace: pvcItem.namespace,
    };
    setSelectedPVC(delPvc);
    setDeleteOpen(true);
  };

  const tableActions = [{ type: "delete", onClick: confirmDeletePVC }];

  const closeDeleteModalAndRefresh = (reloadData: boolean) => {
    setDeleteOpen(false);
    setLoading(true);
  };

  return (
    <Fragment>
      {deleteOpen && (
        <DeletePVC
          deleteOpen={deleteOpen}
          selectedPVC={selectedPVC}
          closeDeleteModalAndRefresh={closeDeleteModalAndRefresh}
        />
      )}
      <PageHeaderWrapper
        label="Persistent Volumes Claims"
        middleComponent={
          <SearchBox
            placeholder={"Search Volumes (PVCs)"}
            onChange={(val) => {
              setFilter(val);
            }}
            value={filter}
          />
        }
      />
      <PageLayout>
        <Grid item xs={12}>
          <TableWrapper
            itemActions={tableActions}
            columns={[
              {
                label: "Name",
                elementKey: "name",
              },
              {
                label: "Namespace",
                elementKey: "namespace",
                width: 90,
              },
              {
                label: "Status",
                elementKey: "status",
                width: 120,
              },
              {
                label: "Tenant",
                renderFullObject: true,
                renderFunction: (record: any) =>
                  `${record.namespace}/${record.tenant}`,
              },
              {
                label: "Capacity",
                elementKey: "capacity",
                width: 90,
              },
              {
                label: "Storage Class",
                elementKey: "storageClass",
              },
            ]}
            isLoading={loading}
            records={filteredRecords}
            entityName="PVCs"
            idField="name"
            customPaperHeight={classes.tableWrapper}
          />
        </Grid>
      </PageLayout>
    </Fragment>
  );
};

export default withStyles(styles)(StorageVolumes);
