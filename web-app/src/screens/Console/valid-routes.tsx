//  This file is part of MinIO Operator
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

import { IAM_PAGES } from "../../common/SecureComponent/permissions";
import {
  DocumentationIcon,
  LicenseIcon,
  MenuItemProps,
  RegisterMenuIcon,
  TenantsOutlineIcon,
} from "mds";
import React from "react";

export const validRoutes = () => {
  let operatorMenus: MenuItemProps[] = [
    {
      group: "Operator",
      id: "Tenants",
      path: IAM_PAGES.TENANTS,
      name: "Tenants",
      icon: <TenantsOutlineIcon />,
    },
    {
      group: "Operator",
      id: "License",
      path: IAM_PAGES.LICENSE,
      name: "License",
      icon: <LicenseIcon />,
    },
    {
      group: "Operator",
      id: "Register",
      path: IAM_PAGES.REGISTER_SUPPORT,
      name: "Register",
      icon: <RegisterMenuIcon />,
    },
    {
      group: "Operator",
      id: "Documentation",
      path: IAM_PAGES.DOCUMENTATION,
      name: "Documentation",
      icon: <DocumentationIcon />,
      onClick: (path) => {
        window.open(
          "https://min.io/docs/minio/linux/index.html?ref=op",
          "_blank"
        );
      },
    },
  ];

  return operatorMenus;
};
