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

import React, { Fragment, useEffect, useState } from "react";
import { Box, DialogContentText } from "@mui/material";
import { Button, HelpBox, InspectMenuIcon, PasswordKeyIcon } from "mds";
import { useNavigate } from "react-router-dom";
import PageLayout from "../Common/Layout/PageLayout";
import InputBoxWrapper from "../Common/FormComponents/InputBoxWrapper/InputBoxWrapper";
import FormSwitchWrapper from "../Common/FormComponents/FormSwitchWrapper/FormSwitchWrapper";
import ModalWrapper from "../Common/ModalWrapper/ModalWrapper";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import {
  deleteDialogStyles,
  modalStyleUtils,
} from "../Common/FormComponents/common/styleLibrary";
import withStyles from "@mui/styles/withStyles";
import { useSelector } from "react-redux";
import {
  deleteCookie,
  encodeURLString,
  getCookieValue,
  performDownload,
} from "../../../common/utils";
import DistributedOnly from "../Common/DistributedOnly/DistributedOnly";
import KeyRevealer from "./KeyRevealer";
import { selDistSet, setErrorSnackMessage } from "../../../systemSlice";
import { useAppDispatch } from "../../../store";
import RegisterCluster from "../Support/RegisterCluster";
import { registeredCluster } from "../../../config";
import PageHeaderWrapper from "../Common/PageHeaderWrapper/PageHeaderWrapper";

const styles = (theme: Theme) =>
  createStyles({
    switchLabel: {
      fontWeight: "normal",
    },
    ...deleteDialogStyles,
    ...modalStyleUtils,
  });

const ExampleBlock = ({
  volumeVal,
  pathVal,
}: {
  volumeVal: string;
  pathVal: string;
}) => {
  return (
    <Box className="code-block-container">
      <Box className="example-code-block">
        <Box
          sx={{
            display: "flex",
            marginBottom: "5px",
            flexFlow: {
              sm: "row",
              xs: "column",
            },
          }}
        >
          <label>Volume/bucket Name :</label> <code>{volumeVal}</code>
        </Box>
        <Box
          sx={{
            display: "flex",
            flexFlow: {
              sm: "row",
              xs: "column",
            },
          }}
        >
          <label>Path : </label>
          <code>{pathVal}</code>
        </Box>
      </Box>
    </Box>
  );
};

