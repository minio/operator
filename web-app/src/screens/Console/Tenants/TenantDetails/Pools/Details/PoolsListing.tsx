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
import { AppState, useAppDispatch } from "../../../../../../store";
import { useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";
import Grid from "@mui/material/Grid";
import { TextField } from "@mui/material";
import InputAdornment from "@mui/material/InputAdornment";
import { AddIcon, Button, SearchIcon } from "mds";
import TableWrapper from "../../../../Common/TableWrapper/TableWrapper";
import { Theme } from "@mui/material/styles";
import {
  actionsTray,
  containerForHeader,
  tableStyles,
  tenantDetailsStyles,
} from "../../../../Common/FormComponents/common/styleLibrary";
import { setSelectedPool } from "../../../tenantsSlice";
import TooltipWrapper from "../../../../Common/TooltipWrapper/TooltipWrapper";
import { Pool } from "../../../../../../api/operatorApi";
import makeStyles from "@mui/styles/makeStyles";
import createStyles from "@mui/styles/createStyles";
import { niceBytesInt } from "../../../../../../common/utils";

interface IPoolsSummary {
  setPoolDetailsView: () => void;
}

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    ...tenantDetailsStyles,
    ...actionsTray,
    ...tableStyles,
    ...containerForHeader,
  }),
);

const PoolsListing = ({ setPoolDetailsView }: IPoolsSummary) => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const classes = useStyles();

  const loadingTenant = useSelector(
    (state: AppState) => state.tenants.loadingTenant,
  );
  const tenant = useSelector((state: AppState) => state.tenants.tenantInfo);

  const [pools, setPools] = useState<Pool[]>([]);
  const [filter, setFilter] = useState<string>("");

  useEffect(() => {
    if (tenant) {
      const resPools = !tenant.pools ? [] : tenant.pools;
      setPools(resPools);
    }
  }, [tenant]);

  const filteredPools = pools.filter((pool) => {
    if (pool.name?.toLowerCase().includes(filter.toLowerCase())) {
      return true;
    }

    return false;
  });

  const listActions = [
    {
      type: "view",
      onClick: (selectedValue: Pool) => {
        dispatch(setSelectedPool(selectedValue.name!));
        setPoolDetailsView();
      },
    },
  ];

  return (
    <Fragment>
      <Grid item xs={12} className={classes.actionsTray}>
        <TextField
          placeholder="Filter"
          className={classes.searchField}
          id="search-resource"
          label=""
          onChange={(event) => {
            setFilter(event.target.value);
          }}
          InputProps={{
            disableUnderline: true,
            startAdornment: (
              <InputAdornment position="start">
                <SearchIcon />
              </InputAdornment>
            ),
          }}
          variant="standard"
        />

        <TooltipWrapper tooltip={"Expand Tenant"}>
          <Button
            id={"expand-tenant"}
            label={"Expand Tenant"}
            onClick={() => {
              navigate(
                `/namespaces/${tenant?.namespace || ""}/tenants/${
                  tenant?.name || ""
                }/add-pool`,
              );
            }}
            icon={<AddIcon />}
            variant={"callAction"}
          />
        </TooltipWrapper>
      </Grid>
      <Grid item xs={12} className={classes.tableBlock}>
        <TableWrapper
          itemActions={listActions}
          columns={[
            { label: "Name", elementKey: "name" },
            {
              label: "Total Capacity",
              elementKey: "capacity",
              renderFullObject: true,
              renderFunction: (x: Pool) =>
                niceBytesInt(
                  x.volumes_per_server *
                    x.servers *
                    x.volume_configuration.size,
                ),
            },
            { label: "Servers", elementKey: "servers" },
            { label: "Volumes/Server", elementKey: "volumes_per_server" },
          ]}
          isLoading={loadingTenant}
          records={filteredPools}
          entityName="Servers"
          idField="name"
          customEmptyMessage="No Pools found"
        />
      </Grid>
    </Fragment>
  );
};

export default PoolsListing;
