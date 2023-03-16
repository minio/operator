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

import React, { Fragment, Suspense } from "react";
import { ApplicationLogo, VersionIcon } from "mds";
import { Box, IconButton } from "@mui/material";
import MenuIcon from "@mui/icons-material/Menu";
import { useSelector } from "react-redux";
import { selOpMode } from "../../../systemSlice";
import TooltipWrapper from "../Common/TooltipWrapper/TooltipWrapper";
import { getLogoVar } from "../../../config";

type MenuToggleProps = {
  isOpen: boolean;
  onToggle: (nextState: boolean) => void;
};
const MenuToggle = ({ isOpen, onToggle }: MenuToggleProps) => {
  const stateClsName = isOpen ? "wide" : "mini";

  const operatorMode = useSelector(selOpMode);

  let logoPlan = getLogoVar();

  return (
    <Box
      className={`${stateClsName}`}
      sx={{
        width: "100%",
        cursor: "pointer",
        "&::after": {
          height: "1px",
          display: "block",
          content: "' '",
          backgroundColor: "#0F446C",
          margin: "0px auto",
        },
        "&.wide:hover": {
          background:
            "transparent linear-gradient(270deg, #00000000 0%, #051d39 53%, #54545400 100%) 0% 0% no-repeat padding-box",
        },
      }}
      onClick={() => {
        if (isOpen) {
          onToggle(false);
        }
      }}
    >
      <Box
        className={`${stateClsName}`}
        sx={{
          marginLeft: "26px",
          marginRight: "8px",
          display: "flex",
          alignItems: "center",
          height: "82px",

          "&.mini": {
            flexFlow: "column",
            display: "flex",
            justifyContent: "center",
            gap: "3px",
            alignItems: "center",
            marginLeft: "auto",
            marginRight: "auto",
          },
          "& .logo": {
            background: "transparent",
            width: "180px",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            "&.wide": {
              flex: "1",
              "& svg": {
                width: "100%",
                maxWidth: 180,
                fill: "white",
              },
            },
            "&.mini": {
              color: "#ffffff",
              "& svg": {
                width: 24,
                fill: "rgba(255, 255, 255, 0.8)",
              },
            },
          },
        }}
      >
        {isOpen ? (
          <TooltipWrapper
            tooltip={"Click to Collapse Menu"}
            placement={"right"}
          >
            <div className={`logo ${stateClsName}`}>
              {!operatorMode ? (
                <Fragment>
                  <ApplicationLogo
                    applicationName={"console"}
                    subVariant={logoPlan}
                    inverse
                  />
                </Fragment>
              ) : (
                <Fragment>
                  <ApplicationLogo applicationName={"operator"} inverse />
                </Fragment>
              )}
            </div>
          </TooltipWrapper>
        ) : (
          <div className={`logo ${stateClsName}`}>
            <Suspense fallback={<div>...</div>}>
              <VersionIcon />
            </Suspense>
          </div>
        )}

        {!isOpen && (
          <IconButton
            className={`${stateClsName}`}
            sx={{
              height: "30px",
              width: "30px",
              "&.mini": {
                "&:hover": {
                  background: "#081C42",
                },
              },

              "&:hover": {
                borderRadius: "50%",
                background: "#073052",
              },
              "& svg": {
                fill: "#ffffff",
                height: "18px",
                width: "18px",
              },
            }}
            onClick={() => {
              onToggle(true);
            }}
            size="small"
          >
            <MenuIcon />
          </IconButton>
        )}
      </Box>
    </Box>
  );
};

export default MenuToggle;
