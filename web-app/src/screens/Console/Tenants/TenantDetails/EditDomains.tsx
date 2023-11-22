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

import React, { useEffect, useState } from "react";
import {
  AddIcon,
  Box,
  Button,
  FormLayout,
  IconButton,
  InputBox,
  RemoveIcon,
} from "mds";
import { modalStyleUtils } from "../../Common/FormComponents/common/styleLibrary";
import {
  ErrorResponseHandler,
  IDomainsRequest,
} from "../../../../common/types";
import {
  setModalErrorSnackMessage,
  setSnackBarMessage,
} from "../../../../systemSlice";
import { useAppDispatch } from "../../../../store";
import ModalWrapper from "../../Common/ModalWrapper/ModalWrapper";
import api from "../../../../common/api";

interface IEditDomains {
  open: boolean;
  closeModalAndRefresh: (update: boolean) => any;
  namespace: string;
  idTenant: string;
  domains: IDomainsRequest | null;
}

const EditDomains = ({
  open,
  closeModalAndRefresh,
  namespace,
  idTenant,
  domains,
}: IEditDomains) => {
  const dispatch = useAppDispatch();
  const [isSending, setIsSending] = useState<boolean>(false);
  const [consoleDomain, setConsoleDomain] = useState<string>("");
  const [minioDomains, setMinioDomains] = useState<string[]>([""]);
  const [consoleDomainValid, setConsoleDomainValid] = useState<boolean>(true);
  const [minioDomainValid, setMinioDomainValid] = useState<boolean[]>([true]);

  useEffect(() => {
    if (domains) {
      const consoleDomainSet = domains.console || "";
      setConsoleDomain(consoleDomainSet);

      if (consoleDomainSet !== "") {
        // We Validate console domain
        const consoleRegExp = new RegExp(
          /^(https?):\/\/([a-zA-Z0-9\-.]+)(:[0-9]+)?(\/[a-zA-Z0-9\-./]*)?$/,
        );

        setConsoleDomainValid(consoleRegExp.test(consoleDomainSet));
      } else {
        setConsoleDomainValid(true);
      }

      if (domains.minio && domains.minio.length > 0) {
        setMinioDomains(domains.minio);

        const minioRegExp = new RegExp(
          /^(https?):\/\/([a-zA-Z0-9\-.]+)(:[0-9]+)?$/,
        );

        const initialValidations = domains.minio.map((domain) => {
          if (domain.trim() !== "") {
            return minioRegExp.test(domain);
          } else {
            return true;
          }
        });

        setMinioDomainValid(initialValidations);
      }
    }
  }, [domains]);

  const closeAction = () => {
    closeModalAndRefresh(false);
  };

  const resetForm = () => {
    setConsoleDomain("");
    setConsoleDomainValid(true);
    setMinioDomains([""]);
    setMinioDomainValid([true]);
  };

  const updateDomainsList = () => {
    setIsSending(true);

    let payload = {
      domains: {
        console: consoleDomain,
        minio: minioDomains.filter((minioDomain) => minioDomain.trim() !== ""),
      },
    };
    api
      .invoke(
        "PUT",
        `/api/v1/namespaces/${namespace}/tenants/${idTenant}/domains`,
        payload,
      )
      .then(() => {
        setIsSending(false);
        dispatch(setSnackBarMessage(`Domains updated successfully`));
        closeModalAndRefresh(true);
      })
      .catch((error: ErrorResponseHandler) => {
        setIsSending(false);
        dispatch(setModalErrorSnackMessage(error));
      });
  };

  const updateMinIODomain = (value: string, index: number) => {
    const cloneDomains = [...minioDomains];
    cloneDomains[index] = value;

    setMinioDomains(cloneDomains);
  };

  const addNewMinIODomain = () => {
    const cloneDomains = [...minioDomains];
    const cloneValidations = [...minioDomainValid];

    cloneDomains.push("");
    cloneValidations.push(true);

    setMinioDomains(cloneDomains);
    setMinioDomainValid(cloneValidations);
  };

  const removeMinIODomain = (removeIndex: number) => {
    const filteredDomains = minioDomains.filter(
      (_, index) => index !== removeIndex,
    );

    const filterValidations = minioDomainValid.filter(
      (_, index) => index !== removeIndex,
    );

    setMinioDomains(filteredDomains);
    setMinioDomainValid(filterValidations);
  };

  const setMinioDomainValidation = (domainValid: boolean, index: number) => {
    const cloneValidation = [...minioDomainValid];
    cloneValidation[index] = domainValid;

    setMinioDomainValid(cloneValidation);
  };
  return (
    <ModalWrapper
      title={`Edit Tenant Domains - ${idTenant}`}
      modalOpen={open}
      onClose={closeAction}
    >
      <Box sx={modalStyleUtils.modalFormScrollable}>
        <FormLayout withBorders={false} containerPadding={false}>
          <InputBox
            id="console_domain"
            name="console_domain"
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
              setConsoleDomain(e.target.value);

              setConsoleDomainValid(e.target.validity.valid);
            }}
            label="Console Domain"
            value={consoleDomain}
            placeholder={"Eg. http://subdomain.domain:port/subpath1/subpath2"}
            pattern={
              "^(https?):\\/\\/([a-zA-Z0-9\\-.]+)(:[0-9]+)?(\\/[a-zA-Z0-9\\-.\\/]*)?$"
            }
            error={
              !consoleDomainValid
                ? "Domain format is incorrect (http|https://subdomain.domain:port/subpath1/subpath2)"
                : ""
            }
          />
          <h4>MinIO Domains</h4>
          <div>
            {minioDomains.map((domain, index) => {
              return (
                <Box
                  key={`minio-domain-key-${index.toString()}`}
                  sx={{
                    display: "flex",
                    gap: 10,
                  }}
                >
                  <InputBox
                    id={`minio-domain-${index.toString()}`}
                    name={`minio-domain-${index.toString()}`}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      updateMinIODomain(e.target.value, index);
                      setMinioDomainValidation(e.target.validity.valid, index);
                    }}
                    label={`MinIO Domain ${index + 1}`}
                    value={domain}
                    placeholder={"Eg. http://subdomain.domain"}
                    pattern={"^(https?):\\/\\/([a-zA-Z0-9\\-.]+)(:[0-9]+)?$"}
                    error={
                      !minioDomainValid[index]
                        ? "MinIO domain format is incorrect (http|https://subdomain.domain)"
                        : ""
                    }
                  />
                  <Box
                    sx={{
                      display: "flex",
                      alignItems: "center",
                      gap: 10,
                      marginBottom: 12,
                    }}
                  >
                    <IconButton
                      size={"small"}
                      onClick={addNewMinIODomain}
                      disabled={index !== minioDomains.length - 1}
                    >
                      <AddIcon />
                    </IconButton>
                    <IconButton
                      size={"small"}
                      onClick={() => removeMinIODomain(index)}
                      disabled={minioDomains.length <= 1}
                    >
                      <RemoveIcon />
                    </IconButton>
                  </Box>
                </Box>
              );
            })}
          </div>
        </FormLayout>
        <Box sx={modalStyleUtils.modalButtonBar}>
          <Button
            id={"clear-edit-domain"}
            type="button"
            variant="regular"
            onClick={resetForm}
            label={"Clear"}
          />
          <Button
            id={"save-domain"}
            type="submit"
            variant="callAction"
            disabled={
              isSending ||
              !consoleDomainValid ||
              minioDomainValid.filter((domain) => !domain).length > 0
            }
            onClick={updateDomainsList}
            label={"Save"}
          />
        </Box>
      </Box>
    </ModalWrapper>
  );
};

export default EditDomains;
