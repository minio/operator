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

export interface ISizeBarItem {
  value: number;
  itemName: string;
  color: string;
}

export interface IUsageBar {
  totalValue: number;
  sizeItems: ISizeBarItem[];
  bgColor?: string;
}

const UsageBar = ({
  totalValue,
  sizeItems,
  bgColor = "#ededed",
}: IUsageBar) => {
  return (
    <div
      style={{
        width: "100%",
        height: 12,
        backgroundColor: bgColor,
        borderRadius: 30,
        display: "flex",
        transitionDuration: "0.3s",
        overflow: "hidden",
      }}
    >
      {sizeItems.map((sizeElement, index) => {
        const itemPercentage = (sizeElement.value * 100) / totalValue;
        return (
          <div
            key={`itemSize-${index.toString()}`}
            style={{
              width: `${itemPercentage}%`,
              height: "100%",
              backgroundColor: sizeElement.color,
              transitionDuration: "0.3s",
            }}
          />
        );
      })}
    </div>
  );
};

export default UsageBar;
