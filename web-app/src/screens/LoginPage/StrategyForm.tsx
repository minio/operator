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
import {
  Button,
  InputBox,
  LockFilledIcon,
  PasswordKeyIcon,
  UserFilledIcon,
  ProgressBar,
  Box,
  Grid,
} from "mds";
import { useSelector } from "react-redux";
import { setAccessKey, setSecretKey, setSTS, setUseSTS } from "./loginSlice";
import { AppState, useAppDispatch } from "../../store";
import { doLoginAsync } from "./loginThunks";

const StrategyForm = () => {
  const dispatch = useAppDispatch();

  const accessKey = useSelector((state: AppState) => state.login.accessKey);
  const secretKey = useSelector((state: AppState) => state.login.secretKey);
  const sts = useSelector((state: AppState) => state.login.sts);
  const useSTS = useSelector((state: AppState) => state.login.useSTS);

  const loginSending = useSelector(
    (state: AppState) => state.login.loginSending,
  );

  const formSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    dispatch(doLoginAsync());
  };

  return (
    <Box
      sx={{
        position: "absolute",
        top: 0,
        left: 0,
        width: "100%",
        height: "100%",
        overflow: "auto",

        "& .form": {
          width: "100%", // Fix IE 11 issue.
        },
        "& .submit": {
          margin: "30px 0px 8px",
          height: 40,
          width: "100%",
          boxShadow: "none",
          padding: "16px 30px",
        },
        "& .submitContainer": {
          textAlign: "right",
          marginTop: 30,
        },
        "& .linearPredef": {
          height: 10,
        },
      }}
    >
      <form className={"form"} noValidate onSubmit={formSubmit}>
        <Grid container>
          <Grid item xs={12} className={"spacerBottom"}>
            <InputBox
              fullWidth
              id="accessKey"
              className={"inputField"}
              value={accessKey}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                dispatch(setAccessKey(e.target.value))
              }
              placeholder={useSTS ? "STS Username" : "Username"}
              name="accessKey"
              autoComplete="username"
              disabled={loginSending}
              startIcon={<UserFilledIcon />}
            />
          </Grid>
          <Grid item xs={12} className={useSTS ? "spacerBottom" : ""}>
            <InputBox
              fullWidth
              value={secretKey}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                dispatch(setSecretKey(e.target.value))
              }
              name="secretKey"
              type="password"
              id="secretKey"
              autoComplete="current-password"
              disabled={loginSending}
              placeholder={useSTS ? "STS Secret" : "Password"}
              startIcon={<LockFilledIcon />}
            />
          </Grid>
          {useSTS && (
            <Grid item xs={12} className={"spacerBottom"}>
              <InputBox
                fullWidth
                id="sts"
                value={sts}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                  dispatch(setSTS(e.target.value))
                }
                placeholder={"STS Token"}
                name="STS"
                autoComplete="sts"
                disabled={loginSending}
                startIcon={<PasswordKeyIcon />}
              />
            </Grid>
          )}
        </Grid>

        <Grid item xs={12} className={"submitContainer"}>
          <Button
            type="submit"
            variant="callAction"
            color="primary"
            id="do-login"
            className={"submit"}
            disabled={
              (!useSTS && (accessKey === "" || secretKey === "")) ||
              (useSTS && sts === "") ||
              loginSending
            }
            label={"Login"}
            fullWidth
          />
        </Grid>
        <Grid item xs={12} className={"linearPredef"}>
          {loginSending && <ProgressBar />}
        </Grid>
        <Grid item xs={12} className={"linearPredef"}>
          <Box
            style={{
              textAlign: "center",
              marginTop: 20,
            }}
          >
            <span
              onClick={() => {
                dispatch(setUseSTS(!useSTS));
              }}
              style={{
                font: "normal normal normal 14px Inter",
                textDecoration: "underline",
                cursor: "pointer",
              }}
            >
              {!useSTS && <Fragment>Use STS</Fragment>}
              {useSTS && <Fragment>Use Credentials</Fragment>}
            </span>
            <span
              onClick={() => {
                dispatch(setUseSTS(!useSTS));
              }}
              style={{
                font: "normal normal normal 12px/15px Inter",
                textDecoration: "none",
                fontWeight: "bold",
                paddingLeft: 4,
              }}
            >
              ➔
            </span>
          </Box>
        </Grid>
      </form>
    </Box>
  );
};

export default StrategyForm;
