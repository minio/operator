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

export interface SubnetInfo {
  account_id: number;
  email: string;
  expires_at: string;
  plan: string;
  storage_capacity: number;
  organization: string;
}

export interface SubnetLoginRequest {
  username?: string;
  password?: string;
  apiKey?: string;
  proxy?: string;
}

export interface SubnetOrganization {
  userId: number;
  accountId: number;
  subscriptionStatus: string;
  isAccountOwner: boolean;
  shortName: string;
  company: string;
}

export interface SubnetLoginResponse {
  registered: boolean;
  mfa_token: string;
  access_token: string;
  organizations: SubnetOrganization[];
}
