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
import CredentialsPrompt from "../../Common/CredentialsPrompt/CredentialsPrompt";
import { resetAddTenantForm } from "./createTenantSlice";
import { useSelector } from "react-redux";
import { AppState, useAppDispatch } from "../../../../store";
import { useNavigate } from "react-router-dom";

const NewTenantCredentials = () => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();

  const showNewCredentials = useSelector(
    (state: AppState) => state.createTenant.showNewCredentials,
  );
  const createdAccount = useSelector(
    (state: AppState) => state.createTenant.createdAccount,
  );

  return (
    <Fragment>
      {showNewCredentials && (
        <CredentialsPrompt
          newServiceAccount={createdAccount}
          open={showNewCredentials}
          closeModal={() => {
            dispatch(resetAddTenantForm());
            navigate("/tenants");
          }}
          entity="Tenant"
        />
      )}
    </Fragment>
  );
};

export default NewTenantCredentials;
