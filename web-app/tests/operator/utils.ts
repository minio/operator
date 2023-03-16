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

const host: string = "http://localhost:9090";

export const loginToOperator = async () => {
  await t
    .navigateTo(`${host}/login`)
    .typeText("#jwt", "anyrandompasswordwillwork")
    .click("#do-login");
};

export const createTenant = async (tenantName: string) => {
  await fillTenantInformation(tenantName);
  await t.click("#wizard-button-Create");
  await checkTenantExists(tenantName);
};

const fillTenantInformation = async (tenantName: string) => {
  await t
    .click("#create-tenant")
    .typeText("#tenant-name", tenantName)
    .typeText("#namespace", tenantName)
    .click("#add-namespace")
    .click("#confirm-ok")
    .wait(1000);
};

const checkTenantExists = async (tenantName: string) => {
  await t
    .wait(1000)
    .click("#close")
    .expect(Selector(`#list-tenant-${tenantName}`).exists)
    .ok();
};

export const deleteTenant = async (tenantName: string) => {
  await goToTenant(tenantName);
  await t
    .click("#delete-tenant")
    .typeText("#retype-tenant", tenantName)
    .click("#confirm-ok")
    .expect(Selector(`#list-tenant-${tenantName}`).exists)
    .notOk();
};

export const goToTenant = async (tenantName: string) => {
  await t.click(Selector(`#list-tenant-${tenantName}`));
};

export const goToVolumesInTenant = async (tenantName: string) => {
  const path: string = `${host}/namespaces/${tenantName}/tenants/${tenantName}/volumes`;
  await redirectToPath(path);
};

export const goToPodsInTenant = async (tenantName: string) => {
  await t.click(`#list-tenant-${tenantName}`).wait(2000);
  await t.click(Selector(`a[href$="/pods"]`));
};

export const goToPodInTenant = async (tenantName: string) => {
  await goToPodsInTenant(tenantName);
  await t.click(Selector("div.ReactVirtualized__Table__row").child(0));
};

export const goToPodSection = async (index: number) => {
  await t
    .expect(Selector(`#simple-tab-${index}`).exists)
    .ok()
    .click(Selector(`#simple-tab-${index}`));
};

export const goToPvcsInTenant = async (tenantName: string) => {
  await t.click(`#list-tenant-${tenantName}`).wait(2000);
  await t.click(Selector(`a[href$="/volumes"]`));
};

export const goToPvcInTenant = async (tenantName: string) => {
  await goToPvcsInTenant(tenantName);
  await t.click(Selector("div.ReactVirtualized__Table__row").child(0));
};

export const goToPvcSection = async (index: number) => {
  await t
    .expect(Selector(`#simple-tab-${index}`).exists)
    .ok()
    .click(Selector(`#simple-tab-${index}`));
};

export const goToMonitoringSection = async (tenantName: string) => {
  await t.click(`#list-tenant-${tenantName}`).wait(2000);
  await t.click(Selector(`a[href$="/monitoring"]`));
};
export const goToLoggingSection = async (tenantName: string) => {
  await t.click(`#list-tenant-${tenantName}`).wait(2000);
  await t.click(Selector(`a[href$="/logging"]`));
};
export const goToLoggingDBSection = async (tenantName: string) => {
  await t.click(Selector(`a[href$="/logging"]`));
  await t.click(Selector("#simple-tab-1"));
};

export const redirectToTenantsList = async () => {
  await redirectToPath(`${host}/tenants`);
};

export const redirectToPath = async (path: string) => {
  await t.navigateTo(path);
};
