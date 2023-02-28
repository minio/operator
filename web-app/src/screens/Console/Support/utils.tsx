import { Box, Grid, Link } from "@mui/material";
import { Fragment, useState } from "react";
import { CopyIcon, SettingsIcon } from "mds";
import FormSwitchWrapper from "../Common/FormComponents/FormSwitchWrapper/FormSwitchWrapper";
import InputBoxWrapper from "../Common/FormComponents/InputBoxWrapper/InputBoxWrapper";
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

export const ProxyConfiguration = () => {
  const proxyConfigurationCommand =
    "mc admin config set {alias} subnet proxy={proxy}";
  const [displaySubnetProxy, setDisplaySubnetProxy] = useState(false);
  return (
    <Fragment>
      <Box
        sx={{
          border: "1px solid #eaeaea",
          borderRadius: "2px",
          display: "flex",
          padding: "23px",
          marginTop: "40px",
          alignItems: "start",
          justifyContent: "space-between",
        }}
      >
        <Box
          sx={{
            display: "flex",
            flexFlow: "column",
          }}
        >
          <Box
            sx={{
              display: "flex",
              "& .min-icon": {
                height: "22px",
                width: "22px",
              },
            }}
          >
            <SettingsIcon />
            <div style={{ marginLeft: "10px", fontWeight: 600 }}>
              Proxy Configuration
            </div>
          </Box>
          <Box
            sx={{
              marginTop: "10px",
              marginBottom: "10px",
              fontSize: "14px",
            }}
          >
            For airgap/firewalled environments it is possible to{" "}
            <Link
              style={{
                color: "#2781B0",
                cursor: "pointer",
              }}
              href="https://min.io/docs/minio/linux/reference/minio-mc-admin/mc-admin-config.html?ref=con"
              target="_blank"
            >
              configure a proxy
            </Link>{" "}
            to connect to SUBNET .
          </Box>
          <Box>
            {displaySubnetProxy && (
              <InputBoxWrapper
                disabled
                id="subnetProxy"
                name="subnetProxy"
                placeholder=""
                onChange={() => {}}
                label=""
                value={proxyConfigurationCommand}
                overlayIcon={<CopyIcon />}
                extraInputProps={{
                  readOnly: true,
                }}
                overlayAction={() =>
                  navigator.clipboard.writeText(proxyConfigurationCommand)
                }
              />
            )}
          </Box>
        </Box>
        <Box
          sx={{
            display: "flex",
          }}
        >
          <FormSwitchWrapper
            value="enableProxy"
            id="enableProxy"
            name="enableProxy"
            checked={displaySubnetProxy}
            onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
              setDisplaySubnetProxy(event.target.checked);
            }}
          />
        </Box>
      </Box>
    </Fragment>
  );
};
