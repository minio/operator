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

import React from "react";
import Grid from "@mui/material/Grid";
import { Box, InputLabel, Tooltip } from "@mui/material";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { Button, CopyIcon, HelpIcon } from "mds";
import { fieldBasic } from "../common/styleLibrary";
import CopyToClipboard from "react-copy-to-clipboard";
import CodeEditor from "@uiw/react-textarea-code-editor";
import TooltipWrapper from "../../TooltipWrapper/TooltipWrapper";

interface ICodeWrapper {
  value: string;
  label?: string;
  mode?: string;
  tooltip?: string;
  classes: any;
  onChange?: (editor: any, data: any, value: string) => any;
  onBeforeChange: (editor: any, data: any, value: string) => any;
  readOnly?: boolean;
  editorHeight?: string;
}

const styles = (theme: Theme) =>
  createStyles({
    ...fieldBasic,
  });

const CodeMirrorWrapper = ({
  value,
  label = "",
  tooltip = "",
  mode = "json",
  classes,
  onBeforeChange,
  readOnly = false,
  editorHeight = "250px",
}: ICodeWrapper) => {
  return (
    <React.Fragment>
      <Grid item xs={12} sx={{ marginBottom: "10px" }}>
        <InputLabel className={classes.inputLabel}>
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
      </Grid>

      <Grid
        item
        xs={12}
        style={{
          maxHeight: editorHeight,
          overflow: "auto",
          border: "1px solid #eaeaea",
        }}
      >
        <CodeEditor
          value={value}
          language={mode}
          onChange={(evn) => {
            onBeforeChange(null, null, evn.target.value);
          }}
          id={"code_wrapper"}
          padding={15}
          style={{
            fontSize: 12,
            backgroundColor: "#fefefe",
            fontFamily:
              "ui-monospace,SFMono-Regular,SF Mono,Consolas,Liberation Mono,Menlo,monospace",
            minHeight: editorHeight || "initial",
            color: "#000000",
          }}
        />
      </Grid>
      <Grid
        item
        xs={12}
        sx={{
          background: "#f7f7f7",
          border: "1px solid #eaeaea",
          borderTop: 0,
        }}
      >
        <Box
          sx={{
            display: "flex",
            alignItems: "center",
            padding: "2px",
            paddingRight: "5px",
            justifyContent: "flex-end",
            "& button": {
              height: "26px",
              width: "26px",
              padding: "2px",
              " .min-icon": {
                marginLeft: "0",
              },
            },
          }}
        >
          <TooltipWrapper tooltip={"Copy to Clipboard"}>
            <CopyToClipboard text={value}>
              <Button
                type={"button"}
                id={"copy-code-mirror"}
                icon={<CopyIcon />}
                color={"primary"}
                variant={"regular"}
              />
            </CopyToClipboard>
          </TooltipWrapper>
        </Box>
      </Grid>
    </React.Fragment>
  );
};

export default withStyles(styles)(CodeMirrorWrapper);
