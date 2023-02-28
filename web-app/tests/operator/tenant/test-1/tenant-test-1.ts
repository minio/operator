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

import { Selector, t } from "testcafe";
import {
  createTenant,
  createTenantWithoutAuditLog,
  deleteTenant,
  goToLoggingDBSection,
  goToLoggingSection,
  goToMonitoringSection,
  goToPodInTenant,
  goToPodSection,
  goToPvcInTenant,
  goToPvcSection,
  loginToOperator,
} from "../../utils";

fixture("For user with default permissions").page("http://localhost:9090");

// Test 1
test("Create Tenant and List Tenants", async (t) => {
  const tenantName = `tenant-${Math.floor(Math.random() * 10000)}`;
  await loginToOperator();
  await createTenant(tenantName);
  await deleteTenant(tenantName);
});

// Test 2
test("Create Tenant Without Audit Log", async (t) => {
  const tenantName = `tenant-${Math.floor(Math.random() * 10000)}`;
  await loginToOperator();
  await createTenantWithoutAuditLog(tenantName);
  await deleteTenant(tenantName);
});

// Test 3
test("Test describe section for PODs in new tenant", async (t) => {
  const tenantName = "storage-lite";
  await loginToOperator();
  await testPODDescribe(tenantName);
});

const testPODDescribe = async (tenantName: string) => {
  await goToPodInTenant(tenantName);
  await goToPodSection(1);
  await checkPodDescribeHasSections();
};

const checkPodDescribeHasSections = async () => {
  await t
    .expect(Selector("#pod-describe-summary").exists)
    .ok()
    .expect(Selector("#pod-describe-annotations").exists)
    .ok()
    .expect(Selector("#pod-describe-labels").exists)
    .ok()
    .expect(Selector("#pod-describe-conditions").exists)
    .ok()
    .expect(Selector("#pod-describe-tolerations").exists)
    .ok()
    .expect(Selector("#pod-describe-volumes").exists)
    .ok()
    .expect(Selector("#pod-describe-containers").exists)
    .ok();
};

// Test 4
test("Test describe section for PVCs in new tenant", async (t) => {
  const tenantName = `storage-lite`;
  await loginToOperator();
  await testPvcDescribe(tenantName);
});

const testPvcDescribe = async (tenantName: string) => {
  await goToPvcInTenant(tenantName);
  await goToPvcSection(1);
  await checkPvcDescribeHasSections();
};

const checkPvcDescribeHasSections = async () => {
  await t
    .expect(Selector("#pvc-describe-summary").exists)
    .ok()
    .expect(Selector("#pvc-describe-annotations").exists)
    .ok()
    .expect(Selector("#pvc-describe-labels").exists)
    .ok();
};

