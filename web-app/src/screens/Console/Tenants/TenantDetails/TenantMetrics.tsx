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

import React, { useState } from "react";
import { Theme } from "@mui/material/styles";
import { useParams } from "react-router-dom";
import { LinearProgress } from "@mui/material";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import {
  containerForHeader,
  tenantDetailsStyles,
} from "../../Common/FormComponents/common/styleLibrary";
import { IAM_PAGES } from "../../../../common/SecureComponent/permissions";

interface ITenantMetrics {
  classes: any;
}

const styles = (theme: Theme) =>
  createStyles({
    ...tenantDetailsStyles,
    iframeStyle: {
      border: "0px",
      flex: "1 1 auto",
      minHeight: "800px",
      width: "100%",
    },
    ...containerForHeader,
  });

const TenantMetrics = ({ classes }: ITenantMetrics) => {
  const { tenantName, tenantNamespace } = useParams();

  const [loading, setLoading] = useState<boolean>(true);

  return (
    <React.Fragment>
      <h1 className={classes.sectionTitle}>Metrics</h1>
      {loading && (
        <div style={{ marginTop: "80px" }}>
          <LinearProgress />
        </div>
      )}
      <iframe
        className={classes.iframeStyle}
        title={"metrics"}
        src={`/api/proxy/${tenantNamespace || ""}/${tenantName || ""}${
          IAM_PAGES.DASHBOARD
        }?cp=y`}
        onLoad={() => {
          setLoading(false);
        }}
      />
    </React.Fragment>
  );
};

export default withStyles(styles)(TenantMetrics);
