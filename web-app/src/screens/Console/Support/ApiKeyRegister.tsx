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

import React, { Fragment, useCallback, useEffect, useState } from "react";
import { Box } from "@mui/material";
import { Button, OnlineRegistrationIcon } from "mds";
import { FormTitle } from "./utils";
import InputBoxWrapper from "../Common/FormComponents/InputBoxWrapper/InputBoxWrapper";
import GetApiKeyModal from "./GetApiKeyModal";
import RegisterHelpBox from "./RegisterHelpBox";
import { SubnetLoginRequest, SubnetLoginResponse } from "../License/types";
import api from "../../../common/api";
import { useAppDispatch } from "../../../store";
import { setErrorSnackMessage } from "../../../systemSlice";
import { ErrorResponseHandler } from "../../../common/types";
import { spacingUtils } from "../Common/FormComponents/common/styleLibrary";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { useNavigate } from "react-router-dom";
import { IAM_PAGES } from "../../../common/SecureComponent/permissions";

interface IApiKeyRegister {
  classes: any;
  registerEndpoint: string;
}

const styles = (theme: Theme) =>
  createStyles({
    sizedLabel: {
      minWidth: "75px",
    },
    ...spacingUtils,
  });

const ApiKeyRegister = ({ classes, registerEndpoint }: IApiKeyRegister) => {
  const navigate = useNavigate();

  const [showApiKeyModal, setShowApiKeyModal] = useState(false);
  const [apiKey, setApiKey] = useState("");
  const [loading, setLoading] = useState(false);
  const [fromModal, setFromModal] = useState(false);
  const dispatch = useAppDispatch();

  const onRegister = useCallback(() => {
    if (loading) {
      return;
    }
    setLoading(true);
    let request: SubnetLoginRequest = { apiKey };
    api
      .invoke("POST", registerEndpoint, request)
      .then((resp: SubnetLoginResponse) => {
        setLoading(false);
        if (resp && resp.registered) {
          navigate(IAM_PAGES.LICENSE);
        }
      })
      .catch((err: ErrorResponseHandler) => {
        dispatch(setErrorSnackMessage(err));
        setLoading(false);
        reset();
      });
  }, [apiKey, dispatch, loading, registerEndpoint, navigate]);

  useEffect(() => {
    if (fromModal) {
      onRegister();
    }
  }, [fromModal, onRegister]);

  const reset = () => {
    setApiKey("");
    setFromModal(false);
  };

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
          title={`Register cluster with API key`}
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
            Use your MinIO Subscription Network API Key to register this
            cluster.
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
              id="api-key"
              name="api-key"
              onChange={(event: React.ChangeEvent<HTMLInputElement>) =>
                setApiKey(event.target.value)
              }
              label="API Key"
              value={apiKey}
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
                id={"get-from-subnet"}
                variant="regular"
                className={classes.spacerRight}
                disabled={loading}
                onClick={() => setShowApiKeyModal(true)}
                label={"Get from SUBNET"}
              />
              <Button
                id={"register"}
                type="submit"
                variant="callAction"
                disabled={loading || apiKey.trim().length === 0}
                onClick={() => onRegister()}
                label={"Register"}
              />
              <GetApiKeyModal
                open={showApiKeyModal}
                closeModal={() => setShowApiKeyModal(false)}
                onSet={(value) => {
                  setApiKey(value);
                  setFromModal(true);
                }}
              />
            </Box>
          </Box>
        </Box>
        <RegisterHelpBox />
      </Box>
    </Fragment>
  );
};

export default withStyles(styles)(ApiKeyRegister);
