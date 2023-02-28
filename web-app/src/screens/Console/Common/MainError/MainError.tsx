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

import React, { useCallback, useEffect, useState } from "react";
import { useSelector } from "react-redux";
import get from "lodash/get";
import { AppState, useAppDispatch } from "../../../../store";
import { Box } from "@mui/material";
import { AlertCloseIcon } from "mds";
import { Portal } from "@mui/base";
import { setErrorSnackMessage } from "../../../../systemSlice";

interface IMainErrorProps {
  isModal?: boolean;
}

let timerI: any;

const startHideTimer = (callbackFunction: () => void) => {
  timerI = setInterval(callbackFunction, 10000);
};

const stopHideTimer = () => {
  clearInterval(timerI);
};

const MainError = ({ isModal = false }: IMainErrorProps) => {
  const dispatch = useAppDispatch();
  const snackBar = useSelector((state: AppState) => {
    return isModal ? state.system.modalSnackBar : state.system.snackBar;
  });
  const [displayErrorMsg, setDisplayErrorMsg] = useState<boolean>(false);

  const closeErrorMessage = useCallback(() => {
    setDisplayErrorMsg(false);
  }, []);

  useEffect(() => {
    if (!displayErrorMsg) {
      dispatch(setErrorSnackMessage({ detailedError: "", errorMessage: "" }));
      clearInterval(timerI);
    }
  }, [dispatch, displayErrorMsg]);

  useEffect(() => {
    if (snackBar.message !== "" && snackBar.type === "error") {
      //Error message received, we trigger the animation
      setDisplayErrorMsg(true);
      startHideTimer(closeErrorMessage);
    }
  }, [closeErrorMessage, snackBar.message, snackBar.type]);

  const message = get(snackBar, "message", "");
  const messageDetails = get(snackBar, "detailedErrorMsg", "");

  if (snackBar.type !== "error" || message === "") {
    return null;
  }

  return (
    <Portal>
      <Box
        sx={{
          "&.alert": {
            border: 0,
            left: 0,
            right: 0,
            top: 0,
            height: "75px",
            position: "fixed",
            color: "#ffffff",
            padding: "0 30px 0 30px",
            zIndex: 10000,
            display: "flex",
            justifyContent: "center",
            alignItems: "center",
            fontWeight: 600,
            backgroundColor: "#C72C48",
            opacity: 0,
            width: "100%",

            "&.show": {
              opacity: 1,
            },
          },
          "& .message-text": {
            flex: 2,
            fontSize: "14px",
            textAlign: {
              md: "center",
              xs: "left",
            },
          },

          "& .close-btn-container": {
            cursor: "pointer",
            border: 0,
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            height: "100%",
            marginLeft: {
              sm: "0px",
              xs: "10px",
            },

            "& .close-btn": {
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              height: "23px",
              width: "23px",
              borderRadius: "50%",

              border: 0,
              backgroundColor: "transparent",
              cursor: "pointer",

              "&:hover,&:focus": {
                border: 0,
                outline: 0,
                backgroundColor: "#ba0202",
              },
              "& .min-icon": {
                height: "11px",
                width: "11px",
                fill: "#ffffff",
              },
            },
          },
        }}
        onMouseOver={stopHideTimer}
        onMouseLeave={() => startHideTimer(closeErrorMessage)}
        className={`alert ${displayErrorMsg ? "show" : ""}`}
      >
        <div className="message-text">
          {messageDetails ? messageDetails : `${message}.`}
        </div>
        <div className="close-btn-container">
          <button className="close-btn" autoFocus onClick={closeErrorMessage}>
            <AlertCloseIcon />
          </button>
        </div>
      </Box>
    </Portal>
  );
};

export default MainError;
