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
import {
  ITolerationEffect,
  ITolerationOperator,
} from "../../../../common/types";
import SelectWrapper, {
  selectorTypes,
} from "../FormComponents/SelectWrapper/SelectWrapper";
import { Grid, SelectChangeEvent } from "@mui/material";
import InputBoxWrapper from "../FormComponents/InputBoxWrapper/InputBoxWrapper";
import InputUnitMenu from "../FormComponents/InputUnitMenu/InputUnitMenu";
import withStyles from "@mui/styles/withStyles";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";

interface ITolerationSelector {
  effect: ITolerationEffect;
  onEffectChange: (value: ITolerationEffect) => void;
  tolerationKey: string;
  onTolerationKeyChange: (value: string) => void;
  operator: ITolerationOperator;
  onOperatorChange: (value: ITolerationOperator) => void;
  value?: string;
  onValueChange: (value: string) => void;
  tolerationSeconds?: number;
  onSecondsChange: (value: number) => void;
  index: number;
  classes: any;
}

const styles = (theme: Theme) =>
  createStyles({
    labelsStyle: {
      fontSize: 18,
      fontWeight: "bold",
      color: "#AEAEAE",
      display: "flex",
      alignItems: "center",
      justifyContent: "center",
      maxWidth: 45,
      marginRight: 10,
    },
    fieldsetStyle: {
      border: "1px solid #EAEAEA",
      borderRadius: 2,
      padding: 10,
      marginBottom: 15,
    },
    firstLevel: {
      marginBottom: 10,
    },
    fieldContainer: {
      marginRight: 10,
    },
    legendStyle: {
      fontSize: 12,
      color: "#696969",
      fontWeight: "bold",
    },
  });

const TolerationSelector = ({
  effect,
  onEffectChange,
  tolerationKey,
  onTolerationKeyChange,
  operator,
  onOperatorChange,
  value,
  onValueChange,
  tolerationSeconds,
  onSecondsChange,
  index,
  classes,
}: ITolerationSelector) => {
  const operatorOptions: selectorTypes[] = [];
  const effectOptions: selectorTypes[] = [];

  for (let operator in ITolerationOperator) {
    operatorOptions.push({
      value: operator,
      label: operator,
    });
  }

  for (let effect in ITolerationEffect) {
    effectOptions.push({
      value: effect,
      label: effect,
    });
  }

  return (
    <Grid item xs={12}>
      <fieldset className={classes.fieldsetStyle}>
        <legend className={classes.legendStyle}>Toleration {index + 1}</legend>
        <Grid container>
          <Grid container className={classes.firstLevel}>
            <Grid item xs className={classes.labelsStyle}>
              If
            </Grid>
            <Grid item xs className={classes.fieldContainer}>
              <InputBoxWrapper
                id={`keyField-${index}`}
                label={""}
                name={`keyField-${index}`}
                value={tolerationKey}
                onChange={(e) => {
                  onTolerationKeyChange(e.target.value);
                }}
                index={index}
                placeholder={"Toleration Key"}
              />
            </Grid>
            {ITolerationOperator[operator] === ITolerationOperator.Equal && (
              <Grid item xs className={classes.labelsStyle}>
                is
              </Grid>
            )}
            <Grid item xs={1} className={classes.fieldContainer}>
              <SelectWrapper
                onChange={(e: SelectChangeEvent<string>) => {
                  onOperatorChange(
                    ITolerationOperator[e.target.value as ITolerationOperator],
                  );
                }}
                id={`operator-${index}`}
                name="operator"
                label={""}
                value={ITolerationOperator[operator]}
                options={operatorOptions}
              />
            </Grid>
            {ITolerationOperator[operator] === ITolerationOperator.Equal && (
              <Grid item xs className={classes.labelsStyle}>
                to
              </Grid>
            )}
            {ITolerationOperator[operator] === ITolerationOperator.Equal && (
              <Grid item xs className={classes.fieldContainer}>
                <InputBoxWrapper
                  id={`valueField-${index}`}
                  label={""}
                  name={`valueField-${index}`}
                  value={value || ""}
                  onChange={(e) => {
                    onValueChange(e.target.value);
                  }}
                  index={index}
                  placeholder={"Toleration Value"}
                />
              </Grid>
            )}
          </Grid>
          <Grid container>
            <Grid item xs className={classes.labelsStyle}>
              then
            </Grid>
            <Grid item xs className={classes.fieldContainer}>
              <SelectWrapper
                onChange={(e: SelectChangeEvent<string>) => {
                  onEffectChange(
                    ITolerationEffect[e.target.value as ITolerationEffect],
                  );
                }}
                id={`effects-${index}`}
                name="effects"
                label={""}
                value={ITolerationEffect[effect]}
                options={effectOptions}
              />
            </Grid>
            <Grid item xs className={classes.labelsStyle}>
              after
            </Grid>
            <Grid item xs className={classes.fieldContainer}>
              <InputBoxWrapper
                id={`seconds-${index}`}
                label={""}
                name={`seconds-${index}`}
                value={tolerationSeconds?.toString() || "0"}
                onChange={(e) => {
                  if (e.target.validity.valid) {
                    onSecondsChange(parseInt(e.target.value));
                  }
                }}
                index={index}
                pattern={"[0-9]*"}
                overlayObject={
                  <InputUnitMenu
                    id={`seconds-${index}`}
                    unitSelected={"seconds"}
                    unitsList={[{ label: "Seconds", value: "seconds" }]}
                    disabled={true}
                  />
                }
              />
            </Grid>
          </Grid>
        </Grid>
      </fieldset>
    </Grid>
  );
};

export default withStyles(styles)(TolerationSelector);
