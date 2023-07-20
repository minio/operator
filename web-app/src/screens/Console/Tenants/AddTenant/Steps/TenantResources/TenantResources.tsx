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

import React, { useEffect, useState } from "react";
import { useSelector } from "react-redux";
import get from "lodash/get";
import NameTenantMain from "./NameTenantMain";
import { IMkEnvs, resourcesConfigurations } from "./utils";
import { selFeatures } from "../../../../consoleSlice";

const TenantResources = () => {
  const features = useSelector(selFeatures);
  const [formRender, setFormRender] = useState<IMkEnvs | null>(null);

  useEffect(() => {
    let setConfiguration = IMkEnvs.default;

    if (features && features.length !== 0) {
      const possibleVariables = Object.keys(resourcesConfigurations);

      possibleVariables.forEach((element) => {
        if (features.includes(element)) {
          setConfiguration = get(
            resourcesConfigurations,
            element,
            IMkEnvs.default,
          );
        }
      });
    }

    setFormRender(setConfiguration);
  }, [features]);

  if (formRender === null) {
    return null;
  }

  return <NameTenantMain formToRender={formRender} />;
};

export default TenantResources;
