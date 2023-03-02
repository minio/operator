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
import { LinearProgress } from "@mui/material";
import { IEvent } from "../../ListTenants/types";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Box from "@mui/material/Box";
import Collapse from "@mui/material/Collapse";
import Typography from "@mui/material/Typography";
import KeyboardArrowDownIcon from "@mui/icons-material/KeyboardArrowDown";
import KeyboardArrowUpIcon from "@mui/icons-material/KeyboardArrowUp";
import TableContainer from "@mui/material/TableContainer";
import Paper from "@mui/material/Paper";

interface IEventsListProps {
  events: IEvent[];
  loading: boolean;
}

const Event = (props: { event: IEvent }) => {
  const { event } = props;
  const [open, setOpen] = React.useState(false);

  return (
    <React.Fragment>
      <TableRow sx={{ "& > *": { borderBottom: "unset" }, cursor: "pointer" }}>
        <TableCell component="th" scope="row" onClick={() => setOpen(!open)}>
          {event.event_type}
        </TableCell>
        <TableCell onClick={() => setOpen(!open)}>{event.reason}</TableCell>
        <TableCell onClick={() => setOpen(!open)}>{event.seen}</TableCell>
        <TableCell onClick={() => setOpen(!open)}>
          {event.message.length >= 30
            ? `${event.message.slice(0, 30)}...`
            : event.message}
        </TableCell>
        <TableCell onClick={() => setOpen(!open)}>
          {open ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
        </TableCell>
      </TableRow>
      <TableRow>
        <TableCell style={{ paddingBottom: 0, paddingTop: 0 }} colSpan={5}>
          <Collapse in={open} timeout="auto" unmountOnExit>
            <Box sx={{ margin: 1 }}>
              <Typography
                style={{
                  background: "#efefef",
                  border: "1px solid #dedede",
                  padding: 4,
                  fontSize: 14,
                  color: "#666666",
                }}
              >
                {event.message}
              </Typography>
            </Box>
          </Collapse>
        </TableCell>
      </TableRow>
    </React.Fragment>
  );
};

const EventsList = ({ events, loading }: IEventsListProps) => {
  if (loading) {
    return <LinearProgress />;
  }
  return (
    <TableContainer component={Paper}>
      <Table aria-label="collapsible table">
        <TableHead>
          <TableRow>
            <TableCell>Type</TableCell>
            <TableCell>Reason</TableCell>
            <TableCell>Age</TableCell>
            <TableCell>Message</TableCell>
            <TableCell />
          </TableRow>
        </TableHead>
        <TableBody>
          {events.map((event) => (
            <Event key={`${event.event_type}-${event.seen}`} event={event} />
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default EventsList;
