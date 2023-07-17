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
import get from "lodash/get";
import { Theme } from "@mui/material/styles";
import {
  Button,
  DownloadIcon,
  ServiceAccountCredentialsIcon,
  WarnIcon,
} from "mds";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { NewServiceAccount } from "./types";
import ModalWrapper from "../ModalWrapper/ModalWrapper";
import Grid from "@mui/material/Grid";
import CredentialItem from "./CredentialItem";
import TooltipWrapper from "../TooltipWrapper/TooltipWrapper";

const styles = (theme: Theme) =>
  createStyles({
    warningBlock: {
      color: "red",
      fontSize: ".85rem",
      margin: ".5rem 0 .5rem 0",
      display: "flex",
      alignItems: "center",
      "& svg ": {
        marginRight: ".3rem",
        height: 16,
        width: 16,
      },
    },
    credentialTitle: {
      padding: ".8rem 0 0 0",
      fontWeight: 600,
      fontSize: ".9rem",
    },
    buttonContainer: {
      display: "flex",
      justifyContent: "flex-end",
      marginTop: "1rem",
    },
    credentialsPanel: {
      overflowY: "auto",
      maxHeight: 350,
    },
    promptTitle: {
      display: "flex",
      alignItems: "center",
    },
    buttonSpacer: {
      marginRight: ".9rem",
    },
  });

interface ICredentialsPromptProps {
  classes: any;
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
  classes,
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
      title={
        <div className={classes.promptTitle}>
          <div>New {entity} Created</div>
        </div>
      }
      titleIcon={<ServiceAccountCredentialsIcon />}
    >
      <Grid container>
        <Grid item xs={12} className={classes.formScrollable}>
          A new {entity} has been created with the following details:
          {!idp && consoleCreds && (
            <React.Fragment>
              <Grid item xs={12} className={classes.credentialsPanel}>
                <div className={classes.credentialTitle}>
                  Console Credentials
                </div>
                {Array.isArray(consoleCreds) &&
                  consoleCreds.map((credentialsPair, index) => {
                    return (
                      <>
                        <CredentialItem
                          label="Access Key"
                          value={credentialsPair.accessKey}
                        />
                        <CredentialItem
                          label="Secret Key"
                          value={credentialsPair.secretKey}
                        />
                      </>
                    );
                  })}
                {!Array.isArray(consoleCreds) && (
                  <>
                    <CredentialItem
                      label="Access Key"
                      value={consoleCreds.accessKey}
                    />
                    <CredentialItem
                      label="Secret Key"
                      value={consoleCreds.secretKey}
                    />
                  </>
                )}
              </Grid>
            </React.Fragment>
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
            <div className={classes.warningBlock}>
              Please Login via the configured external identity provider.
            </div>
          ) : (
            <div className={classes.warningBlock}>
              <WarnIcon />
              <span>
                Write these down, as this is the only time the secret will be
                displayed.
              </span>
            </div>
          )}
        </Grid>
        <Grid item xs={12} className={classes.buttonContainer}>
          {!idp && (
            <>
              <TooltipWrapper
                tooltip={
                  "Download credentials in a JSON file formatted for import using mc alias import. This will only include the default login credentials."
                }
              >
                <Button
                  id={"download-button"}
                  label={"Download for import"}
                  className={classes.buttonSpacer}
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
                    className={classes.buttonSpacer}
                    onClick={downloaddAllCredentials}
                    icon={<DownloadIcon />}
                    variant="callAction"
                    color="primary"
                  />
                </TooltipWrapper>
              )}
            </>
          )}
        </Grid>
      </Grid>
    </ModalWrapper>
  );
};

export default withStyles(styles)(CredentialsPrompt);
