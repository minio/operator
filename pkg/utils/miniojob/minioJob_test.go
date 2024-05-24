// This file is part of MinIO Operator
// Copyright (c) 2024 MinIO, Inc.
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

package miniojob

import (
	"testing"

	"github.com/minio/operator/pkg/apis/job.min.io/v1alpha1"
)

func TestParser(t *testing.T) {
	args := map[string]string{
		"--user":       "a1,b2,c3,d4",
		"user":         "a,b,c,d",
		"group":        "group1,group2,group3",
		"password":     "somepassword",
		"--with-locks": "",
		"--region":     "us-west-2",
		"policy": ` {
              "Version": "2012-10-17",
              "Statement": [
                  {
                      "Effect": "Allow",
                      "Action": [
                          "s3:*"
                      ],
                      "Resource": [
                          "arn:aws:s3:::memes",
                          "arn:aws:s3:::memes/*"
                      ]
                  }
              ]
          }`,
		"name": "mybucketName",
	}
	testCase := []struct {
		command     FieldsFunc
		args        map[string]string
		expect      Arg
		expectError bool
	}{
		{
			command:     FLAGS("--user"),
			args:        copyArgs(args),
			expect:      Arg{Command: "--region=us-west-2 --with-locks"},
			expectError: false,
		},
		{
			command:     FLAGS("user"),
			args:        copyArgs(args),
			expect:      Arg{Command: "--region=us-west-2 --with-locks"},
			expectError: false,
		},
		{
			command:     Key("password"),
			args:        copyArgs(args),
			expect:      Arg{Command: "somepassword"},
			expectError: false,
		},
		{
			command:     KeyFormat("user", "--user $0"),
			args:        copyArgs(args),
			expect:      Arg{Command: "--user a b c d"},
			expectError: false,
		},
		{
			command:     KeyFormat("user", "--user"),
			args:        copyArgs(args),
			expect:      Arg{Command: "--user a b c d"},
			expectError: false,
		},
		{
			command:     ALIAS(),
			args:        copyArgs(args),
			expect:      Arg{Command: "myminio"},
			expectError: false,
		},
		{
			command:     Static("test-static"),
			args:        copyArgs(args),
			expect:      Arg{Command: "test-static"},
			expectError: false,
		},
		{
			command:     OneOf(KeyFormat("user", "--user"), KeyFormat("group", "--group")),
			args:        copyArgs(args),
			expect:      Arg{Command: "--user a b c d"},
			expectError: false,
		},
		{
			command:     OneOf(KeyFormat("miss_user", "--user"), KeyFormat("group", "--group")),
			args:        copyArgs(args),
			expect:      Arg{Command: "--group group1 group2 group3"},
			expectError: false,
		},
		{
			command:     OneOf(KeyFormat("miss_user", "--user"), KeyFormat("miss_group", "--group")),
			args:        copyArgs(args),
			expect:      Arg{Command: "--group group1 group2 group3"},
			expectError: true,
		},
		{
			command:     Sanitize(ALIAS(), Static("/"), Key("name")),
			args:        copyArgs(args),
			expect:      Arg{Command: "myminio/mybucketName"},
			expectError: false,
		},
	}
	for _, tc := range testCase {
		cmd, err := tc.command(tc.args)
		if tc.expectError && err == nil {
			t.Fatalf("expectCommand error")
		}
		if !tc.expectError && err != nil {
			t.Fatalf("expectCommand not a error")
		}
		if !tc.expectError {
			if tc.expect.Command != "" && cmd.Command != tc.expect.Command {
				t.Fatalf("expectCommand %s, but got %s", tc.expect.Command, cmd.Command)
			}
		}
	}
}

func TestAdminPolicyCreate(t *testing.T) {
	mcCommand := "admin/policy/create"
	testCase := []struct {
		name          string
		spec          v1alpha1.CommandSpec
		expectError   bool
		expectCommand string
	}{
		{
			name: "testFull",
			spec: v1alpha1.CommandSpec{
				Operation: mcCommand,
				Args: map[string]string{
					"name":   "mypolicy",
					"policy": "JsonContent",
				},
			},
			expectCommand: "myminio mypolicy JsonContent",
		},
		{
			name: "testError1",
			spec: v1alpha1.CommandSpec{
				Operation: mcCommand,
				Args: map[string]string{
					"name": "mypolicy",
				},
			},
			expectCommand: "",
			expectError:   true,
		},
		{
			name: "testError2",
			spec: v1alpha1.CommandSpec{
				Operation: mcCommand,
				Args: map[string]string{
					"policy": "JsonContent",
				},
			},
			expectCommand: "",
			expectError:   true,
		},
	}
	for _, tc := range testCase {
		command, err := GenerateMinIOIntervalJobCommand(tc.spec, 0)
		if !tc.expectError {
			if err != nil {
				t.Fatal(err)
			}
			if command.Command != tc.expectCommand {
				t.Fatalf("[%s] expectCommand %s, but got %s", tc.name, tc.expectCommand, command.Command)
			}
		} else {
			if err == nil {
				t.Fatalf("[%s] expectCommand error", tc.name)
			}
		}
	}
}

