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
import { useSelector } from "react-redux";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { Grid, InputAdornment, TextField } from "@mui/material";
import { AppState, useAppDispatch } from "../../../store";
import {
  actionsTray,
  containerForHeader,
  searchField,
} from "../Common/FormComponents/common/styleLibrary";
import { IDirectPVVolumes, IVolumesResponse } from "./types";
import { niceBytes } from "../../../common/utils";
import { ErrorResponseHandler } from "../../../common/types";
import { setErrorSnackMessage } from "../../../systemSlice";
import api from "../../../common/api";
import TableWrapper from "../Common/TableWrapper/TableWrapper";
import { SearchIcon } from "mds";
import PageLayout from "../Common/Layout/PageLayout";
import PageHeaderWrapper from "../Common/PageHeaderWrapper/PageHeaderWrapper";

interface IDirectPVVolumesProps {
  classes: any;
}

const styles = (theme: Theme) =>
  createStyles({
    tableWrapper: {
      height: "calc(100vh - 267px)",
    },
    ...actionsTray,
    ...searchField,
    ...containerForHeader,
  });

const DirectPVVolumes = ({ classes }: IDirectPVVolumesProps) => {
  const dispatch = useAppDispatch();

  const selectedDrive = useSelector(
    (state: AppState) => state.directPV.selectedDrive
  );

  const [records, setRecords] = useState<IDirectPVVolumes[]>([]);
  const [filter, setFilter] = useState("");
  const [loading, setLoading] = useState<boolean>(true);

  useEffect(() => {
    if (loading) {
      api
        .invoke("GET", `/api/v1/directpv/volumes?drives=${selectedDrive}`)
        .then((res: IVolumesResponse) => {
          let volumes = get(res, "volumes", []);

          if (!volumes) {
            volumes = [];
          }

          volumes.sort((d1, d2) => {
            if (d1.volume > d2.volume) {
              return 1;
            }

            if (d1.volume < d2.volume) {
              return -1;
            }

            return 0;
          });

          setRecords(volumes);
          setLoading(false);
        })
        .catch((err: ErrorResponseHandler) => {
          setLoading(false);
          dispatch(setErrorSnackMessage(err));
        });
    }
  }, [loading, selectedDrive, dispatch]);

  const filteredRecords: IDirectPVVolumes[] = records.filter((elementItem) =>
    elementItem.drive.includes(filter)
  );

  return (
    <Fragment>
      <PageHeaderWrapper label="Volumes" />
      <PageLayout>
        <Grid item xs={12} className={classes.actionsTray}>
          <TextField
            placeholder="Search Volumes"
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
        <Grid item xs={12}>
          <br />
        </Grid>
        <Grid item xs={12}>
          <TableWrapper
            itemActions={[]}
            columns={[
              {
                label: "Volume",
                elementKey: "volume",
              },
              {
                label: "Capacity",
                elementKey: "capacity",
                renderFunction: niceBytes,
              },
              {
                label: "Node",
                elementKey: "node",
              },
              {
                label: "Drive",
                elementKey: "drive",
              },
            ]}
            isLoading={loading}
            records={filteredRecords}
            entityName="Volumes"
            idField="volume"
            customPaperHeight={classes.tableWrapper}
          />
        </Grid>
      </PageLayout>
    </Fragment>
  );
};

export default withStyles(styles)(DirectPVVolumes);
