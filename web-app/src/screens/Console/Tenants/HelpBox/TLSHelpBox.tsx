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
import React from "react";
import { useSelector } from "react-redux";
import { CertificateIcon, Box, breakPoints } from "mds";
import { useParams } from "react-router-dom";
import { AppState } from "../../../../store";

const FeatureItem = ({
  icon,
  description,
}: {
  icon: any;
  description: string;
}) => {
  return (
    <Box
      sx={{
        display: "flex",
        "& .min-icon": {
          marginRight: "10px",
          height: "23px",
          width: "23px",
          marginBottom: "10px",
        },
      }}
    >
      {icon}{" "}
      <Box className="muted" sx={{ fontSize: "14px", fontStyle: "italic" }}>
        {description}
      </Box>
    </Box>
  );
};
const TLSHelpBox = () => {
  const params = useParams();
  const tenantNameParam = params.tenantName || "";
  const tenantNamespaceParam = params.tenantNamespace || "";
  const namespace = useSelector((state: AppState) => {
    let defaultNamespace = "<namespace>";
    if (tenantNamespaceParam !== "") {
      return tenantNamespaceParam;
    }
    if (state.createTenant.fields.nameTenant.namespace !== "") {
      return state.createTenant.fields.nameTenant.namespace;
    }
    return defaultNamespace;
  });

  const tenantName = useSelector((state: AppState) => {
    let defaultTenantName = "<tenant-name>";
    if (tenantNameParam !== "") {
      return tenantNameParam;
    }

    if (state.createTenant.fields.nameTenant.tenantName !== "") {
      return state.createTenant.fields.nameTenant.tenantName;
    }
    return defaultTenantName;
  });

  return (
    <Box
      sx={{
        flex: 1,
        border: "1px solid #eaeaea",
        borderRadius: "2px",
        display: "flex",
        flexFlow: "column",
        padding: "20px",
        [`@media (max-width: ${breakPoints.sm}px)`]: {
          marginTop: 0,
        },
      }}
    >
      <Box
        sx={{
          display: "flex",
          flexFlow: "column",
        }}
      >
        <FeatureItem
          icon={<CertificateIcon />}
          description={`TLS Certificates Warning`}
        />
        <Box sx={{ fontSize: "14px", marginBottom: "15px" }}>
          Automatic certificate generation is not enabled.
          <br />
          <br />
          If you wish to continue only with <b>custom certificates</b> make sure
          they are valid for the following internode hostnames, i.e.:
          <br />
          <br />
          <Box
            sx={{ fontSize: "14px", fontStyle: "italic" }}
            className={"muted"}
          >
            minio.{namespace}
            <br />
            minio.{namespace}.svc
            <br />
            minio.{namespace}.svc.&#x3C;cluster domain&#x3E;
            <br />
            *.{tenantName}-hl.{namespace}.svc.&#x3C;cluster domain&#x3E;
            <br />
            *.{namespace}.svc.&#x3C;cluster domain&#x3E;
          </Box>
          <br />
          Replace <em>&#x3C;tenant-name&#x3E;</em>,{" "}
          <em>&#x3C;namespace&#x3E;</em> and
          <em>&#x3C;cluster domain&#x3E;</em> with the actual values for your
          MinIO tenant.
          <br />
          <br />
          You can learn more at our{" "}
          <a
            href="https://min.io/docs/minio/kubernetes/upstream/operations/network-encryption.html?ref=op#id5"
            target="_blank"
            rel="noopener"
          >
            documentation
          </a>
          .
        </Box>
      </Box>
    </Box>
  );
};

export default TLSHelpBox;
