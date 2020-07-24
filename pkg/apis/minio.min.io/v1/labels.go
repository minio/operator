/*
 * Copyright (C) 2020, MinIO, Inc.
 *
 * This code is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, version 3,
 * as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License, version 3,
 * along with this program.  If not, see <http://www.gnu.org/licenses/>
 *
 */

package v1

// MinIOPodLabels returns the default labels for MinIO Pod
func (t *Tenant) MinIOPodLabels() map[string]string {
	m := make(map[string]string, 1)
	m[TenantLabel] = t.Name
	return m
}

// KESPodLabels returns the default labels for KES Pod
func (t *Tenant) KESPodLabels() map[string]string {
	m := make(map[string]string, 1)
	m[KESInstanceLabel] = t.KESStatefulSetName()
	return m
}

// ConsolePodLabels returns the default labels for Console Pod
func (t *Tenant) ConsolePodLabels() map[string]string {
	m := make(map[string]string, 1)
	m[ConsoleTenantLabel] = t.ConsoleDeploymentName()
	return m
}
