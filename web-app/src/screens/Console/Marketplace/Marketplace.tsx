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
import SetEmailModal from "./SetEmailModal";
import { PageLayout } from "mds";
import { selFeatures } from "../consoleSlice";
import { useSelector } from "react-redux";
import { Navigate, useNavigate } from "react-router-dom";
import { resourcesConfigurations } from "../Tenants/AddTenant/Steps/TenantResources/utils";
import { selShowMarketplace, showMarketplace } from "../../../systemSlice";
import { useAppDispatch } from "../../../store";
import PageHeaderWrapper from "../Common/PageHeaderWrapper/PageHeaderWrapper";

const Marketplace = () => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const features = useSelector(selFeatures);
  const displayMarketplace = useSelector(selShowMarketplace);
  const [isMPMode, setMPMode] = useState<boolean>(true);

  useEffect(() => {
    let mpMode = false;
    if (features && features.length !== 0) {
      features.forEach((feature) => {
        if (feature in resourcesConfigurations) {
          mpMode = true;
          return;
        }
      });
    }
    setMPMode(mpMode);
  }, [features, displayMarketplace]);

  const getTargetPath = () => {
    let targetPath = "/";
    if (
      localStorage.getItem("redirect-path") &&
      localStorage.getItem("redirect-path") !== ""
    ) {
      targetPath = `${localStorage.getItem("redirect-path")}`;
      localStorage.setItem("redirect-path", "");
    }
    return targetPath;
  };

  const closeModal = () => {
    dispatch(showMarketplace(false));
    navigate(getTargetPath());
  };

  if (!displayMarketplace || !isMPMode) {
    return <Navigate to={{ pathname: getTargetPath() }} />;
  }

  if (features) {
    return (
      <Fragment>
        <PageHeaderWrapper label="Operator Marketplace" />
        <PageLayout>
          <SetEmailModal open={true} closeModal={closeModal} />
        </PageLayout>
      </Fragment>
    );
  }
  return null;
};

export default Marketplace;
