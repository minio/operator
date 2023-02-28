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
import { Theme } from "@mui/material/styles";
import {
  AddIcon,
  Button,
  HelpBox,
  RefreshIcon,
  SearchIcon,
  StorageIcon,
} from "mds";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { Grid, InputAdornment, TextField } from "@mui/material";
import get from "lodash/get";
import GroupIcon from "@mui/icons-material/Group";
import {
  actionsTray,
  containerForHeader,
  searchField,
} from "../Common/FormComponents/common/styleLibrary";
import {
  IDirectPVDrives,
  IDirectPVFormatResItem,
  IDrivesResponse,
} from "./types";
import { niceBytes } from "../../../common/utils";
import { ErrorResponseHandler } from "../../../common/types";
import api from "../../../common/api";
import TableWrapper from "../Common/TableWrapper/TableWrapper";

import withSuspense from "../Common/Components/withSuspense";
import PageLayout from "../Common/Layout/PageLayout";
import PageHeaderWrapper from "../Common/PageHeaderWrapper/PageHeaderWrapper";

const FormatDrives = withSuspense(React.lazy(() => import("./FormatDrives")));
const FormatErrorsResult = withSuspense(
  React.lazy(() => import("./FormatErrorsResult"))
);

interface IDirectPVMain {
  classes: any;
}

const styles = (theme: Theme) =>
  createStyles({
    tableWrapper: {
      height: "calc(100vh - 275px)",
    },
    linkItem: {
      display: "default",
      color: theme.palette.info.main,
      textDecoration: "none",
      "&:hover": {
        textDecoration: "underline",
        color: "#000",
      },
    },
    ...actionsTray,
    ...searchField,
    ...containerForHeader,
  });