export const checkMonitoringToggle = async (tenantName: string) => {
  await goToMonitoringSection(tenantName);
  await t
    .click("#tenant-monitoring")
    .click("#confirm-ok")
    .wait(1000)
    .expect(Selector("#image").exists)
    .notOk()
    .click("#yaml_button")
    .expect(Selector("#code_wrapper").exists)
    .ok();
  await t
    .expect(
      (await Selector("#code_wrapper").textContent).includes("prometheus:")
    )
    .notOk();
  await t
    .click(Selector(`a[href$="/monitoring"]`))
    .click("#tenant-monitoring")
    .click("#confirm-ok")
    .wait(5000)
    .expect(Selector("#prometheus_image").exists)
    .ok()
    .click("#yaml_button")
    .expect(Selector("#code_wrapper").exists)
    .ok();
  await t
    .expect(
      (await Selector("#code_wrapper").textContent).includes("prometheus:")
    )
    .ok();
};
export const checkMonitoringFieldsAcceptValues = async (tenantName: string) => {
  await goToMonitoringSection(tenantName);
  await t
    .typeText("#prometheus_image", "quay.io/prometheus/prometheus:latest", {
      replace: true,
    })
    .typeText("#sidecarImage", "library/alpine:latest", { replace: true })
    .typeText("#initImage", "library/busybox:1.33.1", { replace: true })
    .typeText("#diskCapacityGB", "1", { replace: true })
    .typeText("#cpuRequest", "1", { replace: true })
    .typeText("#memRequest", "1", { replace: true })
    .typeText("#serviceAccountName", "monitoringTestServiceAccountName", {
      replace: true,
    })
    .typeText("#storageClassName", "monitoringTestStorageClassName", {
      replace: true,
    })
    .typeText("#securityContext_runAsUser", "1212", { replace: true })
    .typeText("#securityContext_runAsGroup", "3434", { replace: true })
    .typeText("#securityContext_fsGroup", "5656", { replace: true })
    .expect(Selector("#securityContext_runAsNonRoot").checked)
    .ok()
    .click("#securityContext_runAsNonRoot")
    .expect(Selector("#securityContext_runAsNonRoot").checked)
    .notOk()
    .typeText("#key-Labels-0", "monitoringLabelKey0Test", { replace: true })
    .typeText("#val-Labels-0", "monitoringLabelVal0Test", { replace: true })
    .click("#add-Labels-0")
    .typeText("#key-Annotations-0", "monitoringAnnotationsKey0Test", {
      replace: true,
    })
    .typeText("#val-Annotations-0", "monitoringAnnotationsVal0Test", {
      replace: true,
    })
    .click("#add-Annotations-0")
    .typeText("#key-NodeSelector-0", "monitoringNodeSelectorKey0Test", {
      replace: true,
    })
    .typeText("#val-NodeSelector-0", "monitoringNodeSelectorVal0Test", {
      replace: true,
    })
    .click("#add-NodeSelector-0")
    .expect(Selector("#key-Labels-1").exists)
    .ok()
    .expect(Selector("#key-Annotations-1").exists)
    .ok()
    .expect(Selector("#key-NodeSelector-1").exists)
    .ok()
    .click("#submit_button")
    .click("#yaml_button")
    .expect(Selector("#code_wrapper").exists)
    .ok();
  await t
    .expect(
      (
        await Selector("#code_wrapper").textContent
      ).includes("image: quay.io/prometheus/prometheus:latest")
    )
    .ok()
    .expect(
      (
        await Selector("#code_wrapper").textContent
      ).includes("initimage: library/busybox:1.33.1")
    )
    .ok()
    .expect(
      (
        await Selector("#code_wrapper").textContent
      ).includes("diskCapacityGB: 1")
    )
    .ok()
    .expect((await Selector("#code_wrapper").textContent).includes('cpu: "1"'))
    .ok()
    .expect(
      (await Selector("#code_wrapper").textContent).includes("memory: 1Gi")
    )
    .ok()
    .expect(
      (
        await Selector("#code_wrapper").textContent
      ).includes("serviceAccountName: monitoringTestServiceAccountName")
    )
    .ok()
    .expect(
      (
        await Selector("#code_wrapper").textContent
      ).includes("sidecarimage: library/alpine:latest")
    )
    .ok()
    .expect(
      (
        await Selector("#code_wrapper").textContent
      ).includes("storageClassName: monitoringTestStorageClassName")
    )
    .ok()
    .expect(
      (await Selector("#code_wrapper").textContent).includes("fsGroup: 5656")
    )
    .ok()
    .expect(
      (await Selector("#code_wrapper").textContent).includes("runAsGroup: 3434")
    )
    .ok()
    .expect(
      (await Selector("#code_wrapper").textContent).includes("runAsUser: 1212")
    )
    .ok()
    .expect(
      (
        await Selector("#code_wrapper").textContent
      ).includes("monitoringAnnotationsKey0Test: monitoringAnnotationsVal0Test")
    )
    .ok()
    .expect(
      (
        await Selector("#code_wrapper").textContent
      ).includes(
        "monitoringNodeSelectorKey0Test: monitoringNodeSelectorVal0Test"
      )
    )
    .ok()
    .expect(
      (
        await Selector("#code_wrapper").textContent
      ).includes("monitoringLabelKey0Test: monitoringLabelVal0Test")
    )
    .ok();
};

