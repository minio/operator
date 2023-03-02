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

import React, { Fragment } from "react";
import makeStyles from "@mui/styles/makeStyles";
import Tabs from "@mui/material/Tabs";
import Tab from "@mui/material/Tab";
import { ITabOption } from "./types";

interface ITabSelector {
  selectedTab: number;
  onChange: (newValue: number) => void;
  tabOptions: ITabOption[];
}

const tabSubStyles = makeStyles({
  tabRoot: {
    height: "40px",
    borderBottom: "1px solid #eaeaea",
  },
  root: {
    width: "120px",
    backgroundColor: "transparent",
    paddingTop: 0,
    paddingBottom: 0,
    fontSize: "14px",
    fontWeight: 600,
    color: "#07193E",
    height: "40px",
  },
  selected: {
    "&.MuiTab-selected": {
      backgroundColor: "#F6F7F7 !important",
    },
    "&.MuiTab-wrapper": {
      color: "#07193E",
      fontWeight: 600,
    },
  },
  indicator: {
    background:
      "transparent linear-gradient(90deg, #072B4E 0%, #081C42 100%) 0% 0% no-repeat padding-box;",
    height: 2,
  },
  scroller: {
    maxWidth: 1185,
    position: "relative",
    "&::after": {
      content: '" "',
      backgroundColor: "#EEF1F4",
      height: 2,
      width: "100%",
      display: "block",
    },
  },
});

const TabSelector = ({ selectedTab, onChange, tabOptions }: ITabSelector) => {
  const subStyles = tabSubStyles();

  return (
    <Fragment>
      <Tabs
        indicatorColor="primary"
        textColor="primary"
        aria-label="cluster-tabs"
        variant="scrollable"
        scrollButtons="auto"
        value={selectedTab}
        onChange={(e: React.ChangeEvent<{}>, newValue: number) => {
          onChange(newValue);
        }}
        classes={{
          root: subStyles.tabRoot,
          indicator: subStyles.indicator,
          scroller: subStyles.scroller,
        }}
      >
        {tabOptions.map((option, index) => {
          let tabOptions: ITabOption = {
            label: option.label,
          };

          if (option.value) {
            tabOptions = { ...tabOptions, value: option.value };
          }

          if (option.disabled) {
            tabOptions = { ...tabOptions, disabled: option.disabled };
          }

          return (
            <Tab
              {...tabOptions}
              classes={{
                root: subStyles.root,
                selected: subStyles.selected,
              }}
              id={`simple-tab-${index}`}
              aria-controls={`simple-tabpanel-${index}`}
              key={`tab-${index}-${option.label}`}
            />
          );
        })}
      </Tabs>
    </Fragment>
  );
};

export default TabSelector;
