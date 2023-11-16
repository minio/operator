//  This file is part of MinIO Operator
//  Copyright (c) 2022 MinIO, Inc.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.
import * as React from "react";
import { useEffect, useState } from "react";
import { Box, MenuExpandedIcon } from "mds";
import { useNavigate } from "react-router-dom";
import {
  ActionId,
  ActionImpl,
  KBarAnimator,
  KBarPortal,
  KBarPositioner,
  KBarResults,
  KBarSearch,
  KBarState,
  useKBar,
  useMatches,
  useRegisterActions,
} from "kbar";
import { Action } from "kbar/lib/types";
import { routesAsKbarActions } from "./kbar-actions";
import { useSelector } from "react-redux";
import { selFeatures } from "./consoleSlice";
import { selOpMode } from "../../systemSlice";

const searchStyle = {
  padding: "12px 16px",
  width: "100%",
  boxSizing: "border-box" as React.CSSProperties["boxSizing"],
  outline: "none",
  border: "none",
  color: "#858585",
  boxShadow: "0px 3px 5px #00000017",
  borderRadius: "4px 4px 0px 0px",
  fontSize: "14px",
  backgroundImage: "url(/images/search-icn.svg)",
  backgroundRepeat: "no-repeat",
  backgroundPosition: "95%",
};

const animatorStyle = {
  maxWidth: "600px",
  width: "100%",
  background: "white",
  color: "black",
  borderRadius: "4px",
  overflow: "hidden",
  boxShadow: "0px 3px 20px #00000055",
};

const groupNameStyle = {
  marginLeft: "30px",
  padding: "19px 0px 14px 0px",
  fontSize: "10px",
  textTransform: "uppercase" as const,
  color: "#858585",
  borderBottom: "1px solid #eaeaea",
};

const KBarStateChangeMonitor = ({
  onShow,
  onHide,
}: {
  onShow?: () => void;
  onHide?: () => void;
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const { visualState } = useKBar((state: KBarState) => {
    return {
      visualState: state.visualState,
    };
  });

  useEffect(() => {
    if (visualState === "showing") {
      setIsOpen(true);
    } else {
      setIsOpen(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [visualState]);

  useEffect(() => {
    if (isOpen) {
      onShow?.();
    } else {
      onHide?.();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen]);

  //just to hook into the internal state of KBar. !
  return null;
};

const CommandBar = () => {
  const operatorMode = useSelector(selOpMode);
  const features = useSelector(selFeatures);
  const navigate = useNavigate();

  const initialActions: Action[] = routesAsKbarActions(
    features,
    operatorMode,
    navigate,
  );

  useRegisterActions(initialActions, [operatorMode, features]);

  //fetch buckets everytime the kbar is shown so that new buckets created elsewhere , within first page is also shown

  return (
    <KBarPortal>
      <KBarStateChangeMonitor />
      <KBarPositioner
        style={{
          zIndex: 9999,
          boxShadow: "0px 3px 20px #00000055",
          borderRadius: "4px",
        }}
      >
        <KBarAnimator style={animatorStyle}>
          <KBarSearch style={searchStyle} />
          <RenderResults />
        </KBarAnimator>
      </KBarPositioner>
    </KBarPortal>
  );
};

function RenderResults() {
  const { results, rootActionId } = useMatches();

  return (
    <KBarResults
      items={results}
      onRender={({ item, active }) =>
        typeof item === "string" ? (
          <Box style={groupNameStyle}>{item}</Box>
        ) : (
          <ResultItem
            action={item}
            active={active}
            currentRootActionId={`${rootActionId}`}
          />
        )
      }
    />
  );
}

const ResultItem = React.forwardRef(
  (
    {
      action,
      active,
      currentRootActionId,
    }: {
      action: ActionImpl;
      active: boolean;
      currentRootActionId: ActionId;
    },
    ref: React.Ref<HTMLDivElement>,
  ) => {
    const ancestors = React.useMemo(() => {
      if (!currentRootActionId) return action.ancestors;
      const index = action.ancestors.findIndex(
        (ancestor) => ancestor.id === currentRootActionId,
      );
      // +1 removes the currentRootAction; e.g.
      // if we are on the "Set theme" parent action,
      // the UI should not display "Set theme… > Dark"
      // but rather just "Dark"
      return action.ancestors.slice(index + 1);
    }, [action.ancestors, currentRootActionId]);

    return (
      <div
        ref={ref}
        style={{
          padding: "12px 12px 12px 36px",
          marginTop: "2px",
          background: active ? "#dddddd" : "transparent",
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          cursor: "pointer",
        }}
      >
        <Box
          sx={{
            display: "flex",
            gap: "8px",
            alignItems: "center",
            fontSize: 14,
            flex: 1,
            justifyContent: "space-between",
            "& .min-icon": {
              width: "17px",
              height: "17px",
            },
          }}
        >
          <Box sx={{ height: "15px", width: "15px", marginRight: "36px" }}>
            {action.icon && action.icon}
          </Box>
          <div style={{ display: "flex", flexDirection: "column", flex: 2 }}>
            <Box>
              {ancestors.length > 0 &&
                ancestors.map((ancestor) => (
                  <React.Fragment key={ancestor.id}>
                    <span
                      style={{
                        opacity: 0.5,
                        marginRight: 8,
                      }}
                    >
                      {ancestor.name}
                    </span>
                    <span
                      style={{
                        marginRight: 8,
                      }}
                    >
                      &rsaquo;
                    </span>
                  </React.Fragment>
                ))}
              <span>{action.name}</span>
            </Box>
            {action.subtitle && (
              <span
                style={{
                  fontSize: 12,
                }}
              >
                {action.subtitle}
              </span>
            )}
          </div>
          <Box
            sx={{
              "& .min-icon": {
                width: "15px",
                height: "15px",
                fill: "#8f8b8b",
                transform: "rotate(90deg)",

                "& rect": {
                  fill: "#ffffff",
                },
              },
            }}
          >
            <MenuExpandedIcon />
          </Box>
        </Box>
        {action.shortcut?.length ? (
          <div
            aria-hidden
            style={{ display: "grid", gridAutoFlow: "column", gap: "4px" }}
          >
            {action.shortcut.map((sc) => (
              <kbd
                key={sc}
                style={{
                  padding: "4px 6px",
                  background: "rgba(0 0 0 / .1)",
                  borderRadius: "4px",
                  fontSize: 14,
                }}
              >
                {sc}
              </kbd>
            ))}
          </div>
        ) : null}
      </div>
    );
  },
);

export default CommandBar;
