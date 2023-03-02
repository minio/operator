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
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";

interface IWarningMessage {
  classes: any;
  label: any;
  title: any;
}

const styles = (theme: Theme) =>
  createStyles({
    headerContainer: {
      backgroundColor: "#e78794",
      borderRadius: 3,
      marginBottom: 20,
      padding: 1,
      paddingBottom: 15,
    },
    labelHeadline: {
      color: "#000000",
      fontSize: 14,
      marginLeft: 20,
    },
    labelText: {
      color: "#000000",
      fontSize: 14,
      marginLeft: 20,
      marginRight: 40,
    },
  });

const WarningMessage = ({ classes, label, title }: IWarningMessage) => {
  return (
    <div className={classes.headerContainer}>
      <h4 className={classes.labelHeadline}>{title}</h4>
      <div className={classes.labelText}>{label}</div>
    </div>
  );
};

export default withStyles(styles)(WarningMessage);
