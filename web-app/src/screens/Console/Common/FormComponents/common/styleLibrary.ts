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

// This object contains variables that will be used across form components.

export const containerForHeader = {
  container: {
    position: "relative" as const,
    padding: "20px 35px 0",
    "& h6": {
      color: "#777777",
      fontSize: 30,
    },
    "& p": {
      "& span:not(*[class*='smallUnit'])": {
        fontSize: 16,
      },
    },
  },
  sectionTitle: {
    margin: 0,
    marginBottom: ".8rem",
    fontSize: "1.3rem",
  },
  boxy: {
    border: "#E5E5E5 1px solid",
    borderRadius: 2,
    padding: 40,
    backgroundColor: "#fff",
  },
};

export const actionsTray = {
  label: {
    color: "#07193E",
    fontSize: 13,
    alignSelf: "center" as const,
    whiteSpace: "nowrap" as const,
    "&:not(:first-of-type)": {
      marginLeft: 10,
    },
  },
  actionsTray: {
    display: "flex" as const,
    justifyContent: "space-between" as const,
    marginBottom: "1rem",
    alignItems: "center",
    "& button": {
      flexGrow: 0,
      marginLeft: 8,
    },
  },
};

export const searchField = {
  searchField: {
    flexGrow: 1,
    height: 38,
    background: "#FFFFFF",
    borderRadius: 3,
    border: "#EAEDEE 1px solid",
    display: "flex",
    justifyContent: "center",
    padding: "0 16px",
    "& label, & label.MuiInputLabel-shrink": {
      fontSize: 10,
      transform: "translate(5px, 2px)",
      transformOrigin: "top left",
    },
    "& input": {
      fontSize: 12,
      fontWeight: 700,
      color: "#000",
      "&::placeholder": {
        color: "#858585",
        opacity: 1,
        fontWeight: 400,
      },
    },
    "&:hover": {
      borderColor: "#000",
    },
    "& .min-icon": {
      width: 16,
      height: 16,
    },
    "&:focus-within": {
      borderColor: "rgba(0, 0, 0, 0.87)",
    },
  },
};

export const snackBarCommon = {
  snackBar: {
    backgroundColor: "#081F44",
    fontWeight: 400,
    fontFamily: "Inter, sans-serif",
    fontSize: 14,
    boxShadow: "none" as const,
    "&.MuiPaper-root.MuiSnackbarContent-root": {
      borderRadius: "0px 0px 5px 5px",
    },
    "& div": {
      textAlign: "center" as const,
      padding: "6px 30px",
      width: "100%",
      overflowX: "hidden",
      textOverflow: "ellipsis",
    },
    "&.MuiPaper-root": {
      padding: "0px 20px 0px 20px",
    },
  },
  errorSnackBar: {
    backgroundColor: "#C72C48",
    color: "#fff",
  },
  snackBarExternal: {
    top: -1,
    height: 33,
    position: "fixed" as const,
    minWidth: 348,
    whiteSpace: "nowrap" as const,
    left: 0,
    width: "100%",
    justifyContent: "center" as const,
  },
  snackDiv: {
    top: "17px",
    left: "50%",
    position: "absolute" as const,
  },
  snackBarModal: {
    top: 0,
    position: "absolute" as const,
    minWidth: "348px",
    whiteSpace: "nowrap" as const,
    height: "33px",
    width: "100%",
    justifyContent: "center",
    left: 0,
  },
};

export const spacingUtils: any = {
  spacerRight: {
    marginRight: ".9rem",
  },
  spacerLeft: {
    marginLeft: ".9rem",
  },
  spacerBottom: {
    marginBottom: ".9rem",
  },
  spacerTop: {
    marginTop: ".9rem",
  },
};

export const modalStyleUtils: any = {
  modalButtonBar: {
    marginTop: 15,
    display: "flex",
    alignItems: "center",
    justifyContent: "flex-end",

    "& button": {
      marginRight: 10,
    },
    "& button:last-child": {
      marginRight: 0,
    },
  },
  modalFormScrollable: {
    maxHeight: "calc(100vh - 300px)",
    overflowY: "auto",
    paddingTop: 10,
  },
};
