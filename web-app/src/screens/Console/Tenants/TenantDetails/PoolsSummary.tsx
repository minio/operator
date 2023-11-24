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

import React, { Fragment } from "react";
import { BackLink, SectionTitle, Box } from "mds";
import { useSelector } from "react-redux";
import { useLocation, useNavigate } from "react-router-dom";
import { AppState, useAppDispatch } from "../../../../store";
import { setOpenPoolDetails } from "../tenantsSlice";
import PoolsListing from "./Pools/Details/PoolsListing";
import PoolDetails from "./Pools/Details/PoolDetails";

const PoolsSummary = () => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();

  const { pathname = "" } = useLocation();

  const selectedPool = useSelector(
    (state: AppState) => state.tenants.selectedPool,
  );
  const poolDetailsOpen = useSelector(
    (state: AppState) => state.tenants.poolDetailsOpen,
  );

  return (
    <Fragment>
      {poolDetailsOpen && (
        <Box>
          <BackLink
            label={"Pools list"}
            onClick={() => {
              navigate(pathname);
              dispatch(setOpenPoolDetails(false));
            }}
          />
        </Box>
      )}
      <SectionTitle separator sx={{ marginBottom: 15 }}>
        {poolDetailsOpen ? `Pool Details - ${selectedPool || ""}` : "Pools"}
      </SectionTitle>

      <Box>
        {poolDetailsOpen ? (
          <PoolDetails />
        ) : (
          <PoolsListing
            setPoolDetailsView={() => {
              dispatch(setOpenPoolDetails(true));
            }}
          />
        )}
      </Box>
    </Fragment>
  );
};

export default PoolsSummary;
