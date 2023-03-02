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

package api

import (
	"context"
	"io"
	"time"

	mc "github.com/minio/mc/cmd"
	"github.com/minio/mc/pkg/probe"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/lifecycle"
	"github.com/minio/minio-go/v7/pkg/notification"
	"github.com/minio/minio-go/v7/pkg/sse"
	"github.com/minio/minio-go/v7/pkg/tags"
)

func init() {
	// All minio-go API operations shall be performed only once,
	// another way to look at this is we are turning off retries.
	minio.MaxRetry = 1
}

// MinioClient interface with all functions to be implemented
// by mock when testing, it should include all MinioClient respective api calls
// that are used within this project.
type MinioClient interface {
	listBucketsWithContext(ctx context.Context) ([]minio.BucketInfo, error)
	makeBucketWithContext(ctx context.Context, bucketName, location string, objectLocking bool) error
	setBucketPolicyWithContext(ctx context.Context, bucketName, policy string) error
	removeBucket(ctx context.Context, bucketName string) error
	getBucketNotification(ctx context.Context, bucketName string) (config notification.Configuration, err error)
	getBucketPolicy(ctx context.Context, bucketName string) (string, error)
	listObjects(ctx context.Context, bucket string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo
	getObjectRetention(ctx context.Context, bucketName, objectName, versionID string) (mode *minio.RetentionMode, retainUntilDate *time.Time, err error)
	getObjectLegalHold(ctx context.Context, bucketName, objectName string, opts minio.GetObjectLegalHoldOptions) (status *minio.LegalHoldStatus, err error)
	putObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (info minio.UploadInfo, err error)
	putObjectLegalHold(ctx context.Context, bucketName, objectName string, opts minio.PutObjectLegalHoldOptions) error
	putObjectRetention(ctx context.Context, bucketName, objectName string, opts minio.PutObjectRetentionOptions) error
	statObject(ctx context.Context, bucketName, prefix string, opts minio.GetObjectOptions) (objectInfo minio.ObjectInfo, err error)
	setBucketEncryption(ctx context.Context, bucketName string, config *sse.Configuration) error
	removeBucketEncryption(ctx context.Context, bucketName string) error
	getBucketEncryption(ctx context.Context, bucketName string) (*sse.Configuration, error)
	putObjectTagging(ctx context.Context, bucketName, objectName string, otags *tags.Tags, opts minio.PutObjectTaggingOptions) error
	getObjectTagging(ctx context.Context, bucketName, objectName string, opts minio.GetObjectTaggingOptions) (*tags.Tags, error)
	setObjectLockConfig(ctx context.Context, bucketName string, mode *minio.RetentionMode, validity *uint, unit *minio.ValidityUnit) error
	getBucketObjectLockConfig(ctx context.Context, bucketName string) (mode *minio.RetentionMode, validity *uint, unit *minio.ValidityUnit, err error)
	getObjectLockConfig(ctx context.Context, bucketName string) (lock string, mode *minio.RetentionMode, validity *uint, unit *minio.ValidityUnit, err error)
	getLifecycleRules(ctx context.Context, bucketName string) (lifecycle *lifecycle.Configuration, err error)
	setBucketLifecycle(ctx context.Context, bucketName string, config *lifecycle.Configuration) error
	copyObject(ctx context.Context, dst minio.CopyDestOptions, src minio.CopySrcOptions) (minio.UploadInfo, error)
	GetBucketTagging(ctx context.Context, bucketName string) (*tags.Tags, error)
	SetBucketTagging(ctx context.Context, bucketName string, tags *tags.Tags) error
	RemoveBucketTagging(ctx context.Context, bucketName string) error
}

// MCClient interface with all functions to be implemented
// by mock when testing, it should include all mc/S3Client respective api calls
// that are used within this project.
type MCClient interface {
	addNotificationConfig(ctx context.Context, arn string, events []string, prefix, suffix string, ignoreExisting bool) *probe.Error
	removeNotificationConfig(ctx context.Context, arn string, event string, prefix string, suffix string) *probe.Error
	watch(ctx context.Context, options mc.WatchOptions) (*mc.WatchObject, *probe.Error)
	remove(ctx context.Context, isIncomplete, isRemoveBucket, isBypass, forceDelete bool, contentCh <-chan *mc.ClientContent) <-chan mc.RemoveResult
	list(ctx context.Context, opts mc.ListOptions) <-chan *mc.ClientContent
	get(ctx context.Context, opts mc.GetOptions) (io.ReadCloser, *probe.Error)
	shareDownload(ctx context.Context, versionID string, expires time.Duration) (string, *probe.Error)
	setVersioning(ctx context.Context, status string) *probe.Error
}

// ConsoleCredentialsI interface with all functions to be implemented
// by mock when testing, it should include all needed consoleCredentials.Login api calls
// that are used within this project.
type ConsoleCredentialsI interface {
	Get() (credentials.Value, error)
	Expire()
	GetAccountAccessKey() string
}

// ConsoleCredentials Interface implementation
type ConsoleCredentials struct {
	ConsoleCredentials *credentials.Credentials
	AccountAccessKey   string
}

// GetAccountAccessKey implementation
func (c ConsoleCredentials) GetAccountAccessKey() string {
	return c.AccountAccessKey
}

// Get implements *Login.Get()
func (c ConsoleCredentials) Get() (credentials.Value, error) {
	return c.ConsoleCredentials.Get()
}

// Expire implements *Login.Expire()
func (c ConsoleCredentials) Expire() {
	c.ConsoleCredentials.Expire()
}
