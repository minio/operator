import React, { Fragment, useState } from "react";
import { IMessageEvent, w3cwebsocket as W3CWebSocket } from "websocket";
import { Theme } from "@mui/material/styles";
import { Button } from "mds";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { Grid } from "@mui/material";
import PageLayout from "../Common/Layout/PageLayout";
import CheckboxWrapper from "../Common/FormComponents/CheckboxWrapper/CheckboxWrapper";
import { wsProtocol } from "../../../utils/wsUtils";
import {
  actionsTray,
  containerForHeader,
  inlineCheckboxes,
} from "../Common/FormComponents/common/styleLibrary";
import { useNavigate } from "react-router-dom";
import RegisterCluster from "./RegisterCluster";
import { registeredCluster } from "../../../config";
import PageHeaderWrapper from "../Common/PageHeaderWrapper/PageHeaderWrapper";

const styles = (theme: Theme) =>
  createStyles({
    buttonContainer: {
      display: "flex",
      justifyContent: "flex-end",
      marginTop: 24,
      "& button": {
        marginLeft: 8,
      },
    },
    dropdown: {
      marginBottom: 24,
    },
    checkboxLabel: {
      marginTop: 12,
      marginRight: 4,
      fontSize: 16,
      fontWeight: 500,
    },
    checkboxDisabled: {
      opacity: 0.5,
    },
    inlineCheckboxes: {
      ...inlineCheckboxes.inlineCheckboxes,
      alignItems: "center",

      "@media (max-width: 900px)": {
        flexFlow: "column",
        alignItems: "flex-start",
      },
    },
    ...actionsTray,
    ...containerForHeader,
  });

interface IProfileProps {
  classes: any;
}

var c: any = null;

const Profile = ({ classes }: IProfileProps) => {
  const navigate = useNavigate();

  const [profilingStarted, setProfilingStarted] = useState<boolean>(false);
  const [types, setTypes] = useState<string[]>([
    "cpu",
    "mem",
    "block",
    "mutex",
    "goroutines",
  ]);
  const clusterRegistered = registeredCluster();
  const typesList = [
    { label: "cpu", value: "cpu" },
    { label: "mem", value: "mem" },
    { label: "block", value: "block" },
    { label: "mutex", value: "mutex" },
    { label: "goroutines", value: "goroutines" },
  ];

  const onCheckboxClick = (e: React.ChangeEvent<HTMLInputElement>) => {
    let newArr: string[] = [];
    if (types.indexOf(e.target.value) > -1) {
      newArr = types.filter((type) => type !== e.target.value);
    } else {
      newArr = [...types, e.target.value];
    }
    setTypes(newArr);
  };

  const startProfiling = () => {
    const typeString = types.join(",");

    const url = new URL(window.location.toString());
    const isDev = process.env.NODE_ENV === "development";
    const port = isDev ? "9090" : url.port;

    // check if we are using base path, if not this always is `/`
    const baseLocation = new URL(document.baseURI);
    const baseUrl = baseLocation.pathname;

    const wsProt = wsProtocol(url.protocol);
    c = new W3CWebSocket(
      `${wsProt}://${url.hostname}:${port}${baseUrl}ws/profile?types=${typeString}`
    );

    if (c !== null) {
      c.onopen = () => {
        setProfilingStarted(true);
        c.send("ok");
      };
      c.onmessage = (message: IMessageEvent) => {
        // process received message
        let response = new Blob([message.data], { type: "application/zip" });
        let filename = "profile.zip";
        setProfilingStarted(false);
        var link = document.createElement("a");
        link.href = window.URL.createObjectURL(response);
        link.download = filename;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
      };
      c.onclose = () => {
        console.log("connection closed by server");
        setProfilingStarted(false);
      };
      return () => {
        c.close(1000);
        console.log("closing websockets");
        setProfilingStarted(false);
      };
    }
  };

  const stopProfiling = () => {
    c.close(1000);
    setProfilingStarted(false);
  };

  return (
    <Fragment>
      <PageHeaderWrapper label="Profile" />
      <PageLayout>
        {!clusterRegistered && <RegisterCluster compactMode />}
        <Grid item xs={12} className={classes.boxy}>
          <Grid item xs={12} className={classes.dropdown}>
            <Grid
              item
              xs={12}
              className={`${classes.inlineCheckboxes} ${
                profilingStarted && classes.checkboxDisabled
              }`}
            >
              <div className={classes.checkboxLabel}>Types to profile:</div>
              {typesList.map((t) => (
                <CheckboxWrapper
                  checked={types.indexOf(t.value) > -1}
                  disabled={profilingStarted}
                  key={`checkbox-${t.label}`}
                  id={`checkbox-${t.label}`}
                  label={t.label}
                  name={`checkbox-${t.label}`}
                  onChange={onCheckboxClick}
                  value={t.value}
                />
              ))}
            </Grid>
          </Grid>
          <Grid item xs={12} className={classes.buttonContainer}>
            <Button
              id={"start-profiling"}
              type="submit"
              variant={clusterRegistered ? "callAction" : "regular"}
              disabled={profilingStarted || types.length < 1}
              onClick={() => {
                if (!clusterRegistered) {
                  navigate("/support/register");
                  return;
                }
                startProfiling();
              }}
              label={"Start Profiling"}
            />
            <Button
              id={"stop-profiling"}
              type="submit"
              variant="callAction"
              color="primary"
              disabled={!profilingStarted}
              onClick={() => {
                stopProfiling();
              }}
              label={"Stop Profiling"}
            />
          </Grid>
        </Grid>
      </PageLayout>
    </Fragment>
  );
};

export default withStyles(styles)(Profile);
