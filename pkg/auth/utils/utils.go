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
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"strings"
)

// Do not use:
// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
// It relies on math/rand and therefore not on a cryptographically secure RNG => It must not be used
// for access/secret keys.

// The alphabet of random character string. Each character must be unique.
//
// The RandomCharString implementation requires that: 256 / len(letters) is a natural numbers.
// For example: 256 / 64 = 4. However, 5 > 256/62 > 4 and therefore we must not use a alphabet
// of 62 characters.
// The reason is that if 256 / len(letters) is not a natural number then certain characters become
// more likely then others.
const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ012345"

// RandomCharString returns a random string of characters of n size
func RandomCharString(n int) string {
	random := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, random); err != nil {
		panic(err) // Can only happen if we would run out of entropy.
	}

	var s strings.Builder
	for _, v := range random {
		j := v % byte(len(letters))
		s.WriteByte(letters[j])
	}
	return s.String()
}

// ComputeHmac256 computes an Hmac
func ComputeHmac256(message string, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
