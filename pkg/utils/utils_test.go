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

import "testing"

func TestDecodeInput(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "chinese characters",
			args: args{
				s: "5bCP6aO85by+5bCP6aO85by+5bCP6aO85by+L+Wwj+mjvOW8vuWwj+mjvOW8vuWwj+mjvOW8vg==",
			},
			want:    "小飼弾小飼弾小飼弾/小飼弾小飼弾小飼弾",
			wantErr: false,
		},
		{
			name: "spaces and & symbol",
			args: args{
				s: "YSBhIC0gYSBhICYgYSBhIC0gYSBhIGE=",
			},
			want:    "a a - a a & a a - a a a",
			wantErr: false,
		},
		{
			name: "the infamous fly me to the moon",
			args: args{
				s: "MDIlMjAtJTIwRkxZJTIwTUUlMjBUTyUyMFRIRSUyME1PT04lMjA=",
			},
			want:    "02%20-%20FLY%20ME%20TO%20THE%20MOON%20",
			wantErr: false,
		},
		{
			name: "random symbols",
			args: args{
				s: "IUAjJCVeJiooKV8r",
			},
			want:    "!@#$%^&*()_+",
			wantErr: false,
		},
		{
			name: "name with / symbols",
			args: args{
				s: "dGVzdC90ZXN0Mi/lsI/po7zlvL7lsI/po7zlvL7lsI/po7zlvL4uanBn",
			},
			want:    "test/test2/小飼弾小飼弾小飼弾.jpg",
			wantErr: false,
		},
		{
			name: "decoding fails",
			args: args{
				s: "this should fail",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeBase64(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeBase64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DecodeBase64() got = %v, want %v", got, tt.want)
			}
		})
	}
}
