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
import isString from "lodash/isString";
import { Link } from "react-router-dom";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { IconButton, Tooltip } from "@mui/material";
import CloudIcon from "./TableActionIcons/CloudIcon";
import ConsoleIcon from "./TableActionIcons/ConsoleIcon";
import DisableIcon from "./TableActionIcons/DisableIcon";
import FormatDriveIcon from "./TableActionIcons/FormatDriveIcon";
import {
  DownloadIcon,
  EditIcon,
  IAMPoliciesIcon,
  PreviewIcon,
  ShareIcon,
  TrashIcon,
} from "mds";

const styles = () =>
  createStyles({
    spacing: {
      margin: "0 8px",
    },
    buttonDisabled: {
      "&.MuiButtonBase-root.Mui-disabled": {
        cursor: "not-allowed",
        filter: "grayscale(100%)",
        opacity: "30%",
      },
    },
  });

interface IActionButton {
  label?: string;
  type: string | React.ReactNode;
  onClick?: (id: string) => any;
  to?: string;
  valueToSend: any;
  selected: boolean;
  sendOnlyId?: boolean;
  idField: string;
  disabled: boolean;
  classes: any;
}

const defineIcon = (type: string, selected: boolean) => {
  switch (type) {
    case "view":
      return <PreviewIcon />;
    case "edit":
      return <EditIcon />;
    case "delete":
      return <TrashIcon />;
    case "description":
      return <IAMPoliciesIcon />;
    case "share":
      return <ShareIcon />;
    case "cloud":
      return <CloudIcon active={selected} />;
    case "console":
      return <ConsoleIcon active={selected} />;
    case "download":
      return <DownloadIcon />;
    case "disable":
      return <DisableIcon active={selected} />;
    case "format":
      return <FormatDriveIcon />;
    case "preview":
      return <PreviewIcon />;
  }

  return null;
};

const TableActionButton = ({
  type,
  onClick,
  valueToSend,
  idField,
  selected,
  to,
  sendOnlyId = false,
  disabled = false,
  classes,
  label,
}: IActionButton) => {
  const valueClick = sendOnlyId ? valueToSend[idField] : valueToSend;

  const icon = typeof type === "string" ? defineIcon(type, selected) : type;
  let buttonElement = (
    <IconButton
      aria-label={typeof type === "string" ? type : ""}
      size={"small"}
      className={`${classes.spacing} ${disabled ? classes.buttonDisabled : ""}`}
      disabled={disabled}
      onClick={
        onClick
          ? (e) => {
              e.stopPropagation();
              if (!disabled) {
                onClick(valueClick);
              } else {
                e.preventDefault();
              }
            }
          : () => null
      }
      sx={{
        width: "30px",
        height: "30px",
      }}
    >
      {icon}
    </IconButton>
  );

  if (label && label !== "") {
    buttonElement = <Tooltip title={label}>{buttonElement}</Tooltip>;
  }

  if (onClick) {
    return buttonElement;
  }

  if (isString(to)) {
    if (!disabled) {
      return (
        <Link
          to={`${to}/${valueClick}`}
          onClick={(e) => {
            e.stopPropagation();
          }}
        >
          {buttonElement}
        </Link>
      );
    }

    return buttonElement;
  }

  return null;
};

export default withStyles(styles)(TableActionButton);
