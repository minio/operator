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

import React, { Fragment } from "react";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import TextField from "@mui/material/TextField";
import { searchField } from "../common/styleLibrary";

interface IFilterInputWrapper {
  classes: any;
  value: string;
  onChange: (txtVar: string) => any;
  label: string;
  placeholder?: string;
  id: string;
  name: string;
}

const styles = (theme: Theme) =>
  createStyles({
    searchField: {
      ...searchField.searchField,
      height: 30,
      padding: 0,
      "& input": {
        padding: "0 12px",
        height: 28,
        fontSize: 12,
        fontWeight: 600,
        color: "#393939",
      },
      "&.isDisabled": {
        "&:hover": {
          borderColor: "#EAEDEE",
        },
      },
      "& input.Mui-disabled": {
        backgroundColor: "#EAEAEA",
      },
    },
    labelStyle: {
      color: "#393939",
      fontSize: 12,
      marginBottom: 4,
    },
    buttonKit: {
      display: "flex",
      alignItems: "center",
    },
    fieldContainer: {
      flexGrow: 1,
      margin: "0 15px",
    },
  });

const FilterInputWrapper = ({
  classes,
  label,
  onChange,
  value,
  placeholder = "",
  id,
  name,
}: IFilterInputWrapper) => {
  return (
    <Fragment>
      <div className={classes.fieldContainer}>
        <div className={classes.labelStyle}>{label}</div>
        <div className={classes.buttonKit}>
          <TextField
            placeholder={placeholder}
            id={id}
            name={name}
            label=""
            onChange={(val) => {
              onChange(val.target.value);
            }}
            InputProps={{
              disableUnderline: true,
            }}
            className={classes.searchField}
            value={value}
          />
        </div>
      </div>
    </Fragment>
  );
};

export default withStyles(styles)(FilterInputWrapper);
