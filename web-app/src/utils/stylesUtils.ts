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

import { IEmbeddedCustomStyles } from "../common/types";
import { createTheme } from "@mui/material";

export const getOverrideColorVariants: (
  customStyles: string,
) => false | IEmbeddedCustomStyles = (customStyles) => {
  try {
    return JSON.parse(atob(customStyles)) as IEmbeddedCustomStyles;
  } catch (e) {
    console.error("Error processing override styles, skipping.", e);
    return false;
  }
};

export const generateOverrideTheme = (overrideVars: IEmbeddedCustomStyles) => {
  const theme = createTheme({
    palette: {
      primary: {
        light: overrideVars.buttonStyles.hoverColor || "#073052",
        main: overrideVars.buttonStyles.backgroundColor || "#081C42",
        dark: overrideVars.buttonStyles.activeColor || "#05122B",
        contrastText: overrideVars.buttonStyles.textColor || "#fff",
      },
      secondary: {
        light: "#ff7961",
        main: "#f44336",
        dark: "#ba000d",
        contrastText: "#000",
      },
      background: {
        default: overrideVars.backgroundColor,
      },
      success: {
        main: "#4ccb92",
      },
      warning: {
        main: "#FFBD62",
      },
      error: {
        light: "#e03a48",
        main: "#C83B51",
        contrastText: "#fff",
      },
    },
    typography: {
      fontFamily: ["Inter", "sans-serif"].join(","),
      h1: {
        fontWeight: "bold",
        color: overrideVars.fontColor,
      },
      h2: {
        fontWeight: "bold",
        color: overrideVars.fontColor,
      },
      h3: {
        fontWeight: "bold",
        color: overrideVars.fontColor,
      },
      h4: {
        fontWeight: "bold",
        color: overrideVars.fontColor,
      },
      h5: {
        fontWeight: "bold",
        color: overrideVars.fontColor,
      },
      h6: {
        fontWeight: "bold",
        color: overrideVars.fontColor,
      },
    },
    components: {
      MuiButton: {
        styleOverrides: {
          root: {
            textTransform: "none",
            borderRadius: 3,
            height: 40,
            padding: "0 20px",
            fontSize: 14,
            fontWeight: 600,
            boxShadow: "none",
            "& .min-icon": {
              maxHeight: 18,
            },
            "&.MuiButton-contained.Mui-disabled": {
              backgroundColor: "#EAEDEE",
              fontWeight: 600,
              color: "#767676",
            },
            "& .MuiButton-iconSizeMedium > *:first-of-type": {
              fontSize: 12,
            },
          },
        },
      },
      MuiPaper: {
        styleOverrides: {
          root: {
            backgroundColor: overrideVars.backgroundColor,
            color: overrideVars.fontColor,
          },
          elevation1: {
            boxShadow: "none",
            border: "#EAEDEE 1px solid",
            borderRadius: 3,
          },
        },
      },
      MuiListItem: {
        styleOverrides: {
          root: {
            "&.MuiListItem-root.Mui-selected": {
              background: "inherit",
              "& .MuiTypography-root": {
                fontWeight: "bold",
              },
            },
          },
        },
      },
      MuiTab: {
        styleOverrides: {
          root: {
            textTransform: "none",
          },
        },
      },
    },
    colors: {
      link: "#2781B0",
    },
  });

  return theme;
};