const Inspect = ({ classes }: { classes: any }) => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const distributedSetup = useSelector(selDistSet);

  const [volumeName, setVolumeName] = useState<string>("");
  const [inspectPath, setInspectPath] = useState<string>("");
  const [isEncrypt, setIsEncrypt] = useState<boolean>(true);

  const [decryptionKey, setDecryptionKey] = useState<string>("");

  const [insFileName, setInsFileName] = useState<string>("");

  const [isFormValid, setIsFormValid] = useState<boolean>(false);
  const [volumeError, setVolumeError] = useState<string>("");
  const [pathError, setPathError] = useState<string>("");
  const clusterRegistered = registeredCluster();
  /**
   * Validation Effect
   */
  useEffect(() => {
    let isVolValid;
    let isPathValid;

    isVolValid = volumeName.trim().length > 0;
    if (!isVolValid) {
      setVolumeError("This field is required");
    } else if (volumeName.slice(0, 1) === "/") {
      isVolValid = false;
      setVolumeError("Volume/Bucket name cannot start with /");
    }
    isPathValid = inspectPath.trim().length > 0;
    if (!inspectPath) {
      setPathError("This field is required");
    } else if (inspectPath.slice(0, 1) === "/") {
      isPathValid = false;
      setPathError("Path cannot start with /");
    }
    const isValid = isVolValid && isPathValid;

    if (isVolValid) {
      setVolumeError("");
    }
    if (isPathValid) {
      setPathError("");
    }

    setIsFormValid(isValid);
  }, [volumeName, inspectPath]);

  const makeRequest = async (url: string) => {
    return await fetch(url, { method: "GET" });
  };

  const performInspect = async () => {
    const file = encodeURLString(inspectPath);
    const volume = encodeURLString(volumeName);

    const urlOfInspectApi = `/api/v1/admin/inspect?volume=${volume}&file=${file}&encrypt=${isEncrypt}`;

    makeRequest(urlOfInspectApi)
      .then(async (res) => {
        if (!res.ok) {
          const resErr: any = await res.json();

          dispatch(
            setErrorSnackMessage({
              errorMessage: resErr.message,
              detailedError: resErr.code,
            })
          );
        }
        const blob: Blob = await res.blob();

        //@ts-ignore
        const filename = res.headers.get("content-disposition").split('"')[1];
        const decryptKey = getCookieValue(filename) || "";

        performDownload(blob, filename);
        setInsFileName(filename);
        setDecryptionKey(decryptKey);
      })
      .catch((err) => {
        dispatch(setErrorSnackMessage(err));
      });
  };

  const resetForm = () => {
    setVolumeName("");
    setInspectPath("");
    setIsEncrypt(true);
  };

  const onCloseDecKeyModal = () => {
    deleteCookie(insFileName);
    setDecryptionKey("");
    resetForm();
  };

  return (
    <Fragment>
      <PageHeaderWrapper label={"Inspect"} />
      <PageLayout>
        {!clusterRegistered && <RegisterCluster compactMode />}
        {!distributedSetup ? (
          <DistributedOnly
            iconComponent={<InspectMenuIcon />}
            entity={"Inspect"}
          />
        ) : (
          <Box
            sx={{
              display: "flex",
              alignItems: "flex-start",
              justifyContent: "flex-start",
              border: "1px solid #eaeaea",
              padding: {
                lg: "40px",
                xs: "15px",
              },
              flexWrap: "wrap",
              gap: {
                lg: "55px",
                xs: "20px",
              },
              height: {
                md: "calc(100vh - 120px)",
                xs: "100%",
              },
              flexFlow: {
                lg: "row",
                xs: "column",
              },
            }}
          >
            <Box
              sx={{
                border: "1px solid #eaeaea",
                flex: {
                  md: 2,
                  xs: 1,
                },
                width: {
                  lg: "auto",
                  xs: "100%",
                },
                padding: {
                  lg: "40px",
                  xs: "15px",
                },
              }}
            >
              <form
                noValidate
                autoComplete="off"
                onSubmit={(e: React.FormEvent<HTMLFormElement>) => {
                  e.preventDefault();
                  if (!clusterRegistered) {
                    navigate("/support/register");
                    return;
                  }
                  performInspect();
                }}
              >
                <Box>
                  <InputBoxWrapper
                    id="inspect_volume"
                    name="inspect_volume"
                    extraInputProps={{
                      "data-test-id": "inspect_volume",
                    }}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      setVolumeName(e.target.value);
                    }}
                    label="Volume or Bucket Name"
                    value={volumeName}
                    error={volumeError}
                    required
                    placeholder={"test-bucket"}
                  />
                </Box>
                <Box
                  sx={{
                    marginTop: "15px",
                  }}
                >
                  <InputBoxWrapper
                    id="inspect_path"
                    name="inspect_path"
                    extraInputProps={{
                      "data-test-id": "inspect_path",
                    }}
                    error={pathError}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      setInspectPath(e.target.value);
                    }}
                    label="File or Path to inspect"
                    value={inspectPath}
                    required
                    placeholder={"test*/xl.meta"}
                  />
                </Box>
                <Box
                  sx={{
                    marginTop: "25px",
                  }}
                >
                  <FormSwitchWrapper
                    classes={{
                      inputLabel: classes.switchLabel,
                    }}
                    extraInputProps={{
                      "data-test-id": "inspect_encrypt",
                    }}
                    label="Encrypt"
                    indicatorLabels={["True", "False"]}
                    checked={isEncrypt}
                    value={"true"}
                    id="inspect_encrypt"
                    name="inspect_encrypt"
                    onChange={(e) => {
                      setIsEncrypt(!isEncrypt);
                    }}
                  />
                </Box>
                <Box
                  sx={{
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "flex-end",
                    marginTop: "55px",
                  }}
                >
                  <Button
                    id={"inspect-clear-button"}
                    style={{
                      marginRight: "15px",
                    }}
                    type="button"
                    variant="regular"
                    data-test-id="inspect-clear-button"
                    onClick={resetForm}
                    label={"Clear"}
                  />
                  <Button
                    id={"inspect-start"}
                    type="submit"
                    variant={!clusterRegistered ? "regular" : "callAction"}
                    data-test-id="inspect-submit-button"
                    disabled={!isFormValid}
                    label={"Inspect"}
                  />
                </Box>
              </form>
            </Box>
            <Box
              sx={{
                flex: 1,
                minWidth: {
                  md: "365px",
                  xs: "100%",
                },
                width: "100%",
              }}
            >
              <HelpBox
                title={""}
                iconComponent={null}
                help={
                  <Fragment>
                    <Box
                      sx={{
                        marginTop: "-25px",
                        fontSize: "16px",
                        fontWeight: 600,
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "flex-start",
                        padding: "2px",
                      }}
                    >
                      <Box
                        sx={{
                          backgroundColor: "#07193E",
                          height: "15px",
                          width: "15px",
                          display: "flex",
                          alignItems: "center",
                          justifyContent: "center",
                          borderRadius: "50%",
                          marginRight: "18px",
                          padding: "3px",
                          "& .min-icon": {
                            height: "11px",
                            width: "11px",
                            fill: "#ffffff",
                          },
                        }}
                      >
                        <InspectMenuIcon />
                      </Box>
                      Learn more about the Inspect feature
                    </Box>

                    <Box
                      sx={{
                        marginTop: "16px",
                        fontWeight: 600,
                        fontStyle: "italic",
                        fontSize: "14px",
                      }}
                    >
                      Examples:
                    </Box>

                    <Box
                      sx={{
                        display: "flex",
                        flexFlow: "column",
                        fontSize: "14px",
                        flex: "2",
                        "& .step-number": {
                          color: "#ffffff",
                          height: "25px",
                          width: "25px",
                          background: "#081C42",
                          marginRight: "10px",
                          textAlign: "center",
                          fontWeight: 600,
                          borderRadius: "50%",
                        },

                        "& .step-row": {
                          fontSize: "14px",
                          display: "flex",
                          marginTop: "15px",
                          marginBottom: "15px",

                          "&.step-text": {
                            fontWeight: 400,
                          },
                          "&:before": {
                            content: "' '",
                            height: "7px",
                            width: "7px",
                            backgroundColor: "#2781B0",
                            marginRight: "10px",
                            marginTop: "7px",
                            flexShrink: 0,
                          },
                        },

                        "& .code-block-container": {
                          flex: "1",
                          marginTop: "15px",
                          marginLeft: "35px",

                          "& input": {
                            color: "#737373",
                          },
                        },

                        "& .example-code-block label": {
                          display: "inline-block",
                          width: {
                            sm: "160px",
                            xs: "100%",
                          },
                          fontWeight: 600,
                          fontSize: "14px",
                        },

                        "& code": {
                          width: {
                            sm: "100px",
                            xs: "100%",
                          },
                          paddingLeft: "10px",
                          fontFamily: "monospace",
                          paddingRight: "10px",
                          paddingTop: "3px",
                          paddingBottom: "3px",
                          borderRadius: "2px",
                          border: "1px solid #eaeaea",
                          fontSize: "10px",
                          color: "#082146",
                          fontWeight: 500,
                        },
                        "& .spacer": {
                          marginBottom: "5px",
                        },
                      }}
                    >
                      <Box>
                        <Box className="step-row">
                          <div className="step-text">
                            To Download 'xl.meta' for a specific object from all
                            the drives in a zip file:
                          </div>
                        </Box>

                        <ExampleBlock
                          pathVal={`test*/xl.meta`}
                          volumeVal={`test-bucket`}
                        />
                      </Box>

                      <Box>
                        <Box className="step-row">
                          <div className="step-text">
                            To Download all constituent parts for a specific
                            object, and optionally encrypt the downloaded zip:
                          </div>
                        </Box>

                        <ExampleBlock
                          pathVal={`test*/xl.meta`}
                          volumeVal={`test*/*/part.*`}
                        />
                      </Box>
                      <Box>
                        <Box className="step-row">
                          <div className="step-text">
                            To Download recursively all objects at a prefix.
                            <br />
                            NOTE: This can be an expensive operation use it with
                            caution.
                          </div>
                        </Box>
                        <ExampleBlock
                          pathVal={`test*/xl.meta`}
                          volumeVal={`test/**`}
                        />
                      </Box>
                    </Box>

                    <Box
                      sx={{
                        marginTop: "30px",
                        marginLeft: "15px",
                        fontSize: "14px",
                      }}
                    >
                      You can learn more at our{" "}
                      <a
                        href="https://github.com/minio/minio/tree/master/docs/debugging?ref=con"
                        target="_blank"
                        rel="noopener"
                      >
                        documentation
                      </a>
                      .
                    </Box>
                  </Fragment>
                }
              />
            </Box>
          </Box>
        )}
        {decryptionKey ? (
          <ModalWrapper
            modalOpen={true}
            title="Inspect Decryption Key"
            onClose={onCloseDecKeyModal}
            titleIcon={<PasswordKeyIcon />}
          >
            <DialogContentText component="div">
              <Box>
                This will be displayed only once. It cannot be recovered.
                <br />
                Use secure medium to share this key.
              </Box>
              <form
                noValidate
                onSubmit={() => {
                  return false;
                }}
              >
                <KeyRevealer value={decryptionKey} />
              </form>
            </DialogContentText>
          </ModalWrapper>
        ) : null}
      </PageLayout>
    </Fragment>
  );
};

export default withStyles(styles)(Inspect);
