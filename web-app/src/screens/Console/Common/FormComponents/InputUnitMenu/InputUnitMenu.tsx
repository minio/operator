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
import { selectorTypes } from "../SelectWrapper/SelectWrapper";
import { Menu, MenuItem } from "@mui/material";

interface IInputUnitBox {
  classes: any;
  id: string;
  unitSelected: string;
  unitsList: selectorTypes[];
  disabled?: boolean;
  onUnitChange?: (newValue: string) => void;
}

const styles = (theme: Theme) =>
  createStyles({
    buttonTrigger: {
      border: "#F0F2F2 1px solid",
      borderRadius: 3,
      color: "#838383",
      backgroundColor: "#fff",
      fontSize: 12,
    },
  });

const InputUnitMenu = ({
  classes,
  id,
  unitSelected,
  unitsList,
  disabled = false,
  onUnitChange,
}: IInputUnitBox) => {
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);
  const handleClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };
  const handleClose = (newUnit: string) => {
    setAnchorEl(null);
    if (newUnit !== "" && onUnitChange) {
      onUnitChange(newUnit);
    }
  };

  return (
    <Fragment>
      <button
        id={`${id}-button`}
        aria-controls={`${id}-menu`}
        aria-haspopup="true"
        aria-expanded={open ? "true" : undefined}
        onClick={handleClick}
        className={classes.buttonTrigger}
        disabled={disabled}
        type={"button"}
      >
        {unitSelected}
      </button>
      <Menu
        id={`${id}-menu`}
        aria-labelledby={`${id}-button`}
        anchorEl={anchorEl}
        open={open}
        onClose={() => {
          handleClose("");
        }}
        anchorOrigin={{
          vertical: "bottom",
          horizontal: "center",
        }}
        transformOrigin={{
          vertical: "top",
          horizontal: "center",
        }}
      >
        {unitsList.map((unit) => (
          <MenuItem
            onClick={() => handleClose(unit.value)}
            key={`itemUnit-${unit.value}-${unit.label}`}
          >
            {unit.label}
          </MenuItem>
        ))}
      </Menu>
    </Fragment>
  );
};

export default withStyles(styles)(InputUnitMenu);