const DirectPVMain = ({ classes }: IDirectPVMain) => {
  const [records, setRecords] = useState<IDirectPVDrives[]>([]);
  const [filter, setFilter] = useState("");
  const [checkedDrives, setCheckedDrives] = useState<string[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [formatOpen, setFormatOpen] = useState<boolean>(false);
  const [formatAll, setFormatAll] = useState<boolean>(false);
  const [formatErrorsResult, setFormatErrorsResult] = useState<
    IDirectPVFormatResItem[]
  >([]);
  const [formatErrorsOpen, setFormatErrorsOpen] = useState<boolean>(false);
  const [drivesToFormat, setDrivesToFormat] = useState<string[]>([]);
  const [notAvailable, setNotAvailable] = useState<boolean>(true);

  useEffect(() => {
    if (loading) {
      api
        .invoke("GET", "/api/v1/directpv/drives")
        .then((res: IDrivesResponse) => {
          let drives: IDirectPVDrives[] = get(res, "drives", []);

          if (!drives) {
            drives = [];
          }

          drives = drives.map((item) => {
            const newItem = { ...item };
            newItem.joinName = `${newItem.node}:${newItem.drive}`;

            return newItem;
          });

          drives.sort((d1, d2) => {
            if (d1.drive > d2.drive) {
              return 1;
            }

            if (d1.drive < d2.drive) {
              return -1;
            }

            return 0;
          });

          setRecords(drives);
          setLoading(false);
          setNotAvailable(false);
        })
        .catch((err: ErrorResponseHandler) => {
          setLoading(false);
          setNotAvailable(true);
        });
    }
  }, [loading, notAvailable]);

  const formatAllDrives = () => {
    const allDrives = records.map((item) => {
      return `${item.node}:${item.drive}`;
    });
    setFormatAll(true);
    setDrivesToFormat(allDrives);
    setFormatOpen(true);
  };

  const formatSingleUnit = (driveID: string) => {
    const selectedUnit = [driveID];
    setDrivesToFormat(selectedUnit);
    setFormatAll(false);
    setFormatOpen(true);
  };

  const formatSelectedDrives = () => {
    if (checkedDrives.length > 0) {
      setDrivesToFormat(checkedDrives);
      setFormatAll(false);
      setFormatOpen(true);
    }
  };

  const selectionChanged = (e: React.ChangeEvent<HTMLInputElement>) => {
    const targetD = e.target;
    const value = targetD.value;
    const checked = targetD.checked;

    let elements: string[] = [...checkedDrives]; // We clone the checkedDrives array

    if (checked) {
      // If the user has checked this field we need to push this to checkedDrivesList
      elements.push(value);
    } else {
      // User has unchecked this field, we need to remove it from the list
      elements = elements.filter((element) => element !== value);
    }

    setCheckedDrives(elements);

    return elements;
  };

  const closeFormatModal = (
    refresh: boolean,
    errorsList: IDirectPVFormatResItem[]
  ) => {
    setFormatOpen(false);
    if (refresh) {
      // Errors are present, we trigger the modal box to show these changes.
      if (errorsList && errorsList.length > 0) {
        setFormatErrorsResult(errorsList);
        setFormatErrorsOpen(true);
      }
      setLoading(true);
      setCheckedDrives([]);
    }
  };

  const tableActions = [
    {
      type: "format",
      onClick: formatSingleUnit,
      sendOnlyId: true,
    },
  ];

  const filteredRecords: IDirectPVDrives[] = records.filter((elementItem) =>
    elementItem.drive.includes(filter)
  );

  return (
    <Fragment>
      {formatOpen && (
        <FormatDrives
          closeFormatModalAndRefresh={closeFormatModal}
          deleteOpen={formatOpen}
          allDrives={formatAll}
          drivesToFormat={drivesToFormat}
        />
      )}

      {formatErrorsOpen && (
        <FormatErrorsResult
          errorsList={formatErrorsResult}
          open={formatErrorsOpen}
          onCloseFormatErrorsList={() => {
            setFormatErrorsOpen(false);
          }}
        />
      )}
      <PageHeaderWrapper label="Local Drives" />
      <PageLayout>
        <Grid item xs={12} className={classes.actionsTray}>
          <TextField
            placeholder="Search Drives"
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
            disabled={notAvailable}
            variant="standard"
          />
          <Button
            id={"refresh-directpv-list"}
            color="primary"
            aria-label="Refresh DirectPV List"
            onClick={() => {
              setLoading(true);
            }}
            disabled={notAvailable}
            icon={<RefreshIcon />}
          />
          <Button
            id={"format-selected-drives"}
            variant="callAction"
            disabled={checkedDrives.length <= 0 || notAvailable}
            onClick={formatSelectedDrives}
            label={"Format Selected Drives"}
            icon={<GroupIcon />}
          />
          <Button
            id={"format-all-drives"}
            variant="callAction"
            label={"Format All Drives"}
            onClick={formatAllDrives}
            disabled={notAvailable}
            icon={<AddIcon />}
          />
        </Grid>

        <Grid item xs={12}>
          {notAvailable && !loading ? (
            <HelpBox
              title={"Leverage locally attached drives"}
              iconComponent={<StorageIcon />}
              help={
                <Fragment>
                  We can automatically provision persistent volumes (PVs) on top
                  locally attached drives on your Kubernetes nodes by leveraging
                  DirectPV.
                  <br />
                  <br />
                  For more information{" "}
                  <a
                    href="https://github.com/minio/directpv"
                    rel="noopener"
                    target="_blank"
                    className={classes.linkItem}
                  >
                    Visit DirectPV Documentation
                  </a>
                </Fragment>
              }
            />
          ) : (
            <TableWrapper
              itemActions={tableActions}
              columns={[
                {
                  label: "Drive",
                  elementKey: "drive",
                },
                {
                  label: "Capacity",
                  elementKey: "capacity",
                  renderFunction: niceBytes,
                },
                {
                  label: "Allocated",
                  elementKey: "allocated",
                  renderFunction: niceBytes,
                },
                {
                  label: "Volumes",
                  elementKey: "volumes",
                },
                {
                  label: "Node",
                  elementKey: "node",
                },
                {
                  label: "Status",
                  elementKey: "status",
                },
              ]}
              onSelect={selectionChanged}
              selectedItems={checkedDrives}
              isLoading={loading}
              records={filteredRecords}
              customPaperHeight={classes.tableWrapper}
              entityName="Drives"
              idField="joinName"
            />
          )}
        </Grid>
      </PageLayout>
    </Fragment>
  );
};

export default withStyles(styles)(DirectPVMain);
