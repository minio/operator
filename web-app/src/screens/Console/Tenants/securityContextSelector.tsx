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

import React, { Fragment } from "react";
import { Box, Grid, InputBox, Select, Switch } from "mds";
import { useDispatch } from "react-redux";
import { fsGroupChangePolicyType } from "./types";

interface IEditSecurityContextProps {
  runAsUser: string;
  runAsGroup: string;
  fsGroup: string;
  fsGroupChangePolicy: fsGroupChangePolicyType;
  runAsNonRoot: boolean;
  setRunAsUser: any;
  setRunAsGroup: any;
  setFSGroup: any;
  setRunAsNonRoot: any;
  setFSGroupChangePolicy: any;
}

const SecurityContextSelector = ({
  runAsGroup,
  runAsUser,
  fsGroup,
  fsGroupChangePolicy,
  runAsNonRoot,
  setRunAsUser,
  setRunAsGroup,
  setFSGroup,
  setRunAsNonRoot,
  setFSGroupChangePolicy,
}: IEditSecurityContextProps) => {
  const dispatch = useDispatch();
  return (
    <Fragment>
      <fieldset className={`inputItem`}>
        <legend>Security Context</legend>
        <Box
          sx={{
            "& .multiContainerStackNarrow": {
              display: "flex",
              alignItems: "center",
              justifyContent: "flex-start",
              gap: "8px",
              "@media (max-width: 750px)": {
                flexFlow: "column",
                flexDirection: "column",
              },
            },
            "& .configSectionItem": {
              marginRight: 15,
              marginBottom: 10,
            },
          }}
        >
          <Grid item xs={12}>
            <Box className={`multiContainerStackNarrow`}>
              <Box className={"configSectionItem"}>
                <InputBox
                  type="number"
                  id="securityContext_runAsUser"
                  name="securityContext_runAsUser"
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    dispatch(setRunAsUser(e.target.value));
                  }}
                  label="Run As User"
                  value={runAsUser}
                  required
                  min="0"
                />
              </Box>
              <Box className={"configSectionItem"}>
                <InputBox
                  type="number"
                  id="securityContext_runAsGroup"
                  name="securityContext_runAsGroup"
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    dispatch(setRunAsGroup(e.target.value));
                  }}
                  label="Run As Group"
                  value={runAsGroup}
                  required
                  min="0"
                />
              </Box>
            </Box>
          </Grid>
          <Grid item xs={12}>
            <Box className={`multiContainerStackNarrow `}>
              <Box className={"configSectionItem"}>
                <InputBox
                  type="number"
                  id="securityContext_fsGroup"
                  name="securityContext_fsGroup"
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    dispatch(setFSGroup(e.target.value));
                  }}
                  label="FsGroup"
                  value={fsGroup}
                  required
                  min="0"
                />
              </Box>
              <Box className={"configSectionItem"}>
                <Select
                  label="FsGroupChangePolicy"
                  id="securityContext_fsGroupChangePolicy"
                  name="securityContext_fsGroupChangePolicy"
                  onChange={(value) => {
                    dispatch(setFSGroupChangePolicy(value));
                  }}
                  value={fsGroupChangePolicy}
                  options={[
                    {
                      label: "Always",
                      value: "Always",
                    },
                    {
                      label: "OnRootMismatch",
                      value: "OnRootMismatch",
                    },
                  ]}
                />
              </Box>
            </Box>
          </Grid>
          <Grid item xs={12}>
            <Switch
              value="SecurityContextRunAsNonRoot"
              id="securityContext_runAsNonRoot"
              name="securityContext_runAsNonRoot"
              checked={runAsNonRoot}
              onChange={() => {
                dispatch(setRunAsNonRoot(!runAsNonRoot));
              }}
              label={"Do not run as Root"}
            />
          </Grid>
        </Box>
      </fieldset>
    </Fragment>
  );
};

export default SecurityContextSelector;
