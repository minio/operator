// This file is part of MinIO Operator
// Copyright (C) 2020, MinIO, Inc.
//
// This code is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License, version 3,
// as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License, version 3,
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package resources

// OperatorOptions encapsulates the CLI options for a MinIO Operator
type OperatorOptions struct {
	Name                string
	Image               string
	Namespace           string
	NSToWatch           string
	ClusterDomain       string
	ImagePullSecret     string
	ConsoleImage        string
	ConsoleTLS          bool
	TenantMinIOImage    string
	TenantConsoleImage  string
	TenantKesImage      string
	PrometheusNamespace string
	PrometheusName      string
	STS                 bool
}
