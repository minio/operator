import React from "react";
import { Grid } from "@mui/material";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { pageContentStyles } from "../FormComponents/common/styleLibrary";

const styles = (theme: Theme) =>
  createStyles({
    ...pageContentStyles,
  });

type PageLayoutProps = {
  className?: string;
  classes?: any;
  variant?: "constrained" | "full";
  children: any;
};

const PageLayout = ({
  classes,
  className = "",
  children,
  variant = "constrained",
}: PageLayoutProps) => {
  let style = variant === "constrained" ? { maxWidth: 1220 } : {};
  return (
    <div className={classes.contentSpacer}>
      <Grid container>
        <Grid item xs={12} className={className} style={style}>
          {children}
        </Grid>
      </Grid>
    </div>
  );
};

export default withStyles(styles)(PageLayout);
