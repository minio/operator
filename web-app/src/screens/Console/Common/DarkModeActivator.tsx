// This file is part of MinIO Console Server
// Copyright (c) 2023 MinIO, Inc.
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
import { Button, DarkModeIcon } from "mds";
import { useSelector } from "react-redux";
import { AppState, useAppDispatch } from "../../../store";
import { storeDarkMode } from "../../../utils/stylesUtils";
import { setDarkMode } from "../../../systemSlice";
import TooltipWrapper from "./TooltipWrapper/TooltipWrapper";

const DarkModeActivator = () => {
  const dispatch = useAppDispatch();

  const darkMode = useSelector((state: AppState) => state.system.darkMode);

  const darkModeActivator = () => {
    const currentStatus = !!darkMode;

    dispatch(setDarkMode(!currentStatus));
    storeDarkMode(!currentStatus ? "on" : "off");
  };

  return (
    <TooltipWrapper tooltip={`${darkMode ? "Light" : "Dark"} Mode`}>
      <Button
        id={"dark-mode-activator"}
        icon={<DarkModeIcon />}
        onClick={darkModeActivator}
      />
    </TooltipWrapper>
  );
};

export default DarkModeActivator;
