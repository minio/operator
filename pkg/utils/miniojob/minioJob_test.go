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
		command     FiledsFunc
		args        map[string]string
		expect      MinioJobArg
		expectError bool
	}{
		{
			command:     FLAGS("--user"),
			args:        args,
			expect:      MinioJobArg{Command: "--with-locks --region=us-west-2"},
			expectError: false,
		},
		{
			command:     FLAGS("user"),
			args:        args,
			expect:      MinioJobArg{Command: "--with-locks --region=us-west-2"},
			expectError: false,
		},
		{
			command:     Key("password"),
			args:        args,
			expect:      MinioJobArg{Command: "somepassword"},
			expectError: false,
		},
		{
			command:     KeyForamt("user", "--user $0"),
			args:        args,
			expect:      MinioJobArg{Command: "--user a b c d"},
			expectError: false,
		},
		{
			command:     KeyForamt("user", "--user"),
			args:        args,
			expect:      MinioJobArg{Command: "--user a b c d"},
			expectError: false,
		},
		{
			command:     ALIAS(),
			args:        args,
			expect:      MinioJobArg{Command: "myminio"},
			expectError: false,
		},
		{
			command:     Static("test-static"),
			args:        args,
			expect:      MinioJobArg{Command: "test-static"},
			expectError: false,
		},
		{
			command: File("policy", "json"),
			args:    args,
			expect: MinioJobArg{FileName: "policy",
				FileExt: "json",
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
          }`},
			expectError: false,
		},
		{
			command:     OneOf(KeyForamt("user", "--user"), KeyForamt("group", "--group")),
			args:        args,
			expect:      MinioJobArg{Command: "--user a b c d"},
			expectError: false,
		},
		{
			command:     OneOf(KeyForamt("miss_user", "--user"), KeyForamt("group", "--group")),
			args:        args,
			expect:      MinioJobArg{Command: "--group group1 group2 group3"},
			expectError: false,
		},
		{
			command:     OneOf(KeyForamt("miss_user", "--user"), KeyForamt("miss_group", "--group")),
			args:        args,
			expect:      MinioJobArg{Command: "--group group1 group2 group3"},
			expectError: true,
		},
		{
			command:     NoSpace(ALIAS(), Static("/"), Key("name")),
			args:        args,
			expect:      MinioJobArg{Command: "myminio/mybucketName"},
			expectError: false,
		},
	}
	for _, tc := range testCase {
		cmd, err := tc.command(args)
		if tc.expectError && err == nil {
			t.Fatalf("expect error")
		}
		if !tc.expectError && err != nil {
			t.Fatalf("expect not a error")
		}
		if !tc.expectError {
			if tc.expect.Command != "" && cmd.Command != tc.expect.Command {
				t.Fatalf("expect %s, but got %s", tc.expect, cmd.Command)
			}
			if tc.expect.FileName != "" {
				if tc.expect.FileContext != cmd.FileContext {
					t.Fatalf("expect %s, but got %s", tc.expect.FileContext, cmd.FileContext)
				}
				if tc.expect.FileExt != cmd.FileExt {
					t.Fatalf("expect %s, but got %s", tc.expect.FileExt, cmd.FileExt)
				}
				if tc.expect.FileName != cmd.FileName {
					t.Fatalf("expect %s, but got %s", tc.expect.FileName, cmd.FileName)
				}
			}
		}
	}
}