const checkLoggingToggle = async (tenantName: string) => {
  await goToLoggingSection(tenantName);
  await t
    .click("#tenant_logging")
    .click("#confirm-ok")
    .wait(3000)
    .expect(Selector("#image").exists)
    .notOk()
    .click("#yaml_button")
    .expect(Selector("#code_wrapper").exists)
    .ok();
  await t
    .expect((await Selector("#code_wrapper").textContent).includes("log:"))
    .notOk();
  await t
    .click(Selector(`a[href$="/logging"]`))
    .click("#tenant_logging")
    .click("#confirm-ok")
    .wait(3000)
    .expect(Selector("#image").exists)
    .ok()
    .click("#yaml_button")
    .expect(Selector("#code_wrapper").exists)
    .ok();
  await t
    .expect((await Selector("#code_wrapper").textContent).includes("log:"))
    .ok();
};

const checkLoggingFieldsAcceptValues = async (tenantName: string) => {
  await goToLoggingSection(tenantName);
  await t
    .wait(3000)
    .typeText("#image", "minio/operator:v4.4.22", { replace: true })
    .typeText("#diskCapacityGB", "3", { replace: true })
    .typeText("#cpuRequest", "3", { replace: true })
    .typeText("#memRequest", "3", { replace: true })
    .typeText("#serviceAccountName", "loggingTestServiceAccountName", {
      replace: true,
    })
    .typeText("#securityContext_runAsUser", "1111", { replace: true })
    .typeText("#securityContext_runAsGroup", "2222", { replace: true })
    .typeText("#securityContext_fsGroup", "3333", { replace: true })
    .expect(Selector("#securityContext_runAsNonRoot").checked)
    .notOk()
    .click("#securityContext_runAsNonRoot")
    .expect(Selector("#securityContext_runAsNonRoot").checked)
    .ok()
    .typeText("#key-Labels-0", "loggingLabelKey0Test", { replace: true })
    .typeText("#val-Labels-0", "loggingLabelVal0Test", { replace: true })
    .click("#add-Labels-0")
    .typeText("#key-Annotations-0", "loggingAnnotationsKey0Test", {
      replace: true,
    })
    .typeText("#val-Annotations-0", "loggingAnnotationsVal0Test", {
      replace: true,
    })
    .click("#add-Annotations-0")
    .typeText("#key-NodeSelector-0", "loggingNodeSelectorKey0Test", {
      replace: true,
    })
    .typeText("#val-NodeSelector-0", "loggingNodeSelectorVal0Test", {
      replace: true,
    })
    .click("#add-NodeSelector-0")
    .expect(Selector("#key-Labels-1").exists)
    .ok()
    .expect(Selector("#key-Annotations-1").exists)
    .ok()
    .expect(Selector("#key-NodeSelector-1").exists)
    .ok()
    .click("#submit_button")
    .click("#yaml_button")
    .expect(Selector("#code_wrapper").exists)
    .ok();
  await t
    .expect(
      (
        await Selector("#code_wrapper").textContent
      ).includes("image: minio/operator:v4.4.22")
    )
    .ok()
    .expect(
      (
        await Selector("#code_wrapper").textContent
      ).includes("diskCapacityGB: 3")
    )
    .ok()
    .expect((await Selector("#code_wrapper").textContent).includes('cpu: "3"'))
    .ok()
    .expect(
      (await Selector("#code_wrapper").textContent).includes('memory: "3"')
    )
    .ok()
    .expect(
      (
        await Selector("#code_wrapper").textContent
      ).includes("serviceAccountName: loggingTestServiceAccountName")
    )
    .ok()
    .expect(
      (await Selector("#code_wrapper").textContent).includes("fsGroup: 3333")
    )
    .ok()
    .expect(
      (await Selector("#code_wrapper").textContent).includes("runAsGroup: 2222")
    )
    .ok()
    .expect(
      (await Selector("#code_wrapper").textContent).includes("runAsUser: 1111")
    )
    .ok()
    .expect(
      (
        await Selector("#code_wrapper").textContent
      ).includes("AnnotationsKey0Test: loggingAnnotationsVal0Test")
    )
    .ok()
    .expect(
      (
        await Selector("#code_wrapper").textContent
      ).includes("NodeSelectorKey0Test: loggingNodeSelectorVal0Test")
    )
    .ok()
    .expect(
      (
        await Selector("#code_wrapper").textContent
      ).includes("loggingLabelKey0Test: loggingLabelVal0Test")
    )
    .ok();
};

