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
import { Box, Grid } from "mds";
import { useSelector } from "react-redux";
import {
  AutoSizer,
  CellMeasurer,
  CellMeasurerCache,
  List,
} from "react-virtualized";
import styled from "styled-components";
import get from "lodash/get";
import { ErrorResponseHandler } from "../../../../../common/types";
import { AppState, useAppDispatch } from "../../../../../store";
import { setErrorSnackMessage } from "../../../../../systemSlice";
import SearchBox from "../../../Common/SearchBox";
import api from "../../../../../common/api";

interface IPodLogsProps {
  tenant: string;
  namespace: string;
  podName: string;
  propLoading: boolean;
}

const LogsItem = styled.div(({ theme }) => ({
  "& .highlighted": {
    "& span": {
      backgroundColor: get(theme, "signalColors.warning", "#FFBD62"),
    },
  },
  "& .ansidefault": {
    color: get(theme, "fontColor", "#000"),
    lineHeight: "16px",
  },
}));

const PodLogs = ({
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
        <LogsItem
          key={index}
          className={`${highlightedLine ? "highlight" : ""}`}
        >
          <span className={"tab"}>{substr}</span>
        </LogsItem>
      );
    } else {
      // for all remaining set default class
      return (
        <LogsItem
          key={index}
          className={`${highlightedLine ? "highlight" : ""}`}
        >
          <span className={"ansidefault"}>{substr}</span>
        </LogsItem>
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
    <Fragment>
      <Grid
        item
        xs={12}
        sx={{
          display: "flex" as const,
          justifyContent: "space-between" as const,
          marginBottom: "1rem",
          alignItems: "center",
          gap: 10,
          "& button": {
            flexGrow: 0,
            marginLeft: 8,
          },
        }}
      >
        <SearchBox
          value={highlight}
          placeholder="Highlight Line"
          onChange={(value) => {
            setHighlight(value);
          }}
        />
      </Grid>
      <Grid item xs={12}>
        <Box
          sx={{
            minHeight: 400,
            height: "calc(100vh - 310px)",
            overflow: "hidden",
            fontSize: 13,
            padding: "25px 45px 0",
          }}
          useBackground
          withBorders
        >
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
        </Box>
      </Grid>
    </Fragment>
  );
};

export default PodLogs;
