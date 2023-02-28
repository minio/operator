// This file is part of MinIO Operator
// Copyright (c) 2023 MinIO, Inc.
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
import { Box } from "@mui/material";
import { FormTitle } from "./utils";
import { Button, OnlineRegistrationIcon, UsersIcon } from "mds";
import InputBoxWrapper from "../Common/FormComponents/InputBoxWrapper/InputBoxWrapper";
import VisibilityOffIcon from "@mui/icons-material/VisibilityOff";
import RemoveRedEyeIcon from "@mui/icons-material/RemoveRedEye";
import RegisterHelpBox from "./RegisterHelpBox";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import { spacingUtils } from "../Common/FormComponents/common/styleLibrary";
import makeStyles from "@mui/styles/makeStyles";
import { useSelector } from "react-redux";
import { selOpMode } from "../../../systemSlice";
import { AppState, useAppDispatch } from "../../../store";
import {
  setShowPassword,
  setSubnetEmail,
  setSubnetPassword,
} from "./registerSlice";
import { subnetLogin } from "./registerThunks";

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    sizedLabel: {
      minWidth: "75px",
    },
    ...spacingUtils,
  })
);

const OnlineRegistration = () => {
  const classes = useStyles();
  const dispatch = useAppDispatch();

  const operatorMode = useSelector(selOpMode);
  const subnetPassword = useSelector(
    (state: AppState) => state.register.subnetPassword
  );
  const subnetEmail = useSelector(
    (state: AppState) => state.register.subnetEmail
  );
  const showPassword = useSelector(
    (state: AppState) => state.register.showPassword
  );
  const loading = useSelector((state: AppState) => state.register.loading);

  return (
    <Fragment>
      <Box
        sx={{
          "& .title-text": {
            marginLeft: "27px",
            fontWeight: 600,
          },
        }}
      >
        <FormTitle
          icon={<OnlineRegistrationIcon />}
          title={`Online activation of MinIO Subscription Network License`}
        />
      </Box>
      <Box
        sx={{
          display: "flex",
          flexFlow: {
            xs: "column",
            md: "row",
          },
        }}
      >
        <Box
          sx={{
            display: "flex",
            flexFlow: "column",
            flex: "2",
          }}
        >
          <Box
            sx={{
              fontSize: "16px",
              display: "flex",
              flexFlow: "column",
              marginTop: "30px",
              marginBottom: "30px",
            }}
          >
            Use your MinIO Subscription Network login credentials to register
            this cluster.
          </Box>
          <Box
            sx={{
              flex: "1",
            }}
          >
            <InputBoxWrapper
              className={classes.spacerBottom}
              classes={{
                inputLabel: classes.sizedLabel,
              }}
              id="subnet-email"
              name="subnet-email"
              onChange={(event: React.ChangeEvent<HTMLInputElement>) =>
                dispatch(setSubnetEmail(event.target.value))
              }
              label="Email"
              value={subnetEmail}
              overlayIcon={<UsersIcon />}
            />
            <InputBoxWrapper
              className={classes.spacerBottom}
              classes={{
                inputLabel: classes.sizedLabel,
              }}
              id="subnet-password"
              name="subnet-password"
              onChange={(event: React.ChangeEvent<HTMLInputElement>) =>
                dispatch(setSubnetPassword(event.target.value))
              }
              label="Password"
              type={showPassword ? "text" : "password"}
              value={subnetPassword}
              overlayIcon={
                showPassword ? <VisibilityOffIcon /> : <RemoveRedEyeIcon />
              }
              overlayAction={() => dispatch(setShowPassword(!showPassword))}
            />

            <Box
              sx={{
                display: "flex",
                alignItems: "center",
                justifyContent: "flex-end",
                "& button": {
                  marginLeft: "8px",
                },
              }}
            >
              <Button
                id={"sign-up"}
                type="submit"
                className={classes.spacerRight}
                variant="regular"
                onClick={(e) => {
                  e.preventDefault();
                  window.open(
                    `https://min.io/signup?ref=${operatorMode ? "op" : "con"}`,
                    "_blank"
                  );
                }}
                label={"Sign up"}
              />
              <Button
                id={"register-credentials"}
                type="submit"
                variant="callAction"
                disabled={
                  loading ||
                  subnetEmail.trim().length === 0 ||
                  subnetPassword.trim().length === 0
                }
                onClick={() => dispatch(subnetLogin())}
                label={"Register"}
              />
            </Box>
          </Box>
        </Box>
        <RegisterHelpBox />
      </Box>
    </Fragment>
  );
};

export default OnlineRegistration;
