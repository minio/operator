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

package logger

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
)

func testServer(w http.ResponseWriter, r *http.Request) {
}

func TestInitializeLogger(t *testing.T) {
	testServerWillStart := make(chan interface{})
	http.HandleFunc("/", testServer)
	go func() {
		close(testServerWillStart)
		err := http.ListenAndServe("127.0.0.1:1337", nil)
		if err != nil {
			return
		}
	}()
	<-testServerWillStart

	loggerWebhookEnable := fmt.Sprintf("%s_TEST", EnvLoggerWebhookEnable)
	loggerWebhookEndpoint := fmt.Sprintf("%s_TEST", EnvLoggerWebhookEndpoint)
	loggerWebhookAuthToken := fmt.Sprintf("%s_TEST", EnvLoggerWebhookAuthToken)
	loggerWebhookClientCert := fmt.Sprintf("%s_TEST", EnvLoggerWebhookClientCert)
	loggerWebhookClientKey := fmt.Sprintf("%s_TEST", EnvLoggerWebhookClientKey)
	loggerWebhookQueueSize := fmt.Sprintf("%s_TEST", EnvLoggerWebhookQueueSize)

	auditWebhookEnable := fmt.Sprintf("%s_TEST", EnvAuditWebhookEnable)
	auditWebhookEndpoint := fmt.Sprintf("%s_TEST", EnvAuditWebhookEndpoint)
	auditWebhookAuthToken := fmt.Sprintf("%s_TEST", EnvAuditWebhookAuthToken)
	auditWebhookClientCert := fmt.Sprintf("%s_TEST", EnvAuditWebhookClientCert)
	auditWebhookClientKey := fmt.Sprintf("%s_TEST", EnvAuditWebhookClientKey)
	auditWebhookQueueSize := fmt.Sprintf("%s_TEST", EnvAuditWebhookQueueSize)

	type args struct {
		ctx       context.Context
		transport *http.Transport
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		setEnvVars   func()
		unsetEnvVars func()
	}{
		{
			name: "logger or auditlog is not enabled",
			args: args{
				ctx:       context.Background(),
				transport: http.DefaultTransport.(*http.Transport).Clone(),
			},
			wantErr: false,
			setEnvVars: func() {
			},
			unsetEnvVars: func() {
			},
		},
		{
			name: "logger webhook initialized correctly",
			args: args{
				ctx:       context.Background(),
				transport: http.DefaultTransport.(*http.Transport).Clone(),
			},
			wantErr: false,
			setEnvVars: func() {
				os.Setenv(loggerWebhookEnable, "on")
				os.Setenv(loggerWebhookEndpoint, "http://127.0.0.1:1337/logger")
				os.Setenv(loggerWebhookAuthToken, "test")
				os.Setenv(loggerWebhookClientCert, "")
				os.Setenv(loggerWebhookClientKey, "")
				os.Setenv(loggerWebhookQueueSize, "1000")
			},
			unsetEnvVars: func() {
				os.Unsetenv(loggerWebhookEnable)
				os.Unsetenv(loggerWebhookEndpoint)
				os.Unsetenv(loggerWebhookAuthToken)
				os.Unsetenv(loggerWebhookClientCert)
				os.Unsetenv(loggerWebhookClientKey)
				os.Unsetenv(loggerWebhookQueueSize)
			},
		},
		{
			name: "logger webhook failed to initialize",
			args: args{
				ctx:       context.Background(),
				transport: http.DefaultTransport.(*http.Transport).Clone(),
			},
			wantErr: true,
			setEnvVars: func() {
				os.Setenv(loggerWebhookEnable, "on")
				os.Setenv(loggerWebhookEndpoint, "https://aklsjdakljdjkalsd.com")
				os.Setenv(loggerWebhookAuthToken, "test")
				os.Setenv(loggerWebhookClientCert, "")
				os.Setenv(loggerWebhookClientKey, "")
				os.Setenv(loggerWebhookQueueSize, "1000")
			},
			unsetEnvVars: func() {
				os.Unsetenv(loggerWebhookEnable)
				os.Unsetenv(loggerWebhookEndpoint)
				os.Unsetenv(loggerWebhookAuthToken)
				os.Unsetenv(loggerWebhookClientCert)
				os.Unsetenv(loggerWebhookClientKey)
				os.Unsetenv(loggerWebhookQueueSize)
			},
		},
		{
			name: "auditlog webhook initialized correctly",
			args: args{
				ctx:       context.Background(),
				transport: http.DefaultTransport.(*http.Transport).Clone(),
			},
			wantErr: false,
			setEnvVars: func() {
				os.Setenv(auditWebhookEnable, "on")
				os.Setenv(auditWebhookEndpoint, "http://127.0.0.1:1337/audit")
				os.Setenv(auditWebhookAuthToken, "test")
				os.Setenv(auditWebhookClientCert, "")
				os.Setenv(auditWebhookClientKey, "")
				os.Setenv(auditWebhookQueueSize, "1000")
			},
			unsetEnvVars: func() {
				os.Unsetenv(auditWebhookEnable)
				os.Unsetenv(auditWebhookEndpoint)
				os.Unsetenv(auditWebhookAuthToken)
				os.Unsetenv(auditWebhookClientCert)
				os.Unsetenv(auditWebhookClientKey)
				os.Unsetenv(auditWebhookQueueSize)
			},
		},
		{
			name: "auditlog webhook failed to initialize",
			args: args{
				ctx:       context.Background(),
				transport: http.DefaultTransport.(*http.Transport).Clone(),
			},
			wantErr: true,
			setEnvVars: func() {
				os.Setenv(auditWebhookEnable, "on")
				os.Setenv(auditWebhookEndpoint, "https://aklsjdakljdjkalsd.com")
				os.Setenv(auditWebhookAuthToken, "test")
				os.Setenv(auditWebhookClientCert, "")
				os.Setenv(auditWebhookClientKey, "")
				os.Setenv(auditWebhookQueueSize, "1000")
			},
			unsetEnvVars: func() {
				os.Unsetenv(auditWebhookEnable)
				os.Unsetenv(auditWebhookEndpoint)
				os.Unsetenv(auditWebhookAuthToken)
				os.Unsetenv(auditWebhookClientCert)
				os.Unsetenv(auditWebhookClientKey)
				os.Unsetenv(auditWebhookQueueSize)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnvVars != nil {
				tt.setEnvVars()
			}
			if err := InitializeLogger(tt.args.ctx, tt.args.transport); (err != nil) != tt.wantErr {
				t.Errorf("InitializeLogger() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.unsetEnvVars != nil {
				tt.unsetEnvVars()
			}
		})
	}
}

func TestEnableJSON(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "enable json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			EnableJSON()
			if !IsJSON() {
				t.Errorf("EnableJSON() = %v, want %v", IsJSON(), true)
			}
		})
	}
}

func TestEnableQuiet(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "enable quiet",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			EnableQuiet()
			if !IsQuiet() {
				t.Errorf("EnableQuiet() = %v, want %v", IsQuiet(), true)
			}
		})
	}
}

func TestEnableAnonymous(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "enable anonymous",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			EnableAnonymous()
			if !IsAnonymous() {
				t.Errorf("EnableAnonymous() = %v, want %v", IsAnonymous(), true)
			}
		})
	}
}
