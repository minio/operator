// This file is part of MinIO Operator
// Copyright (c) 2023 MinIO, Inc.
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
import { Button, CallHomeMenuIcon, CircleIcon, Grid } from "mds";
import { LinearProgress } from "@mui/material";
import api from "../../../common/api";
import { ICallHomeResponse } from "./types";
import { ErrorResponseHandler } from "../../../common/types";
import { setErrorSnackMessage, setSnackBarMessage } from "../../../systemSlice";
import { useAppDispatch } from "../../../store";
import ModalWrapper from "../Common/ModalWrapper/ModalWrapper";

interface ICallHomeConfirmation {
  onClose: (refresh: boolean) => any;
  open: boolean;
  diagStatus: boolean;
  logsStatus: boolean;
  disable?: boolean;
}

const CallHomeConfirmation = ({
  onClose,
  diagStatus,
  logsStatus,
  open,
  disable = false,
}: ICallHomeConfirmation) => {
  const dispatch = useAppDispatch();

  const [loading, setLoading] = useState<boolean>(false);

  const onConfirmAction = () => {
    setLoading(true);
    api
      .invoke("PUT", `/api/v1/support/callhome`, {
        diagState: disable ? false : diagStatus,
        logsState: disable ? false : logsStatus,
      })
      .then((res: ICallHomeResponse) => {
        dispatch(setSnackBarMessage("Configuration saved successfully"));
        setLoading(false);
        onClose(true);
      })
      .catch((err: ErrorResponseHandler) => {
        setLoading(false);
        dispatch(setErrorSnackMessage(err));
      });
  };

  return (
    <ModalWrapper
      modalOpen={open}
      title={disable ? "Disable Call Home" : "Edit Call Home Configurations"}
      onClose={() => onClose(false)}
      titleIcon={<CallHomeMenuIcon />}
    >
      {disable ? (
        <Fragment>
          Please Acknowledge that after doing this action, we will no longer
          receive updated cluster information automatically, losing the
          potential benefits that Call Home provides to your MinIO cluster.
          <Grid item xs={12} sx={{ margin: "15px 0" }}>
            Are you sure you want to disable SUBNET Call Home?
          </Grid>
          <br />
          {loading && (
            <Grid
              item
              xs={12}
              sx={{
                marginBottom: 10,
              }}
            >
              <LinearProgress />
            </Grid>
          )}
          <Grid
            item
            xs={12}
            sx={{
              display: "flex",
              justifyContent: "flex-end",
            }}
          >
            <Button
              id={"reset"}
              type="button"
              variant="regular"
              disabled={loading}
              onClick={() => onClose(false)}
              label={"Cancel"}
              sx={{
                marginRight: 10,
              }}
            />
            <Button
              id={"save-lifecycle"}
              type="submit"
              variant={"secondary"}
              color="primary"
              disabled={loading}
              label={"Yes, Disable Call Home"}
              onClick={onConfirmAction}
            />
          </Grid>
        </Fragment>
      ) : (
        <Fragment>
          Are you sure you want to change the following configurations for
          SUBNET Call Home:
          <Grid
            item
            sx={{
              margin: "20px 0",
              display: "flex",
              flexDirection: "column",
              gap: 15,
            }}
          >
            <Grid item sx={{ display: "flex", alignItems: "center", gap: 10 }}>
              <CircleIcon
                style={{ fill: diagStatus ? "#4CCB92" : "#C83B51", width: 20 }}
              />
              <span>
                <strong>{diagStatus ? "Enable" : "Disable"}</strong> - Send
                Diagnostics Information to SUBNET
              </span>
            </Grid>
            <Grid item sx={{ display: "flex", alignItems: "center", gap: 10 }}>
              <CircleIcon
                style={{ fill: logsStatus ? "#4CCB92" : "#C83B51", width: 20 }}
              />
              <span>
                <strong>{logsStatus ? "Enable" : "Disable"}</strong> - Send Logs
                Information to SUBNET
              </span>
            </Grid>
          </Grid>
          <Grid item xs={12} sx={{ margin: "15px 0" }}>
            Please Acknowledge that the information provided will only be
            available in your SUBNET Account and it will not be shared to other
            persons or entities besides MinIO team and you.
          </Grid>
          {loading && (
            <Grid
              item
              xs={12}
              sx={{
                marginBottom: 10,
              }}
            >
              <LinearProgress />
            </Grid>
          )}
          <Grid
            item
            xs={12}
            sx={{
              display: "flex",
              justifyContent: "flex-end",
            }}
          >
            <Button
              id={"reset"}
              type="button"
              variant="regular"
              disabled={loading}
              onClick={() => onClose(false)}
              label={"Cancel"}
              sx={{
                marginRight: 10,
              }}
            />
            <Button
              id={"save-lifecycle"}
              type="submit"
              variant={"callAction"}
              color="primary"
              disabled={loading}
              label={"Yes, Save this Configuration"}
              onClick={onConfirmAction}
            />
          </Grid>
        </Fragment>
      )}
    </ModalWrapper>
  );
};

export default CallHomeConfirmation;
