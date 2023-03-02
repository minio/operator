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

import React from "react";
import { Box } from "@mui/material";
import { FormTitle } from "./utils";
import SelectWrapper from "../Common/FormComponents/SelectWrapper/SelectWrapper";
import { setLoading, setSelectedSubnetOrganization } from "./registerSlice";
import { Button } from "mds";
import RegisterHelpBox from "./RegisterHelpBox";
import { useSelector } from "react-redux";
import { AppState, useAppDispatch } from "../../../store";
import { callRegister } from "./registerThunks";

const ClusterRegistrationForm = () => {
  const dispatch = useAppDispatch();

  const subnetAccessToken = useSelector(
    (state: AppState) => state.register.subnetAccessToken
  );
  const selectedSubnetOrganization = useSelector(
    (state: AppState) => state.register.selectedSubnetOrganization
  );
  const subnetOrganizations = useSelector(
    (state: AppState) => state.register.subnetOrganizations
  );
  const loading = useSelector((state: AppState) => state.register.loading);

  return (
    <Box
      sx={{
        display: "flex",
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
            marginTop: "15px",
            marginBottom: "15px",
            "& .title-text": {
              marginLeft: "0px",
            },
          }}
        >
          <FormTitle title={`Register MinIO cluster`} />
        </Box>
        <Box>
          <SelectWrapper
            id="subnet-organization"
            name="subnet-organization"
            onChange={(e) =>
              dispatch(setSelectedSubnetOrganization(e.target.value as string))
            }
            label="Select an organization"
            value={selectedSubnetOrganization}
            options={subnetOrganizations.map((organization) => ({
              label: organization.company,
              value: organization.accountId.toString(),
            }))}
          />
          <Box
            sx={{
              display: "flex",
              alignItems: "center",
              justifyContent: "flex-end",
              marginTop: "15px",
            }}
          >
            <Button
              id={"register-cluster"}
              onClick={() => () => {
                if (loading) {
                  return;
                }
                dispatch(setLoading(true));
                if (subnetAccessToken && selectedSubnetOrganization) {
                  dispatch(
                    callRegister({
                      token: subnetAccessToken,
                      account_id: selectedSubnetOrganization,
                    })
                  );
                }
              }}
              disabled={loading || subnetAccessToken.trim().length === 0}
              variant="callAction"
              label={"Register"}
            />
          </Box>
        </Box>
      </Box>
      <RegisterHelpBox />
    </Box>
  );
};

export default ClusterRegistrationForm;
