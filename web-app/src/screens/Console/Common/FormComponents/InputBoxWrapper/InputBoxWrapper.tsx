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
import React, { ClipboardEvent, useState } from "react";
import {
  Grid,
  IconButton,
  InputLabel,
  TextField,
  TextFieldProps,
  Tooltip,
} from "@mui/material";
import { OutlinedInputProps } from "@mui/material/OutlinedInput";
import { InputProps as StandardInputProps } from "@mui/material/Input";
import VisibilityOffIcon from "@mui/icons-material/VisibilityOff";
import RemoveRedEyeIcon from "@mui/icons-material/RemoveRedEye";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import makeStyles from "@mui/styles/makeStyles";
import withStyles from "@mui/styles/withStyles";
import {
  fieldBasic,
  inputFieldStyles,
  tooltipHelper,
} from "../common/styleLibrary";
import { HelpIcon } from "mds";
import clsx from "clsx";

interface InputBoxProps {
  label: string;
  classes: any;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onKeyPress?: (e: any) => void;
  onFocus?: () => void;
  onPaste?: (e: ClipboardEvent<HTMLInputElement>) => void;
  value: string | boolean;
  id: string;
  name: string;
  disabled?: boolean;
  multiline?: boolean;
  type?: string;
  tooltip?: string;
  autoComplete?: string;
  index?: number;
  error?: string;
  required?: boolean;
  placeholder?: string;
  min?: string;
  max?: string;
  overlayId?: string;
  overlayIcon?: any;
  overlayAction?: () => void;
  overlayObject?: any;
  extraInputProps?: StandardInputProps["inputProps"];
  noLabelMinWidth?: boolean;
  pattern?: string;
  autoFocus?: boolean;
  className?: string;
}

const styles = (theme: Theme) =>
  createStyles({
    ...fieldBasic,
    ...tooltipHelper,
    textBoxContainer: {
      flexGrow: 1,
      position: "relative",
    },
    overlayAction: {
      position: "absolute",
      right: 5,
      top: 6,
      "& svg": {
        maxWidth: 15,
        maxHeight: 15,
      },
      "&.withLabel": {
        top: 5,
      },
    },
  });

const inputStyles = makeStyles((theme: Theme) =>
  createStyles({
    ...inputFieldStyles,
  }),
);

function InputField(props: TextFieldProps) {
  const classes = inputStyles();

  return (
    <TextField
      InputProps={{ classes } as Partial<OutlinedInputProps>}
      {...props}
    />
  );
}

const InputBoxWrapper = ({
  label,
  onChange,
  value,
  id,
  name,
  type = "text",
  autoComplete = "off",
  disabled = false,
  multiline = false,
  tooltip = "",
  index = 0,
  error = "",
  required = false,
  placeholder = "",
  min,
  max,
  overlayId,
  overlayIcon = null,
  overlayObject = null,
  extraInputProps = {},
  overlayAction,
  noLabelMinWidth = false,
  pattern = "",
  autoFocus = false,
  classes,
  className = "",
  onKeyPress,
  onFocus,
  onPaste,
}: InputBoxProps) => {
  let inputProps: any = { "data-index": index, ...extraInputProps };
  const [toggleTextInput, setToggleTextInput] = useState<boolean>(false);

  if (type === "number" && min) {
    inputProps["min"] = min;
  }

  if (type === "number" && max) {
    inputProps["max"] = max;
  }

  if (pattern !== "") {
    inputProps["pattern"] = pattern;
  }

  let inputBoxWrapperIcon = overlayIcon;
  let inputBoxWrapperType = type;

  if (type === "password" && overlayIcon === null) {
    inputBoxWrapperIcon = toggleTextInput ? (
      <VisibilityOffIcon />
    ) : (
      <RemoveRedEyeIcon />
    );
    inputBoxWrapperType = toggleTextInput ? "text" : "password";
  }

  return (
    <React.Fragment>
      <Grid
        container
        className={clsx(
          className !== "" ? className : "",
          error !== "" ? classes.errorInField : classes.inputBoxContainer,
        )}
      >
        {label !== "" && (
          <InputLabel
            htmlFor={id}
            className={
              noLabelMinWidth ? classes.noMinWidthLabel : classes.inputLabel
            }
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

        <div className={classes.textBoxContainer}>
          <InputField
            id={id}
            name={name}
            fullWidth
            value={value}
            autoFocus={autoFocus}
            disabled={disabled}
            onChange={onChange}
            type={inputBoxWrapperType}
            multiline={multiline}
            autoComplete={autoComplete}
            inputProps={inputProps}
            error={error !== ""}
            helperText={error}
            placeholder={placeholder}
            className={classes.inputRebase}
            onKeyPress={onKeyPress}
            onFocus={onFocus}
            onPaste={onPaste}
          />
          {inputBoxWrapperIcon && (
            <div
              className={`${classes.overlayAction} ${
                label !== "" ? "withLabel" : ""
              }`}
            >
              <IconButton
                onClick={
                  overlayAction
                    ? () => {
                        overlayAction();
                      }
                    : () => setToggleTextInput(!toggleTextInput)
                }
                id={overlayId}
                size={"small"}
                disableFocusRipple={false}
                disableRipple={false}
                disableTouchRipple={false}
              >
                {inputBoxWrapperIcon}
              </IconButton>
            </div>
          )}
          {overlayObject && (
            <div
              className={`${classes.overlayAction} ${
                label !== "" ? "withLabel" : ""
              }`}
            >
              {overlayObject}
            </div>
          )}
        </div>
      </Grid>
    </React.Fragment>
  );
};

export default withStyles(styles)(InputBoxWrapper);
