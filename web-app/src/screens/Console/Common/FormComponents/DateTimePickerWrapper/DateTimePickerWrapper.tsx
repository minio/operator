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
import { DateTime } from "luxon";
import { HelpIcon, OpenListIcon } from "mds";
import { Grid, InputLabel, TextField, Tooltip } from "@mui/material";
import { DateTimePicker, LocalizationProvider } from "@mui/x-date-pickers";
import { AdapterLuxon } from "@mui/x-date-pickers/AdapterLuxon";
import InputAdornment from "@mui/material/InputAdornment";
import ScheduleIcon from "@mui/icons-material/Schedule";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { fieldBasic, tooltipHelper } from "../common/styleLibrary";

interface IDateTimePicker {
  value: DateTime | null;
  onChange: (value: DateTime | null) => void;
  classes: any;
  forSearchBlock?: boolean;
  forFilterContained?: boolean;
  label?: string;
  required?: boolean;
  tooltip?: string;
  id: string;
  disabled?: boolean;
  noInputIcon?: boolean;
  classNamePrefix?: string;
  openPickerIcon?: any;
}

const styles = (theme: Theme) =>
  createStyles({
    dateSelectorOverride: {
      height: 40,
      border: "#EAEDEE 1px solid",
      marginLeft: 15,
      backgroundColor: "#fff",
      padding: "0 16px",
      borderRadius: 5,
      "&.MuiInput-underline:hover:not(.Mui-disabled):before": {
        borderBottom: 0,
      },
      "&:hover": {
        borderColor: "#000",
        "&:before, &:after": {
          borderColor: "transparent",
          borderBottom: 0,
        },
      },
      "&:before, &:after": {
        borderColor: "transparent",
        borderBottom: 0,
      },
      "& input": {
        fontSize: 12,
        fontWeight: 600,
        color: "#393939",
      },
    },
    dateSelectorFilterOverride: {
      width: 180,
      height: 42,
      marginLeft: 20,
      padding: 0,
      borderRadius: 5,
      "&.MuiInput-underline:hover:not(.Mui-disabled):before": {
        borderBottom: 0,
      },
      "&:hover": {
        "&:before, &:after": {
          borderColor: "transparent",
          borderBottom: 0,
        },
      },
      "&:before, &:after": {
        borderColor: "transparent",
        borderBottom: 0,
      },
      "& input": {
        fontSize: 12,
        fontWeight: "bold",
        color: "#081C42",
      },
      "@media (max-width: 900px)": {
        width: 103,
      },
    },
    dateSelectorFormOverride: {
      width: "100%",
      maxWidth: 840,
    },
    parentDateOverride: {
      flexGrow: 1,
    },
    textBoxContainer: {
      flexGrow: 1,
    },
    openListIcon: {
      color: "#9D9E9D",
      width: 8,
      marginTop: 2,
    },
    paperOverride: {
      "& .MuiCalendarPicker-root": {
        padding: "0 22px",
        "& > div": {
          padding: 0,
          "& > div > div.PrivatePickersFadeTransitionGroup-root:first-of-type":
            {
              color: "#0A224C",
              fontWeight: "bold",
            },
          "& > div > div.PrivatePickersFadeTransitionGroup-root:last-of-type": {
            color: "#9D9E9D",
            "& +  button.MuiButtonBase-root": {
              color: "#9D9E9D",
            },
          },
          "& > div:nth-child(2)": {
            "& > div": {
              width: 0,
            },
            "& > button.MuiButtonBase-root": {
              color: "#0A224C",
            },
          },
        },
      },
      "& .MuiTypography-root.MuiTypography-caption": {
        width: 26,
        height: 26,
        margin: 5,
        color: "#BCBCBC",
        fontSize: 10,
      },
      "& button.MuiPickersDay-root": {
        fontWeight: "bold",
        width: 26,
        height: 26,
        margin: 5,
        textAlign: "center",
        "&.Mui-selected": {
          backgroundColor: theme.palette.primary.main,
          color: "#fff",
        },
      },
      "& div.MuiPickersDay-hiddenDaySpacingFiller": {
        width: 26,
        height: 26,
        margin: 5,
      },
      "& div.PrivatePickersSlideTransition-root": {
        minHeight: 240,
      },
      "& div.MuiCalendarPicker-viewTransitionContainer": {
        borderTop: "#F0F3F5 1px solid",
      },
      "& .MuiClockPicker-arrowSwitcher": {
        marginRight: 10,
        marginTop: -1,
        "& > div": {
          width: 0,
        },
        "& > button.MuiButtonBase-root": {
          color: "#0A224C",
        },
        "& + div > div": {
          width: 255,
          height: 255,
          backgroundColor: "#fff",
          marginTop: 30,
          marginBottom: 14,
          border: "#F0F3F5 3px solid",
          "& > div:nth-child(2)": {
            backgroundColor: "#B4B5B4",
            width: 12,
            height: 12,
            "&::before": {
              content: "' '",
              width: 35,
              height: 35,
              display: "block",
              position: "absolute",
              border: "#F0F3F5 3px solid",
              top: -12,
              left: -12,
              borderRadius: "100%",
            },
          },
          "& > div:nth-child(3)": {
            backgroundColor: "#B4B5B4",
            width: 4,
          },
          "& > div:last-of-type": {
            marginTop: 15,
            "& > span": {
              color: "#0A224C",
              "&.Mui-selected": {
                color: "#fff",
              },
              "&[aria-label='1 hours'], &[aria-label='2 hours'], &[aria-label='3 hours'], &[aria-label='4 hours'], &[aria-label='5 hours'], &[aria-label='6 hours'], &[aria-label='7 hours'], &[aria-label='8 hours'], &[aria-label='9 hours'], &[aria-label='10 hours'], &[aria-label='11 hours'], &[aria-label='12 hours']":
                {
                  fontWeight: "bold",
                  fontSize: 20,
                  marginTop: -1,
                },
            },
          },
        },
      },
    },
    ...fieldBasic,
    ...tooltipHelper,
  });

