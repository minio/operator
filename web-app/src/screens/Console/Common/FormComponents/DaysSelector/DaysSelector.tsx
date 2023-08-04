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

import React, { Fragment, useEffect, useState } from "react";
import Grid from "@mui/material/Grid";
import InputLabel from "@mui/material/InputLabel";
import { DateTime } from "luxon";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { fieldBasic, tooltipHelper } from "../common/styleLibrary";
import InputBoxWrapper from "../InputBoxWrapper/InputBoxWrapper";
import { LinkIcon } from "mds";

interface IDaysSelector {
  classes: any;
  id: string;
  initialDate: Date;
  maxDays?: number;
  label: string;
  entity: string;
  onChange: (newDate: string, isValid: boolean) => void;
}

const styles = (theme: Theme) =>
  createStyles({
    ...fieldBasic,
    ...tooltipHelper,
    labelContainer: {
      display: "flex",
      alignItems: "center",
      marginBottom: 15,
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
    dateInputContainer: {
      margin: "0 10px",
    },
    durationInputs: {
      display: "flex",
      alignItems: "center",
      justifyContent: "flex-start",
    },

    validityIndicator: {
      display: "flex",
      alignItems: "center",
      justifyContent: "flex-start",
      marginTop: 25,
      marginLeft: 10,
    },
    invalidDurationText: {
      marginTop: 15,
      display: "flex",
      color: "red",
      fontSize: 11,
    },
    reverseInput: {
      flexFlow: "row-reverse",
      "& > label": {
        fontWeight: 400,
        marginLeft: 15,
        marginRight: 25,
      },
    },
    validityText: {
      fontSize: 14,
      marginTop: 15,
      display: "flex",
      alignItems: "center",
      justifyContent: "center",
      "@media (max-width: 900px)": {
        flexFlow: "column",
      },
      "& > .min-icon": {
        color: "#5E5E5E",
        width: 15,
        height: 15,
        marginRight: 10,
      },
    },
    validTill: {
      fontWeight: "bold",
      marginLeft: 15,
    },
  });

const calculateNewTime = (
  initialDate: Date,
  days: number,
  hours: number,
  minutes: number,
) => {
  return DateTime.fromJSDate(initialDate).plus({ days, hours, minutes });
};

const DaysSelector = ({
  classes,
  id,
  initialDate,
  label,
  maxDays,
  entity,
  onChange,
}: IDaysSelector) => {
  const [selectedDays, setSelectedDays] = useState<number>(7);
  const [selectedHours, setSelectedHours] = useState<number>(0);
  const [selectedMinutes, setSelectedMinutes] = useState<number>(0);
  const [validDate, setValidDate] = useState<boolean>(true);
  const [dateSelected, setDateSelected] = useState<DateTime>(DateTime.now());

  useEffect(() => {
    if (
      !isNaN(selectedHours) &&
      !isNaN(selectedDays) &&
      !isNaN(selectedMinutes)
    ) {
      setDateSelected(
        calculateNewTime(
          initialDate,
          selectedDays,
          selectedHours,
          selectedMinutes,
        ),
      );
    }
  }, [initialDate, selectedDays, selectedHours, selectedMinutes]);

  useEffect(() => {
    if (validDate) {
      const formattedDate = dateSelected.toFormat("yyyy-MM-dd HH:mm:ss");
      onChange(formattedDate.split(" ").join("T"), true);
    } else {
      onChange("0000-00-00", false);
    }
  }, [dateSelected, onChange, validDate]);

  // Basic validation for inputs
  useEffect(() => {
    let valid = true;
    if (
      selectedDays < 0 ||
      (maxDays && selectedDays > maxDays) ||
      isNaN(selectedDays)
    ) {
      valid = false;
    }

    if (selectedHours < 0 || selectedHours > 23 || isNaN(selectedHours)) {
      valid = false;
    }

    if (selectedMinutes < 0 || selectedMinutes > 59 || isNaN(selectedMinutes)) {
      valid = false;
    }

    if (
      maxDays &&
      selectedDays === maxDays &&
      (selectedHours !== 0 || selectedMinutes !== 0)
    ) {
      valid = false;
    }

    setValidDate(valid);
  }, [
    dateSelected,
    maxDays,
    onChange,
    selectedDays,
    selectedHours,
    selectedMinutes,
  ]);

  const extraInputProps = {
    style: {
      textAlign: "center" as const,
      paddingRight: 10,
      paddingLeft: 10,
      width: 25,
    },
    className: "removeArrows" as const,
  };

  return (
    <Fragment>
      <Grid container className={classes.fieldContainer}>
        <Grid item xs={12} className={classes.labelContainer}>
          <InputLabel
            htmlFor={id}
            className={classes.inputLabel}
            sx={{ marginLeft: "10px" }}
          >
            <span>{label}</span>
          </InputLabel>
        </Grid>
        <Grid item xs={12} className={classes.durationInputs}>
          <Grid item className={classes.dateInputContainer}>
            <InputBoxWrapper
              id={id}
              className={classes.reverseInput}
              type="number"
              min="0"
              max={maxDays ? maxDays.toString() : "999"}
              label="Days"
              name={id}
              onChange={(e) => {
                setSelectedDays(parseInt(e.target.value));
              }}
              value={selectedDays.toString()}
              extraInputProps={extraInputProps}
              noLabelMinWidth
            />
          </Grid>
          <Grid item className={classes.dateInputContainer}>
            <InputBoxWrapper
              id={id}
              className={classes.reverseInput}
              type="number"
              min="0"
              max="23"
              label="Hours"
              name={id}
              onChange={(e) => {
                setSelectedHours(parseInt(e.target.value));
              }}
              value={selectedHours.toString()}
              extraInputProps={extraInputProps}
              noLabelMinWidth
            />
          </Grid>
          <Grid item className={classes.dateInputContainer}>
            <InputBoxWrapper
              id={id}
              className={classes.reverseInput}
              type="number"
              min="0"
              max="59"
              label="Minutes"
              name={id}
              onChange={(e) => {
                setSelectedMinutes(parseInt(e.target.value));
              }}
              value={selectedMinutes.toString()}
              extraInputProps={extraInputProps}
              noLabelMinWidth
            />
          </Grid>
        </Grid>
        <Grid
          item
          xs={12}
          className={`${classes.validityIndicator} ${classes.formFieldRow}`}
        >
          {validDate ? (
            <div className={classes.validityText}>
              <LinkIcon />
              <div className={classes.validityLabel}>
                {entity} will be available until:
              </div>{" "}
              <div className={classes.validTill}>
                {dateSelected.toFormat("MM/dd/yyyy HH:mm:ss")}
              </div>
            </div>
          ) : (
            <div className={classes.invalidDurationText}>
              Please select a valid duration.
            </div>
          )}
        </Grid>
      </Grid>
    </Fragment>
  );
};

export default withStyles(styles)(DaysSelector);
