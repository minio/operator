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
import { DropdownSelector, SelectorType } from "mds";
import styled from "styled-components";
import get from "lodash/get";

interface IInputUnitBox {
  id: string;
  unitSelected: string;
  unitsList: SelectorType[];
  disabled?: boolean;
  onUnitChange?: (newValue: string) => void;
}

const UnitMenuButton = styled.button(({ theme }) => ({
  border: `1px solid ${get(theme, "borderColor", "#E2E2E2")}`,
  borderRadius: 3,
  color: get(theme, "secondaryText", "#5B5C5C"),
  backgroundColor: get(theme, "boxBackground", "#FBFAFA"),
  fontSize: 12,
}));

const InputUnitMenu = ({
  id,
  unitSelected,
  unitsList,
  disabled = false,
  onUnitChange,
}: IInputUnitBox) => {
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);
  const handleClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };
  const handleClose = (newUnit: string) => {
    setAnchorEl(null);
    if (newUnit !== "" && onUnitChange) {
      onUnitChange(newUnit);
    }
  };

  return (
    <Fragment>
      <UnitMenuButton
        id={`${id}-button`}
        aria-controls={`${id}-menu`}
        aria-haspopup="true"
        aria-expanded={open ? "true" : undefined}
        onClick={handleClick}
        disabled={disabled}
        type={"button"}
      >
        {unitSelected}
      </UnitMenuButton>
      <DropdownSelector
        id={"upload-main-menu"}
        options={unitsList}
        selectedOption={""}
        onSelect={(value) => handleClose(value)}
        hideTriggerAction={() => {
          handleClose("");
        }}
        open={open}
        anchorEl={anchorEl}
        anchorOrigin={"end"}
      />
    </Fragment>
  );
};

export default InputUnitMenu;
