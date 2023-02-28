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
import InputBoxWrapper from "../Common/FormComponents/InputBoxWrapper/InputBoxWrapper";
import FormSwitchWrapper from "../Common/FormComponents/FormSwitchWrapper/FormSwitchWrapper";
import SelectWrapper from "../Common/FormComponents/SelectWrapper/SelectWrapper";
import { Grid, SelectChangeEvent } from "@mui/material";
import { useDispatch } from "react-redux";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { fsGroupChangePolicyType } from "./types";

interface IEditSecurityContextProps {
  classes: any;
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

const styles = (theme: Theme) =>
  createStyles({
    configSectionItem: {
      marginRight: 15,
      marginBottom: 15,
      "& .multiContainer": {
        border: "1px solid red",
      },
    },
  });

const SecurityContextSelector = ({
  classes,
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
      <fieldset className={`${classes.fieldGroup} ${classes.fieldSpaceTop} `}>
        <legend className={classes.descriptionText}>Security Context</legend>

        <Grid item xs={12}>
          <div className={`${classes.multiContainerStackNarrow} `}>
            <div className={classes.configSectionItem}>
              <InputBoxWrapper
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
            </div>
            <div className={classes.configSectionItem}>
              <InputBoxWrapper
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
            </div>
          </div>
        </Grid>
        <Grid item xs={12}>
          <div className={`${classes.multiContainerStackNarrow} `}>
            <div className={classes.configSectionItem}>
              <InputBoxWrapper
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
            </div>

            <div className={classes.configSectionItem}>
              <SelectWrapper
                label="FsGroupChangePolicy"
                id="securityContext_fsGroupChangePolicy"
                name="securityContext_fsGroupChangePolicy"
                onChange={(e: SelectChangeEvent<string>) => {
                  dispatch(setFSGroupChangePolicy(e.target.value));
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
            </div>
          </div>
        </Grid>
        <Grid item xs={12}>
          <div className={classes.multiContainer}>
            <FormSwitchWrapper
              value="SecurityContextRunAsNonRoot"
              id="securityContext_runAsNonRoot"
              name="securityContext_runAsNonRoot"
              checked={runAsNonRoot}
              onChange={() => {
                dispatch(setRunAsNonRoot(!runAsNonRoot));
              }}
              label={"Do not run as Root"}
            />
          </div>
        </Grid>
      </fieldset>
    </Fragment>
  );
};

export default withStyles(styles)(SecurityContextSelector);