func TestMCConfigSet(t *testing.T) {
	mcCommand := "admin/config/set"
	testCase := []struct {
		name          string
		spec          v1alpha1.CommandSpec
		expectCommand string
		expectError   bool
	}{
		{
			name: "testFull",
			spec: v1alpha1.CommandSpec{
				Operation: mcCommand,
				Args: map[string]string{
					"webhookName": "webhook1",
					"endpoint":    "endpoint1",
					"auth_token":  "token1",
					"client_cert": "cert1",
					"client_key":  "key1",
				},
			},
			expectCommand: "myminio webhook1 endpoint=\"endpoint1\" auth_token=\"token1\" client_cert=\"cert1\" client_key=\"key1\"",
		},
		{
			name: "notify_mysql",
			spec: v1alpha1.CommandSpec{
				Operation: mcCommand,
				Args: map[string]string{
					"webhookName": "notify_mysql",
					"dsn_string":  "username:password@tcp(mysql.example.com:3306)/miniodb",
					"table":       "minioevents",
					"format":      "namespace",
				},
			},
			expectCommand: "myminio notify_mysql dsn_string=\"username:password@tcp(mysql.example.com:3306)/miniodb\" format=\"namespace\" table=\"minioevents\"",
		},
		{
			name: "notify_amqp",
			spec: v1alpha1.CommandSpec{
				Operation: mcCommand,
				Args: map[string]string{
					"webhookName": "notify_amqp:primary",
					"url":         "user:password@amqp://amqp-endpoint.example.net:5672",
				},
			},
			expectCommand: "myminio notify_amqp:primary url=\"user:password@amqp://amqp-endpoint.example.net:5672\"",
		},
		{
			name: "notify_elasticsearch",
			spec: v1alpha1.CommandSpec{
				Operation: mcCommand,
				Args: map[string]string{
					"webhookName": "notify_elasticsearch:primary",
					"url":         "user:password@https://elasticsearch-endpoint.example.net:9200",
					"index":       "bucketevents",
					"format":      "namespace",
				},
			},
			expectCommand: "myminio notify_elasticsearch:primary format=\"namespace\" index=\"bucketevents\" url=\"user:password@https://elasticsearch-endpoint.example.net:9200\"",
		},
		{
			name: "identity_ldap",
			spec: v1alpha1.CommandSpec{
				Operation: mcCommand,
				Args: map[string]string{
					"webhookName":             "identity_ldap",
					"enabled":                 "true",
					"server_addr":             "ad-ldap.example.net/",
					"lookup_bind_dn":          "cn=miniolookupuser,dc=example,dc=net",
					"lookup_bind_dn_password": "userpassword",
					"user_dn_search_base_dn":  "dc=example,dc=net",
					"user_dn_search_filter":   "(&(objectCategory=user)(sAMAccountName=%s))",
				},
			},
			expectCommand: "myminio identity_ldap enabled=\"true\" lookup_bind_dn=\"cn=miniolookupuser,dc=example,dc=net\" lookup_bind_dn_password=\"userpassword\" server_addr=\"ad-ldap.example.net/\" user_dn_search_base_dn=\"dc=example,dc=net\" user_dn_search_filter=\"(&(objectCategory=user)(sAMAccountName=%s))\"",
		},
	}
	for _, tc := range testCase {
		command, err := GenerateMinIOIntervalJobCommand(tc.spec, 0)
		if !tc.expectError {
			if err != nil {
				t.Fatal(err)
			}
			if command.Command != tc.expectCommand {
				t.Fatalf("[%s] expectCommand %s, but got %s", tc.name, tc.expectCommand, command.Command)
			}
		} else {
			if err == nil {
				t.Fatalf("[%s] expectCommand error", tc.name)
			}
		}
	}
}

func TestSupportcallhome(t *testing.T) {
	mcCommand := "support/callhome"
	testCase := []struct {
		name          string
		spec          v1alpha1.CommandSpec
		expectCommand string
		expectError   bool
	}{
		{
			name: "testEnable",
			spec: v1alpha1.CommandSpec{
				Operation: mcCommand,
				Args: map[string]string{
					"action": "enable",
					"--logs": "",
					"--diag": "",
				},
			},
			expectCommand: "enable myminio --diag --logs",
		},
		{
			name: "testDisable",
			spec: v1alpha1.CommandSpec{
				Operation: mcCommand,
				Args: map[string]string{
					"action": "disable",
					"--logs": "",
					"--diag": "",
				},
			},
			expectCommand: "disable myminio --diag --logs",
		},
		{
			name: "testNoAction",
			spec: v1alpha1.CommandSpec{
				Operation: mcCommand,
				Args: map[string]string{
					"--logs": "",
					"--diag": "",
				},
			},
			expectCommand: "",
			expectError:   true,
		},
	}
	for _, tc := range testCase {
		command, err := GenerateMinIOIntervalJobCommand(tc.spec, 0)
		if !tc.expectError {
			if err != nil {
				t.Fatal(err)
			}
			if command.Command != tc.expectCommand {
				t.Fatalf("[%s] expectCommand %s, but got %s", tc.name, tc.expectCommand, command.Command)
			}
		} else {
			if err == nil {
				t.Fatalf("[%s] expectCommand error", tc.name)
			}
		}
	}
}

func copyArgs(args map[string]string) map[string]string {
	newArgs := make(map[string]string)
	for key, val := range args {
		newArgs[key] = val
	}
	return newArgs
}
