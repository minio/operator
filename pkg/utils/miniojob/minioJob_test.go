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

import "testing"

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
			args:        args,
			expect:      Arg{Command: "--with-locks --region=us-west-2"},
			expectError: false,
		},
		{
			command:     FLAGS("user"),
			args:        args,
			expect:      Arg{Command: "--with-locks --region=us-west-2"},
			expectError: false,
		},
		{
			command:     Key("password"),
			args:        args,
			expect:      Arg{Command: "somepassword"},
			expectError: false,
		},
		{
			command:     KeyForamt("user", "--user $0"),
			args:        args,
			expect:      Arg{Command: "--user a b c d"},
			expectError: false,
		},
		{
			command:     KeyForamt("user", "--user"),
			args:        args,
			expect:      Arg{Command: "--user a b c d"},
			expectError: false,
		},
		{
			command:     ALIAS(),
			args:        args,
			expect:      Arg{Command: "myminio"},
			expectError: false,
		},
		{
			command:     Static("test-static"),
			args:        args,
			expect:      Arg{Command: "test-static"},
			expectError: false,
		},
		{
			command: File("policy", "json"),
			args:    args,
			expect: Arg{
				FileName: "policy",
				FileExt:  "json",
				FileContext: `{
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
			},
			expectError: false,
		},
		{
			command:     OneOf(KeyForamt("user", "--user"), KeyForamt("group", "--group")),
			args:        args,
			expect:      Arg{Command: "--user a b c d"},
			expectError: false,
		},
		{
			command:     OneOf(KeyForamt("miss_user", "--user"), KeyForamt("group", "--group")),
			args:        args,
			expect:      Arg{Command: "--group group1 group2 group3"},
			expectError: false,
		},
		{
			command:     OneOf(KeyForamt("miss_user", "--user"), KeyForamt("miss_group", "--group")),
			args:        args,
			expect:      Arg{Command: "--group group1 group2 group3"},
			expectError: true,
		},
		{
			command:     Sanitize(ALIAS(), Static("/"), Key("name")),
			args:        args,
			expect:      Arg{Command: "myminio/mybucketName"},
			expectError: false,
		},
	}
	for _, tc := range testCase {
		cmd, err := tc.command(args)
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
			if tc.expect.FileName != "" {
				if tc.expect.FileContext != cmd.FileContext {
					t.Fatalf("expectCommand %s, but got %s", tc.expect.FileContext, cmd.FileContext)
				}
				if tc.expect.FileExt != cmd.FileExt {
					t.Fatalf("expectCommand %s, but got %s", tc.expect.FileExt, cmd.FileExt)
				}
				if tc.expect.FileName != cmd.FileName {
					t.Fatalf("expectCommand %s, but got %s", tc.expect.FileName, cmd.FileName)
				}
			}
		}
	}
}

func TestAdminPolicyCreate(t *testing.T) {
	mcCommand := "admin/policy/create"
	funcs := JobOperation[mcCommand]
	testCase := []struct {
		name             string
		args             map[string]string
		expectError      bool
		expectCommand    string
		expectFileNumber int
	}{
		{
			name: "testFull",
			args: map[string]string{
				"name":   "mypolicy",
				"policy": "JsonContent",
			},
			expectCommand:    "myminio mypolicy /temp/policy.json",
			expectFileNumber: 1,
		},
		{
			name: "testError1",
			args: map[string]string{
				"name": "mypolicy",
			},
			expectCommand: "",
			expectError:   true,
		},
		{
			name: "testError2",
			args: map[string]string{
				"policy": "JsonContent",
			},
			expectCommand: "",
			expectError:   true,
		},
	}
	for _, tc := range testCase {
		command, err := GenerateMinIOIntervalJobCommand(mcCommand, 0, nil, "test", tc.args, funcs)
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
	funcs := JobOperation[mcCommand]
	testCase := []struct {
		name             string
		args             map[string]string
		expectCommand    string
		expectError      bool
		expectFileNumber int
	}{
		{
			name: "testFull",
			args: map[string]string{
				"webhookName": "webhook1",
				"endpoint":    "endpoint1",
				"auth_token":  "token1",
				"client_cert": "cert1",
				"client_key":  "key1",
			},
			expectCommand:    "myminio webhook1 endpoint=\"endpoint1\" auth_token=\"token1\" client_key=\"/temp/client_key.key\" client_cert=\"/temp/client_cert.pem\"",
			expectFileNumber: 2,
		},
		{
			name: "testOptionFile",
			args: map[string]string{
				"webhookName": "webhook1",
				"endpoint":    "endpoint1",
				"auth_token":  "token1",
				"client_key":  "key1",
			},
			expectCommand:    "myminio webhook1 endpoint=\"endpoint1\" auth_token=\"token1\" client_key=\"/temp/client_key.key\"",
			expectFileNumber: 1,
		},
		{
			name: "testOptionKeyValue",
			args: map[string]string{
				"webhookName": "webhook1",
				"endpoint":    "endpoint1",
				"client_key":  "key1",
			},
			expectCommand:    "myminio webhook1 endpoint=\"endpoint1\" client_key=\"/temp/client_key.key\"",
			expectFileNumber: 1,
		},
	}
	for _, tc := range testCase {
		command, err := GenerateMinIOIntervalJobCommand(mcCommand, 0, nil, "test", tc.args, funcs)
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
