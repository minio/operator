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
import { DateTime, Duration } from "luxon";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { ICertificateInfo } from "../../Tenants/types";
import LanguageIcon from "@mui/icons-material/Language";
import Chip from "@mui/material/Chip";
import {
  Box,
  Container,
  Divider,
  Grid,
  List,
  ListItem,
  ListItemAvatar,
  ListItemText,
  Typography,
} from "@mui/material";
import EventBusyIcon from "@mui/icons-material/EventBusy";
import AccessTimeIcon from "@mui/icons-material/AccessTime";
import { CertificateIcon } from "mds";

const styles = (theme: Theme) =>
  createStyles({
    certificateIcon: {
      float: "left",
      paddingTop: "5px !important",
      paddingRight: "10px !important",
    },
    certificateInfo: { float: "right" },
    certificateWrapper: {
      height: "auto",
      margin: 5,
      border: "1px solid #E2E2E2",
      userSelect: "text",
      borderRadius: 4,
      "& h6": {
        fontWeight: "bold",
      },
      "& div": {
        padding: 0,
      },
    },
    certificateExpiry: {
      color: "#616161",
      display: "flex",
      alignItems: "center",
      flexWrap: "wrap",
      marginBottom: 5,
      "& .label": {
        fontWeight: "bold",
      },
    },
    certificateDomains: {
      color: "#616161",
      "& .label": {
        fontWeight: "bold",
      },
    },
    certificatesList: {
      border: "1px solid #E2E2E2",
      borderRadius: 4,
      color: "#616161",
      textTransform: "lowercase",
      overflowY: "scroll",
      maxHeight: 145,
      marginBottom: 10,
    },
    certificatesListItem: {
      padding: "0px 16px",
      borderBottom: "1px solid #E2E2E2",
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
    certificateExpiring: {
      color: "orange",
      "& .label": {
        fontWeight: "bold",
      },
    },
    certificateExpired: {
      color: "red",
      "& .label": {
        fontWeight: "bold",
      },
    },
  });

interface ITLSCertificate {
  classes: any;
  certificateInfo: ICertificateInfo;
  onDelete: any;
}

const TLSCertificate = ({
  classes,
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
      certificateExpiration = classes.certificateExpiring;
    }
    if (daysToExpiry < 10) {
      certificateExpiration = classes.certificateExpired;
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
    <Chip
      key={certificateInfo.name}
      variant="outlined"
      color="primary"
      className={classes.certificateWrapper}
      label={
        <Container>
          <Grid item xs={1} className={classes.certificateIcon}>
            <CertificateIcon />
          </Grid>
          <Grid item xs={11} className={classes.certificateInfo}>
            <Typography variant="subtitle1" display="block" gutterBottom>
              {certificateInfo.name}
            </Typography>
            <Box className={classes.certificateExpiry}>
              <EventBusyIcon color="inherit" fontSize="small" />
              &nbsp;
              <span className={"label"}>Expiry:&nbsp;</span>
              <span>{expiry.toFormat("yyyy/MM/dd")}</span>
            </Box>
            <Box className={classes.certificateExpiry}>
              <AccessTimeIcon color="inherit" fontSize="small" />
              &nbsp;
              <span className={"label"}>Expires in:&nbsp;</span>
              <span className={certificateExpiration}>{daysToExpiryHuman}</span>
            </Box>
            <Divider />
            <br />
            <Box className={classes.certificateDomains}>
              <span className="label">{`${certificates.length} Domain (s):`}</span>
            </Box>
            <List className={classes.certificatesList}>
              {certificates.map((dom, index) => (
                <ListItem
                  key={`${dom}-${index}`}
                  className={classes.certificatesListItem}
                >
                  <ListItemAvatar>
                    <LanguageIcon />
                  </ListItemAvatar>
                  <ListItemText primary={dom} />
                </ListItem>
              ))}
            </List>
          </Grid>
        </Container>
      }
      onDelete={onDelete}
    />
  );
};

export default withStyles(styles)(TLSCertificate);
