// This file is part of MinIO Operator
// Copyright (c) 2021 MinIO, Inc.
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

import React, { Fragment, useState } from "react";
import { Button, ConfirmDeleteIcon, PageLayout, SectionTitle, Grid } from "mds";
import ConfirmDialog from "./ModalWrapper/ConfirmDialog";
import PageHeaderWrapper from "./PageHeaderWrapper/PageHeaderWrapper";

const ComponentsScreen = () => {
  const [dialogOpen, setDialogOpen] = useState<boolean>(false);

  return (
    <Fragment>
      <PageHeaderWrapper label={"Components"} />

      <PageLayout>
        <Grid container>
          <Grid item xs={12}>
            <SectionTitle>Confirm Dialogs</SectionTitle>
          </Grid>
          <Grid item xs={12}>
            <p>Used to confirm a non-idempotent action.</p>
          </Grid>
          <Grid item xs={12}>
            <Button
              id={"open-dialog-test"}
              type="button"
              variant={"regular"}
              onClick={() => {
                setDialogOpen(true);
              }}
              label={"Open Dialog"}
            />
            <ConfirmDialog
              title={`Delete Bucket`}
              confirmText={"Delete"}
              isOpen={dialogOpen}
              titleIcon={<ConfirmDeleteIcon />}
              isLoading={false}
              onConfirm={() => {
                setDialogOpen(false);
              }}
              onClose={() => {
                setDialogOpen(false);
              }}
              confirmationContent={
                <Fragment>
                  Are you sure you want to delete bucket <b>bucket</b>
                  ? <br />A bucket can only be deleted if it's empty.
                </Fragment>
              }
            />
          </Grid>
        </Grid>
      </PageLayout>
    </Fragment>
  );
};

export default ComponentsScreen;
