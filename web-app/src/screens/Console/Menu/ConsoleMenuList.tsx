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

import React, { Fragment, useEffect, useState } from "react";
import { Box } from "@mui/material";
import { useLocation } from "react-router-dom";
import ListItem from "@mui/material/ListItem";
import ListItemIcon from "@mui/material/ListItemIcon";
import { LogoutIcon } from "mds";
import ListItemText from "@mui/material/ListItemText";
import List from "@mui/material/List";
import {
  LogoutItemIconStyle,
  menuItemContainerStyles,
  menuItemMiniStyles,
  menuItemTextStyles,
} from "./MenuStyleUtils";
import MenuItem from "./MenuItem";

import { IAM_PAGES } from "../../../common/SecureComponent/permissions";
import MenuSectionHeader from "./MenuSectionHeader";

const ConsoleMenuList = ({
  menuItems,
  isOpen,
  displayHeaders = false,
}: {
  menuItems: any[];
  isOpen: boolean;
  displayHeaders?: boolean;
}) => {
  const stateClsName = isOpen ? "wide" : "mini";
  const { pathname = "" } = useLocation();
  let groupToSelect = pathname.slice(1, pathname.length); //single path
  if (groupToSelect.indexOf("/") !== -1) {
    groupToSelect = groupToSelect.slice(0, groupToSelect.indexOf("/")); //nested path
  }

  const [expandGroup, setExpandGroup] = useState(IAM_PAGES.BUCKETS);
  const [selectedMenuItem, setSelectedMenuItem] =
    useState<string>(groupToSelect);

  const [previewMenuGroup, setPreviewMenuGroup] = useState<string>("");

  useEffect(() => {
    setExpandGroup(groupToSelect);
    setSelectedMenuItem(groupToSelect);
  }, [groupToSelect]);

  let basename = document.baseURI.replace(window.location.origin, "");
  let header = "";

  return (
    <Box
      className={`${stateClsName} wrapper`}
      sx={{
        display: "flex",
        flexFlow: "column",
        justifyContent: "space-between",
        height: "100%",
        flex: 1,
        paddingRight: "8px",

        "&.wide": {
          marginLeft: "30px",
        },

        "&.mini": {
          marginLeft: "10px",
          "& .menuHeader": {
            display: "none",
          },
        },
      }}
    >
      <List
        sx={{
          flex: 1,
          paddingTop: 0,

          "&.mini": {
            padding: 0,
            display: "flex",
            alignItems: "center",
            flexFlow: "column",

            "& .main-menu-item": {
              marginBottom: "20px",
            },
          },
        }}
        className={`${stateClsName} group-wrapper main-list`}
      >
        <React.Fragment>
          {(menuItems || []).map((menuGroup: any, index) => {
            if (menuGroup) {
              let grHeader = null;

              if (menuGroup.group !== header && displayHeaders) {
                grHeader = <MenuSectionHeader label={menuGroup.group} />;
                header = menuGroup.group;
              }

              return (
                <Fragment key={`${menuGroup.id}-${index.toString()}`}>
                  {grHeader}
                  <MenuItem
                    stateClsName={stateClsName}
                    page={menuGroup}
                    id={menuGroup.id}
                    selectedMenuItem={selectedMenuItem}
                    setSelectedMenuItem={setSelectedMenuItem}
                    pathValue={pathname}
                    onExpand={setExpandGroup}
                    expandedGroup={expandGroup}
                    previewMenuGroup={previewMenuGroup}
                    setPreviewMenuGroup={setPreviewMenuGroup}
                  />
                </Fragment>
              );
            }
            return null;
          })}
        </React.Fragment>
      </List>
      {/* List of Bottom anchored menus */}
      <List
        sx={{
          paddingTop: 0,
          "&.mini": {
            padding: 0,
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            flexFlow: "column",
          },
        }}
        className={`${stateClsName} group-wrapper bottom-list`}
      >
        <ListItem
          button
          component="a"
          href={`${window.location.origin}${basename}logout`}
          disableRipple
          sx={{
            ...menuItemContainerStyles,
            ...menuItemMiniStyles,
            marginBottom: "3px",
          }}
          className={`$ ${stateClsName} bottom-menu-item`}
        >
          <ListItemIcon
            sx={{
              ...LogoutItemIconStyle,
            }}
          >
            <LogoutIcon />
          </ListItemIcon>
          <ListItemText
            primary="Sign Out"
            id={"logout"}
            sx={{ ...menuItemTextStyles }}
            className={stateClsName}
          />
        </ListItem>
      </List>
    </Box>
  );
};
export default ConsoleMenuList;
