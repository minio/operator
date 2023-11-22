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

import React, { Fragment, useState } from "react";
import { InfoIcon, InputBox } from "mds";
import { ISetEmailModalProps } from "./types";
import { ErrorResponseHandler } from "../../../common/types";
import { setErrorSnackMessage, setSnackBarMessage } from "../../../systemSlice";
import { useAppDispatch } from "../../../store";
import { euTimezones } from "./euTimezones";
import ConfirmDialog from "../Common/ModalWrapper/ConfirmDialog";
import useApi from "../Common/Hooks/useApi";

const reEmail =
  // eslint-disable-next-line
  /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;

const SetEmailModal = ({ open, closeModal }: ISetEmailModalProps) => {
  const dispatch = useAppDispatch();

  const onError = (err: ErrorResponseHandler) => {
    dispatch(setErrorSnackMessage(err));
    closeModal();
  };

  const onSuccess = (res: any) => {
    let msg = `Email ${email} has been saved`;
    dispatch(setSnackBarMessage(msg));
    closeModal();
  };

  const [isLoading, invokeApi] = useApi(onSuccess, onError);
  const [email, setEmail] = useState<string>("");
  const [isEmailSet, setIsEmailSet] = useState<boolean>(false);

  const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    let v = event.target.value;
    setIsEmailSet(reEmail.test(v));
    setEmail(v);
  };

  const onConfirm = () => {
    const isInEU = isEU();
    invokeApi("POST", "/api/v1/mp-integration", { email, isInEU });
  };

  const isEU = () => {
    const tz = Intl.DateTimeFormat().resolvedOptions().timeZone;
    return euTimezones.includes(tz.toLocaleLowerCase());
  };

  return open ? (
    <ConfirmDialog
      title={"Register Email"}
      confirmText={"Register"}
      isOpen={open}
      titleIcon={<InfoIcon />}
      isLoading={isLoading}
      cancelText={"Later"}
      onConfirm={onConfirm}
      onClose={closeModal}
      confirmButtonProps={{
        color: "info",
        disabled: !isEmailSet || isLoading,
      }}
      confirmationContent={
        <Fragment>
          <p>
            Your Marketplace subscription includes support access from the
            <a
              href="https://min.io/product/subnet"
              target="_blank"
              rel="noopener"
            >
              MinIO Subscription Network (SUBNET)
            </a>
            .
            <br />
            Enter your email to register now.
          </p>
          <p>
            To register later, contact{" "}
            <a href="mailto: support@min.io">support@min.io</a>.
          </p>
          <InputBox
            id="set-mp-email"
            name="set-mp-email"
            onChange={handleInputChange}
            label={""}
            placeholder="Enter email"
            type={"email"}
            value={email}
          />
        </Fragment>
      }
    />
  ) : null;
};

export default SetEmailModal;
