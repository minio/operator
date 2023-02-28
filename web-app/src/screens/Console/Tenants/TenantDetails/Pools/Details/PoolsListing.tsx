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
import withStyles from "@mui/styles/withStyles";
import { AppState, useAppDispatch } from "../../../../../../store";
import { useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";
import { IPool } from "../../../ListTenants/types";
import Grid from "@mui/material/Grid";
import { TextField } from "@mui/material";
import InputAdornment from "@mui/material/InputAdornment";
import { AddIcon, Button, SearchIcon } from "mds";
import TableWrapper from "../../../../Common/TableWrapper/TableWrapper";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import {
  actionsTray,
  containerForHeader,
  tableStyles,
  tenantDetailsStyles,
} from "../../../../Common/FormComponents/common/styleLibrary";
import { setSelectedPool } from "../../../tenantsSlice";
import TooltipWrapper from "../../../../Common/TooltipWrapper/TooltipWrapper";

interface IPoolsSummary {
  classes: any;
  setPoolDetailsView: () => void;
}

const styles = (theme: Theme) =>
  createStyles({
    ...tenantDetailsStyles,
    ...actionsTray,
    ...tableStyles,
    ...containerForHeader,
  });

const PoolsListing = ({ classes, setPoolDetailsView }: IPoolsSummary) => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();

  const loadingTenant = useSelector(
    (state: AppState) => state.tenants.loadingTenant
  );
  const tenant = useSelector((state: AppState) => state.tenants.tenantInfo);

  const [pools, setPools] = useState<IPool[]>([]);
  const [filter, setFilter] = useState<string>("");

  useEffect(() => {
    if (tenant) {
      const resPools = !tenant.pools ? [] : tenant.pools;
      setPools(resPools);
    }
  }, [tenant]);

  const filteredPools = pools.filter((pool) => {
    if (pool.name.toLowerCase().includes(filter.toLowerCase())) {
      return true;
    }

    return false;
  });

  const listActions = [
    {
      type: "view",
      onClick: (selectedValue: IPool) => {
        dispatch(setSelectedPool(selectedValue.name));
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
                }/add-pool`
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
            { label: "Capacity", elementKey: "capacity" },
            { label: "# of Instances", elementKey: "servers" },
            { label: "# of Drives", elementKey: "volumes" },
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

export default withStyles(styles)(PoolsListing);
