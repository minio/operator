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

import React, { Fragment, useEffect } from "react";
import { useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";
import {
  Button,
  InputBox,
  Loader,
  LockIcon,
  LoginWrapper,
  LogoutIcon,
  RefreshIcon,
  Grid,
  ProgressBar,
  Select,
  Box,
} from "mds";
import { loginStrategyType, redirectRule } from "./types";
import { AppState, useAppDispatch } from "../../store";
import {
  doLoginAsync,
  getFetchConfigurationAsync,
  getVersionAsync,
} from "./loginThunks";
import { resetForm, setJwt } from "./loginSlice";
import { redirectRules } from "../../utils/sortFunctions";
import { getLogoVar } from "../../config";
import MainError from "../Console/Common/MainError/MainError";
import StrategyForm from "./StrategyForm";

export interface LoginStrategyRoutes {
  [key: string]: string;
}

export interface LoginStrategyPayload {
  [key: string]: any;
}

export const loginStrategyEndpoints: LoginStrategyRoutes = {
  form: "/api/v1/login",
  "service-account": "/api/v1/login/operator",
};

export const getTargetPath = () => {
  let targetPath = "/";
  if (
    localStorage.getItem("redirect-path") &&
    localStorage.getItem("redirect-path") !== ""
  ) {
    targetPath = `${localStorage.getItem("redirect-path")}`;
    localStorage.setItem("redirect-path", "");
  }
  return targetPath;
};

const Login = () => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();

  const jwt = useSelector((state: AppState) => state.login.jwt);
  const loginStrategy = useSelector(
    (state: AppState) => state.login.loginStrategy,
  );
  const loginSending = useSelector(
    (state: AppState) => state.login.loginSending,
  );
  const loadingFetchConfiguration = useSelector(
    (state: AppState) => state.login.loadingFetchConfiguration,
  );
  const loadingVersion = useSelector(
    (state: AppState) => state.login.loadingVersion,
  );
  const navigateTo = useSelector((state: AppState) => state.login.navigateTo);

  const isK8S = useSelector((state: AppState) => state.login.isK8S);

  useEffect(() => {
    if (navigateTo !== "") {
      dispatch(resetForm());
      navigate(navigateTo);
    }
  }, [navigateTo, dispatch, navigate]);

  const formSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    dispatch(doLoginAsync());
  };

  useEffect(() => {
    if (loadingFetchConfiguration) {
      dispatch(getFetchConfigurationAsync());
    }
  }, [loadingFetchConfiguration, dispatch]);

  useEffect(() => {
    if (loadingVersion) {
      dispatch(getVersionAsync());
    }
  }, [dispatch, loadingVersion]);

  let loginComponent;

  switch (loginStrategy.loginStrategy) {
    case loginStrategyType.form: {
      loginComponent = <StrategyForm />;
      break;
    }
    case loginStrategyType.redirect:
    case loginStrategyType.redirectServiceAccount: {
      let redirectItems: redirectRule[] = [];

      if (
        loginStrategy.redirectRules &&
        loginStrategy.redirectRules.length > 0
      ) {
        redirectItems = [...loginStrategy.redirectRules].sort(redirectRules);
      }

      if (
        loginStrategy.redirectRules &&
        loginStrategy.redirectRules.length > 1
      ) {
        const rules = redirectItems.map((r) => ({
          icon: <LogoutIcon />,
          value: r.redirect,
          label: r.displayName,
        }));

        loginComponent = (
          <Fragment>
            <div className={"loginSsoText"}>Login with SSO:</div>
            <Select
              id="alternativeMethods"
              name="alternativeMethods"
              fixedLabel="Other Authentication Methods"
              options={rules}
              onChange={(newValue) => {
                if (newValue) {
                  window.location.href = newValue;
                }
              }}
              value={""}
            />
          </Fragment>
        );
      } else if (redirectItems.length === 1) {
        loginComponent = (
          <div className={"submit, ssoSubmit"}>
            <Button
              key={`login-button`}
              variant="callAction"
              id="sso-login"
              label={
                redirectItems[0].displayName === ""
                  ? "Login with SSO"
                  : redirectItems[0].displayName
              }
              onClick={() => (window.location.href = redirectItems[0].redirect)}
              fullWidth
            />
          </div>
        );
      } else {
        loginComponent = (
          <div className={"loginStrategyMessage"}>
            Cannot retrieve redirect from login strategy
          </div>
        );
      }
      break;
    }
    case loginStrategyType.serviceAccount: {
      loginComponent = (
        <Fragment>
          <form
            style={{
              width: "100%", // Fix IE 11 issue.
            }}
            noValidate
            onSubmit={formSubmit}
          >
            <Grid container>
              <Grid item xs={12}>
                <InputBox
                  required
                  fullWidth
                  id="jwt"
                  value={jwt}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                    dispatch(setJwt(e.target.value))
                  }
                  name="jwt"
                  autoComplete="off"
                  disabled={loginSending}
                  placeholder={"Enter JWT"}
                  startIcon={<LockIcon />}
                />
              </Grid>
            </Grid>
            <Grid
              item
              xs={12}
              sx={{
                textAlign: "right",
                marginTop: 30,
              }}
            >
              <Button
                variant="callAction"
                id="do-login"
                disabled={jwt === "" || loginSending}
                label={"Login"}
                fullWidth
              />
            </Grid>
            <Grid
              item
              xs={12}
              sx={{
                height: 10,
              }}
            >
              {loginSending && <ProgressBar barHeight={5} />}
            </Grid>
          </form>
        </Fragment>
      );
      break;
    }
    default:
      loginComponent = (
        <Box
          sx={{
            textAlign: "center",
            "& .loadingLoginStrategy": {
              textAlign: "center",
              width: 40,
              height: 40,
            },
            "& .buttonRetry": {
              display: "flex",
              justifyContent: "center",
            },
          }}
        >
          {loadingFetchConfiguration ? (
            <Loader className={"loadingLoginStrategy"} />
          ) : (
            <Fragment>
              <div>
                <p style={{ color: "#000", textAlign: "center" }}>
                  An error has occurred
                  <br />
                  The backend cannot be reached.
                </p>
              </div>
              <div className={"buttonRetry"}>
                <Button
                  onClick={() => {
                    dispatch(getFetchConfigurationAsync());
                  }}
                  icon={<RefreshIcon />}
                  iconLocation={"end"}
                  variant="regular"
                  id="retry"
                  label={"Retry"}
                />
              </div>
            </Fragment>
          )}
        </Box>
      );
  }

  const logoVar = getLogoVar();

  let docsURL = "https://min.io/docs/minio/linux/index.html?ref=op";
  if (isK8S) {
    docsURL = "https://min.io/docs/minio/kubernetes/upstream/index.html?ref=op";
  }

  return (
    <Fragment>
      <MainError />
      <LoginWrapper
        logoProps={{ applicationName: "operator", subVariant: logoVar }}
        form={loginComponent}
        backgroundAnimation={false}
        formFooter={
          <Fragment>
            <a href={docsURL} target="_blank" rel="noopener">
              Documentation
            </a>
            <span className={"separator"}>|</span>
            <a
              href="https://github.com/minio/minio"
              target="_blank"
              rel="noopener"
            >
              Github
            </a>
            <span className={"separator"}>|</span>
            <a
              href="https://subnet.min.io/?ref=op"
              target="_blank"
              rel="noopener"
            >
              Support
            </a>
            <span className={"separator"}>|</span>
            <a
              href="https://min.io/download/?ref=op"
              target="_blank"
              rel="noopener"
            >
              Download
            </a>
          </Fragment>
        }
        promoHeader={<Fragment>Multi-Cloud Object&nbsp;Store</Fragment>}
        promoInfo={
          <Fragment>
            MinIO's high-performance, Kubernetes-native object store is licensed
            under GNU AGPL v3 and is available on every cloud - public, private
            and edge. For more information on the terms of the license or to
            learn more about commercial licensing options visit the{" "}
            <a href={"https://min.io/pricing"}>pricing page</a>.
          </Fragment>
        }
      />
    </Fragment>
  );
};

export default Login;
