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
import { AddIcon, Button, HelpBox, RefreshIcon, TenantsIcon } from "mds";
import Grid from "@mui/material/Grid";
import { LinearProgress, SelectChangeEvent } from "@mui/material";
import { Theme } from "@mui/material/styles";
import { NewServiceAccount } from "../../Common/CredentialsPrompt/types";
import {
  actionsTray,
  containerForHeader,
  searchField,
} from "../../Common/FormComponents/common/styleLibrary";
import TenantListItem from "./TenantListItem";
import AButton from "../../Common/AButton/AButton";

import withSuspense from "../../Common/Components/withSuspense";
import VirtualizedList from "../../Common/VirtualizedList/VirtualizedList";
import SearchBox from "../../Common/SearchBox";
import PageLayout from "../../Common/Layout/PageLayout";
import { setErrorSnackMessage } from "../../../../systemSlice";
import SelectWrapper from "../../Common/FormComponents/SelectWrapper/SelectWrapper";
import { useNavigate } from "react-router-dom";
import { useAppDispatch } from "../../../../store";
import TooltipWrapper from "../../Common/TooltipWrapper/TooltipWrapper";
import PageHeaderWrapper from "../../Common/PageHeaderWrapper/PageHeaderWrapper";
import makeStyles from "@mui/styles/makeStyles";
import { api } from "../../../../api";
import {
  Error,
  HttpResponse,
  ListTenantsResponse,
  TenantList,
} from "../../../../api/operatorApi";

const CredentialsPrompt = withSuspense(
  React.lazy(() => import("../../Common/CredentialsPrompt/CredentialsPrompt")),
);

const useStyles = makeStyles((theme: Theme) => ({
  ...actionsTray,
  ...searchField,
  ...containerForHeader,
  tenantsList: {
    height: "calc(100vh - 195px)",
  },
  sortByContainer: {
    display: "flex",
    justifyContent: "flex-end",
    marginBottom: 10,
  },
  innerSort: {
    maxWidth: 200,
    width: "95%",
    display: "flex",
    flexDirection: "row",
    alignItems: "center",
  },
  sortByLabel: {
    whiteSpace: "nowrap",
    fontSize: 14,
    color: "#838383",
    fontWeight: "bold",
    marginRight: 10,
  },
}));

