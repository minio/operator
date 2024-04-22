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

import React, { Fragment, useCallback, useEffect, useState } from "react";
import {
  AddIcon,
  Box,
  Button,
  ConfirmModalIcon,
  DeleteIcon,
  FormLayout,
  Grid,
  IconButton,
  InformativeMessage,
  InputBox,
  LDAPIcon,
  Loader,
  OIDCIcon,
  RadioGroup,
  SectionTitle,
  Switch,
  Tooltip,
  UsersIcon,
} from "mds";
import { useSelector } from "react-redux";
import { modalStyleUtils } from "../../Common/FormComponents/common/styleLibrary";
import {
  ITenantIdentityProviderResponse,
  ITenantSetAdministratorsRequest,
} from "../types";
import { clearValidationError } from "../utils";
import {
  commonFormValidation,
  IValidation,
} from "../../../../utils/validationFunctions";
import {
  setErrorSnackMessage,
  setSnackBarMessage,
} from "../../../../systemSlice";
import { AppState, useAppDispatch } from "../../../../store";
import { ErrorResponseHandler } from "../../../../common/types";
import ConfirmDialog from "../../Common/ModalWrapper/ConfirmDialog";
import api from "../../../../common/api";

const TenantIdentityProvider = () => {
  const dispatch = useAppDispatch();

  const tenant = useSelector((state: AppState) => state.tenants.tenantInfo);
  const loadingTenant = useSelector(
    (state: AppState) => state.tenants.loadingTenant,
  );

  const [isSending, setIsSending] = useState<boolean>(false);
  const [dialogOpen, setDialogOpen] = useState<boolean>(false);
  const [idpSelection, setIdpSelection] = useState<string>("Built-in");
  const [openIDConfigurationURL, setOpenIDConfigurationURL] =
    useState<string>("");
  const [openIDClientID, setOpenIDClientID] = useState<string>("");
  const [openIDSecretID, setOpenIDSecretID] = useState<string>("");
  const [openIDCallbackURL, setOpenIDCallbackURL] = useState<string>("");
  const [openIDClaimName, setOpenIDClaimName] = useState<string>("");
  const [openIDScopes, setOpenIDScopes] = useState<string>("");
  const [ADURL, setADURL] = useState<string>("");
  const [ADLookupBindDN, setADLookupBindDN] = useState<string>("");
  const [ADLookupBindPassword, setADLookupBindPassword] = useState<string>("");
  const [ADUserDNSearchBaseDN, setADUserDNSearchBaseDN] = useState<string>("");
  const [ADUserDNSearchFilter, setADUserDNSearchFilter] = useState<string>("");
  const [ADGroupSearchBaseDN, setADGroupSearchBaseDN] = useState<string>("");
  const [ADGroupSearchFilter, setADGroupSearchFilter] = useState<string>("");
  const [ADSkipTLS, setADSkipTLS] = useState<boolean>(false);
  const [ADServerInsecure, setADServerInsecure] = useState<boolean>(false);
  const [ADServerStartTLS, setADServerStartTLS] = useState<boolean>(false);
  const [ADUserDNs, setADUserDNs] = useState<string[]>([""]);
  const [ADGroupDNs, setADGroupDNs] = useState<string[]>([""]);
  const [validationErrors, setValidationErrors] = useState<any>({});
  const cleanValidation = (fieldName: string) => {
    setValidationErrors(clearValidationError(validationErrors, fieldName));
  };
  const [isFormValid, setIsFormValid] = useState<boolean>(false);

  // Validation
  useEffect(() => {
    let identityProviderValidation: IValidation[] = [];

    if (idpSelection === "OpenID") {
      identityProviderValidation = [
        ...identityProviderValidation,
        {
          fieldKey: "openID_CONFIGURATION_URL",
          required: true,
          value: openIDConfigurationURL,
        },
        {
          fieldKey: "openID_clientID",
          required: true,
          value: openIDClientID,
        },
        {
          fieldKey: "openID_secretID",
          required: true,
          value: openIDSecretID,
        },
        {
          fieldKey: "openID_claimName",
          required: false,
          value: openIDClaimName,
        },
      ];
    }

    if (idpSelection === "AD") {
      identityProviderValidation = [
        ...identityProviderValidation,
        {
          fieldKey: "AD_URL",
          required: true,
          value: ADURL,
        },
        {
          fieldKey: "ad_lookupBindDN",
          required: true,
          value: ADLookupBindDN,
        },
      ];
    }

    const commonVal = commonFormValidation(identityProviderValidation);

    setIsFormValid(Object.keys(commonVal).length === 0);

    setValidationErrors(commonVal);
  }, [
    idpSelection,
    openIDConfigurationURL,
    openIDClientID,
    openIDSecretID,
    openIDClaimName,
    ADURL,
    ADLookupBindDN,
  ]);

  const getTenantIdentityProviderInfo = useCallback(() => {
    api
      .invoke(
        "GET",
        `/api/v1/namespaces/${tenant?.namespace}/tenants/${tenant?.name}/identity-provider`,
      )
      .then((res: ITenantIdentityProviderResponse) => {
        if (res) {
          if (res.oidc) {
            setIdpSelection("OpenID");
            setOpenIDConfigurationURL(res.oidc.configuration_url);
            setOpenIDClientID(res.oidc.client_id);
            setOpenIDSecretID(res.oidc.secret_id);
            setOpenIDCallbackURL(res.oidc.callback_url);
            setOpenIDClaimName(res.oidc.claim_name);
            setOpenIDScopes(res.oidc.scopes);
          } else if (res.active_directory) {
            setIdpSelection("AD");
            setADURL(res.active_directory.url);
            setADLookupBindDN(res.active_directory.lookup_bind_dn);
            setADLookupBindPassword(res.active_directory.lookup_bind_password);
            setADUserDNSearchBaseDN(
              res.active_directory.user_dn_search_base_dn,
            );
            setADUserDNSearchFilter(res.active_directory.user_dn_search_filter);
            setADGroupSearchBaseDN(res.active_directory.group_search_base_dn);
            setADGroupSearchFilter(res.active_directory.group_search_filter);
            setADSkipTLS(res.active_directory.skip_tls_verification);
            setADServerInsecure(res.active_directory.server_insecure);
            setADServerStartTLS(res.active_directory.server_start_tls);
          }
        }
      })
      .catch((err: ErrorResponseHandler) => {
        dispatch(setErrorSnackMessage(err));
      });
  }, [tenant, dispatch]);

  useEffect(() => {
    if (tenant) {
      getTenantIdentityProviderInfo();
    }
  }, [tenant, getTenantIdentityProviderInfo]);

  const updateTenantIdentityProvider = () => {
    setIsSending(true);
    let payload: ITenantIdentityProviderResponse = {};
    switch (idpSelection) {
      case "AD":
        payload.active_directory = {
          url: ADURL,
          lookup_bind_dn: ADLookupBindDN,
          lookup_bind_password: ADLookupBindPassword,
          user_dn_search_base_dn: ADUserDNSearchBaseDN,
          user_dn_search_filter: ADUserDNSearchFilter,
          group_search_base_dn: ADGroupSearchBaseDN,
          group_search_filter: ADGroupSearchFilter,
          skip_tls_verification: ADSkipTLS,
          server_insecure: ADServerInsecure,
          server_start_tls: ADServerStartTLS,
        };
        break;
      case "OpenID":
        payload.oidc = {
          configuration_url: openIDConfigurationURL,
          client_id: openIDClientID,
          secret_id: openIDSecretID,
          callback_url: openIDCallbackURL,
          claim_name: openIDClaimName,
          scopes: openIDScopes,
        };
        break;
      default:
      // Built-in IDP will be used by default
    }

    api
      .invoke(
        "POST",
        `/api/v1/namespaces/${tenant?.namespace}/tenants/${tenant?.name}/identity-provider`,
        payload,
      )
      .then(() => {
        setIsSending(false);
        // Close confirmation modal
        setDialogOpen(false);
        getTenantIdentityProviderInfo();
      })
      .catch((err: ErrorResponseHandler) => {
        dispatch(setErrorSnackMessage(err));
        setIsSending(false);
      });
  };

  const setAdministrators = () => {
    setIsSending(true);
    let payload: ITenantSetAdministratorsRequest = {};
    switch (idpSelection) {
      case "AD":
        payload = {
          user_dns: ADUserDNs.filter((user) => user.trim() !== ""),
          group_dns: ADGroupDNs.filter((group) => group.trim() !== ""),
        };
        break;
      default:
      // Built-in IDP will be used by default
    }

    api
      .invoke(
        "POST",
        `/api/v1/namespaces/${tenant?.namespace}/tenants/${tenant?.name}/set-administrators`,
        payload,
      )
      .then(() => {
        setIsSending(false);
        setADGroupDNs([""]);
        setADUserDNs([""]);
        getTenantIdentityProviderInfo();
        dispatch(setSnackBarMessage(`Administrators added successfully`));
      })
      .catch((err: ErrorResponseHandler) => {
        dispatch(setErrorSnackMessage(err));
        setIsSending(false);
      });
  };

  return (
    <Fragment>
      <ConfirmDialog
        title={"Save and Restart"}
        confirmText={"Restart"}
        cancelText="Cancel"
        titleIcon={<ConfirmModalIcon />}
        isLoading={isSending}
        onClose={() => setDialogOpen(false)}
        isOpen={dialogOpen}
        onConfirm={updateTenantIdentityProvider}
        confirmationContent={
          <Fragment>
            Are you sure you want to save the changes and restart the service?
          </Fragment>
        }
      />
      {loadingTenant ? (
        <Box
          sx={{
            textAlign: "center",
          }}
        >
          <Loader />
        </Box>
      ) : (
        <Fragment>
          <Box>
            <SectionTitle separator sx={{ marginBottom: 15 }}>
              Identity Provider
            </SectionTitle>
          </Box>
          <FormLayout>
            <RadioGroup
              currentValue={idpSelection}
              id="idp-options"
              name="idp-options"
              label="Protocol"
              onChange={(e) => {
                setIdpSelection(e.target.value);
              }}
              selectorOptions={[
                { label: "Built-in", value: "Built-in", icon: <UsersIcon /> },
                { label: "Open ID", value: "OpenID", icon: <OIDCIcon /> },
                {
                  label: "LDAP / Active Directory",
                  value: "AD",
                  icon: <LDAPIcon />,
                },
              ]}
            />

            {idpSelection === "OpenID" && (
              <Fragment>
                <InputBox
                  id="openID_CONFIGURATION_URL"
                  name="openID_CONFIGURATION_URL"
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    setOpenIDConfigurationURL(e.target.value);
                    cleanValidation("openID_CONFIGURATION_URL");
                  }}
                  label="Configuration URL"
                  value={openIDConfigurationURL}
                  placeholder="https://your-identity-provider.com/.well-known/openid-configuration"
                  error={validationErrors["openID_CONFIGURATION_URL"] || ""}
                  required
                />
                <InputBox
                  id="openID_clientID"
                  name="openID_clientID"
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    setOpenIDClientID(e.target.value);
                    cleanValidation("openID_clientID");
                  }}
                  label="Client ID"
                  value={openIDClientID}
                  error={validationErrors["openID_clientID"] || ""}
                  required
                />
                <InputBox
                  type={"password"}
                  id="openID_secretID"
                  name="openID_secretID"
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    setOpenIDSecretID(e.target.value);
                    cleanValidation("openID_secretID");
                  }}
                  label="Secret ID"
                  value={openIDSecretID}
                  error={validationErrors["openID_secretID"] || ""}
                  required
                />
                <InputBox
                  id="openID_claimName"
                  name="openID_claimName"
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    setOpenIDClaimName(e.target.value);
                    cleanValidation("openID_claimName");
                  }}
                  label="Claim Name"
                  value={openIDClaimName}
                  placeholder="policy"
                  error={validationErrors["openID_claimName"] || ""}
                />
                <InputBox
                  id="openID_scopes"
                  name="openID_scopes"
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    setOpenIDScopes(e.target.value);
                    cleanValidation("openID_scopes");
                  }}
                  label="Scopes"
                  value={openIDScopes}
                />
              </Fragment>
            )}

            {idpSelection === "AD" && (
              <Fragment>
                <InputBox
                  id="AD_URL"
                  name="AD_URL"
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    setADURL(e.target.value);
                    cleanValidation("AD_URL");
                  }}
                  label="LDAP Server Address"
                  value={ADURL}
                  placeholder="ldap-server:636"
                  error={validationErrors["AD_URL"] || ""}
                  required
                />
                <Switch
                  value="ad_skipTLS"
                  id="ad_skipTLS"
                  name="ad_skipTLS"
                  checked={ADSkipTLS}
                  onChange={(e) => {
                    const targetD = e.target;
                    const checked = targetD.checked;
                    setADSkipTLS(checked);
                  }}
                  label={"Skip TLS Verification"}
                />
                <Switch
                  value="ad_serverInsecure"
                  id="ad_serverInsecure"
                  name="ad_serverInsecure"
                  checked={ADServerInsecure}
                  onChange={(e) => {
                    const targetD = e.target;
                    const checked = targetD.checked;
                    setADServerInsecure(checked);
                  }}
                  label={"Server Insecure"}
                />
                {ADServerInsecure ? (
                  <InformativeMessage
                    title={"Warning"}
                    message={
                      "All traffic with Active Directory will be\n" +
                      "                    unencrypted"
                    }
                    variant={"warning"}
                    sx={{ marginBottom: 15 }}
                  />
                ) : null}
                <Switch
                  value="ad_serverStartTLS"
                  id="ad_serverStartTLS"
                  name="ad_serverStartTLS"
                  checked={ADServerStartTLS}
                  onChange={(e) => {
                    const targetD = e.target;
                    const checked = targetD.checked;
                    setADServerStartTLS(checked);
                  }}
                  label={"Start TLS connection to AD/LDAP server"}
                />
                <InputBox
                  id="ad_lookupBindDN"
                  name="ad_lookupBindDN"
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    setADLookupBindDN(e.target.value);
                    cleanValidation("ad_lookupBindDN");
                  }}
                  label="Lookup Bind DN"
                  value={ADLookupBindDN}
                  placeholder="cn=admin,dc=min,dc=io"
                  error={validationErrors["ad_lookupBindDN"] || ""}
                  required
                />
                <InputBox
                  type={"password"}
                  id="ad_lookupBindPassword"
                  name="ad_lookupBindPassword"
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    setADLookupBindPassword(e.target.value);
                  }}
                  label="Lookup Bind Password"
                  value={ADLookupBindPassword}
                  placeholder="admin"
                />
                <InputBox
                  id="ad_userDNSearchBaseDN"
                  name="ad_userDNSearchBaseDN"
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    setADUserDNSearchBaseDN(e.target.value);
                  }}
                  label="User DN Search Base DN"
                  value={ADUserDNSearchBaseDN}
                  placeholder="dc=min,dc=io"
                />
                <InputBox
                  id="ad_userDNSearchFilter"
                  name="ad_userDNSearchFilter"
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    setADUserDNSearchFilter(e.target.value);
                  }}
                  label="User DN Search Filter"
                  value={ADUserDNSearchFilter}
                  placeholder="(sAMAcountName=%s)"
                />
                <InputBox
                  id="ad_groupSearchBaseDN"
                  name="ad_groupSearchBaseDN"
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    setADGroupSearchBaseDN(e.target.value);
                  }}
                  label="Group Search Base DN"
                  value={ADGroupSearchBaseDN}
                  placeholder="ou=hwengg,dc=min,dc=io;ou=swengg,dc=min,dc=io"
                />
                <InputBox
                  id="ad_groupSearchFilter"
                  name="ad_groupSearchFilter"
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    setADGroupSearchFilter(e.target.value);
                  }}
                  label="Group Search Filter"
                  value={ADGroupSearchFilter}
                  placeholder="(&(objectclass=groupOfNames)(member=%s))"
                />
              </Fragment>
            )}

            <Box sx={modalStyleUtils.modalButtonBar}>
              <Button
                id={"save-idp"}
                type="submit"
                variant="callAction"
                color="primary"
                disabled={!isFormValid || isSending}
                onClick={() => setDialogOpen(true)}
                label={"Save"}
              />
            </Box>

            {idpSelection === "AD" && (
              <Fragment>
                <SectionTitle separator>User & Group management</SectionTitle>
                <br />
                <fieldset className={"inputItem"}>
                  <legend>
                    List of user DNs (Distinguished Names) to be added as Tenant
                    Administrators
                  </legend>
                  <Grid item xs={12}>
                    {ADUserDNs.map((_, index) => {
                      return (
                        <Fragment key={`identityField-${index.toString()}`}>
                          <Box
                            sx={{
                              display: "flex",
                              marginBottom: 10,
                            }}
                          >
                            <InputBox
                              id={`ad-userdn-${index.toString()}`}
                              label={""}
                              placeholder=""
                              name={`ad-userdn-${index.toString()}`}
                              value={ADUserDNs[index]}
                              onChange={(
                                e: React.ChangeEvent<HTMLInputElement>,
                              ) => {
                                setADUserDNs(
                                  ADUserDNs.map((group, i) =>
                                    i === index ? e.target.value : group,
                                  ),
                                );
                              }}
                              index={index}
                              key={`csv-ad-userdn-${index.toString()}`}
                              error={
                                validationErrors[
                                  `ad-userdn-${index.toString()}`
                                ] || ""
                              }
                            />
                            <Box
                              sx={{
                                marginLeft: 10,
                                display: "flex",
                                height: 38,
                              }}
                            >
                              <Tooltip tooltip="Add User" aria-label="add">
                                <IconButton
                                  size={"small"}
                                  onClick={() => {
                                    setADUserDNs([...ADUserDNs, ""]);
                                  }}
                                >
                                  <AddIcon />
                                </IconButton>
                              </Tooltip>
                              <Tooltip tooltip="Remove" aria-label="add">
                                <IconButton
                                  size={"small"}
                                  style={{ marginLeft: 16 }}
                                  onClick={() => {
                                    if (ADUserDNs.length > 1) {
                                      setADUserDNs(
                                        ADUserDNs.filter((_, i) => i !== index),
                                      );
                                    }
                                  }}
                                >
                                  <DeleteIcon />
                                </IconButton>
                              </Tooltip>
                            </Box>
                          </Box>
                        </Fragment>
                      );
                    })}
                  </Grid>
                </fieldset>
                <fieldset className={"inputItem"}>
                  <legend>
                    List of group DNs (Distinguished Names) to be added as
                    Tenant Administrators
                  </legend>
                  <Grid item xs={12}>
                    {ADGroupDNs.map((_, index) => {
                      return (
                        <Fragment key={`identityField-${index.toString()}`}>
                          <Box
                            sx={{
                              display: "flex",
                              marginBottom: 10,
                            }}
                          >
                            <InputBox
                              id={`ad-groupdn-${index.toString()}`}
                              label={""}
                              placeholder=""
                              name={`ad-groupdn-${index.toString()}`}
                              value={ADGroupDNs[index]}
                              onChange={(
                                e: React.ChangeEvent<HTMLInputElement>,
                              ) => {
                                setADGroupDNs(
                                  ADGroupDNs.map((group, i) =>
                                    i === index ? e.target.value : group,
                                  ),
                                );
                              }}
                              index={index}
                              key={`csv-ad-groupdn-${index.toString()}`}
                              error={
                                validationErrors[
                                  `ad-groupdn-${index.toString()}`
                                ] || ""
                              }
                            />
                            <Box
                              sx={{
                                marginLeft: 10,
                                display: "flex",
                                height: 38,
                              }}
                            >
                              <Tooltip tooltip="Add Group" aria-label="add">
                                <IconButton
                                  size={"small"}
                                  onClick={() => {
                                    setADGroupDNs([...ADGroupDNs, ""]);
                                  }}
                                >
                                  <AddIcon />
                                </IconButton>
                              </Tooltip>
                              <Tooltip tooltip="Remove" aria-label="add">
                                <IconButton
                                  size={"small"}
                                  style={{ marginLeft: 16 }}
                                  onClick={() => {
                                    if (ADGroupDNs.length > 1) {
                                      setADGroupDNs(
                                        ADGroupDNs.filter(
                                          (_, i) => i !== index,
                                        ),
                                      );
                                    }
                                  }}
                                >
                                  <DeleteIcon />
                                </IconButton>
                              </Tooltip>
                            </Box>
                          </Box>
                        </Fragment>
                      );
                    })}
                  </Grid>
                </fieldset>
                <br />
                <Box sx={modalStyleUtils.modalButtonBar}>
                  <Button
                    id={"add-additional-dns"}
                    type="submit"
                    variant="callAction"
                    disabled={!isFormValid || isSending}
                    onClick={() => setAdministrators()}
                    label={"Add additional DNs"}
                  />
                </Box>
              </Fragment>
            )}
          </FormLayout>
        </Fragment>
      )}
    </Fragment>
  );
};

export default TenantIdentityProvider;
