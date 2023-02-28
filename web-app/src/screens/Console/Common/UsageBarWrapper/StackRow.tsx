import React from "react";
import { Stack } from "@mui/material";

const StackRow = ({
  children = null,
  ...restProps
}: {
  children?: any;
  [x: string]: any;
}) => {
  return (
    <Stack
      direction={{ xs: "column", sm: "row" }}
      justifyContent="space-between"
      margin={"5px 0 5px 0"}
      spacing={{ xs: 1, sm: 2, md: 4 }}
      {...restProps}
    >
      {children}
    </Stack>
  );
};
export default StackRow;
