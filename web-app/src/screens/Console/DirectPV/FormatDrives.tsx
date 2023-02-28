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

import React, { Fragment, useState } from "react";
import { DialogContentText, Grid, LinearProgress } from "@mui/material";
import { IDirectPVFormatResItem, IDirectPVFormatResult } from "./types";
import { ErrorResponseHandler } from "../../../common/types";
import api from "../../../common/api";
import InputBoxWrapper from "../Common/FormComponents/InputBoxWrapper/InputBoxWrapper";
import PredefinedList from "../Common/FormComponents/PredefinedList/PredefinedList";
import FormSwitchWrapper from "../Common/FormComponents/FormSwitchWrapper/FormSwitchWrapper";
import ConfirmDialog from "../Common/ModalWrapper/ConfirmDialog";
import { FormatDrivesIcon } from "mds";
import { setErrorSnackMessage } from "../../../systemSlice";
import { useAppDispatch } from "../../../store";

interface IFormatAllDrivesProps {
  closeFormatModalAndRefresh: (
    refresh: boolean,
    formatIssuesList: IDirectPVFormatResItem[]
  ) => void;
  deleteOpen: boolean;
  allDrives: boolean;
  drivesToFormat: string[];
}

const FormatDrives = ({
  closeFormatModalAndRefresh,
  deleteOpen,
  allDrives,
  drivesToFormat,
}: IFormatAllDrivesProps) => {
  const dispatch = useAppDispatch();
  const [deleteLoading, setDeleteLoading] = useState<boolean>(false);
  const [formatAll, setFormatAll] = useState<string>("");
  const [force, setForce] = useState<boolean>(false);

  const removeRecord = () => {
    if (deleteLoading) {
      return;
    }
    setDeleteLoading(true);
    api
      .invoke("POST", `/api/v1/directpv/drives/format`, {
        drives: drivesToFormat,
        force,
      })
      .then((res: IDirectPVFormatResult) => {
        setDeleteLoading(false);
        closeFormatModalAndRefresh(true, res.formatIssuesList);
      })
      .catch((err: ErrorResponseHandler) => {
        setDeleteLoading(false);
        dispatch(setErrorSnackMessage(err));
      });
  };
  return (
    <ConfirmDialog
      title={`Format ${allDrives ? "All " : ""} Drives`}
      confirmText={`Format Drive${
        drivesToFormat.length > 1 || allDrives ? "s" : ""
      }`}
      confirmButtonProps={{
        disabled: formatAll !== "YES, PROCEED",
      }}
      isOpen={deleteOpen}
      isLoading={deleteLoading}
      onConfirm={removeRecord}
      onClose={() => {
        closeFormatModalAndRefresh(false, []);
      }}
      titleIcon={<FormatDrivesIcon />}
      confirmationContent={
        <React.Fragment>
          <DialogContentText>
            {!allDrives && (
              <Fragment>
                <PredefinedList
                  label={`Selected Drive${
                    drivesToFormat.length > 1 ? "s" : ""
                  }`}
                  content={drivesToFormat.join(", ")}
                />
                <br />
              </Fragment>
            )}
            <Grid item xs={12}>
              <FormSwitchWrapper
                value="force"
                id="force"
                name="force"
                checked={force}
                onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                  setForce(event.target.checked);
                }}
                label={"Force Format"}
                indicatorLabels={["Yes", "No"]}
              />
            </Grid>
            Are you sure you want to format{" "}
            {allDrives ? <strong>All</strong> : "the selected"} drive
            {drivesToFormat.length > 1 || allDrives ? "s" : ""}?.
            <br />
            <br />
            <strong>
              All information contained will be erased and cannot be recovered
            </strong>
            <br />
            <br />
            To continue please type <b>YES, PROCEED</b> in the box.
            <Grid item xs={12}>
              <InputBoxWrapper
                id="format-confirm"
                name="format-confirm"
                onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                  setFormatAll(event.target.value);
                }}
                label=""
                value={formatAll}
              />
            </Grid>
          </DialogContentText>
          {deleteLoading && <LinearProgress />}
        </React.Fragment>
      }
    />
  );
};

export default FormatDrives;
