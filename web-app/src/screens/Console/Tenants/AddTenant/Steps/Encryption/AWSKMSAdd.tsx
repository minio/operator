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

import React, { Fragment, useCallback, useEffect, useState } from "react";
import Grid from "@mui/material/Grid";
import InputBoxWrapper from "../../../../Common/FormComponents/InputBoxWrapper/InputBoxWrapper";
import { useSelector } from "react-redux";
import { AppState, useAppDispatch } from "../../../../../../store";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import {
  createTenantCommon,
  formFieldStyles,
  modalBasic,
  wizardCommon,
} from "../../../../Common/FormComponents/common/styleLibrary";
import makeStyles from "@mui/styles/makeStyles";
import {
  commonFormValidation,
  IValidation,
} from "../../../../../../utils/validationFunctions";
import { isPageValid, updateAddField } from "../../createTenantSlice";
import { clearValidationError } from "../../../utils";

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    ...createTenantCommon,
    ...formFieldStyles,
    ...modalBasic,
    ...wizardCommon,
  })
);

const AWSKMSAdd = () => {
  const dispatch = useAppDispatch();
  const classes = useStyles();

  const encryptionTab = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.encryptionTab
  );
  const awsEndpoint = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.awsEndpoint
  );
  const awsRegion = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.awsRegion
  );
  const awsKMSKey = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.awsKMSKey
  );
  const awsAccessKey = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.awsAccessKey
  );
  const awsSecretKey = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.awsSecretKey
  );
  const awsToken = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.awsToken
  );
  const [validationErrors, setValidationErrors] = useState<any>({});

  // Validation
  useEffect(() => {
    let encryptionValidation: IValidation[] = [];

    if (!encryptionTab) {
      encryptionValidation = [
        ...encryptionValidation,
        {
          fieldKey: "aws_endpoint",
          required: true,
          value: awsEndpoint,
        },
        {
          fieldKey: "aws_region",
          required: true,
          value: awsRegion,
        },
        {
          fieldKey: "aws_accessKey",
          required: true,
          value: awsAccessKey,
        },
        {
          fieldKey: "aws_secretKey",
          required: true,
          value: awsSecretKey,
        },
      ];
    }

    const commonVal = commonFormValidation(encryptionValidation);

    dispatch(
      isPageValid({
        pageName: "encryption",
        valid: Object.keys(commonVal).length === 0,
      })
    );

    setValidationErrors(commonVal);
  }, [
    encryptionTab,
    awsEndpoint,
    awsRegion,
    awsSecretKey,
    awsAccessKey,
    dispatch,
  ]);

  // Common
  const updateField = useCallback(
    (field: string, value: any) => {
      dispatch(
        updateAddField({ pageName: "encryption", field: field, value: value })
      );
    },
    [dispatch]
  );

  const cleanValidation = (fieldName: string) => {
    setValidationErrors(clearValidationError(validationErrors, fieldName));
  };

  return (
    <Fragment>
      <Grid item xs={12} className={classes.formFieldRow}>
        <InputBoxWrapper
          id="aws_endpoint"
          name="aws_endpoint"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("awsEndpoint", e.target.value);
            cleanValidation("aws_endpoint");
          }}
          label="Endpoint"
          tooltip="Endpoint is the AWS SecretsManager endpoint. AWS SecretsManager endpoints have the following schema: secrestmanager[-fips].<region>.amanzonaws.com"
          value={awsEndpoint}
          error={validationErrors["aws_endpoint"] || ""}
          required
        />
      </Grid>
      <Grid item xs={12} className={classes.formFieldRow}>
        <InputBoxWrapper
          id="aws_region"
          name="aws_region"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("awsRegion", e.target.value);
            cleanValidation("aws_region");
          }}
          label="Region"
          tooltip="Region is the AWS region the SecretsManager is located"
          value={awsRegion}
          error={validationErrors["aws_region"] || ""}
          required
        />
      </Grid>
      <Grid item xs={12} className={classes.formFieldRow}>
        <InputBoxWrapper
          id="aws_kmsKey"
          name="aws_kmsKey"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("awsKMSKey", e.target.value);
          }}
          label="KMS Key"
          tooltip="KMSKey is the AWS-KMS key ID (CMK-ID) used to en/decrypt secrets managed by the SecretsManager. If empty, the default AWS KMS key is used"
          value={awsKMSKey}
        />
      </Grid>
      <Grid item xs={12}>
        <fieldset className={classes.fieldGroup}>
          <legend className={classes.descriptionText}>Credentials</legend>
          <Grid item xs={12} className={classes.formFieldRow}>
            <InputBoxWrapper
              id="aws_accessKey"
              name="aws_accessKey"
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                updateField("awsAccessKey", e.target.value);
                cleanValidation("aws_accessKey");
              }}
              label="Access Key"
              tooltip="AccessKey is the access key for authenticating to AWS"
              value={awsAccessKey}
              error={validationErrors["aws_accessKey"] || ""}
              required
            />
          </Grid>
          <Grid item xs={12} className={classes.formFieldRow}>
            <InputBoxWrapper
              id="aws_secretKey"
              name="aws_secretKey"
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                updateField("awsSecretKey", e.target.value);
                cleanValidation("aws_secretKey");
              }}
              label="Secret Key"
              tooltip="SecretKey is the secret key for authenticating to AWS"
              value={awsSecretKey}
              error={validationErrors["aws_secretKey"] || ""}
              required
            />
          </Grid>
          <Grid item xs={12} className={classes.formFieldRow}>
            <InputBoxWrapper
              id="aws_token"
              name="aws_token"
              tooltip="SessionToken is an optional session token for authenticating to AWS when using STS"
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                updateField("awsToken", e.target.value);
              }}
              label="Token"
              value={awsToken}
            />
          </Grid>
        </fieldset>
      </Grid>
    </Fragment>
  );
};

export default AWSKMSAdd;
