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

import React, { useState } from "react";
import { Box, FormLayout, InfoIcon, InputBox, UsersIcon, LockIcon } from "mds";
import { ErrorResponseHandler } from "../../../common/types";
import { useAppDispatch } from "../../../store";
import { setErrorSnackMessage } from "../../../systemSlice";
import ConfirmDialog from "../Common/ModalWrapper/ConfirmDialog";
import useApi from "../Common/Hooks/useApi";

interface IGetApiKeyModalProps {
  open: boolean;
  closeModal: () => void;
  onSet: (apiKey: string) => void;
}

const GetApiKeyModal = ({ open, closeModal, onSet }: IGetApiKeyModalProps) => {
  const dispatch = useAppDispatch();
  const [email, setEmail] = useState<string>("");
  const [password, setPassword] = useState("");
  const [mfaToken, setMfaToken] = useState("");
  const [subnetOTP, setSubnetOTP] = useState("");

  const onError = (err: ErrorResponseHandler) => {
    dispatch(setErrorSnackMessage(err));
    closeModal();
    setEmail("");
    setPassword("");
    setMfaToken("");
    setSubnetOTP("");
  };

  const onSuccess = (res: any) => {
    if (res.mfa_token) {
      setMfaToken(res.mfa_token);
    } else if (res.access_token) {
      invokeApi("GET", `/api/v1/subnet/apikey?token=${res.access_token}`);
    } else {
      onSet(res.apiKey);
      closeModal();
    }
  };

  const [isLoading, invokeApi] = useApi(onSuccess, onError);

  const onConfirm = () => {
    if (mfaToken !== "") {
      invokeApi("POST", "/api/v1/subnet/login/mfa", {
        username: email,
        otp: subnetOTP,
        mfa_token: mfaToken,
      });
    } else {
      invokeApi("POST", "/api/v1/subnet/login", { username: email, password });
    }
  };

  const getDialogContent = () => {
    if (mfaToken === "") {
      return getCredentialsDialog();
    }
    return getMFADialog();
  };

  const getCredentialsDialog = () => {
    return (
      <FormLayout withBorders={false} containerPadding={false}>
        <InputBox
          id="subnet-email"
          name="subnet-email"
          onChange={(event: React.ChangeEvent<HTMLInputElement>) =>
            setEmail(event.target.value)
          }
          label="Email"
          value={email}
          overlayIcon={<UsersIcon />}
        />
        <InputBox
          id="subnet-password"
          name="subnet-password"
          onChange={(event: React.ChangeEvent<HTMLInputElement>) =>
            setPassword(event.target.value)
          }
          label="Password"
          type={"password"}
          value={password}
        />
      </FormLayout>
    );
  };

  const getMFADialog = () => {
    return (
      <Box sx={{ display: "flex" }}>
        <Box sx={{ display: "flex", flexFlow: "column", flex: "2" }}>
          <Box
            sx={{
              fontSize: 14,
              display: "flex",
              flexFlow: "column",
              marginTop: 20,
              marginBottom: 20,
            }}
          >
            Two-Factor Authentication
          </Box>

          <Box>
            Please enter the 6-digit verification code that was sent to your
            email address. This code will be valid for 5 minutes.
          </Box>

          <Box
            sx={{
              flex: "1",
              marginTop: "30px",
            }}
          >
            <InputBox
              overlayIcon={<LockIcon />}
              id="subnet-otp"
              name="subnet-otp"
              onChange={(event: React.ChangeEvent<HTMLInputElement>) =>
                setSubnetOTP(event.target.value)
              }
              placeholder=""
              label=""
              value={subnetOTP}
            />
          </Box>
        </Box>
      </Box>
    );
  };

  return open ? (
    <ConfirmDialog
      title={"Get API Key from SUBNET"}
      confirmText={"Get API Key"}
      isOpen={open}
      titleIcon={<InfoIcon />}
      isLoading={isLoading}
      cancelText={"Cancel"}
      onConfirm={onConfirm}
      onClose={closeModal}
      confirmButtonProps={{
        variant: "callAction",
        disabled: !email || !password || isLoading,
        hidden: true,
      }}
      cancelButtonProps={{
        disabled: isLoading,
      }}
      confirmationContent={getDialogContent()}
    />
  ) : null;
};

export default GetApiKeyModal;
