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
import {
  ProgressBar,
  Table,
  TableBody,
  TableHeadCell,
  TableCell,
  TableHead,
  TableRow,
  Box,
  ExpandCaret,
  CollapseCaret,
} from "mds";
import { IEvent } from "../../ListTenants/types";

interface IEventsListProps {
  events: IEvent[];
  loading: boolean;
}

const Event = (props: { event: IEvent }) => {
  const { event } = props;
  const [open, setOpen] = React.useState(false);

  return (
    <React.Fragment>
      <TableRow sx={{ cursor: "pointer" }}>
        <TableHeadCell
          scope="row"
          onClick={() => setOpen(!open)}
          sx={{ borderBottom: 0 }}
        >
          {event.event_type}
        </TableHeadCell>
        <TableCell onClick={() => setOpen(!open)} sx={{ borderBottom: 0 }}>
          {event.reason}
        </TableCell>
        <TableCell onClick={() => setOpen(!open)} sx={{ borderBottom: 0 }}>
          {event.seen}
        </TableCell>
        <TableCell onClick={() => setOpen(!open)} sx={{ borderBottom: 0 }}>
          {event.message.length >= 30
            ? `${event.message.slice(0, 30)}...`
            : event.message}
        </TableCell>
        <TableCell onClick={() => setOpen(!open)} sx={{ borderBottom: 0 }}>
          {open ? <CollapseCaret /> : <ExpandCaret />}
        </TableCell>
      </TableRow>
      <TableRow>
        <TableCell style={{ paddingBottom: 0, paddingTop: 0 }} colSpan={5}>
          {open && (
            <Box useBackground sx={{ padding: 10, marginBottom: 10 }}>
              {event.message}
            </Box>
          )}
        </TableCell>
      </TableRow>
    </React.Fragment>
  );
};

const EventsList = ({ events, loading }: IEventsListProps) => {
  if (loading) {
    return <ProgressBar />;
  }
  return (
    <Box withBorders customBorderPadding={"0px"}>
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
    </Box>
  );
};

export default EventsList;
