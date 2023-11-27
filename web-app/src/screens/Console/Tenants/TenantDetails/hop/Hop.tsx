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
import { Box, IconButton, Loader, LoginIcon, RefreshIcon } from "mds";
import { Link, useNavigate, useParams } from "react-router-dom";
import get from "lodash/get";
import styled from "styled-components";
import PageHeaderWrapper from "../../../Common/PageHeaderWrapper/PageHeaderWrapper";

const HopContainer = styled.div(({ theme }) => ({
  position: "absolute",
  left: 0,
  top: 80,
  height: "calc(100vh - 81px)",
  width: "100%",
  borderTop: `1px solid ${get(theme, "borderColor", "#E2E2E2")}`,
  "& .loader": {
    width: 100,
    margin: "auto",
    marginTop: 80,
  },
  "& .iframeStyle": {
    border: 0,
    position: "absolute",
    height: "calc(100vh - 77px)",
    width: "100%",
  },
}));

const Hop = () => {
  const navigate = useNavigate();
  const params = useParams();

  const [loading, setLoading] = useState<boolean>(true);

  const tenantName = params.tenantName || "";
  const tenantNamespace = params.tenantNamespace || "";
  const consoleFrame = React.useRef<HTMLIFrameElement>(null);

  return (
    <Fragment>
      <PageHeaderWrapper
        label={
          <Fragment>
            <Link to={"/tenants"}>Tenants</Link>
            {` > `}
            <Link to={`/namespaces/${tenantNamespace}/tenants/${tenantName}`}>
              {tenantName}
            </Link>
            {` > Management`}
          </Fragment>
        }
        actions={
          <React.Fragment>
            <IconButton
              onClick={() => {
                if (
                  consoleFrame !== null &&
                  consoleFrame.current !== null &&
                  consoleFrame.current.contentDocument !== null
                ) {
                  const loc =
                    consoleFrame.current.contentDocument.location.toString();

                  let add = "&";

                  if (loc.indexOf("?") < 0) {
                    add = `?`;
                  }

                  if (loc.indexOf("cp=y") < 0) {
                    const next = `${loc}${add}cp=y`;
                    consoleFrame.current.contentDocument.location.replace(next);
                  } else {
                    consoleFrame.current.contentDocument.location.reload();
                  }
                }
              }}
              size="large"
            >
              <RefreshIcon />
            </IconButton>
            <IconButton
              onClick={() => {
                navigate(
                  `/namespaces/${tenantNamespace}/tenants/${tenantName}`,
                );
              }}
              size="large"
            >
              <LoginIcon />
            </IconButton>
          </React.Fragment>
        }
      />
      <HopContainer>
        {loading && (
          <Box className={"loader"}>
            <Loader />
          </Box>
        )}
        <iframe
          ref={consoleFrame}
          className={"iframeStyle"}
          title={"metrics"}
          src={`/api/hop/${tenantNamespace}/${tenantName}/?cp=y`}
          onLoad={(val) => {
            setLoading(false);
          }}
        />
      </HopContainer>
    </Fragment>
  );
};

export default Hop;