const ListTenants = () => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const classes = useStyles();

  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [filterTenants, setFilterTenants] = useState<string>("");
  const [records, setRecords] = useState<TenantList[]>([]);
  const [showNewCredentials, setShowNewCredentials] = useState<boolean>(false);
  const [createdAccount, setCreatedAccount] =
    useState<NewServiceAccount | null>(null);
  const [sortValue, setSortValue] = useState<string>("name");

  const closeCredentialsModal = () => {
    setShowNewCredentials(false);
    setCreatedAccount(null);
  };

  const filteredRecords = records.filter((b: any) => {
    if (filterTenants === "") {
      return true;
    } else {
      if (b.name.indexOf(filterTenants) >= 0) {
        return true;
      } else {
        return false;
      }
    }
  });

  filteredRecords.sort((a, b) => {
    switch (sortValue) {
      case "capacity":
        if (!a.capacity || !b.capacity) {
          return 0;
        }

        if (a.capacity > b.capacity) {
          return 1;
        }

        if (a.capacity < b.capacity) {
          return -1;
        }

        return 0;
      case "usage":
        if (!a.capacity_usage || !b.capacity_usage) {
          return 0;
        }

        if (a.capacity_usage > b.capacity_usage) {
          return 1;
        }

        if (a.capacity_usage < b.capacity_usage) {
          return -1;
        }

        return 0;
      case "active_status":
        if (a.health_status === "red" && b.health_status !== "red") {
          return 1;
        }

        if (a.health_status !== "red" && b.health_status === "red") {
          return -1;
        }

        return 0;
      case "failing_status":
        if (a.health_status === "green" && b.health_status !== "green") {
          return 1;
        }

        if (a.health_status !== "green" && b.health_status === "green") {
          return -1;
        }

        return 0;
      default:
        if (a.name! > b.name!) {
          return 1;
        }
        if (a.name! < b.name!) {
          return -1;
        }
        return 0;
    }
  });

  useEffect(() => {
    if (isLoading) {
      const fetchRecords = () => {
        api.tenants
          .listAllTenants()
          .then((res: HttpResponse<ListTenantsResponse, Error>) => {
            if (!res.data) {
              setIsLoading(false);
              return;
            }
            let resTenants: TenantList[] =
              (res.data.tenants as TenantList[]) ?? [];

            setRecords(resTenants);
            setIsLoading(false);
          })
          .catch((err) => {
            dispatch(setErrorSnackMessage(err));
            setIsLoading(false);
          });
      };
      fetchRecords();
    }
  }, [isLoading, dispatch]);

  useEffect(() => {
    setIsLoading(true);
  }, []);

  const renderItemLine = (index: number) => {
    const tenant = filteredRecords[index] || null;

    if (tenant) {
      return <TenantListItem tenant={tenant} />;
    }

    return null;
  };

  return (
    <Fragment>
      {showNewCredentials && (
        <CredentialsPrompt
          newServiceAccount={createdAccount}
          open={showNewCredentials}
          closeModal={() => {
            closeCredentialsModal();
          }}
          entity="Tenant"
        />
      )}
      <PageHeaderWrapper
        label="Tenants"
        middleComponent={
          <SearchBox
            placeholder={"Filter Tenants"}
            onChange={(val) => {
              setFilterTenants(val);
            }}
            value={filterTenants}
          />
        }
        actions={
          <Grid
            item
            xs={12}
            sx={{ display: "flex", justifyContent: "flex-end" }}
          >
            <TooltipWrapper tooltip={"Refresh Tenant List"}>
              <Button
                id={"refresh-tenant-list"}
                onClick={() => {
                  setIsLoading(true);
                }}
                icon={<RefreshIcon />}
                variant={"regular"}
              />
            </TooltipWrapper>
            <TooltipWrapper tooltip={"Create Tenant"}>
              <Button
                id={"create-tenant"}
                label={"Create Tenant"}
                onClick={() => {
                  navigate("/tenants/add");
                }}
                icon={<AddIcon />}
                variant={"callAction"}
              />
            </TooltipWrapper>
          </Grid>
        }
      />
      <PageLayout>
        <Grid item xs={12} className={classes.tenantsList}>
          {isLoading && <LinearProgress />}
          {!isLoading && (
            <Fragment>
              {filteredRecords.length !== 0 && (
                <Fragment>
                  <Grid item xs={12} className={classes.sortByContainer}>
                    <div className={classes.innerSort}>
                      <span className={classes.sortByLabel}>Sort by</span>
                      <SelectWrapper
                        id={"sort-by"}
                        label={""}
                        value={sortValue}
                        onChange={(e: SelectChangeEvent<string>) => {
                          setSortValue(e.target.value as string);
                        }}
                        name={"sort-by"}
                        options={[
                          { label: "Name", value: "name" },
                          {
                            label: "Capacity",
                            value: "capacity",
                          },
                          {
                            label: "Usage",
                            value: "usage",
                          },
                          {
                            label: "Active Status",
                            value: "active_status",
                          },
                          {
                            label: "Failing Status",
                            value: "failing_status",
                          },
                        ]}
                      />
                    </div>
                  </Grid>
                  <VirtualizedList
                    rowRenderFunction={renderItemLine}
                    totalItems={filteredRecords.length}
                  />
                </Fragment>
              )}
              {filteredRecords.length === 0 && (
                <Grid
                  container
                  justifyContent={"center"}
                  alignContent={"center"}
                  alignItems={"center"}
                >
                  <Grid item xs={8}>
                    <HelpBox
                      iconComponent={<TenantsIcon />}
                      title={"Tenants"}
                      help={
                        <Fragment>
                          Tenant is the logical structure to represent a MinIO
                          deployment. A tenant can have different size and
                          configurations from other tenants, even a different
                          storage class.
                          <br />
                          <br />
                          To get started,&nbsp;
                          <AButton
                            onClick={() => {
                              navigate("/tenants/add");
                            }}
                          >
                            Create a Tenant.
                          </AButton>
                        </Fragment>
                      }
                    />
                  </Grid>
                </Grid>
              )}
            </Fragment>
          )}
        </Grid>
      </PageLayout>
    </Fragment>
  );
};

export default ListTenants;
