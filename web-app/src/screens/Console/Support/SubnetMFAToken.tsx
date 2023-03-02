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
import InputBoxWrapper from "../Common/FormComponents/InputBoxWrapper/InputBoxWrapper";
import LockOutlinedIcon from "@mui/icons-material/LockOutlined";
import { setSubnetOTP } from "./registerSlice";
import { Button } from "mds";
import RegisterHelpBox from "./RegisterHelpBox";
import { AppState, useAppDispatch } from "../../../store";
import { useSelector } from "react-redux";
import { subnetLoginWithMFA } from "./registerThunks";

const SubnetMFAToken = () => {
  const dispatch = useAppDispatch();

  const subnetMFAToken = useSelector(
    (state: AppState) => state.register.subnetMFAToken
  );
  const subnetOTP = useSelector((state: AppState) => state.register.subnetOTP);
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
            fontSize: "16px",
            display: "flex",
            flexFlow: "column",
            marginTop: "30px",
            marginBottom: "30px",
          }}
        >
          Two-Factor Authentication
        </Box>

        <Box>
          Please enter the 6-digit verification code that was sent to your email
          address. This code will be valid for 5 minutes.
        </Box>

        <Box
          sx={{
            flex: "1",
            marginTop: "30px",
          }}
        >
          <InputBoxWrapper
            overlayIcon={<LockOutlinedIcon />}
            id="subnet-otp"
            name="subnet-otp"
            onChange={(event: React.ChangeEvent<HTMLInputElement>) =>
              dispatch(setSubnetOTP(event.target.value))
            }
            placeholder=""
            label=""
            value={subnetOTP}
          />
        </Box>
        <Box
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "flex-end",
          }}
        >
          <Button
            id={"verify"}
            onClick={() => dispatch(subnetLoginWithMFA())}
            disabled={
              loading ||
              subnetOTP.trim().length === 0 ||
              subnetMFAToken.trim().length === 0
            }
            variant="callAction"
            label={"Verify"}
          />
        </Box>
      </Box>

      <RegisterHelpBox />
    </Box>
  );
};
export default SubnetMFAToken;
