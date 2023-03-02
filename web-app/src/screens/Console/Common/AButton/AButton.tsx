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
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { IconButtonProps } from "@mui/material";

const styles = (theme: Theme) =>
  createStyles({
    root: {
      padding: 0,
      margin: 0,
      border: 0,
      backgroundColor: "transparent",
      textDecoration: "underline",
      cursor: "pointer",
      fontSize: "inherit",
      color: theme.palette.info.main,
      fontFamily: "Inter, sans-serif",
    },
  });

interface IAButton extends IconButtonProps {
  classes: any;
  children: any;
}

const AButton = ({ classes, children, ...rest }: IAButton) => {
  return (
    <button {...rest} className={classes.root}>
      {children}
    </button>
  );
};

export default withStyles(styles)(AButton);
