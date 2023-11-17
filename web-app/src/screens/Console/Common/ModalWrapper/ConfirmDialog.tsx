// This file is part of MinIO Operator
// Copyright (c) 2023 MinIO, Inc.
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
import { Box, Button, ModalBox } from "mds";

interface ButtonProps {
  label?: string;
  variant?: "regular" | "callAction" | "secondary";
  icon?: React.ReactNode;
  iconLocation?: "start" | "end";
  fullWidth?: boolean;
  disabled?: boolean;
  onClick?: React.MouseEventHandler<HTMLButtonElement>;
}

type ConfirmDialogProps = {
  isOpen?: boolean;
  onClose: () => void;
  onCancel?: () => void;
  onConfirm: () => void;
  title: string;
  isLoading?: boolean;
  confirmationContent: React.ReactNode | React.ReactNode[];
  cancelText?: string;
  confirmText?: string;
  confirmButtonProps?: ButtonProps &
    React.ButtonHTMLAttributes<HTMLButtonElement>;
  cancelButtonProps?: ButtonProps &
    React.ButtonHTMLAttributes<HTMLButtonElement>;
  titleIcon?: React.ReactNode;
  confirmationButtonSimple?: boolean;
};

const ConfirmDialog = ({
  isOpen = false,
  onClose,
  onCancel,
  onConfirm,
  title = "",
  isLoading,
  confirmationContent,
  cancelText = "Cancel",
  confirmText = "Confirm",
  confirmButtonProps = undefined,
  cancelButtonProps = undefined,
  titleIcon = null,
  confirmationButtonSimple = false,
}: ConfirmDialogProps) => {
  return (
    <ModalBox
      title={title}
      titleIcon={titleIcon}
      onClose={onClose}
      open={isOpen}
      customMaxWidth={510}
    >
      <Box>{confirmationContent}</Box>
      <Box
        sx={{
          display: "flex",
          justifyContent: "flex-end",
          gap: 10,
          marginTop: 20,
        }}
      >
        <Button
          onClick={onCancel || onClose}
          disabled={isLoading}
          type="button"
          {...cancelButtonProps}
          variant="regular"
          id={"confirm-cancel"}
          label={cancelText}
        />

        <Button
          id={"confirm-ok"}
          onClick={onConfirm}
          label={confirmText}
          disabled={isLoading}
          variant={"secondary"}
          {...confirmButtonProps}
        />
      </Box>
    </ModalBox>
  );
};

export default ConfirmDialog;
