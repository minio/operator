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
import { InputBox, SearchIcon } from "mds";
import { CSSObject } from "styled-components";

type SearchBoxProps = {
  placeholder?: string;
  value: string;
  onChange: (value: string) => void;
  overrideClass?: any;
  id?: string;
  label?: string;
  sx?: CSSObject;
};

const SearchBox = ({
  placeholder = "",
  onChange,
  overrideClass,
  value,
  id = "search-resource",
  label = "",
  sx,
}: SearchBoxProps) => {
  return (
    <InputBox
      placeholder={placeholder}
      className={overrideClass ? overrideClass : ""}
      id={id}
      label={label}
      onChange={(e) => {
        onChange(e.target.value);
      }}
      value={value}
      startIcon={<SearchIcon />}
      sx={sx}
    />
  );
};

export default SearchBox;
