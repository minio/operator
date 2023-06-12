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

import React, {
  Fragment,
  Suspense,
  useEffect,
  useLayoutEffect,
  useState,
} from "react";
import { Theme } from "@mui/material/styles";
import { MainContainer } from "mds";
import debounce from "lodash/debounce";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { LinearProgress } from "@mui/material";
import CssBaseline from "@mui/material/CssBaseline";
import Snackbar from "@mui/material/Snackbar";
import { Navigate, Route, Routes, useLocation } from "react-router-dom";
import { useSelector } from "react-redux";
import { AppState, useAppDispatch } from "../../store";
import { snackBarCommon } from "./Common/FormComponents/common/styleLibrary";
import AppMenu from "./Menu/AppMenu";
import MainError from "./Common/MainError/MainError";
import {
  CONSOLE_UI_RESOURCE,
  IAM_PAGES,
  IAM_PAGES_PERMISSIONS,
} from "../../common/SecureComponent/permissions";
import { hasPermission } from "../../common/SecureComponent";
import { IRouteRule } from "./Menu/types";
import LoadingComponent from "../../common/LoadingComponent";
import EditPool from "./Tenants/TenantDetails/Pools/EditPool/EditPool";
import ComponentsScreen from "./Common/ComponentsScreen";
import { menuOpen, setSnackBarMessage } from "../../systemSlice";
import { selFeatures, selSession } from "./consoleSlice";

const Hop = React.lazy(() => import("./Tenants/TenantDetails/hop/Hop"));
const RegisterOperator = React.lazy(() => import("./Support/RegisterOperator"));

const AddTenant = React.lazy(() => import("./Tenants/AddTenant/AddTenant"));

const ListTenants = React.lazy(
  () => import("./Tenants/ListTenants/ListTenants")
);

const IconsScreen = React.lazy(() => import("./Common/IconsScreen"));

const TenantDetails = React.lazy(
  () => import("./Tenants/TenantDetails/TenantDetails")
);
const License = React.lazy(() => import("./License/License"));
const Marketplace = React.lazy(() => import("./Marketplace/Marketplace"));
const AddPool = React.lazy(
  () => import("./Tenants/TenantDetails/Pools/AddPool/AddPool")
);

const styles = (theme: Theme) =>
  createStyles({
    root: {
      display: "flex",
      "& .MuiPaper-root.MuiSnackbarContent-root": {
        borderRadius: "0px 0px 5px 5px",
        boxShadow: "none",
      },
    },
    content: {
      flexGrow: 1,
      height: "100vh",
      overflow: "auto",
      position: "relative",
    },
    warningBar: {
      background: theme.palette.primary.main,
      color: "white",
      heigh: "60px",
      widht: "100%",
      lineHeight: "60px",
      display: "flex",
      justifyContent: "center",
      alignItems: "center",
      "& button": {
        marginLeft: 8,
      },
    },
    progress: {
      height: "3px",
      backgroundColor: "#eaeaea",
    },
    ...snackBarCommon,
  });

interface IConsoleProps {
  classes: any;
}

