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

import React, { Fragment, useEffect, useState } from "react";
import { Box } from "@mui/material";
import { Button, CallHomeMenuIcon, HelpBox, Loader } from "mds";
import { Link, useNavigate } from "react-router-dom";
import PageLayout from "../Common/Layout/PageLayout";
import api from "../../../common/api";
import { ErrorResponseHandler } from "../../../common/types";
import { setErrorSnackMessage } from "../../../systemSlice";
import { useAppDispatch } from "../../../store";
import { ICallHomeResponse } from "./types";
import { registeredCluster } from "../../../config";
import CallHomeConfirmation from "./CallHomeConfirmation";
import RegisterCluster from "./RegisterCluster";
import FormSwitchWrapper from "../Common/FormComponents/FormSwitchWrapper/FormSwitchWrapper";
import PageHeaderWrapper from "../Common/PageHeaderWrapper/PageHeaderWrapper";

const PromoLabels = ({ title, text }: { title: string; text: string }) => {
  return (
    <div style={{ marginTop: 15 }}>
      <div style={{ marginBottom: 10, fontWeight: "bold" }}>{title}</div>
      <div style={{ color: "#969696", fontSize: 12, marginBottom: 40 }}>
        {text}
      </div>
    </div>
  );
};

const CallHome = () => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();

  const [loading, setLoading] = useState<boolean>(true);
  const [showConfirmation, setShowConfirmation] = useState<boolean>(false);
  const [diagEnabled, setDiagEnabled] = useState<boolean>(false);
  const [oDiagEnabled, setODiagEnabled] = useState<boolean>(false);
  const [oLogsEnabled, setOLogsEnabled] = useState<boolean>(false);
  const [logsEnabled, setLogsEnabled] = useState<boolean>(false);
  const [disableMode, setDisableMode] = useState<boolean>(false);

  const clusterRegistered = registeredCluster();

  useEffect(() => {
    if (loading) {
      api
        .invoke("GET", `/api/v1/support/callhome`)
        .then((res: ICallHomeResponse) => {
          setLoading(false);

          setDiagEnabled(!!res.diagnosticsStatus);
          setLogsEnabled(!!res.logsStatus);

          setODiagEnabled(!!res.diagnosticsStatus);
          setOLogsEnabled(!!res.logsStatus);
        })
        .catch((err: ErrorResponseHandler) => {
          setLoading(false);
          dispatch(setErrorSnackMessage(err));
        });
    }
  }, [loading, dispatch]);

  const callHomeClose = (refresh: boolean) => {
    if (refresh) {
      setLoading(true);
    }
    setShowConfirmation(false);
  };

  const confirmCallHomeAction = () => {
    if (!clusterRegistered) {
      navigate("/support/register");
      return;
    }
    setDisableMode(false);
    setShowConfirmation(true);
  };

  const disableCallHomeAction = () => {
    setDisableMode(true);
    setShowConfirmation(true);
  };

  let mainVariant: "regular" | "callAction" = "regular";

  if (
    clusterRegistered &&
    (diagEnabled !== oDiagEnabled || logsEnabled !== oLogsEnabled)
  ) {
    mainVariant = "callAction";
  }

  return (
    <Fragment>
      {showConfirmation && (
        <CallHomeConfirmation
          onClose={callHomeClose}
          open={showConfirmation}
          logsStatus={logsEnabled}
          diagStatus={diagEnabled}
          disable={disableMode}
        />
      )}
      <PageHeaderWrapper label="Call Home" />
      <PageLayout>
        {!clusterRegistered && <RegisterCluster compactMode />}
        <Box
          sx={{
            display: "flex",
            alignItems: "flex-start",
            justifyContent: "flex-start",
            border: "1px solid #eaeaea",
            padding: {
              lg: "40px",
              xs: "15px",
            },
            flexWrap: "wrap",
            gap: {
              lg: "55px",
              xs: "20px",
            },
            height: {
              md: "calc(100vh - 120px)",
              xs: "100%",
            },
            flexFlow: {
              lg: "row",
              xs: "column",
            },
          }}
        >
          <Box
            sx={{
              border: "1px solid #eaeaea",
              flex: {
                md: 2,
                xs: 1,
              },
              width: {
                lg: "auto",
                xs: "100%",
              },
              padding: {
                lg: "40px",
                xs: "15px",
              },
            }}
          >
            {loading ? (
              <span style={{ marginLeft: 5 }}>
                <Loader style={{ width: 16, height: 16 }} />
              </span>
            ) : (
              <Fragment>
                <div style={{ marginBottom: 25 }}>
                  <FormSwitchWrapper
                    value="enableDiag"
                    id="enableDiag"
                    name="enableDiag"
                    checked={diagEnabled}
                    onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                      setDiagEnabled(event.target.checked);
                    }}
                    label={"Daily Health Report"}
                    disabled={!clusterRegistered}
                  />
                  <PromoLabels
                    title={"When you enable diagnostics"}
                    text={
                      "Daily Health Report enables you to proactively identify potential issues in your deployment before they escalate."
                    }
                  />
                </div>
                <div>
                  <FormSwitchWrapper
                    value="enableLogs"
                    id="enableLogs"
                    name="enableLogs"
                    checked={logsEnabled}
                    onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                      setLogsEnabled(event.target.checked);
                    }}
                    label={"Live Error Logs"}
                    disabled={!clusterRegistered}
                  />
                  <PromoLabels
                    title={"When you enable logs"}
                    text={
                      "Live Error Logs will enable MinIO's support team and automatic diagnostics system to catch failures early."
                    }
                  />
                </div>
                <Box
                  sx={{
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "flex-end",
                    marginTop: "55px",
                    gap: "0px 10px",
                  }}
                >
                  {(oDiagEnabled || oLogsEnabled) && (
                    <Button
                      id={"callhome-action"}
                      variant={"secondary"}
                      data-test-id="call-home-toggle-button"
                      onClick={disableCallHomeAction}
                      disabled={loading}
                    >
                      Disable Call Home
                    </Button>
                  )}
                  <Button
                    id={"callhome-action"}
                    type="button"
                    variant={mainVariant}
                    data-test-id="call-home-toggle-button"
                    onClick={confirmCallHomeAction}
                    disabled={loading}
                  >
                    Save Configuration
                  </Button>
                </Box>
              </Fragment>
            )}
          </Box>
          <Box
            sx={{
              flex: 1,
              minWidth: {
                md: "365px",
                xs: "100%",
              },
              width: "100%",
            }}
          >
            <HelpBox
              title={""}
              iconComponent={null}
              help={
                <Fragment>
                  <Box
                    sx={{
                      marginTop: "-25px",
                      fontSize: "16px",
                      fontWeight: 600,
                      display: "flex",
                      alignItems: "center",
                      justifyContent: "flex-start",
                      padding: "2px",
                    }}
                  >
                    <Box
                      sx={{
                        backgroundColor: "#07193E",
                        height: "15px",
                        width: "15px",
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                        borderRadius: "50%",
                        marginRight: "18px",
                        padding: "3px",
                        "& .min-icon": {
                          height: "11px",
                          width: "11px",
                          fill: "#ffffff",
                        },
                      }}
                    >
                      <CallHomeMenuIcon />
                    </Box>
                    Learn more about Call Home
                  </Box>

                  <Box
                    sx={{
                      display: "flex",
                      flexFlow: "column",
                      fontSize: "14px",
                      flex: "2",
                      marginTop: "10px",
                    }}
                  >
                    <Box>
                      Enabling Call Home sends cluster health & status to your
                      registered MinIO Subscription Network account every 24
                      hours.
                      <br />
                      <br />
                      This helps the MinIO support team to provide quick
                      incident responses along with suggestions for possible
                      improvements that can be made to your MinIO instances.
                      <br />
                      <br />
                      Your cluster must be{" "}
                      <Link to={"/support/register"}>registered</Link> in the
                      MinIO Subscription Network (SUBNET) before enabling this
                      feature.
                    </Box>
                  </Box>
                </Fragment>
              }
            />
          </Box>
        </Box>
      </PageLayout>
    </Fragment>
  );
};

export default CallHome;
