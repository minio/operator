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

import React, {
  forwardRef,
  useEffect,
  useImperativeHandle,
  useState,
} from "react";
import clsx from "clsx";
import Grid from "@mui/material/Grid";
import { SelectChangeEvent } from "@mui/material";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import InputLabel from "@mui/material/InputLabel";
import Tooltip from "@mui/material/Tooltip";
import FormControl from "@mui/material/FormControl";
import Select from "@mui/material/Select";
import MenuItem from "@mui/material/MenuItem";
import InputBase from "@mui/material/InputBase";
import { fieldBasic, tooltipHelper } from "../common/styleLibrary";
import { HelpIcon } from "mds";
import FormSwitchWrapper from "../FormSwitchWrapper/FormSwitchWrapper";
import { days, months, validDate, years } from "./utils";

const styles = (theme: Theme) =>
  createStyles({
    dateInput: {
      "&:not(:last-child)": {
        marginRight: 22,
      },
    },
    ...fieldBasic,
    ...tooltipHelper,
    labelContainer: {
      flex: 1,
    },
    fieldContainer: {
      ...fieldBasic.fieldContainer,
      display: "flex",
      alignItems: "center",
      justifyContent: "space-between",
      paddingBottom: 10,
      marginTop: 11,
      marginBottom: 6,
    },
    fieldContainerBorder: {
      borderBottom: "#9c9c9c 1px solid",
      marginBottom: 20,
    },
  });

const SelectStyled = withStyles((theme: Theme) =>
  createStyles({
    root: {
      "& .MuiSelect-icon": {
        color: "#000",
        "&.Mui-disabled": {
          color: "#9c9c9c",
        },
      },
    },
    input: {
      borderBottom: 0,
      fontSize: 12,
    },
  }),
)(InputBase);

interface IDateSelectorProps {
  classes: any;
  id: string;
  label: string;
  disableOptions?: boolean;
  addSwitch?: boolean;
  tooltip?: string;
  borderBottom?: boolean;
  value?: string;
  onDateChange: (date: string, isValid: boolean) => any;
}

const DateSelector = forwardRef(
  (
    {
      classes,
      id,
      label,
      disableOptions = false,
      addSwitch = false,
      tooltip = "",
      borderBottom = false,
      onDateChange,
      value = "",
    }: IDateSelectorProps,
    ref: any,
  ) => {
    useImperativeHandle(ref, () => ({ resetDate }));

    const [dateEnabled, setDateEnabled] = useState<boolean>(false);
    const [month, setMonth] = useState<string>("");
    const [day, setDay] = useState<string>("");
    const [year, setYear] = useState<string>("");

    useEffect(() => {
      // verify if there is a current value
      // assume is in the format "2021-12-30"
      if (value !== "") {
        const valueSplit = value.split("-");
        setYear(valueSplit[0]);
        setMonth(valueSplit[1]);
        // Turn to single digit to be displayed on dropdown buttons
        setDay(`${parseInt(valueSplit[2])}`);
      }
    }, [value]);

    useEffect(() => {
      const [isValid, dateString] = validDate(year, month, day);
      onDateChange(dateString, isValid);
    }, [month, day, year, onDateChange]);

    const resetDate = () => {
      setMonth("");
      setDay("");
      setYear("");
    };

    const isDateDisabled = () => {
      if (disableOptions) {
        return disableOptions;
      } else if (addSwitch) {
        return !dateEnabled;
      } else {
        return false;
      }
    };

    const onMonthChange = (e: SelectChangeEvent<string>) => {
      setMonth(e.target.value as string);
    };

    const onDayChange = (e: SelectChangeEvent<string>) => {
      setDay(e.target.value as string);
    };

    const onYearChange = (e: SelectChangeEvent<string>) => {
      setYear(e.target.value as string);
    };

    return (
      <Grid
        item
        xs={12}
        className={clsx(classes.fieldContainer, {
          [classes.fieldContainerBorder]: borderBottom,
        })}
      >
        <div className={classes.labelContainer}>
          <Grid container>
            <InputLabel htmlFor={id} className={classes.inputLabel}>
              <span>{label}</span>
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
            {addSwitch && (
              <FormSwitchWrapper
                indicatorLabels={["Specific Date", "Default (7 Days)"]}
                checked={dateEnabled}
                value={"date_enabled"}
                id="date-status"
                name="date-status"
                onChange={(e) => {
                  setDateEnabled(e.target.checked);
                  if (!e.target.checked) {
                    onDateChange("", true);
                  }
                }}
                switchOnly
              />
            )}
          </Grid>
        </div>
        <div>
          <FormControl
            disabled={isDateDisabled()}
            className={classes.dateInput}
          >
            <Select
              id={`${id}-month`}
              name={`${id}-month`}
              value={month}
              displayEmpty
              onChange={onMonthChange}
              input={<SelectStyled />}
            >
              <MenuItem value="" disabled>
                {"<Month>"}
              </MenuItem>
              {months.map((option) => (
                <MenuItem
                  value={option.value}
                  key={`select-${id}-monthOP-${option.label}`}
                >
                  {option.label}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
          <FormControl
            disabled={isDateDisabled()}
            className={classes.dateInput}
          >
            <Select
              id={`${id}-day`}
              name={`${id}-day`}
              value={day}
              displayEmpty
              onChange={onDayChange}
              input={<SelectStyled />}
            >
              <MenuItem value="" disabled>
                {"<Day>"}
              </MenuItem>
              {days.map((dayNumber) => (
                <MenuItem
                  value={dayNumber}
                  key={`select-${id}-dayOP-${dayNumber}`}
                >
                  {dayNumber}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
          <FormControl
            disabled={isDateDisabled()}
            className={classes.dateInput}
          >
            <Select
              id={`${id}-year`}
              name={`${id}-year`}
              value={year}
              displayEmpty
              onChange={onYearChange}
              input={<SelectStyled />}
            >
              <MenuItem value="" disabled>
                {"<Year>"}
              </MenuItem>
              {years.map((year) => (
                <MenuItem value={year} key={`select-${id}-yearOP-${year}`}>
                  {year}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        </div>
      </Grid>
    );
  },
);

export default withStyles(styles)(DateSelector);
