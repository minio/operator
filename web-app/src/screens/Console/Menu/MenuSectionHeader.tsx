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

interface IMenuSectionHeader {
  label: string;
}

const MenuSectionHeader = ({ label }: IMenuSectionHeader) => {
  return (
    <div
      style={{
        fontSize: 15,
        color: "#b3c9d9",
        marginTop: 20,
        marginBottom: 10,
        borderBottom: "#0F446C 1px solid",
        marginRight: 25,
        paddingBottom: 5,
      }}
      className={"menuHeader"}
    >
      {label}
    </div>
  );
};

export default MenuSectionHeader;