const checkLoggingDBFieldsAcceptValues = async (tenantName: string) => {
  await goToLoggingDBSection(tenantName);
  await t
    .typeText("#dbImage", "library/postgres:13", { replace: true })
    .typeText("#dbInitImage", "library/busybox:1.33.1", { replace: true })
    .typeText("#dbCPURequest", "4", { replace: true })
    .typeText("#dbMemRequest", "4", { replace: true })
    .typeText("#securityContext_runAsUser", "4444", { replace: true })
    .typeText("#securityContext_runAsGroup", "5555", { replace: true })
    .typeText("#securityContext_fsGroup", "6666", { replace: true })
    .expect(Selector("#securityContext_runAsNonRoot").checked)
    .notOk()
    .click("#securityContext_runAsNonRoot")
    .expect(Selector("#securityContext_runAsNonRoot").checked)
    .ok()
    .typeText("#key-dbLabels-0", "loggingdbLabelKey0Test", { replace: true })
    .typeText("#val-dbLabels-0", "loggingdbLabelVal0Test", { replace: true })
    .click("#add-dbLabels-0")
    .typeText("#key-dbAnnotations-0", "loggingdbAnnotationsKey0Test", {
      replace: true,
    })
    .typeText("#val-dbAnnotations-0", "loggingdbAnnotationsVal0Test", {
      replace: true,
    })
    .click("#add-dbAnnotations-0")
    .typeText("#key-DBNodeSelector-0", "loggingdbNodeSelectorKey0Test", {
      replace: true,
    })
    .typeText("#val-DBNodeSelector-0", "loggingdbNodeSelectorVal0Test", {
      replace: true,
    })
    .click("#add-DBNodeSelector-0")
    .expect(Selector("#key-dbLabels-1").exists)
    .ok()
    .expect(Selector("#key-dbAnnotations-1").exists)
    .ok()
    .expect(Selector("#key-DBNodeSelector-1").exists)
    .ok()
    .click("#remove-dbLabels-1")
    .click("#remove-dbAnnotations-1")
    .click("#remove-DBNodeSelector-1")
    .expect(Selector("#key-dbLabels-1").exists)
    .notOk()
    .expect(Selector("#key-dbAnnotations-1").exists)
    .notOk()
    .expect(Selector("#key-DBNodeSelector-1").exists)
    .notOk()
    .click("#submit_button")
    .click("#yaml_button")
    .expect(Selector("#code_wrapper").exists)
    .ok();
  await t
    .expect(
      (
        await Selector("#code_wrapper").textContent
      ).includes("image: library/postgres:13")
    )
    .ok()
    .expect((await Selector("#code_wrapper").textContent).includes('cpu: "4"'))
    .ok()
    .expect(
      (await Selector("#code_wrapper").textContent).includes('memory: "4"')
    )
    .ok()
    .expect(
      (await Selector("#code_wrapper").textContent).includes("fsGroup: 6666")
    )
    .ok()
    .expect(
      (await Selector("#code_wrapper").textContent).includes("runAsGroup: 5555")
    )
    .ok()
    .expect(
      (await Selector("#code_wrapper").textContent).includes("runAsUser: 4444")
    )
    .ok()
    .expect(
      (
        await Selector("#code_wrapper").textContent
      ).includes("loggingdbAnnotationsKey0Test: loggingdbAnnotationsVal0Test")
    )
    .ok()
    .expect(
      (
        await Selector("#code_wrapper").textContent
      ).includes("loggingdbNodeSelectorKey0Test: loggingdbNodeSelectorVal0Test")
    )
    .ok()
    .expect(
      (
        await Selector("#code_wrapper").textContent
      ).includes("loggingdbLabelKey0Test: loggingdbLabelVal0Test")
    )
    .ok();
};

// Test 5
test("Test Prometheus monitoring can be disabled and enabled", async (t) => {
  const tenantName = `storage-lite`;
  await loginToOperator();
  await checkMonitoringToggle(tenantName);
});