const DateTimePickerWrapper = ({
  value,
  onChange,
  classes,
  forSearchBlock = false,
  forFilterContained = false,
  label,
  tooltip = "",
  required,
  id,
  disabled = false,
  noInputIcon = false,
  classNamePrefix = "",
  openPickerIcon,
}: IDateTimePicker) => {
  let adornment = {};

  if (!noInputIcon) {
    adornment = {
      startAdornment: (
        <InputAdornment position="start">
          <ScheduleIcon />
        </InputAdornment>
      ),
    };
  }

  if (forFilterContained) {
    adornment = {
      endAdornment: (
        <InputAdornment position="end">
          <OpenListIcon className={classes.openListIcon} />
        </InputAdornment>
      ),
    };
  }

  const classOverridden = `${classNamePrefix}date-time-input  ${
    forSearchBlock ? classes.dateSelectorOverride : ""
  } ${
    forFilterContained && !forSearchBlock
      ? classes.dateSelectorFilterOverride
      : ""
  }`;

  const clsName = forSearchBlock
    ? classes.parentDateOverride
    : classes.dateSelectorFormOverride;

  const inputItem = (
    <LocalizationProvider dateAdapter={AdapterLuxon}>
      <DateTimePicker
        value={value}
        onChange={onChange}
        InputProps={{
          ...adornment,
          className: classOverridden,
        }}
        components={{
          OpenPickerIcon: openPickerIcon,
        }}
        label=""
        className={clsName}
        disabled={disabled}
        renderInput={(props: any) => (
          <TextField id={id} variant="standard" {...props} disabled />
        )}
        ampm={false}
        PopperProps={{
          className: classes.paperOverride,
        }}
        inputFormat={"LL/dd/yyyy HH:mm"}
      />
    </LocalizationProvider>
  );

  if (forSearchBlock) {
    return inputItem;
  }

  const containerCls = !forFilterContained ? classes.fieldContainer : "";
  return (
    <Fragment>
      <Grid
        item
        xs={12}
        className={`${containerCls} ${classNamePrefix}input-field-container`}
      >
        {label !== "" && (
          <InputLabel
            htmlFor={id}
            className={`${classes.inputLabel} ${classNamePrefix}input-label`}
          >
            <span>
              {label}
              {required ? "*" : ""}
            </span>
            {tooltip !== "" && (
              <div className={classes.tooltipContainer}>
                <Tooltip title={tooltip} placement="top-start">
                  <div className={classes.tooltip}>
                    <HelpIcon />
                  </div>
                </Tooltip>
              </div>
            )}
          </InputLabel>
        )}

        <div
          className={`${classes.textBoxContainer} ${classNamePrefix}input-wrapper`}
        >
          {inputItem}
        </div>
      </Grid>
    </Fragment>
  );
};

export default withStyles(styles)(DateTimePickerWrapper);
