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

import React, { Fragment, useState } from "react";
import { SectionTitle, ProgressBar } from "mds";
import styled from "styled-components";
import { useParams } from "react-router-dom";
import { IAM_PAGES } from "../../../../common/SecureComponent/permissions";

const IFrameContainer = styled.iframe(() => ({
  border: "0px",
  flex: "1 1 auto",
  minHeight: "800px",
  width: "100%",
}));

const TenantMetrics = () => {
  const { tenantName, tenantNamespace } = useParams();

  const [loading, setLoading] = useState<boolean>(true);

  return (
    <Fragment>
      <SectionTitle separator sx={{ marginBottom: 15 }}>
        Metrics
      </SectionTitle>
      {loading && (
        <div style={{ marginTop: "80px" }}>
          <ProgressBar />
        </div>
      )}
      <IFrameContainer
        title={"metrics"}
        src={`/api/proxy/${tenantNamespace || ""}/${tenantName || ""}${
          IAM_PAGES.DASHBOARD
        }?cp=y`}
        onLoad={() => {
          setLoading(false);
        }}
      />
    </Fragment>
  );
};

export default TenantMetrics;
