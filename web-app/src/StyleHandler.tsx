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
import { useSelector } from "react-redux";
import { GlobalStyles, ThemeHandler } from "mds";
import { AppState } from "./store";

import "./index.css";
import "react-virtualized/styles.css";

interface IStyleHandler {
  children: React.ReactNode;
}

const StyleHandler = ({ children }: IStyleHandler) => {
  const darkMode = useSelector((state: AppState) => state.system.darkMode);

  return (
    <Fragment>
      <GlobalStyles />
      <ThemeHandler darkMode={darkMode}>{children}</ThemeHandler>
    </Fragment>
  );
};

export default StyleHandler;
