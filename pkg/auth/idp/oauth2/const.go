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

package oauth2

// Environment constants for console IDP/SSO configuration
const (
	ConsoleMinIOServer           = "CONSOLE_MINIO_SERVER"
	ConsoleIDPURL                = "CONSOLE_IDP_URL"
	ConsoleIDPClientID           = "CONSOLE_IDP_CLIENT_ID"
	ConsoleIDPSecret             = "CONSOLE_IDP_SECRET"
	ConsoleIDPCallbackURL        = "CONSOLE_IDP_CALLBACK"
	ConsoleIDPCallbackURLDynamic = "CONSOLE_IDP_CALLBACK_DYNAMIC"
	ConsoleIDPHmacPassphrase     = "CONSOLE_IDP_HMAC_PASSPHRASE"
	ConsoleIDPHmacSalt           = "CONSOLE_IDP_HMAC_SALT"
	ConsoleIDPScopes             = "CONSOLE_IDP_SCOPES"
	ConsoleIDPUserInfo           = "CONSOLE_IDP_USERINFO"
	ConsoleIDPTokenExpiration    = "CONSOLE_IDP_TOKEN_EXPIRATION"
)
