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

import * as React from "react";
import { SVGProps } from "react";

const EncryptionIcon = (props: SVGProps<SVGSVGElement>) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="255.209"
    height="255.209"
    viewBox="0 0 255.209 255.209"
    className={`min-icon`}
    fill={"currentcolor"}
    {...props}
  >
    <path
      id="KMS"
      d="M175.664,255.209V228.695H79.546v26.515H46.4V228.695H3a3,3,0,0,1-3-3V3A3,3,0,0,1,3,0H252.21a3,3,0,0,1,3,3V225.694a3,3,0,0,1-3,3h-43.4v26.515ZM23.2,29.83V198.865a9.954,9.954,0,0,0,9.943,9.943H222.065a9.954,9.954,0,0,0,9.943-9.943V29.83a9.954,9.954,0,0,0-9.943-9.943H33.144A9.954,9.954,0,0,0,23.2,29.83ZM222.065,198.866h0Zm-188.921,0V29.83H222.065V198.865H33.144ZM69.224,88.258a26.52,26.52,0,1,0,34.909,34.375h33.071a2,2,0,0,0,2-2V104.747a2,2,0,0,0-2-2H104.134A26.545,26.545,0,0,0,69.224,88.258ZM59.659,112.69a19.886,19.886,0,1,1,19.886,19.886A19.887,19.887,0,0,1,59.659,112.69Z"
    />
  </svg>
);

export default EncryptionIcon;
