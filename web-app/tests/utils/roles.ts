import { readFileSync } from "fs";
import { Role, Selector } from "testcafe";

const data = readFileSync(__dirname + "/../constants/timestamp.txt", "utf-8");
const unixTimestamp = data.trim();

const loginUrl = "http://localhost:9090/login";
// diagnostics/watch/trace need to run in port 9090 (through the server) to work
const loginUrlServer = "http://localhost:9090/login";
const submitButton = Selector("form button");

export const admin = Role(
  loginUrl,
  async (t) => {
    await t
      .typeText("#accessKey", "minioadmin")
      .typeText("#secretKey", "minioadmin")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const bucketAssignPolicy = Role(
  loginUrl,
  async (t) => {
    await t
      .typeText("#accessKey", "bucketassignpolicy-" + unixTimestamp)
      .typeText("#secretKey", "bucketassignpolicy")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const bucketRead = Role(
  loginUrl,
  async (t) => {
    await t
      .typeText("#accessKey", "bucketread-" + unixTimestamp)
      .typeText("#secretKey", "bucketread")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const bucketWrite = Role(
  loginUrl,
  async (t) => {
    await t
      .typeText("#accessKey", "bucketwrite-" + unixTimestamp)
      .typeText("#secretKey", "bucketwrite")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const bucketReadWrite = Role(
  loginUrl,
  async (t) => {
    await t
      .typeText("#accessKey", "bucketreadwrite-" + unixTimestamp)
      .typeText("#secretKey", "bucketreadwrite")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const bucketObjectTags = Role(
  loginUrl,
  async (t) => {
    await t
      .typeText("#accessKey", "bucketobjecttags-" + unixTimestamp)
      .typeText("#secretKey", "bucketobjecttags")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const bucketCannotTag = Role(
  loginUrl,
  async (t) => {
    await t
      .typeText("#accessKey", "bucketcannottag-" + unixTimestamp)
      .typeText("#secretKey", "bucketcannottag")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const bucketSpecific = Role(
  loginUrl,
  async (t) => {
    await t
      .typeText("#accessKey", "bucketspecific-" + unixTimestamp)
      .typeText("#secretKey", "bucketspecific")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const bucketWritePrefixOnly = Role(
  loginUrl,
  async (t) => {
    await t
      .typeText("#accessKey", "bucketwriteprefixonlypolicy-" + unixTimestamp)
      .typeText("#secretKey", "bucketwriteprefixonlypolicy")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const dashboard = Role(
  loginUrl,
  async (t) => {
    await t
      .typeText("#accessKey", "dashboard-" + unixTimestamp)
      .typeText("#secretKey", "dashboard")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const diagnostics = Role(
  loginUrlServer,
  async (t) => {
    await t
      .typeText("#accessKey", "diagnostics-" + unixTimestamp)
      .typeText("#secretKey", "diagnostics")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const groups = Role(
  loginUrl,
  async (t) => {
    await t
      .typeText("#accessKey", "groups-" + unixTimestamp)
      .typeText("#secretKey", "groups1234")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const heal = Role(
  loginUrlServer,
  async (t) => {
    await t
      .typeText("#accessKey", "heal-" + unixTimestamp)
      .typeText("#secretKey", "heal1234")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const iamPolicies = Role(
  loginUrl,
  async (t) => {
    await t
      .typeText("#accessKey", "iampolicies-" + unixTimestamp)
      .typeText("#secretKey", "iampolicies")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const logs = Role(
  loginUrl,
  async (t) => {
    await t
      .typeText("#accessKey", "logs-" + unixTimestamp)
      .typeText("#secretKey", "logs1234")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const notificationEndpoints = Role(
  loginUrl,
  async (t) => {
    await t
      .typeText("#accessKey", "notificationendpoints-" + unixTimestamp)
      .typeText("#secretKey", "notificationendpoints")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const settings = Role(
  loginUrl,
  async (t) => {
    await t
      .typeText("#accessKey", "settings-" + unixTimestamp)
      .typeText("#secretKey", "settings")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const tiers = Role(
  loginUrl,
  async (t) => {
    await t
      .typeText("#accessKey", "tiers-" + unixTimestamp)
      .typeText("#secretKey", "tiers1234")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const trace = Role(
  loginUrlServer,
  async (t) => {
    await t
      .typeText("#accessKey", "trace-" + unixTimestamp)
      .typeText("#secretKey", "trace1234")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const users = Role(
  loginUrl,
  async (t) => {
    await t
      .typeText("#accessKey", "users-" + unixTimestamp)
      .typeText("#secretKey", "users1234")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const watch = Role(
  loginUrlServer,
  async (t) => {
    await t
      .typeText("#accessKey", "watch-" + unixTimestamp)
      .typeText("#secretKey", "watch1234")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const deleteObjectWithPrefixOnly = Role(
  loginUrl,
  async (t) => {
    await t
      .typeText("#accessKey", "delete-object-with-prefix-" + unixTimestamp)
      .typeText("#secretKey", "deleteobjectwithprefix1234")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const conditions1 = Role(
  loginUrl,
  async (t) => {
    await t
      .typeText("#accessKey", "conditions-" + unixTimestamp)
      .typeText("#secretKey", "conditions1234")
      .click(submitButton);
  },
  { preserveUrl: true }
);

export const conditions2 = Role(
  loginUrl,
  async (t) => {
    await t
      .typeText("#accessKey", "conditions-2-" + unixTimestamp)
      .typeText("#secretKey", "conditions1234")
      .click(submitButton);
  },
  { preserveUrl: true }
);
