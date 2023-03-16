//  This file is part of MinIO Console Server
//  Copyright (c) 2022 MinIO, Inc.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

import { IMenuItem } from "./Menu/types";
import { NavLink } from "react-router-dom";
import {
  CONSOLE_UI_RESOURCE,
  IAM_PAGES,
  IAM_PAGES_PERMISSIONS,
} from "../../common/SecureComponent/permissions";
import {
  DocumentationIcon,
  LicenseIcon,
  RegisterMenuIcon,
  TenantsOutlineIcon,
} from "mds";
import { hasPermission } from "../../common/SecureComponent";
import React from "react";

export const validRoutes = (
  features: string[] | null | undefined,
  operatorMode: boolean
) => {
  let operatorMenus: IMenuItem[] = [
    {
      group: "Operator",
      type: "item",
      id: "Tenants",
      component: NavLink,
      to: IAM_PAGES.TENANTS,
      name: "Tenants",
      icon: TenantsOutlineIcon,
      forceDisplay: true,
    },
    {
      group: "Operator",
      type: "item",
      id: "License",
      component: NavLink,
      to: IAM_PAGES.LICENSE,
      name: "License",
      icon: LicenseIcon,
      forceDisplay: true,
    },
    {
      group: "Operator",
      type: "item",
      id: "Register",
      component: NavLink,
      to: IAM_PAGES.REGISTER_SUPPORT,
      name: "Register",
      icon: RegisterMenuIcon,
      forceDisplay: true,
    },
    {
      group: "Operator",
      type: "item",
      id: "Documentation",
      component: NavLink,
      to: IAM_PAGES.DOCUMENTATION,
      name: "Documentation",
      icon: DocumentationIcon,
      forceDisplay: true,
      onClick: (
        e:
          | React.MouseEvent<HTMLLIElement>
          | React.MouseEvent<HTMLAnchorElement>
          | React.MouseEvent<HTMLDivElement>
      ) => {
        e.preventDefault();
        window.open(
          "https://min.io/docs/minio/linux/index.html?ref=op",
          "_blank"
        );
      },
    },
  ];

  const allowedItems = operatorMenus.filter((item: IMenuItem) => {
    if (item.children && item.children.length > 0) {
      const c = item.children?.filter((childItem: IMenuItem) => {
        return (
          ((childItem.customPermissionFnc
            ? childItem.customPermissionFnc()
            : hasPermission(
                CONSOLE_UI_RESOURCE,
                IAM_PAGES_PERMISSIONS[childItem.to ?? ""]
              )) ||
            childItem.forceDisplay) &&
          !childItem.fsHidden
        );
      });
      return c.length > 0;
    }

    const res =
      ((item.customPermissionFnc
        ? item.customPermissionFnc()
        : hasPermission(
            CONSOLE_UI_RESOURCE,
            IAM_PAGES_PERMISSIONS[item.to ?? ""]
          )) ||
        item.forceDisplay) &&
      !item.fsHidden;
    return res;
  });
  return allowedItems;
};