const Console = ({ classes }: IConsoleProps) => {
  const dispatch = useAppDispatch();
  const { pathname = "" } = useLocation();
  const open = useSelector((state: AppState) => state.system.sidebarOpen);
  const session = useSelector(selSession);
  const features = useSelector(selFeatures);
  const snackBarMessage = useSelector(
    (state: AppState) => state.system.snackBar
  );
  const loadingProgress = useSelector(
    (state: AppState) => state.system.loadingProgress
  );

  const [openSnackbar, setOpenSnackbar] = useState<boolean>(false);

  const obOnly = !!features?.includes("object-browser-only");

  // Layout effect to be executed after last re-render for resizing only
  useLayoutEffect(() => {
    // Debounce to not execute constantly
    const debounceSize = debounce(() => {
      if (open && window.innerWidth <= 1024) {
        dispatch(menuOpen(false));
      }
    }, 300);

    // Added event listener for window resize
    window.addEventListener("resize", debounceSize);

    // We remove the listener on component unmount
    return () => window.removeEventListener("resize", debounceSize);
  });

  const operatorConsoleRoutes: IRouteRule[] = [
    {
      component: ListTenants,
      path: IAM_PAGES.TENANTS,
      forceDisplay: true,
    },
    {
      component: AddTenant,
      path: IAM_PAGES.TENANTS_ADD,
      forceDisplay: true,
    },
    {
      component: TenantDetails,
      path: IAM_PAGES.NAMESPACE_TENANT,
      forceDisplay: true,
    },
    {
      component: Hop,
      path: IAM_PAGES.NAMESPACE_TENANT_HOP,
      forceDisplay: true,
    },
    {
      component: AddPool,
      path: IAM_PAGES.NAMESPACE_TENANT_POOLS_ADD,
      forceDisplay: true,
    },
    {
      component: EditPool,
      path: IAM_PAGES.NAMESPACE_TENANT_POOLS_EDIT,
      forceDisplay: true,
    },
    {
      component: License,
      path: IAM_PAGES.LICENSE,
      forceDisplay: true,
    },
    {
      component: RegisterOperator,
      path: IAM_PAGES.REGISTER_SUPPORT,
      forceDisplay: true,
    },
    {
      component: Marketplace,
      path: IAM_PAGES.OPERATOR_MARKETPLACE,
      forceDisplay: true,
    },
  ];

  let routes = operatorConsoleRoutes;

  const allowedRoutes = routes.filter((route: any) =>
    obOnly
      ? route.path.includes("buckets")
      : (route.forceDisplay ||
          (route.customPermissionFnc
            ? route.customPermissionFnc()
            : hasPermission(
                CONSOLE_UI_RESOURCE,
                IAM_PAGES_PERMISSIONS[route.path]
              ))) &&
        !route.fsHidden
  );

  const closeSnackBar = () => {
    setOpenSnackbar(false);
    dispatch(setSnackBarMessage(""));
  };

  useEffect(() => {
    if (snackBarMessage.message === "") {
      setOpenSnackbar(false);
      return;
    }
    // Open SnackBar
    if (snackBarMessage.type !== "error") {
      setOpenSnackbar(true);
    }
  }, [snackBarMessage]);

  let hideMenu = false;
  if (features?.includes("hide-menu")) {
    hideMenu = true;
  } else if (pathname.endsWith("/hop")) {
    hideMenu = true;
  } else if (obOnly) {
    hideMenu = true;
  }

  return (
    <MainContainer menu={!hideMenu ? <AppMenu /> : <Fragment />}>
      {session && session.status === "ok" ? (
        <div className={classes.root}>
          <CssBaseline />

          <main className={classes.content}>
            {loadingProgress < 100 && (
              <LinearProgress
                className={classes.progress}
                variant="determinate"
                value={loadingProgress}
              />
            )}
            <MainError />
            <div className={classes.snackDiv}>
              <Snackbar
                open={openSnackbar}
                onClose={() => {
                  closeSnackBar();
                }}
                autoHideDuration={
                  snackBarMessage.type === "error" ? 10000 : 5000
                }
                message={snackBarMessage.message}
                className={classes.snackBarExternal}
                ContentProps={{
                  className: `${classes.snackBar} ${
                    snackBarMessage.type === "error"
                      ? classes.errorSnackBar
                      : ""
                  }`,
                }}
              />
            </div>

            <Routes>
              {allowedRoutes.map((route: any) => (
                <Route
                  key={route.path}
                  path={`${route.path}/*`}
                  element={
                    <Suspense fallback={<LoadingComponent />}>
                      <route.component {...route.props} />
                    </Suspense>
                  }
                />
              ))}
              <Route
                key={"icons"}
                path={"icons"}
                element={
                  <Suspense fallback={<LoadingComponent />}>
                    <IconsScreen />
                  </Suspense>
                }
              />
              <Route
                key={"components"}
                path={"components"}
                element={
                  <Suspense fallback={<LoadingComponent />}>
                    <ComponentsScreen />
                  </Suspense>
                }
              />
              <Route
                path={"*"}
                element={
                  <Fragment>
                    {allowedRoutes.length > 0 ? (
                      <Navigate to={allowedRoutes[0].path} />
                    ) : (
                      <Fragment />
                    )}
                  </Fragment>
                }
              />
            </Routes>
          </main>
        </div>
      ) : (
        <Fragment />
      )}
    </MainContainer>
  );
};

export default withStyles(styles)(Console);
