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
import { Grid, InputBox, Select, SelectorType } from "mds";
import get from "lodash/get";
import styled from "styled-components";
import {
  ITolerationEffect,
  ITolerationOperator,
} from "../../../../common/types";
import InputUnitMenu from "../FormComponents/InputUnitMenu/InputUnitMenu";

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
}

const TolerationBase = styled.div(({ theme }) => ({
  flexGrow: "1",
  flexBasis: "100%",
  width: "100%",
  "& .labelsStyle": {
    fontSize: 18,
    fontWeight: "bold",
    color: get(theme, "secondaryText", "#AEAEAE"),
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    maxWidth: 45,
    marginRight: 10,
  },
  "& .firstLevel": {
    width: "100%",
    marginBottom: 10,
  },
  "& .secondLevel": {
    width: "100%",
  },
  "& .fieldContainer": {
    marginRight: 10,
  },
}));

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
}: ITolerationSelector) => {
  const operatorOptions: SelectorType[] = [];
  const effectOptions: SelectorType[] = [];

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
      <fieldset>
        <legend>Toleration {index + 1}</legend>
        <TolerationBase>
          <Grid container>
            <Grid container className={"firstLevel"}>
              <Grid item xs className={"labelsStyle"}>
                If
              </Grid>
              <Grid item xs className={"fieldContainer"}>
                <InputBox
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
                <Grid item xs className={"labelsStyle"}>
                  is
                </Grid>
              )}
              <Grid item xs={1} className={"fieldContainer"}>
                <Select
                  onChange={(value) => {
                    onOperatorChange(
                      ITolerationOperator[value as ITolerationOperator],
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
                <Grid item xs className={"labelsStyle"}>
                  to
                </Grid>
              )}
              {ITolerationOperator[operator] === ITolerationOperator.Equal && (
                <Grid item xs className={"fieldContainer"}>
                  <InputBox
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
            <Grid container className={"secondLevel"}>
              <Grid item xs className={"labelsStyle"}>
                then
              </Grid>
              <Grid item xs className={"fieldContainer"}>
                <Select
                  onChange={(value) => {
                    onEffectChange(
                      ITolerationEffect[value as ITolerationEffect],
                    );
                  }}
                  id={`effects-${index}`}
                  name="effects"
                  label={""}
                  value={ITolerationEffect[effect]}
                  options={effectOptions}
                />
              </Grid>
              <Grid item xs className={"labelsStyle"}>
                after
              </Grid>
              <Grid item xs className={"fieldContainer"}>
                <InputBox
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
        </TolerationBase>
      </fieldset>
    </Grid>
  );
};

export default TolerationSelector;
