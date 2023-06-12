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
import { AppState, useAppDispatch } from "../../../store";
import { validRoutes } from "../valid-routes";
import { menuOpen } from "../../../systemSlice";
import { getLogoVar } from "../../../config";
import { Menu } from "mds";
import { useLocation, useNavigate } from "react-router-dom";

const AppMenu = () => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const { pathname } = useLocation();
  let logoPlan = getLogoVar();

  const sidebarOpen = useSelector(
    (state: AppState) => state.system.sidebarOpen
  );

  const routes = validRoutes();

  return (
    <Menu
      isOpen={sidebarOpen}
      applicationLogo={{
        applicationName: "operator",
        subVariant: logoPlan,
      }}
      callPathAction={(path) => navigate(path)}
      collapseAction={() => dispatch(menuOpen(!sidebarOpen))}
      currentPath={pathname}
      displayGroupTitles
      options={routes}
      signOutAction={() => navigate("/logout")}
    />
  );
};

export default AppMenu;
