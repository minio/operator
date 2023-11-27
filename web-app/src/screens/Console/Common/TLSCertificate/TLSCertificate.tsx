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
import { AlertCloseIcon, Box, CertificateIcon, IconButton } from "mds";
import { DateTime, Duration } from "luxon";
import styled from "styled-components";
import get from "lodash/get";
import { ICertificateInfo } from "../../Tenants/types";
import LanguageIcon from "@mui/icons-material/Language";
import EventBusyIcon from "@mui/icons-material/EventBusy";
import AccessTimeIcon from "@mui/icons-material/AccessTime";

const CertificateContainer = styled.div(({ theme }) => ({
  position: "relative",
  margin: 0,
  userSelect: "none",
  appearance: "none",
  maxWidth: "100%",
  fontFamily: "'Inter', sans-serif",
  fontSize: 13,
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  gap: 6,
  border: `1px solid ${get(theme, "borderColor", "#E2E2E2")}`,
  borderRadius: 3,
  padding: "5px 10px",
  "& .certificateName": {
    display: "flex",
    alignItems: "center",
    gap: 5,
    fontWeight: "bold",
    color: get(theme, "signalColors.main", "#07193E"),
  },
  "& .deleteTagButton": {
    backgroundColor: "transparent",
    border: 0,
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    padding: 0,
    cursor: "pointer",
    opacity: 0.6,
    "&:hover": {
      opacity: 1,
    },
    "& svg": {
      fill: get(theme, `tag.grey.background`, "#07193E"),
      width: 10,
      height: 10,
      minWidth: 10,
      minHeight: 10,
    },
  },
  "& .certificateContainer": {
    margin: "5px 10px",
  },
  "& .certificateExpiry": {
    color: get(theme, "secondaryText", "#5B5C5C"),
    display: "flex",
    alignItems: "center",
    flexWrap: "wrap",
    marginBottom: 5,
    "& .label": {
      fontWeight: "bold",
    },
  },
  "& .certificateDomains": {
    color: get(theme, "secondaryText", "#5B5C5C"),
    "& .label": {
      fontWeight: "bold",
    },
  },
  "& .certificatesList": {
    border: `1px solid ${get(theme, "borderColor", "#E2E2E2")}`,
    borderRadius: 4,
    color: get(theme, "secondaryText", "#5B5C5C"),
    textTransform: "lowercase",
    overflowY: "scroll",
    maxHeight: 145,
    marginTop: 3,
    marginBottom: 5,
    padding: 0,
    "& li": {
      listStyle: "none",
      padding: "5px 10px",
      margin: 0,
      display: "flex",
      alignItems: "center",
      "&:before": {
        content: "' '",
      },
    },
  },
  "& .certificatesListItem": {
    padding: "0px 16px",
    borderBottom: `1px solid ${get(theme, "borderColor", "#E2E2E2")}`,
    "& div": {
      minWidth: 0,
    },
    "& svg": {
      fontSize: 12,
      marginRight: 10,
      opacity: 0.5,
    },
    "& span": {
      fontSize: 12,
    },
  },
  "& .certificateExpiring": {
    color: get(theme, "signalColors.warning", "#FFBD62"),
    "& .label": {
      fontWeight: "bold",
    },
  },
  "& .certificateExpired": {
    color: get(theme, "signalColors.danger", "#C51B3F"),
    "& .label": {
      fontWeight: "bold",
    },
  },
  "& .closeIcon": {
    transform: "scale(0.8)",
  },
}));

interface ITLSCertificate {
  certificateInfo: ICertificateInfo;
  onDelete: any;
}

const TLSCertificate = ({
  certificateInfo,
  onDelete = () => {},
}: ITLSCertificate) => {
  const certificates = certificateInfo.domains || [];

  const expiry = DateTime.fromISO(certificateInfo.expiry);
  const now = DateTime.utc();
  // Expose error on Tenant if certificate is near expiration or expired
  let daysToExpiry: number = 0;
  let daysToExpiryHuman: string = "";
  let certificateExpiration: string = "";
  if (expiry) {
    let durationToExpiry = expiry.diff(now);
    daysToExpiry = durationToExpiry.as("days");
    daysToExpiryHuman = durationToExpiry
      .minus(Duration.fromObject({ days: 1 }))
      .shiftTo("days")
      .toHuman({ listStyle: "long", maximumFractionDigits: 0 });
    if (daysToExpiry >= 10 && daysToExpiry < 30) {
      certificateExpiration = "certificateExpiring";
    }
    if (daysToExpiry < 10) {
      certificateExpiration = "certificateExpired";
      if (daysToExpiry < 2) {
        daysToExpiryHuman = durationToExpiry
          .minus(Duration.fromObject({ minutes: 1 }))
          .shiftTo("hours", "minutes")
          .toHuman({ listStyle: "long", maximumFractionDigits: 0 });
        if (durationToExpiry.as("minutes") <= 1) {
          daysToExpiryHuman = "EXPIRED";
        }
      }
    }
  }

  return (
    <CertificateContainer>
      <Box>
        <Box className={"certificateName"}>
          <CertificateIcon />
          <span>{certificateInfo.name}</span>
        </Box>
        <Box className={"certificateContainer"}>
          <Box className={"certificateExpiry"}>
            <EventBusyIcon color="inherit" fontSize="small" />
            &nbsp;
            <span className={"label"}>Expiry:&nbsp;</span>
            <span>{expiry.toFormat("yyyy/MM/dd")}</span>
          </Box>
          <Box className={"certificateExpiry"}>
            <AccessTimeIcon color="inherit" fontSize="small" />
            &nbsp;
            <span className={"label"}>Expires in:&nbsp;</span>
            <span className={certificateExpiration}>{daysToExpiryHuman}</span>
          </Box>
          <hr style={{ marginBottom: 12 }} />
          <Box className={"certificateDomains"}>
            <span className="label">{`${certificates.length} Domain (s):`}</span>
          </Box>
          <ul className={"certificatesList"}>
            {certificates.map((dom, index) => (
              <li key={`${dom}-${index}`} className={"certificatesListItem"}>
                <LanguageIcon />
                <span>{dom}</span>
              </li>
            ))}
          </ul>
        </Box>
      </Box>
      <IconButton size={"small"} onClick={onDelete} className={"closeIcon"}>
        <AlertCloseIcon />
      </IconButton>
    </CertificateContainer>
  );
};

export default TLSCertificate;
