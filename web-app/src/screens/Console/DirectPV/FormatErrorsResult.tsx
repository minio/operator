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

import React from "react";
import { Grid, Theme } from "@mui/material";
import { Button, DriveFormatErrorsIcon } from "mds";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import ModalWrapper from "../Common/ModalWrapper/ModalWrapper";
import TableWrapper from "../Common/TableWrapper/TableWrapper";
import { IDirectPVFormatResItem } from "./types";
import { modalStyleUtils } from "../Common/FormComponents/common/styleLibrary";
import { encodeURLString } from "../../../common/utils";

interface IFormatErrorsProps {
  open: boolean;
  onCloseFormatErrorsList: () => void;
  errorsList: IDirectPVFormatResItem[];
  classes: any;
}

const styles = (theme: Theme) =>
  createStyles({
    errorsList: {
      height: "calc(100vh - 280px)",
    },
    ...modalStyleUtils,
  });

const download = (filename: string, text: string) => {
  let element = document.createElement("a");
  element.setAttribute(
    "href",
    "data:application/json;charset=utf-8," + encodeURLString(text)
  );
  element.setAttribute("download", filename);

  element.style.display = "none";
  document.body.appendChild(element);

  element.click();

  document.body.removeChild(element);
};

const FormatErrorsResult = ({
  open,
  onCloseFormatErrorsList,
  errorsList,
  classes,
}: IFormatErrorsProps) => {
  return (
    <ModalWrapper
      modalOpen={open}
      title={"Format Errors"}
      onClose={onCloseFormatErrorsList}
      titleIcon={<DriveFormatErrorsIcon />}
    >
      <Grid container>
        <Grid item xs={12} className={classes.modalFormScrollable}>
          There were some issues trying to format the selected CSI Drives,
          please fix the issues and try again.
          <br />
          <TableWrapper
            columns={[
              {
                label: "Node",
                elementKey: "node",
              },
              { label: "Drive", elementKey: "drive" },
              { label: "Message", elementKey: "error" },
            ]}
            entityName="Format Errors"
            idField="drive"
            records={errorsList}
            isLoading={false}
            customPaperHeight={classes.errorsList}
            textSelectable
            noBackground
          />
        </Grid>
        <Grid item xs={12} className={classes.modalButtonBar}>
          <Button
            id={"download-results"}
            variant="regular"
            onClick={() => {
              download("csiFormatErrors.json", JSON.stringify([...errorsList]));
            }}
            label={"Download"}
          />
          <Button
            id={"finish"}
            onClick={onCloseFormatErrorsList}
            variant="callAction"
            label={"Donw"}
          />
        </Grid>
      </Grid>
    </ModalWrapper>
  );
};

export default withStyles(styles)(FormatErrorsResult);
