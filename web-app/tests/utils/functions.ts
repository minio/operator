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

import * as roles from "./roles";
import * as elements from "./elements";
import * as constants from "./constants";
import { Selector } from "testcafe";

import * as Minio from "minio";

export const setUpBucket = (t, modifier) => {
  return setUpNamedBucket(t, `${constants.TEST_BUCKET_NAME}-${modifier}`);
};

export const setUpNamedBucket = (t, name) => {
  const minioClient = new Minio.Client({
    endPoint: "localhost",
    port: 9000,
    useSSL: false,
    accessKey: "minioadmin",
    secretKey: "minioadmin",
  });
  return new Promise((resolve, reject) => {
    minioClient.makeBucket(name, "us-east-1", (err) => {
      if (err) {
        console.log(err);
      }
      resolve("done: " + err);
    });
  });
};

export const uploadObjectToBucket = (t, modifier, objectName, objectPath) => {
  const bucketName = `${constants.TEST_BUCKET_NAME}-${modifier}`;
  return uploadNamedObjectToBucket(t, bucketName, objectName, objectPath);
};

export const uploadNamedObjectToBucket = (
  t,
  modifier,
  objectName,
  objectPath,
) => {
  const bucketName = modifier;
  const minioClient = new Minio.Client({
    endPoint: "localhost",
    port: 9000,
    useSSL: false,
    accessKey: "minioadmin",
    secretKey: "minioadmin",
  });
  return new Promise((resolve, reject) => {
    minioClient.fPutObject(bucketName, objectName, objectPath, {}, (err) => {
      if (err) {
        console.log(err);
      }
      resolve("done");
    });
  });
};

export const setVersioned = (t, modifier) => {
  return setVersionedBucket(t, `${constants.TEST_BUCKET_NAME}-${modifier}`);
};

export const setVersionedBucket = (t, name) => {
  const minioClient = new Minio.Client({
    endPoint: "localhost",
    port: 9000,
    useSSL: false,
    accessKey: "minioadmin",
    secretKey: "minioadmin",
  });

  return new Promise((resolve, reject) => {
    minioClient
      .setBucketVersioning(name, { Status: "Enabled" })
      .then(resolve)
      .catch(resolve);
  });
};

export const namedManageButtonFor = (name) => {
  return Selector("div").withAttribute("id", `manageBucket-${name}`);
};

export const manageButtonFor = (modifier) => {
  return namedManageButtonFor(`${constants.TEST_BUCKET_NAME}-${modifier}`);
};

export const cleanUpNamedBucket = (t, name) => {
  const minioClient = new Minio.Client({
    endPoint: "localhost",
    port: 9000,
    useSSL: false,
    accessKey: "minioadmin",
    secretKey: "minioadmin",
  });

  return minioClient.removeBucket(name);
};

export const cleanUpBucket = (t, modifier) => {
  return cleanUpNamedBucket(t, `${constants.TEST_BUCKET_NAME}-${modifier}`);
};

export const namedTestBucketBrowseButtonFor = (name) => {
  return Selector("span").withAttribute("id", `browse-${name}`);
};

export const testBucketBrowseButtonFor = (modifier) => {
  return namedTestBucketBrowseButtonFor(
    `${constants.TEST_BUCKET_NAME}-${modifier}`,
  );
};

export const uploadFilesButton = () => {
  return Selector("button").withText("Upload Files");
};

export const cleanUpNamedBucketAndUploads = (t, bucket) => {
  return new Promise((resolve, reject) => {
    const minioClient = new Minio.Client({
      endPoint: "localhost",
      port: 9000,
      useSSL: false,
      accessKey: "minioadmin",
      secretKey: "minioadmin",
    });

    var stream = minioClient.listObjects(bucket, "", true);

    let proms = [];
    stream.on("data", function (obj) {
      proms.push(minioClient.removeObject(bucket, obj.name));
    });

    stream.on("end", () => {
      Promise.all(proms).then(() => {
        minioClient.removeBucket(bucket).then(resolve).catch(resolve);
      });
    });
  });
};

export const cleanUpBucketAndUploads = (t, modifier) => {
  const bucket = `${constants.TEST_BUCKET_NAME}-${modifier}`;
  return cleanUpNamedBucketAndUploads(t, bucket);
};

export const createUser = (t) => {
  return t
    .useRole(roles.admin)
    .navigateTo(`http://localhost:9090/identity/users/add-user`)
    .typeText(elements.usersAccessKeyInput, constants.TEST_USER_NAME)
    .typeText(elements.usersSecretKeyInput, constants.TEST_PASSWORD)
    .click(elements.saveButton);
};

export const cleanUpUser = (t) => {
  const userListItem = Selector(".ReactVirtualized__Table__rowColumn").withText(
    constants.TEST_USER_NAME,
  );

  const userDeleteIconButton = userListItem
    .nextSibling()
    .child("button")
    .withAttribute("aria-label", "delete");

  return t
    .useRole(roles.admin)
    .navigateTo("http://localhost:9090/identity/users")
    .click(userDeleteIconButton)
    .click(elements.deleteButton);
};
