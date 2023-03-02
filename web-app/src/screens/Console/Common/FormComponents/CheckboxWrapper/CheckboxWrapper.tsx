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
import React from "react";
import { Checkbox, Grid, InputLabel, Tooltip } from "@mui/material";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import {
  checkboxIcons,
  fieldBasic,
  tooltipHelper,
} from "../common/styleLibrary";
import { HelpIcon } from "mds";

interface CheckBoxProps {
  label: string;
  classes: any;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  value: string | boolean;
  id: string;
  name: string;
  disabled?: boolean;
  tooltip?: string;
  overrideLabelClasses?: string;
  overrideCheckboxStyles?: React.CSSProperties;
  index?: number;
  noTopMargin?: boolean;
  checked: boolean;
  className?: string;
}

const styles = (theme: Theme) =>
  createStyles({
    ...fieldBasic,
    ...tooltipHelper,
    ...checkboxIcons,
    fieldContainer: {
      ...fieldBasic.fieldContainer,
      display: "flex",
      justifyContent: "flex-start",
      alignItems: "center",
      margin: "15px 0",
      marginBottom: 0,
      flexBasis: "initial",
      flexWrap: "nowrap",
    },
    noTopMargin: {
      marginTop: 0,
    },
  });

const CheckboxWrapper = ({
  label,
  onChange,
  value,
  id,
  name,
  checked = false,
  disabled = false,
  noTopMargin = false,
  tooltip = "",
  overrideLabelClasses = "",
  overrideCheckboxStyles,
  classes,
  className,
}: CheckBoxProps) => {
  return (
    <React.Fragment>
      <Grid
        item
        xs={12}
        className={`${classes.fieldContainer} ${
          noTopMargin ? classes.noTopMargin : ""
        } ${className ? className : ""}`}
      >
        <div>
          <Checkbox
            name={name}
            id={id}
            value={value}
            color="primary"
            inputProps={{ "aria-label": "secondary checkbox" }}
            checked={checked}
            onChange={onChange}
            checkedIcon={<span className={classes.checkedIcon} />}
            icon={<span className={classes.unCheckedIcon} />}
            disabled={disabled}
            disableRipple
            disableFocusRipple
            focusRipple={false}
            centerRipple={false}
            disableTouchRipple
            style={overrideCheckboxStyles || {}}
          />
        </div>
        {label !== "" && (
          <InputLabel
            htmlFor={id}
            className={`${classes.noMinWidthLabel} ${overrideLabelClasses}`}
          >
            <span>{label}</span>
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
      </Grid>
    </React.Fragment>
  );
};

export default withStyles(styles)(CheckboxWrapper);
