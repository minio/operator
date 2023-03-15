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
import { useSelector } from "react-redux";
import { Drawer } from "@mui/material";
import withStyles from "@mui/styles/withStyles";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import clsx from "clsx";
import { AppState, useAppDispatch } from "../../../store";

import MenuToggle from "./MenuToggle";
import ConsoleMenuList from "./ConsoleMenuList";
import { validRoutes } from "../valid-routes";
import { menuOpen, selOpMode } from "../../../systemSlice";
import { selFeatures } from "../consoleSlice";

const drawerWidth = 250;

const styles = (theme: Theme) =>
  createStyles({
    drawer: {
      width: drawerWidth,
      flexShrink: 0,
      whiteSpace: "nowrap",
      background:
        "transparent linear-gradient(90deg, #073052 0%, #081C42 100%) 0% 0% no-repeat padding-box !important",
      boxShadow: "0px 3px 7px #00000014",
      "& .MuiPaper-root": {
        backgroundColor: "inherit",
      },
      "& ::-webkit-scrollbar": {
        width: "5px",
      },
      "& ::-webkit-scrollbar-track": {
        background: "#F0F0F0",
        borderRadius: 0,
        boxShadow: "inset 0px 0px 0px 0px #F0F0F0",
      },
      "& ::-webkit-scrollbar-thumb": {
        background: "#5A6375",
        borderRadius: 0,
      },
      "& ::-webkit-scrollbar-thumb:hover": {
        background: "#081C42",
      },
    },
    drawerOpen: {
      width: drawerWidth,
      transition: theme.transitions.create("width", {
        easing: theme.transitions.easing.sharp,
        duration: theme.transitions.duration.enteringScreen,
      }),
    },
    drawerClose: {
      transition: theme.transitions.create("width", {
        easing: theme.transitions.easing.sharp,
        duration: theme.transitions.duration.leavingScreen,
      }),
      overflowX: "hidden",
      [theme.breakpoints.up("xs")]: {
        width: 75,
      },
    },
  });

interface IMenuProps {
  classes?: any;
}

const Menu = ({ classes }: IMenuProps) => {
  const dispatch = useAppDispatch();
  const features = useSelector(selFeatures);

  const sidebarOpen = useSelector(
    (state: AppState) => state.system.sidebarOpen
  );
  const operatorMode = useSelector(selOpMode);

  const allowedMenuItems = validRoutes(features, operatorMode);

  return (
    <Drawer
      id="app-menu"
      variant="permanent"
      className={clsx(classes.drawer, {
        [classes.drawerOpen]: sidebarOpen,
        [classes.drawerClose]: !sidebarOpen,
      })}
      classes={{
        paper: clsx({
          [classes.drawerOpen]: sidebarOpen,
          [classes.drawerClose]: !sidebarOpen,
        }),
      }}
    >
      <MenuToggle
        onToggle={(nextState) => {
          dispatch(menuOpen(nextState));
        }}
        isOpen={sidebarOpen}
      />

      <ConsoleMenuList
        menuItems={allowedMenuItems}
        isOpen={sidebarOpen}
        displayHeaders={!operatorMode}
      />
    </Drawer>
  );
};

export default withStyles(styles)(Menu);
