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

export interface IValidation {
  fieldKey: string;
  required: boolean;
  pattern?: RegExp;
  customPatternMessage?: string;
  customValidation?: boolean; // The validation to trigger the error
  customValidationMessage?: string;
  value: string;
}

export const commonFormValidation = (fieldsValidate: IValidation[]) => {
  let returnErrors: any = {};

  fieldsValidate.forEach((field) => {
    if (
      field.required &&
      typeof field.value !== "undefined" &&
      field.value.trim &&
      field.value.trim() === ""
    ) {
      returnErrors[field.fieldKey] = "Field cannot be empty";
      return;
    }
    // if it's not required and the value is empty, we are done here
    if (
      !field.required &&
      typeof field.value !== "undefined" &&
      field.value.trim &&
      field.value.trim() === ""
    ) {
      return;
    }

    if (field.customValidation && field.customValidationMessage) {
      returnErrors[field.fieldKey] = field.customValidationMessage;
      return;
    }

    if (field.pattern && field.customPatternMessage) {
      const rgx = new RegExp(field.pattern, "g");

      if (
        field.value &&
        field.value.trim() !== "" &&
        !field.value.match(rgx) &&
        typeof field.value !== "undefined"
      ) {
        returnErrors[field.fieldKey] = field.customPatternMessage;
      }
      return;
    }
  });

  return returnErrors;
};
