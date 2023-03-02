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

package utils

import (
	"crypto/sha1"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/pbkdf2"
)

func TestRandomCharString(t *testing.T) {
	funcAssert := assert.New(t)
	// Test-1 : RandomCharString() should return string with expected length
	length := 32
	token := RandomCharString(length)
	funcAssert.Equal(length, len(token))
	// Test-2 : RandomCharString() should output random string, new generated string should not be equal to the previous one
	newToken := RandomCharString(length)
	funcAssert.NotEqual(token, newToken)
}

func TestComputeHmac256(t *testing.T) {
	funcAssert := assert.New(t)
	// Test-1 : ComputeHmac256() should return the right Hmac256 string based on a derived key
	derivedKey := pbkdf2.Key([]byte("secret"), []byte("salt"), 4096, 32, sha1.New)
	message := "hello world"
	expectedHmac := "5r32q7W+0hcBnqzQwJJUDzVGoVivXGSodTcHSqG/9Q8="
	hmac := ComputeHmac256(message, derivedKey)
	funcAssert.Equal(hmac, expectedHmac)
}
