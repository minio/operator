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

import React, { useEffect, useState } from "react";
import { useSelector } from "react-redux";
import { Theme } from "@mui/material/styles";
import {
  AutoSizer,
  CellMeasurer,
  CellMeasurerCache,
  List,
} from "react-virtualized";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { TextField } from "@mui/material";
import Grid from "@mui/material/Grid";
import Paper from "@mui/material/Paper";
import InputAdornment from "@mui/material/InputAdornment";
import api from "../../../../../common/api";
import { SearchIcon } from "mds";
import {
  actionsTray,
  containerForHeader,
  searchField,
} from "../../../Common/FormComponents/common/styleLibrary";
import { ErrorResponseHandler } from "../../../../../common/types";
import { AppState, useAppDispatch } from "../../../../../store";
import { setErrorSnackMessage } from "../../../../../systemSlice";

interface IPodLogsProps {
  classes: any;
  tenant: string;
  namespace: string;
  podName: string;
  propLoading: boolean;
}

const styles = (theme: Theme) =>
  createStyles({
    logList: {
      background: "#fff",
      minHeight: 400,
      height: "calc(100vh - 304px)",
      overflow: "auto",
      fontSize: 13,
      padding: "25px 45px 0",
      border: "1px solid #EAEDEE",
      borderRadius: 4,
    },

    ...searchField,
    actionsTray: {
      ...actionsTray.actionsTray,
      padding: "15px 0 0",
    },
    logerror_tab: {
      color: "#A52A2A",
      paddingLeft: 25,
    },
    ansidefault: {
      color: "#000",
      lineHeight: "16px",
    },
    highlight: {
      "& span": {
        backgroundColor: "#082F5238",
      },
    },
    ...containerForHeader,
  });

const PodLogs = ({
  classes,
  tenant,
  namespace,
  podName,
  propLoading,
}: IPodLogsProps) => {
  const dispatch = useAppDispatch();
  const loadingTenant = useSelector(
    (state: AppState) => state.tenants.loadingTenant,
  );
  const [highlight, setHighlight] = useState<string>("");
  const [logLines, setLogLines] = useState<string[]>([]);
  const [loading, setLoading] = useState<boolean>(true);

  const cache = new CellMeasurerCache({
    minWidth: 5,
    fixedHeight: false,
  });

  useEffect(() => {
    if (propLoading) {
      setLoading(true);
    }
  }, [propLoading]);

  useEffect(() => {
    if (loadingTenant) {
      setLoading(true);
    }
  }, [loadingTenant]);

  const renderLog = (logMessage: string, index: number) => {
    if (!logMessage) {
      return null;
    }
    // remove any non ascii characters, exclude any control codes
    logMessage = logMessage.replace(/([^\x20-\x7F])/g, "");

    // regex for terminal colors like e.g. `[31;4m `
    const tColorRegex = /((\[[0-9;]+m))/g;

    // get substring if there was a match for to split what
    // is going to be colored and what not, here we add color
    // only to the first match.
    let substr = logMessage.replace(tColorRegex, "");

    // in case highlight is set, we select the line that contains the requested string
    let highlightedLine =
      highlight !== ""
        ? logMessage.toLowerCase().includes(highlight.toLowerCase())
        : false;

    // if starts with multiple spaces add padding
    if (substr.startsWith("   ")) {
      return (
        <div
          key={index}
          className={`${highlightedLine ? classes.highlight : ""}`}
        >
          <span className={classes.tab}>{substr}</span>
        </div>
      );
    } else {
      // for all remaining set default class
      return (
        <div
          key={index}
          className={`${highlightedLine ? classes.highlight : ""}`}
        >
          <span className={classes.ansidefault}>{substr}</span>
        </div>
      );
    }
  };

  useEffect(() => {
    if (loading) {
      api
        .invoke(
          "GET",
          `/api/v1/namespaces/${namespace}/tenants/${tenant}/pods/${podName}`,
        )
        .then((res: string) => {
          setLogLines(res.split("\n"));
          setLoading(false);
        })
        .catch((err: ErrorResponseHandler) => {
          dispatch(setErrorSnackMessage(err));
          setLoading(false);
        });
    }
  }, [loading, podName, namespace, tenant, dispatch]);

  function cellRenderer({ columnIndex, key, parent, index, style }: any) {
    return (
      // @ts-ignore
      <CellMeasurer
        cache={cache}
        columnIndex={columnIndex}
        key={key}
        parent={parent}
        rowIndex={index}
      >
        <div
          style={{
            ...style,
          }}
        >
          {renderLog(logLines[index], index)}
        </div>
      </CellMeasurer>
    );
  }

  return (
    <React.Fragment>
      <Grid item xs={12} className={classes.actionsTray}>
        <TextField
          placeholder="Highlight Line"
          className={classes.searchField}
          id="search-resource"
          label=""
          onChange={(val) => {
            setHighlight(val.target.value);
          }}
          InputProps={{
            disableUnderline: true,
            startAdornment: (
              <InputAdornment position="start">
                <SearchIcon />
              </InputAdornment>
            ),
          }}
          variant="standard"
        />
      </Grid>
      <Grid item xs={12}>
        <br />
      </Grid>
      <Grid item xs={12}>
        <Paper>
          <div className={classes.logList}>
            {logLines.length >= 1 && (
              // @ts-ignore
              <AutoSizer>
                {({ width, height }) => (
                  // @ts-ignore
                  <List
                    rowHeight={(item) => cache.rowHeight(item)}
                    overscanRowCount={15}
                    rowCount={logLines.length}
                    rowRenderer={cellRenderer}
                    width={width}
                    height={height}
                  />
                )}
              </AutoSizer>
            )}
          </div>
        </Paper>
      </Grid>
    </React.Fragment>
  );
};

export default withStyles(styles)(PodLogs);
