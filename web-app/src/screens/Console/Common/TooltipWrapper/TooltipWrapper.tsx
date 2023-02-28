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

import React, { cloneElement } from "react";
import { Tooltip } from "@mui/material";

interface ITooltipWrapperProps {
  tooltip: string;
  children: any;
  errorProps?: any;
  placement?:
    | "bottom-end"
    | "bottom-start"
    | "bottom"
    | "left-end"
    | "left-start"
    | "left"
    | "right-end"
    | "right-start"
    | "right"
    | "top-end"
    | "top-start"
    | "top";
}

const TooltipWrapper = ({
  tooltip,
  children,
  errorProps = null,
  placement,
}: ITooltipWrapperProps) => {
  return (
    <Tooltip title={tooltip} placement={placement}>
      <span>
        {errorProps ? cloneElement(children, { ...errorProps }) : children}
      </span>
    </Tooltip>
  );
};

export default TooltipWrapper;
