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

import React, { Fragment } from "react";
import styled from "styled-components";
import get from "lodash/get";

const LinkElement = styled.a(({ theme }) => ({
  color: get(theme, "linkColor", "#2781B0"),
  fontWeight: 600,
}));

const LicenseFAQ = () => {
  return (
    <Fragment>
      <h2>What is the GNU AGPL v3?</h2>
      <p>
        The GNU AGPL v3 is short for the "GNU Affero General Public License v3."
        It is a common open source license certified by the Free Software
        Foundation and the Open Source Initiative. You can get a copy of the GNU
        AGPL v3 license with MinIO source code or at&nbsp;
        <LinkElement
          href={"https://min.io/compliance?ref=op"}
          target={"_blank"}
        >
          https://www.gnu.org/licenses/agpl-3.0.en.html
        </LinkElement>
        .
      </p>
      <h2>What does it mean for me to comply with the GNU AGPL v3?</h2>
      <p>
        When you host or distribute MinIO over a network, the AGPL v3 applies to
        you. Any distribution or copying of MinIO software modified or not has
        to comply with the obligations specified in the AGPL v3. Otherwise, you
        may risk infringing MinIO’s copyrights.
      </p>
      <h2>Making combined or derivative works of MinIO</h2>
      <p>
        Combining MinIO software as part of a larger software stack triggers
        your GNU AGPL v3 obligations.
      </p>
      <p>
        The method of combining does not matter. When MinIO is linked to a
        larger software stack in any form, including statically, dynamically,
        pipes, or containerized and invoked remotely, the AGPL v3 applies to
        your use. What triggers the AGPL v3 obligations is the exchanging data
        between the larger stack and MinIO.
      </p>
      <h2>Talking to your Legal Counsel</h2>
      <p>
        If you have questions, we recommend that you talk to your own attorney
        for legal advice. Purchasing a commercial license from MinIO removes the
        AGPL v3 obligations from MinIO software.
      </p>
    </Fragment>
  );
};
export default LicenseFAQ;
