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

import React, { Fragment } from "react";

interface IInformationItemProps {
  label: string;
  value: string;
  unit?: string;
  variant?: "normal" | "faded";
}

const InformationItem = ({
  label,
  value,
  unit,
  variant = "normal",
}: IInformationItemProps) => {
  return (
    <div style={{ margin: "0px 20px" }}>
      <div style={{ textAlign: "center" }}>
        <span
          style={{
            fontSize: 18,
            color: variant === "normal" ? "#000" : "#999",
            fontWeight: 400,
          }}
        >
          {value}
        </span>
        {unit && (
          <Fragment>
            {" "}
            <span
              style={{ fontSize: 12, color: "#8F9090", fontWeight: "bold" }}
            >
              {unit}
            </span>
          </Fragment>
        )}
      </div>
      <div
        style={{
          textAlign: "center",
          color: variant === "normal" ? "#767676" : "#bababa",
          fontSize: 12,
          whiteSpace: "nowrap",
        }}
      >
        {label}
      </div>
    </div>
  );
};

export default InformationItem;
