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
import {
  Button,
  OnlineRegistrationIcon,
  Box,
  breakPoints,
  InputBox,
} from "mds";
import { useNavigate } from "react-router-dom";
import { FormTitle } from "./utils";
import { SubnetLoginRequest, SubnetLoginResponse } from "../License/types";
import { useAppDispatch } from "../../../store";
import { setErrorSnackMessage } from "../../../systemSlice";
import { ErrorResponseHandler } from "../../../common/types";
import { IAM_PAGES } from "../../../common/SecureComponent/permissions";
import api from "../../../common/api";
import GetApiKeyModal from "./GetApiKeyModal";
import RegisterHelpBox from "./RegisterHelpBox";

interface IApiKeyRegister {
  registerEndpoint: string;
}

const ApiKeyRegister = ({ registerEndpoint }: IApiKeyRegister) => {
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
          flexFlow: "row",

          [`@media (max-width: ${breakPoints.sm}px)`]: {
            flexFlow: "column",
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
              fontSize: 16,
              display: "flex",
              flexFlow: "column",
              marginTop: 30,
              marginBottom: 30,
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
            <InputBox
              id="api-key"
              name="api-key"
              onChange={(event: React.ChangeEvent<HTMLInputElement>) =>
                setApiKey(event.target.value)
              }
              label="API Key"
              value={apiKey}
              sx={{
                minWidth: "75px",
                marginBottom: 15,
              }}
            />

            <Box
              sx={{
                display: "flex",
                alignItems: "center",
                justifyContent: "flex-end",
                gap: 10,
              }}
            >
              <Button
                id={"get-from-subnet"}
                variant="regular"
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

export default ApiKeyRegister;
