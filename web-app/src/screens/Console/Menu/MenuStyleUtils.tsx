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

export const menuItemContainerStyles: any = {
  padding: "7px 0",
  "div:nth-of-type(2)": {
    flexGrow: 0,
    marginRight: "15px",
  },

  "&.active": {
    background:
      "transparent linear-gradient(270deg, #00000000 0%, #005F81 53%, #54545400 100%) 0% 0% no-repeat padding-box",
    backgroundBlendMode: "multiply",

    "& span": {
      color: "#fff",
    },
    "& svg": {
      fill: "#fff",
    },

    "& div:nth-of-type(1)": {
      border: "1px solid #fff",
    },

    "&:hover, &:focus": {
      "& div:nth-of-type(1)": {
        background: "none",
        "& svg": {
          fill: "#fff",
        },
      },
    },
  },
};
export const menuItemIconStyles: any = {
  width: 30,
  minWidth: 30,
  height: 30,
  background: "#00274D",
  border: "1px solid #002148",
  display: "flex",
  alignItems: "center",
  borderRadius: "50%",
  justifyContent: "center",
  "& svg": {
    width: 12,
    height: 12,
    fill: "#8399AB",
  },
  "&.active": {
    "& span": {
      color: "#fff",
    },
    "& svg": {
      fill: "#fff",
    },
  },
};

export const LogoutItemIconStyle: any = {
  width: 40,
  minWidth: 40,
  height: 40,
  background: "#00274D",
  border: "2px solid #002148",
  display: "flex",
  alignItems: "center",
  borderRadius: "50%",
  justifyContent: "center",
  "& svg": {
    width: 16,
    height: 16,
    fill: "#8399AB",
  },
  "&.active": {
    "& span": {
      color: "#fff",
    },
    "& svg": {
      fill: "#fff",
    },
  },
};

export const menuItemTextStyles: any = {
  color: "#8399AB",
  fontSize: "14px",
  marginLeft: "18px",
  display: "flex",
  position: "relative",

  "& span": {
    fontSize: "14px",
  },
  "&.mini": {
    display: "none",
  },
};

export const menuItemMiniStyles: any = {
  "&.mini": {
    padding: 0,
    display: "flex",
    alignItems: "center",
    justifyContent: "center",

    "& .group-icon": {
      display: "none",
    },

    "&.active": {
      ".menu-icon": {
        border: "none",
      },
    },
  },

  "&.bottom-menu-item": {
    marginBottom: "5px",
  },
};

export const menuItemStyle: any = {
  paddingLeft: "8px",
  paddingRight: "5px",
  paddingBottom: "8px",
  borderRadius: "2px",
  marginTop: "2px",
  "&.active": {
    ".menu-icon": {
      border: "1px solid #fff",
      borderRadius: "50%",
      background: "#072549",
      "& svg": {
        fill: "#fff",
      },
    },
    "& span": {
      color: "#fff",
    },
  },
  "& .menu-icon": {
    padding: "5px",
    maxWidth: "28px",
    minWidth: "28px",
    height: "28px",
    background: "none",
    "& svg": {
      width: "12px",
      height: "12px",
      fill: "#8399AB",
    },
  },
  "&:hover, &:focus": {
    "& .menu-icon": {
      background: "#072549",
      borderRadius: "50%",
      "& svg": {
        fill: "#c7c3c3",
      },
    },
  },
};
