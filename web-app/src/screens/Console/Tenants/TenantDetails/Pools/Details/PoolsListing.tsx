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
import { AddIcon, Button, DataTable, Grid } from "mds";
import { useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";
import { AppState, useAppDispatch } from "../../../../../../store";
import { actionsTray } from "../../../../Common/FormComponents/common/styleLibrary";
import { setSelectedPool } from "../../../tenantsSlice";
import { Pool } from "../../../../../../api/operatorApi";
import { niceBytesInt } from "../../../../../../common/utils";
import TooltipWrapper from "../../../../Common/TooltipWrapper/TooltipWrapper";
import SearchBox from "../../../../Common/SearchBox";

interface IPoolsSummary {
  setPoolDetailsView: () => void;
}

const PoolsListing = ({ setPoolDetailsView }: IPoolsSummary) => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();

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
      <Grid item xs={12} sx={actionsTray.actionsTray}>
        <SearchBox
          value={filter}
          onChange={(value) => {
            setFilter(value);
          }}
          placeholder={"Filter"}
          id="search-resource"
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
      <Grid item xs={12}>
        <DataTable
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
          customPaperHeight={"calc(100vh - 400px)"}
        />
      </Grid>
    </Fragment>
  );
};

export default PoolsListing;
