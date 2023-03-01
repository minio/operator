// This file is part of MinIO Operator
// Copyright (c) 2023 MinIO, Inc.
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

package api

import (
	"strings"
	"time"

	xhttp "github.com/minio/operator/pkg/http"
	"github.com/minio/operator/pkg/utils"
	"github.com/minio/pkg/env"
)

// GetMinioImage returns the image URL to be used when deploying a MinIO instance, if there is
// a preferred image to be used (configured via ENVIRONMENT VARIABLES) GetMinioImage will return that
// if not, GetMinioImage will try to obtain the image URL for the latest version of MinIO and return that
func GetMinioImage() (*string, error) {
	image := strings.TrimSpace(env.Get(MinioImage, ""))
	// if there is a preferred image configured by the user we'll always return that
	if image != "" {
		return &image, nil
	}
	client := GetConsoleHTTPClient("")
	client.Timeout = 5 * time.Second
	latestMinIOImage, errLatestMinIOImage := utils.GetLatestMinIOImage(
		&xhttp.Client{
			Client: client,
		})

	if errLatestMinIOImage != nil {
		return nil, errLatestMinIOImage
	}
	return latestMinIOImage, nil
}
