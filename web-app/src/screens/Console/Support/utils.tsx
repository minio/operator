import { Box, Grid, Link } from "@mui/material";
import { Fragment } from "react";
import RegistrationStatusBanner from "./RegistrationStatusBanner";

export const FormTitle = ({
  icon = null,
  title,
}: {
  icon?: any;
  title: any;
}) => {
  return (
    <Box
      sx={{
        display: "flex",
        alignItems: "center",
        justifyContent: "flex-start",
      }}
    >
      {icon}
      <div className="title-text">{title}</div>
    </Box>
  );
};

export const ClusterRegistered = ({ email }: { email: string }) => {
  return (
    <Fragment>
      <RegistrationStatusBanner email={email} />
      <Grid item xs={12} marginTop={"25px"}>
        <Box
          sx={{
            padding: "20px",
            "& a": {
              color: "#2781B0",
              cursor: "pointer",
            },
          }}
        >
          Login to{" "}
          <Link
            href="https://subnet.min.io"
            target="_blank"
            style={{
              color: "#2781B0",
              cursor: "pointer",
            }}
          >
            SUBNET
          </Link>{" "}
          to avail support for this MinIO cluster
        </Box>
      </Grid>
    </Fragment>
  );
};
