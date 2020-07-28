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

package server

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type tenantJWTClaims struct {
	jwt.StandardClaims
}

func GenerateJWTForTenant(tenant string, operator string, key interface{}) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tenantJWTClaims{
		StandardClaims: jwt.StandardClaims{
			Audience: "operator",
			Id:       tenant + "-" + operator, // Unique ID for token
			IssuedAt: time.Now().Unix(),
			Issuer:   operator,
			Subject:  tenant, // Principal that is subject of the JWT
		},
	})
	return token.SignedString(key)
}

func DecodeJWTGetTenant(token string, key interface{}) (string, error) {
	claims := tenantJWTClaims{}
	_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	return claims.Subject, err
}
