//  This file is part of MinIO Console Server
//  Copyright (c) 2022 MinIO, Inc.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

import hasPermission from "../accessControl";
import { store } from "../../../store";
import { IAM_PAGES, IAM_PAGES_PERMISSIONS, IAM_SCOPES } from "../permissions";
import { saveSessionResponse } from "../../../screens/Console/consoleSlice";

const setPolicy1 = () => {
  store.dispatch(
    saveSessionResponse({
      distributedMode: true,
      features: ["log-search"],
      permissions: {
        "arn:aws:s3:::testcafe": [
          "admin:CreateUser",
          "s3:GetBucketLocation",
          "s3:ListBucket",
          "admin:CreateServiceAccount",
        ],
        "arn:aws:s3:::testcafe/*": [
          "admin:CreateServiceAccount",
          "admin:CreateUser",
          "s3:GetObject",
          "s3:ListBucket",
        ],
        "arn:aws:s3:::testcafe/write/*": [
          "admin:CreateServiceAccount",
          "admin:CreateUser",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:GetObject",
          "s3:ListBucket",
        ],
        "console-ui": ["admin:CreateServiceAccount", "admin:CreateUser"],
      },
      operator: false,
      status: "ok",
    })
  );
};
const setPolicy2 = () => {
  store.dispatch(
    saveSessionResponse({
      distributedMode: true,
      operator: false,
      features: [],
      permissions: {
        "arn:aws:s3:::bucket-svc": [
          "admin:CreateServiceAccount",
          "s3:GetBucketLocation",
          "s3:ListBucket",
          "s3:ListBucketMultipartUploads",
          "s3:ListMultipartUploadParts",
          "admin:CreateUser",
        ],
        "arn:aws:s3:::bucket-svc/prefix1/*": [
          "admin:CreateUser",
          "admin:CreateServiceAccount",
          "s3:GetObject",
          "s3:PutObject",
        ],
        "arn:aws:s3:::bucket-svc/prefix1/ini*": [
          "admin:CreateServiceAccount",
          "s3:*",
          "admin:CreateUser",
        ],
        "arn:aws:s3:::bucket-svc/prefix1/jars*": [
          "admin:CreateUser",
          "admin:CreateServiceAccount",
          "s3:*",
        ],
        "arn:aws:s3:::bucket-svc/prefix1/logs*": [
          "admin:CreateUser",
          "admin:CreateServiceAccount",
          "s3:*",
        ],
        "console-ui": ["admin:CreateServiceAccount", "admin:CreateUser"],
      },
      status: "ok",
    })
  );
};
const setPolicy3 = () => {
  store.dispatch(
    saveSessionResponse({
      distributedMode: true,
      features: [],
      permissions: {
        "arn:aws:s3:::testbucket-*": [
          "admin:CreateServiceAccount",
          "s3:*",
          "admin:CreateUser",
        ],
        "console-ui": ["admin:CreateServiceAccount", "admin:CreateUser"],
      },
      status: "ok",
      operator: false,
    })
  );
};

const setPolicy4 = () => {
  store.dispatch(
    saveSessionResponse({
      distributedMode: true,
      features: [],
      permissions: {
        "arn:aws:s3:::test/*": ["s3:ListBucket"],
        "arn:aws:s3:::test": ["s3:GetBucketLocation"],
        "arn:aws:s3:::test/digitalinsights/xref_cust_guid_actd*": ["s3:*"],
      },
      status: "ok",
      operator: false,
    })
  );
};

test("Upload button disabled", () => {
  setPolicy1();
  expect(hasPermission("testcafe", ["s3:PutObject"])).toBe(false);
});

test("Upload button enabled valid prefix", () => {
  setPolicy1();
  expect(hasPermission("testcafe/write", ["s3:PutObject"], false, true)).toBe(
    true
  );
});

test("Can Browse Bucket", () => {
  setPolicy2();
  expect(
    hasPermission(
      "bucket-svc",
      IAM_PAGES_PERMISSIONS[IAM_PAGES.OBJECT_BROWSER_VIEW]
    )
  ).toBe(true);
});

test("Can List Objects In Bucket", () => {
  setPolicy2();
  expect(hasPermission("bucket-svc", [IAM_SCOPES.S3_LIST_BUCKET])).toBe(true);
});

test("Can create bucket for policy with a wildcard", () => {
  setPolicy3();
  expect(hasPermission("*", [IAM_SCOPES.S3_CREATE_BUCKET])).toBe(true);
});

test("Can browse a bucket for a policy with a wildcard", () => {
  setPolicy3();
  expect(
    hasPermission(
      "testbucket-0",
      IAM_PAGES_PERMISSIONS[IAM_PAGES.OBJECT_BROWSER_VIEW]
    )
  ).toBe(true);
});

test("Can delete an object inside a bucket prefix", () => {
  setPolicy4();
  expect(
    hasPermission(
      [
        "xref_cust_guid_actd-v1.jpg",
        "test/digitalinsights/xref_cust_guid_actd-v1.jpg",
      ],
      [IAM_SCOPES.S3_DELETE_OBJECT]
    )
  ).toBe(true);
});

test("Can't delete an object inside a bucket prefix", () => {
  setPolicy4();
  expect(
    hasPermission(
      ["xref_cust_guid_actd-v1.jpg", "test/xref_cust_guid_actd-v1.jpg"],
      [IAM_SCOPES.S3_DELETE_OBJECT]
    )
  ).toBe(false);
});
