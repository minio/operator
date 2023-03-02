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
import { Box } from "@mui/material";
import SectionTitle from "./SectionTitle";

type Props = {
  title: string;
  icon: React.ReactNode;
  helpbox?: React.ReactNode;
  children: React.ReactNode;
};

const FormLayout: React.FC<Props> = ({ children, title, helpbox, icon }) => {
  return (
    <Box
      sx={{
        display: "grid",
        padding: "25px",
        gap: "25px",
        gridTemplateColumns: {
          md: "2fr 1.2fr",
          xs: "1fr",
        },
        border: "1px solid #eaeaea",
      }}
    >
      <Box>
        <SectionTitle icon={icon}>{title}</SectionTitle>
        <Box sx={{ height: 16 }} />
        {children}
      </Box>

      {helpbox}
    </Box>
  );
};

export default FormLayout;
