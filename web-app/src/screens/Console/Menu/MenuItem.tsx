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

import React, { Suspense, useCallback } from "react";
import {
  Collapse,
  ListItem,
  ListItemIcon,
  ListItemText,
  Tooltip,
} from "@mui/material";
import {
  menuItemContainerStyles,
  menuItemIconStyles,
  menuItemMiniStyles,
  menuItemStyle,
  menuItemTextStyles,
} from "./MenuStyleUtils";
import List from "@mui/material/List";
import { MenuCollapsedIcon, MenuExpandedIcon } from "mds";
import { hasPermission } from "../../../common/SecureComponent";
import {
  CONSOLE_UI_RESOURCE,
  IAM_PAGES_PERMISSIONS,
} from "../../../common/SecureComponent/permissions";

const MenuItem = ({
  page,
  stateClsName = "",
  onExpand,
  selectedMenuItem,
  pathValue = "",
  expandedGroup = "",
  setSelectedMenuItem,
  id = `${Math.random()}`,
  setPreviewMenuGroup,
  previewMenuGroup,
}: {
  page: any;
  stateClsName?: string;
  setSelectedMenuItem: (value: string) => void;
  selectedMenuItem?: any;
  pathValue?: string;
  onExpand: (id: any) => void;
  expandedGroup?: string;
  id?: string;
  setPreviewMenuGroup: (value: string) => void;
  previewMenuGroup: string;
}) => {
  const childrenMenuList = page?.children?.filter(
    (item: any) =>
      ((item.customPermissionFnc
        ? item.customPermissionFnc()
        : hasPermission(CONSOLE_UI_RESOURCE, IAM_PAGES_PERMISSIONS[item.to])) ||
        item.forceDisplay) &&
      !item.fsHidden,
  );

  let hasChildren = childrenMenuList?.length;

  const expandCollapseHandler = useCallback(
    (e: any) => {
      e.preventDefault();
      if (previewMenuGroup === page.id) {
        setPreviewMenuGroup("");
      } else if (page.id !== selectedMenuItem) {
        setPreviewMenuGroup(page.id);
        onExpand("");
      }

      if (page.id === selectedMenuItem && selectedMenuItem === expandedGroup) {
        onExpand(null);
      } else if (page.id === selectedMenuItem) {
        onExpand(selectedMenuItem);
        setPreviewMenuGroup("");
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [page, selectedMenuItem, previewMenuGroup, expandedGroup],
  );

  const selectMenuHandler = useCallback(
    (e: any) => {
      onExpand(page.id);
      setSelectedMenuItem(page.id);
      page.onClick && page.onClick(e);
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [page],
  );

  const onClickHandler = hasChildren
    ? expandCollapseHandler
    : selectMenuHandler;

  const isActiveGroup = expandedGroup === page.id;
  const activeClsName =
    pathValue.includes(selectedMenuItem) && page.id === selectedMenuItem
      ? "active"
      : "";

  return (
    <React.Fragment>
      <ListItem
        key={page.to}
        button
        onClick={(e: any) => {
          onClickHandler(e);
          setSelectedMenuItem(selectedMenuItem);
        }}
        component={page.component}
        to={page.to}
        id={id}
        className={`${activeClsName} ${stateClsName} main-menu-item `}
        disableRipple
        sx={{
          ...menuItemContainerStyles,
          marginTop: "5px",
          ...menuItemMiniStyles,

          "& .expanded-icon": {
            border: "1px solid #35393c",
          },
        }}
      >
        {page.icon && (
          <Tooltip title={`${page.name}`} placement="right">
            <ListItemIcon
              sx={{ ...menuItemIconStyles }}
              className={`${
                isActiveGroup && hasChildren ? "expanded-icon" : ""
              }`}
            >
              <Suspense fallback={<div>...</div>}>
                <page.icon />
              </Suspense>
            </ListItemIcon>
          </Tooltip>
        )}
        {page.name && (
          <ListItemText
            className={stateClsName}
            sx={{ ...menuItemTextStyles }}
            primary={page.name}
            secondary={page.badge ? <page.badge /> : null}
          />
        )}

        {hasChildren ? (
          isActiveGroup || previewMenuGroup === page.id ? (
            <MenuExpandedIcon
              height={15}
              width={15}
              className="group-icon"
              style={{ color: "#8399AB" }}
            />
          ) : (
            <MenuCollapsedIcon
              height={15}
              width={15}
              className="group-icon"
              style={{ color: "#8399AB" }}
            />
          )
        ) : null}
      </ListItem>

      {(isActiveGroup || previewMenuGroup === page.id) && hasChildren ? (
        <Collapse
          key={page.id}
          id={`${page.id}-children`}
          in={true}
          timeout="auto"
          unmountOnExit
        >
          <List
            component="div"
            disablePadding
            key={page.id}
            sx={{
              marginLeft: "15px",
              "&.mini": {
                marginLeft: "0px",
              },
            }}
            className={`${stateClsName}`}
          >
            {childrenMenuList.map((item: any) => {
              return (
                <ListItem
                  key={item.to}
                  button
                  component={item.component}
                  to={item.to}
                  onClick={(e: any) => {
                    if (page.id) {
                      setPreviewMenuGroup("");
                      setSelectedMenuItem(page.id);
                    }
                  }}
                  disableRipple
                  sx={{
                    ...menuItemStyle,
                    ...menuItemMiniStyles,
                  }}
                  className={`${stateClsName}`}
                >
                  {item.icon && (
                    <Tooltip title={`${item.name}`} placement="right">
                      <ListItemIcon
                        sx={{
                          background: "#00274D",
                          display: "flex",
                          alignItems: "center",
                          justifyContent: "center",

                          "& svg": {
                            fill: "#fff",
                            minWidth: "12px",
                            maxWidth: "12px",
                          },
                        }}
                        className="menu-icon"
                      >
                        <Suspense fallback={<div>...</div>}>
                          <item.icon />
                        </Suspense>
                        {item.badge ? <item.badge /> : null}
                      </ListItemIcon>
                    </Tooltip>
                  )}
                  {item.name && (
                    <ListItemText
                      className={stateClsName}
                      sx={{ ...menuItemTextStyles, marginLeft: "16px" }}
                      primary={item.name}
                    />
                  )}
                </ListItem>
              );
            })}
          </List>
        </Collapse>
      ) : null}
    </React.Fragment>
  );
};

export default MenuItem;
