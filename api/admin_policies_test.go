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
	"errors"
	"testing"

	"github.com/minio/operator/models"
	"github.com/stretchr/testify/assert"
)

func TestSetPolicy(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	funcAssert := assert.New(t)
	adminClient := AdminClientMock{}
	policyName := "readOnly"
	entityName := "alevsk"
	entityObject := models.PolicyEntityUser
	minioSetPolicyMock = func(policyName, entityName string, isGroup bool) error {
		return nil
	}
	// Test-1 : SetPolicy() set policy to user
	function := "SetPolicy()"
	err := SetPolicy(ctx, adminClient, policyName, entityName, entityObject)
	if err != nil {
		t.Errorf("Failed on %s:, error occurred: %s", function, err.Error())
	}
	// Test-2 : SetPolicy() set policy to group
	entityObject = models.PolicyEntityGroup
	err = SetPolicy(ctx, adminClient, policyName, entityName, entityObject)
	if err != nil {
		t.Errorf("Failed on %s:, error occurred: %s", function, err.Error())
	}
	// Test-3 : SetPolicy() set policy to user and get error
	entityObject = models.PolicyEntityUser
	minioSetPolicyMock = func(policyName, entityName string, isGroup bool) error {
		return errors.New("error")
	}
	if err := SetPolicy(ctx, adminClient, policyName, entityName, entityObject); funcAssert.Error(err) {
		funcAssert.Equal("error", err.Error())
	}
	// Test-4 : SetPolicy() set policy to group and get error
	entityObject = models.PolicyEntityGroup
	minioSetPolicyMock = func(policyName, entityName string, isGroup bool) error {
		return errors.New("error")
	}
	if err := SetPolicy(ctx, adminClient, policyName, entityName, entityObject); funcAssert.Error(err) {
		funcAssert.Equal("error", err.Error())
	}
}
