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
import IconButton from "@mui/material/IconButton";
import Snackbar from "@mui/material/Snackbar";
import { Dialog, DialogContent, DialogTitle } from "@mui/material";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import {
  deleteDialogStyles,
  snackBarCommon,
} from "../FormComponents/common/styleLibrary";
import { AppState, useAppDispatch } from "../../../../store";
import CloseIcon from "@mui/icons-material/Close";
import MainError from "../MainError/MainError";
import { setModalSnackMessage } from "../../../../systemSlice";

interface IModalProps {
  classes: any;
  onClose: () => void;
  modalOpen: boolean;
  title: string | React.ReactNode;
  children: any;
  wideLimit?: boolean;
  noContentPadding?: boolean;
  titleIcon?: React.ReactNode;
}

const styles = (theme: Theme) =>
  createStyles({
    ...deleteDialogStyles,
    content: {
      padding: 25,
      paddingBottom: 0,
    },
    customDialogSize: {
      width: "100%",
      maxWidth: 765,
    },
    ...snackBarCommon,
  });

const ModalWrapper = ({
  onClose,
  modalOpen,
  title,
  children,
  classes,
  wideLimit = true,
  noContentPadding,
  titleIcon = null,
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

  const customSize = wideLimit
    ? {
        classes: {
          paper: classes.customDialogSize,
        },
      }
    : { maxWidth: "lg" as const, fullWidth: true };

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
    <Dialog
      open={modalOpen}
      classes={classes}
      {...customSize}
      scroll={"paper"}
      onClose={(event, reason) => {
        if (reason !== "backdropClick") {
          onClose(); // close on Esc but not on click outside
        }
      }}
      className={classes.root}
    >
      <DialogTitle className={classes.title}>
        <div className={classes.titleText}>
          {titleIcon} {title}
        </div>
        <div className={classes.closeContainer}>
          <IconButton
            aria-label="close"
            id={"close"}
            className={classes.closeButton}
            onClick={onClose}
            disableRipple
            size="small"
          >
            <CloseIcon />
          </IconButton>
        </div>
      </DialogTitle>

      <MainError isModal={true} />
      <Snackbar
        open={openSnackbar}
        className={classes.snackBarModal}
        onClose={() => {
          closeSnackBar();
        }}
        message={message}
        ContentProps={{
          className: `${classes.snackBar} ${
            modalSnackMessage && modalSnackMessage.type === "error"
              ? classes.errorSnackBar
              : ""
          }`,
        }}
        autoHideDuration={
          modalSnackMessage && modalSnackMessage.type === "error" ? 10000 : 5000
        }
      />
      <DialogContent className={noContentPadding ? "" : classes.content}>
        {children}
      </DialogContent>
    </Dialog>
  );
};

export default withStyles(styles)(ModalWrapper);
