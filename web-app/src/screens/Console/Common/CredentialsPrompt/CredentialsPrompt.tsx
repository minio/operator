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

import React, { Fragment } from "react";
import get from "lodash/get";
import styled from "styled-components";
import {
  Box,
  Button,
  DownloadIcon,
  ServiceAccountCredentialsIcon,
  WarnIcon,
  Grid,
} from "mds";
import { NewServiceAccount } from "./types";
import ModalWrapper from "../ModalWrapper/ModalWrapper";
import CredentialItem from "./CredentialItem";
import TooltipWrapper from "../TooltipWrapper/TooltipWrapper";
import { modalStyleUtils } from "../FormComponents/common/styleLibrary";

const WarningBlock = styled.div(({ theme }) => ({
  color: get(theme, "signalColors.danger", "#C51B3F"),
  fontSize: ".85rem",
  margin: ".5rem 0 .5rem 0",
  display: "flex",
  alignItems: "center",
  "& svg ": {
    marginRight: ".3rem",
    height: 16,
    width: 16,
  },
}));

interface ICredentialsPromptProps {
  newServiceAccount: NewServiceAccount | null;
  open: boolean;
  entity: string;
  closeModal: () => void;
}

const download = (filename: string, text: string) => {
  let element = document.createElement("a");
  element.setAttribute("href", "data:text/plain;charset=utf-8," + text);
  element.setAttribute("download", filename);

  element.style.display = "none";
  document.body.appendChild(element);

  element.click();
  document.body.removeChild(element);
};

const CredentialsPrompt = ({
  newServiceAccount,
  open,
  closeModal,
  entity,
}: ICredentialsPromptProps) => {
  if (!newServiceAccount) {
    return null;
  }
  const consoleCreds = get(newServiceAccount, "console", null);
  const idp = get(newServiceAccount, "idp", false);

  const downloadImport = () => {
    let consoleExtras = {};

    if (consoleCreds) {
      if (!Array.isArray(consoleCreds)) {
        consoleExtras = {
          url: consoleCreds.url,
          accessKey: consoleCreds.accessKey,
          secretKey: consoleCreds.secretKey,
          api: "s3v4",
          path: "auto",
        };
      } else {
        const cCreds = consoleCreds.map((itemMap) => {
          return {
            url: itemMap.url,
            accessKey: itemMap.accessKey,
            secretKey: itemMap.secretKey,
            api: "s3v4",
            path: "auto",
          };
        });
        consoleExtras = cCreds[0];
      }
    } else {
      consoleExtras = {
        url: newServiceAccount.url,
        accessKey: newServiceAccount.accessKey,
        secretKey: newServiceAccount.secretKey,
        api: "s3v4",
        path: "auto",
      };
    }

    download(
      "credentials.json",
      JSON.stringify({
        ...consoleExtras,
      }),
    );
  };

  const downloaddAllCredentials = () => {
    let allCredentials = {};
    if (
      consoleCreds &&
      Array.isArray(consoleCreds) &&
      consoleCreds.length > 1
    ) {
      const cCreds = consoleCreds.map((itemMap) => {
        return {
          accessKey: itemMap.accessKey,
          secretKey: itemMap.secretKey,
        };
      });
      allCredentials = cCreds;
    }
    download(
      "all_credentials.json",
      JSON.stringify({
        ...allCredentials,
      }),
    );
  };

  return (
    <ModalWrapper
      modalOpen={open}
      onClose={() => {
        closeModal();
      }}
      title={`New ${entity} Created`}
      titleIcon={<ServiceAccountCredentialsIcon />}
    >
      <Grid container>
        <Grid item xs={12}>
          A new {entity} has been created with the following details:
          {!idp && consoleCreds && (
            <Fragment>
              <Grid
                item
                xs={12}
                sx={{
                  overflowY: "auto",
                  maxHeight: 350,
                }}
              >
                <Box
                  sx={{
                    padding: ".8rem 0 0 0",
                    fontWeight: 600,
                    fontSize: ".9rem",
                  }}
                >
                  Console Credentials
                </Box>
                {Array.isArray(consoleCreds) &&
                  consoleCreds.map((credentialsPair, index) => {
                    return (
                      <Fragment>
                        <CredentialItem
                          label="Access Key"
                          value={credentialsPair.accessKey}
                        />
                        <CredentialItem
                          label="Secret Key"
                          value={credentialsPair.secretKey}
                        />
                      </Fragment>
                    );
                  })}
                {!Array.isArray(consoleCreds) && (
                  <Fragment>
                    <CredentialItem
                      label="Access Key"
                      value={consoleCreds.accessKey}
                    />
                    <CredentialItem
                      label="Secret Key"
                      value={consoleCreds.secretKey}
                    />
                  </Fragment>
                )}
              </Grid>
            </Fragment>
          )}
          {(consoleCreds === null || consoleCreds === undefined) && (
            <>
              <CredentialItem
                label="Access Key"
                value={newServiceAccount.accessKey || ""}
              />
              <CredentialItem
                label="Secret Key"
                value={newServiceAccount.secretKey || ""}
              />
            </>
          )}
          {idp ? (
            <WarningBlock>
              Please Login via the configured external identity provider.
            </WarningBlock>
          ) : (
            <WarningBlock>
              <WarnIcon />
              <span>
                Write these down, as this is the only time the secret will be
                displayed.
              </span>
            </WarningBlock>
          )}
        </Grid>
        <Grid item xs={12} sx={{ ...modalStyleUtils.modalButtonBar }}>
          {!idp && (
            <Fragment>
              <TooltipWrapper
                tooltip={
                  "Download credentials in a JSON file formatted for import using mc alias import. This will only include the default login credentials."
                }
              >
                <Button
                  id={"download-button"}
                  label={"Download for import"}
                  onClick={downloadImport}
                  icon={<DownloadIcon />}
                  variant="callAction"
                />
              </TooltipWrapper>

              {Array.isArray(consoleCreds) && consoleCreds.length > 1 && (
                <TooltipWrapper
                  tooltip={
                    "Download all access credentials to a JSON file. NOTE: This file is not formatted for import using mc alias import. If you plan to import this alias from the file, please use the Download for Import button. "
                  }
                >
                  <Button
                    id={"download-all-button"}
                    label={"Download all access credentials"}
                    onClick={downloaddAllCredentials}
                    icon={<DownloadIcon />}
                    variant="callAction"
                    color="primary"
                  />
                </TooltipWrapper>
              )}
            </Fragment>
          )}
        </Grid>
      </Grid>
    </ModalWrapper>
  );
};

export default CredentialsPrompt;
