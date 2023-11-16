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
import { Snackbar } from "mds";
import { useSelector } from "react-redux";
import get from "lodash/get";
import { AppState, useAppDispatch } from "../../../../store";
import {
  setErrorSnackMessage,
  setModalSnackMessage,
} from "../../../../systemSlice";

interface IMainErrorProps {
  isModal?: boolean;
}

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
      dispatch(setModalSnackMessage(""));
    }
  }, [dispatch, displayErrorMsg]);

  useEffect(() => {
    if (snackBar.message !== "" && snackBar.type === "error") {
      //Error message received, we trigger the animation
      setDisplayErrorMsg(true);
    }
  }, [closeErrorMessage, snackBar.message, snackBar.type]);

  const message = get(snackBar, "message", "");
  const messageDetails = get(snackBar, "detailedErrorMsg", "");

  if (snackBar.type !== "error" || message === "") {
    return null;
  }

  return (
    <Snackbar
      onClose={closeErrorMessage}
      open={displayErrorMsg}
      variant={"error"}
      message={messageDetails ? messageDetails : `${message}.`}
      autoHideDuration={10}
      closeButton
    />
  );
};

export default MainError;
