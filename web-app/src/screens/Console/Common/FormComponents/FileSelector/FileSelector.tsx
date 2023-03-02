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

import React, { useState } from "react";
import get from "lodash/get";
import { Grid, InputLabel, Tooltip } from "@mui/material";
import IconButton from "@mui/material/IconButton";
import AttachFileIcon from "@mui/icons-material/AttachFile";
import CancelIcon from "@mui/icons-material/Cancel";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import {
  fieldBasic,
  fileInputStyles,
  tooltipHelper,
} from "../common/styleLibrary";
import { fileProcess } from "./utils";
import { HelpIcon } from "mds";
import ErrorBlock from "../../../../shared/ErrorBlock";

interface InputBoxProps {
  label: string;
  classes: any;
  onChange: (e: string, i: string) => void;
  id: string;
  name: string;
  disabled?: boolean;
  tooltip?: string;
  required?: boolean;
  error?: string;
  accept?: string;
  value?: string;
}

const styles = (theme: Theme) =>
  createStyles({
    ...fieldBasic,
    ...tooltipHelper,
    valueString: {
      maxWidth: 350,
      whiteSpace: "nowrap",
      overflow: "hidden",
      textOverflow: "ellipsis",
      marginTop: 2,
    },
    fileInputField: {
      margin: "13px 0",
      "@media (max-width: 900px)": {
        flexFlow: "column",
      },
    },
    ...fileInputStyles,
    inputLabel: {
      ...fieldBasic.inputLabel,
      fontWeight: "normal",
    },
    textBoxContainer: {
      ...fieldBasic.textBoxContainer,
      maxWidth: "100%",
      border: "1px solid #eaeaea",
      paddingLeft: "15px",
    },
  });

const FileSelector = ({
  label,
  classes,
  onChange,
  id,
  name,
  disabled = false,
  tooltip = "",
  required,
  error = "",
  accept = "",
  value = "",
}: InputBoxProps) => {
  const [showFileSelector, setShowSelector] = useState(false);

  return (
    <React.Fragment>
      <Grid
        item
        xs={12}
        className={`${classes.fileInputField} ${classes.fieldBottom} ${
          classes.fieldContainer
        } ${error !== "" ? classes.errorInField : ""}`}
      >
        {label !== "" && (
          <InputLabel
            htmlFor={id}
            className={`${error !== "" ? classes.fieldLabelError : ""} ${
              classes.inputLabel
            }`}
          >
            <span>
              {label}
              {required ? "*" : ""}
            </span>
            {tooltip !== "" && (
              <div className={classes.tooltipContainer}>
                <Tooltip title={tooltip} placement="top-start">
                  <div className={classes.tooltip}>
                    <HelpIcon />
                  </div>
                </Tooltip>
              </div>
            )}
          </InputLabel>
        )}

        {showFileSelector || value === "" ? (
          <div className={classes.textBoxContainer}>
            <input
              type="file"
              name={name}
              onChange={(e) => {
                const fileName = get(e, "target.files[0].name", "");
                fileProcess(e, (data: any) => {
                  onChange(data, fileName);
                });
              }}
              accept={accept}
              required={required}
              disabled={disabled}
              className={classes.fileInputField}
            />

            {value !== "" && (
              <IconButton
                color="primary"
                aria-label="upload picture"
                component="span"
                onClick={() => {
                  setShowSelector(false);
                }}
                disableRipple={false}
                disableFocusRipple={false}
                size="small"
              >
                <CancelIcon />
              </IconButton>
            )}

            {error !== "" && <ErrorBlock errorMessage={error} />}
          </div>
        ) : (
          <div className={classes.fileReselect}>
            <div className={classes.valueString}>{value}</div>
            <IconButton
              color="primary"
              aria-label="upload picture"
              component="span"
              onClick={() => {
                setShowSelector(true);
              }}
              disableRipple={false}
              disableFocusRipple={false}
              size="small"
            >
              <AttachFileIcon />
            </IconButton>
          </div>
        )}
      </Grid>
    </React.Fragment>
  );
};

export default withStyles(styles)(FileSelector);
