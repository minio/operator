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

import { Selector } from "testcafe";
import { IAM_PAGES } from "../../src/common/SecureComponent/permissions";

//----------------------------------------------------
// General sidebar element
//----------------------------------------------------
export const sidebarItem = Selector(".MuiPaper-root").find("ul").child("a");
export const logoutItem = Selector(".MuiPaper-root").find("ul").child("div");

//----------------------------------------------------
// Specific sidebar elements
//----------------------------------------------------
export const monitoringElement = Selector(".MuiPaper-root")
  .find("ul")
  .child("#tools");
export const monitoringChildren = Selector("#tools-children");
export const dashboardElement = monitoringChildren
  .find("a")
  .withAttribute("href", IAM_PAGES.DASHBOARD);
export const logsElement = monitoringChildren
  .find("a")
  .withAttribute("href", "/tools/logs");
export const traceElement = monitoringChildren
  .find("a")
  .withAttribute("href", "/tools/trace");
export const drivesElement = monitoringChildren
  .find("a")
  .withAttribute("href", "/tools/heal");
export const watchElement = monitoringChildren
  .find("a")
  .withAttribute("href", "/tools/watch");

export const bucketsElement = sidebarItem.withAttribute("href", "/buckets");

export const serviceAcctsElement = sidebarItem.withAttribute(
  "href",
  IAM_PAGES.ACCOUNT,
);

export const identityElement = Selector(".MuiPaper-root")
  .find("ul")
  .child("#identity");
export const identityChildren = Selector("#identity-children");

export const usersElement = identityChildren
  .find("a")
  .withAttribute("href", IAM_PAGES.USERS);
export const groupsElement = identityChildren
  .find("a")
  .withAttribute("href", IAM_PAGES.GROUPS);

export const iamPoliciesElement = sidebarItem.withAttribute(
  "href",
  IAM_PAGES.POLICIES,
);

export const configurationsElement = Selector(".MuiPaper-root")
  .find("ul")
  .child("#configurations");

export const notificationEndpointsElement = Selector(".MuiPaper-root")
  .find("ul")
  .child("#lambda");

export const tiersElement = Selector(".MuiPaper-root")
  .find("ul")
  .child("#tiers");

export const diagnosticsElement = Selector(".MuiPaper-root")
  .find("ul")
  .child("#diagnostics");
export const performanceElement = Selector(".MuiPaper-root")
  .find("ul")
  .child("#performance");
export const profileElement = Selector(".MuiPaper-root")
  .find("ul")
  .child("#profile");
export const inspectElement = sidebarItem.withAttribute(
  "href",
  "/support/inspect",
);

export const licenseElement = sidebarItem.withAttribute("href", "/license");
