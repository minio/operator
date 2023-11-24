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

import React, { Fragment, useEffect, useState } from "react";
import { useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";
import {
  Button,
  CodeEditor,
  Grid,
  InformativeMessage,
  ProgressBar,
  SectionTitle,
} from "mds";
import api from "../../../../common/api";
import { ErrorResponseHandler } from "../../../../common/types";
import { setModalErrorSnackMessage } from "../../../../systemSlice";
import { AppState, useAppDispatch } from "../../../../store";
import { getTenantAsync } from "../thunks/tenantDetailsAsync";

interface ITenantYAML {
  yaml: string;
}

const TenantYAML = () => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();

  const tenant = useSelector((state: AppState) => state.tenants.currentTenant);
  const namespace = useSelector(
    (state: AppState) => state.tenants.currentNamespace,
  );

  const [addLoading, setAddLoading] = useState<boolean>(false);
  const [loading, setLoading] = useState<boolean>(false);
  const [tenantYaml, setTenantYaml] = useState<string>("");
  const [errorMessage, setErrorMessage] = useState<string>("");

  const updateTenant = (event: React.FormEvent) => {
    event.preventDefault();
    if (addLoading) {
      return;
    }
    setAddLoading(true);
    setErrorMessage("");
    api
      .invoke("PUT", `/api/v1/namespaces/${namespace}/tenants/${tenant}/yaml`, {
        yaml: tenantYaml,
      })
      .then(() => {
        setAddLoading(false);
        dispatch(getTenantAsync());
        setErrorMessage("");
        navigate(`/namespaces/${namespace}/tenants/${tenant}/summary`);
      })
      .catch((err: ErrorResponseHandler) => {
        setAddLoading(false);
        const errMessage = err?.message || "" || err.errorMessage;
        setErrorMessage(errMessage);
      });
  };

  useEffect(() => {
    if (namespace && tenant) {
      api
        .invoke("GET", `/api/v1/namespaces/${namespace}/tenants/${tenant}/yaml`)
        .then((res: ITenantYAML) => {
          setLoading(false);
          setTenantYaml(res.yaml);
        })
        .catch((err: ErrorResponseHandler) => {
          setLoading(false);
          dispatch(setModalErrorSnackMessage(err));
        });
    }
  }, [tenant, namespace, dispatch]);

  useEffect(() => {}, []);

  const validSave = tenantYaml.trim() !== "";

  return (
    <Fragment>
      {addLoading ||
        (loading && (
          <Grid item xs={12}>
            <ProgressBar />
          </Grid>
        ))}

      {!loading && (
        <form
          noValidate
          autoComplete="off"
          onSubmit={(e: React.FormEvent<HTMLFormElement>) => {
            e.preventDefault();
            updateTenant(e);
          }}
        >
          <Grid container>
            <Grid item xs={12}>
              <SectionTitle>Tenant Specification</SectionTitle>
            </Grid>
            {errorMessage ? (
              <Grid item xs={12}>
                <InformativeMessage
                  title={"Error"}
                  message={errorMessage}
                  variant={"error"}
                />
              </Grid>
            ) : null}
            <Grid item xs={12}>
              <CodeEditor
                value={tenantYaml}
                mode={"yaml"}
                onChange={(value) => {
                  setTenantYaml(value);
                }}
                editorHeight={"550px"}
              />
            </Grid>
            <Grid
              item
              xs={12}
              style={{
                display: "flex",
                justifyContent: "flex-end",
                paddingTop: 16,
              }}
            >
              <Button
                id={"cancel-tenant-yaml"}
                type="button"
                variant="regular"
                disabled={addLoading}
                onClick={() => {
                  navigate(
                    `/namespaces/${namespace}/tenants/${tenant}/summary`,
                  );
                }}
                label={"Cancel"}
              />
              <Button
                id={"save-tenant-yaml"}
                type="submit"
                variant="callAction"
                disabled={addLoading || !validSave}
                style={{ marginLeft: 8 }}
                label={"Save"}
              />
            </Grid>
          </Grid>
        </form>
      )}
    </Fragment>
  );
};

export default TenantYAML;
