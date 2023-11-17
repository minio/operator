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

import React from "react";
import { Box, Button, CopyIcon, InputLabel, ReadBox } from "mds";
import CopyToClipboard from "react-copy-to-clipboard";
import { useAppDispatch } from "../../../../store";
import { setModalSnackMessage } from "../../../../systemSlice";

interface ICredentialsItem {
  label?: string;
  value?: string;
}

const CredentialItem = ({ label = "", value = "" }: ICredentialsItem) => {
  const dispatch = useAppDispatch();

  return (
    <Box sx={{ marginTop: 12 }}>
      <InputLabel>{label}</InputLabel>
      <ReadBox
        actionButton={
          <CopyToClipboard text={value}>
            <Button
              id={"copy-path"}
              variant="regular"
              onClick={() => {
                dispatch(setModalSnackMessage(`${label} copied to clipboard`));
              }}
              style={{
                marginRight: "5px",
                width: "28px",
                height: "28px",
                padding: "0px",
              }}
              icon={<CopyIcon />}
            />
          </CopyToClipboard>
        }
      >
        {value}
      </ReadBox>
    </Box>
  );
};

export default CredentialItem;
