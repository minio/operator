import React, { Fragment } from "react";
import Grid from "@mui/material/Grid";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { predefinedList } from "../common/styleLibrary";

interface IPredefinedList {
  classes: any;
  label?: string;
  content: any;
  multiLine?: boolean;
  actionButton?: React.ReactNode;
}

const styles = (theme: Theme) =>
  createStyles({
    ...predefinedList,
  });

const PredefinedList = ({
  classes,
  label = "",
  content,
  multiLine = false,
  actionButton,
}: IPredefinedList) => {
  return (
    <Fragment>
      <Grid className={classes.prefinedContainer}>
        {label !== "" && (
          <Grid item xs={12} className={classes.predefinedTitle}>
            {label}
          </Grid>
        )}
        <Grid
          item
          xs={12}
          className={`${classes.predefinedList} ${
            actionButton ? classes.includesActionButton : ""
          }`}
        >
          <Grid
            item
            xs={12}
            className={
              multiLine ? classes.innerContentMultiline : classes.innerContent
            }
          >
            {content}
          </Grid>
          {actionButton && (
            <div className={classes.overlayShareOption}>{actionButton}</div>
          )}
        </Grid>
      </Grid>
    </Fragment>
  );
};

export default withStyles(styles)(PredefinedList);
