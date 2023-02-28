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

import makeStyles from "@mui/styles/makeStyles";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import { TextFieldProps } from "@mui/material";
import TextField from "@mui/material/TextField";
import React from "react";

const inputStyles = makeStyles((theme: Theme) =>
  createStyles({
    root: {
      "& .MuiOutlinedInput-root": {
        paddingLeft: 0,
        "& svg": {
          marginLeft: 4,
          height: 14,
          color: theme.palette.primary.main,
        },
        "& input": {
          padding: 10,
          fontSize: 14,
          paddingLeft: 0,
          "&::placeholder": {
            fontSize: 12,
          },
          "@media (max-width: 900px)": {
            padding: 10,
          },
        },
        "& fieldset": {},

        "& fieldset:hover": {
          borderBottom: "2px solid #000000",
          borderRadius: 0,
        },
      },
    },
  })
);

export const LoginField = (props: TextFieldProps) => {
  const classes = inputStyles();

  return (
    <TextField
      classes={{
        root: classes.root,
      }}
      variant="standard"
      {...props}
    />
  );
};
