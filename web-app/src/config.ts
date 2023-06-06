// This file is part of MinIO Operator
// Copyright (c) 2023 MinIO, Inc.
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

export const MinIOPlan =
  (
    document.head.querySelector(
      "[name~=minio-license][content]"
    ) as HTMLMetaElement
  )?.content || "AGPL";

type LogoVar = "simple" | "AGPL" | "standard" | "enterprise";

export const getLogoVar = (): LogoVar => {
  let logoVar: LogoVar;
  switch (MinIOPlan) {
    case "enterprise":
      logoVar = "enterprise";
      break;
    case "STANDARD":
      logoVar = "standard";
      break;
    default:
      logoVar = "AGPL";
      break;
  }
  return logoVar;
};

export const registeredCluster = (): boolean => {
  const plan = getLogoVar();
  return plan === "standard" || plan === "enterprise";
};
