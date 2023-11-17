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
import React, { useEffect, useState } from "react";
import { useSelector } from "react-redux";
import { ModalBox, Snackbar } from "mds";
import { CSSObject } from "styled-components";
import { AppState, useAppDispatch } from "../../../../store";
import { setModalSnackMessage } from "../../../../systemSlice";
import MainError from "../MainError/MainError";

interface IModalProps {
  onClose: () => void;
  modalOpen: boolean;
  title: string | React.ReactNode;
  children: any;
  wideLimit?: boolean;
  titleIcon?: React.ReactNode;
  iconColor?: "default" | "delete" | "accept";
  sx?: CSSObject;
}

const ModalWrapper = ({
  onClose,
  modalOpen,
  title,
  children,
  wideLimit = true,
  titleIcon = null,
  iconColor = "default",
  sx,
}: IModalProps) => {
  const dispatch = useAppDispatch();
  const [openSnackbar, setOpenSnackbar] = useState<boolean>(false);

  const modalSnackMessage = useSelector(
    (state: AppState) => state.system.modalSnackBar,
  );

  useEffect(() => {
    dispatch(setModalSnackMessage(""));
  }, [dispatch]);

  useEffect(() => {
    if (modalSnackMessage) {
      if (modalSnackMessage.message === "") {
        setOpenSnackbar(false);
        return;
      }
      // Open SnackBar
      if (modalSnackMessage.type !== "error") {
        setOpenSnackbar(true);
      }
    }
  }, [modalSnackMessage]);

  const closeSnackBar = () => {
    setOpenSnackbar(false);
    dispatch(setModalSnackMessage(""));
  };

  let message = "";

  if (modalSnackMessage) {
    message = modalSnackMessage.detailedErrorMsg;
    if (
      modalSnackMessage.detailedErrorMsg === "" ||
      modalSnackMessage.detailedErrorMsg.length < 5
    ) {
      message = modalSnackMessage.message;
    }
  }

  return (
    <ModalBox
      onClose={onClose}
      open={modalOpen}
      title={title}
      titleIcon={titleIcon}
      widthLimit={wideLimit}
      sx={sx}
      iconColor={iconColor}
    >
      <MainError isModal={true} />
      <Snackbar
        onClose={closeSnackBar}
        open={openSnackbar}
        message={message}
        mode={"inline"}
        variant={modalSnackMessage.type === "error" ? "error" : "default"}
        autoHideDuration={modalSnackMessage.type === "error" ? 10 : 5}
        condensed
      />
      {children}
    </ModalBox>
  );
};

export default ModalWrapper;
