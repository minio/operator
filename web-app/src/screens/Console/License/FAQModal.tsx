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
import ModalWrapper from "../Common/ModalWrapper/ModalWrapper";
import LicenseFAQ from "./LicenseFAQ";
import { useSelector } from "react-redux";
import { AppState, useAppDispatch } from "../../../store";
import { closeFAQModal } from "./licenseSlice";

const FAQModal = () => {
  const dispatch = useAppDispatch();
  const isOpen = useSelector((state: AppState) => state.license.faqModalOpen);

  return (
    <ModalWrapper
      modalOpen={isOpen}
      title="License FAQ"
      onClose={() => {
        dispatch(closeFAQModal());
      }}
    >
      <LicenseFAQ />
    </ModalWrapper>
  );
};

export default FAQModal;
