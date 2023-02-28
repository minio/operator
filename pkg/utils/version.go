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

package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/minio/operator/pkg/http"
)

// ErrCantDetermineMinIOImage when not able to find MinIO version
var ErrCantDetermineMinIOImage = errors.New("can't determine MinIO Image")

// GetLatestMinIOImage returns the latest docker image for MinIO if found on the internet
func GetLatestMinIOImage(client http.ClientI) (*string, error) {
	resp, err := client.Get("https://dl.min.io/server/minio/release/linux-amd64/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`minio\.(RELEASE.*?Z)"`)
	// look for a single match
	matches := re.FindAllStringSubmatch(string(body), 1)
	for i := range matches {
		release := matches[i][1]
		dockerImage := fmt.Sprintf("minio/minio:%s", release)
		return &dockerImage, nil
	}
	return nil, ErrCantDetermineMinIOImage
}
