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

export const fileProcess = (evt: any, callback: any) => {
  const file = evt.target.files[0];
  const reader = new FileReader();
  reader.readAsDataURL(file);

  reader.onload = () => {
    // reader.readAsDataURL(file) output will be something like: data:application/x-x509-ca-cert;base64,LS0tLS1CRUdJTiBDRVJUSU
    // we care only about the actual base64 part (everything after "data:application/x-x509-ca-cert;base64,")
    const fileBase64 = reader.result;
    if (fileBase64) {
      const fileArray = fileBase64.toString().split("base64,");

      if (fileArray.length === 2) {
        callback(fileArray[1]);
      }
    }
  };
};
