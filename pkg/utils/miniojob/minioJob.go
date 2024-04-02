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
	"fmt"
	"sort"
	"strings"
)

// ArgType - arg type
type ArgType int

const (
	// ArgTypeKey - key=value print value
	ArgTypeKey ArgType = iota
	// ArgTypeFile - key=value print /temp/value.ext
	ArgTypeFile
	// ArgTypeKeyFile - key=value print key="/temp/value.ext"
	ArgTypeKeyFile
)

// Arg - parse the arg result
type Arg struct {
	Command     string
	FileName    string
	FileExt     string
	FileContext string
	ArgType     ArgType
}

// FieldsFunc - alias function
type FieldsFunc func(args map[string]string) (Arg, error)

// Key - key=value|value1,value2,value3
func Key(key string) FieldsFunc {
	return KeyForamt(key, "$0")
}

// FLAGS - --key=""|value|value1,value2,value3
func FLAGS(ignoreKeys ...string) FieldsFunc {
	return prefixKeyForamt("-", ignoreKeys...)
}

// ALIAS - myminio
func ALIAS() FieldsFunc {
	return Static("myminio")
}

// Static - some static value
func Static(val string) FieldsFunc {
	return func(args map[string]string) (Arg, error) {
		return Arg{Command: val}, nil
	}
}

// File - fName is the the key, value is content, ext is the file ext
func File(fName string, ext string) FieldsFunc {
	return func(args map[string]string) (out Arg, err error) {
		if args == nil {
			return out, fmt.Errorf("args is nil")
		}
		for key, val := range args {
			if key == fName {
				if val == "" {
					return out, fmt.Errorf("value is empty")
				}
				out.FileName = fName
				out.FileExt = ext
				out.FileContext = strings.TrimSpace(val)
				out.ArgType = ArgTypeFile
				return out, nil
			}
		}
		return out, fmt.Errorf("file %s not found", fName)
	}
}

// KeyValue - match key and putout the key, like endpoint="https://webhook-1.example.net"
func KeyValue(key string) FieldsFunc {
	return func(args map[string]string) (out Arg, err error) {
		if args == nil {
			return out, fmt.Errorf("args is nil")
		}
		val, ok := args[key]
		if !ok {
			return out, fmt.Errorf("key %s not found", key)
		}
		out.Command = fmt.Sprintf(`%s="%s"`, key, val)
		return out, nil
	}
}

// KeyFile - match key and putout the key, like client_cert="[here is content]"
func KeyFile(key string, ext string) FieldsFunc {
	return func(args map[string]string) (out Arg, err error) {
		if args == nil {
			return out, fmt.Errorf("args is nil")
		}
		val, ok := args[key]
		if !ok {
			return out, fmt.Errorf("key %s not found", key)
		}
		out.FileName = key
		out.FileExt = ext
		out.FileContext = strings.TrimSpace(val)
		out.ArgType = ArgTypeKeyFile
		return out, nil
	}
}

// Option - ignore the error
func Option(opt FieldsFunc) FieldsFunc {
	return func(args map[string]string) (out Arg, err error) {
		if args == nil {
			return out, nil
		}
		out, _ = opt(args)
		return out, nil
	}
}

// KeyForamt - match key and get outPut to replace $0 to output the value
// if format not contain $0, will add $0 to the end
func KeyForamt(key string, format string) FieldsFunc {
	return func(args map[string]string) (out Arg, err error) {
		if args == nil {
			return out, fmt.Errorf("args is nil")
		}
		if !strings.Contains(format, "$0") {
			format = fmt.Sprintf("%s %s", format, "$0")
		}
		val, ok := args[key]
		if !ok {
			return out, fmt.Errorf("key %s not found", key)
		}
		out.Command = strings.ReplaceAll(format, "$0", strings.ReplaceAll(val, ",", " "))
		return out, nil
	}
}

// OneOf - one of the funcs must be found
// mc admin policy attach OneOf(--user | --group) = mc admin policy attach --user user or mc admin policy attach --group group
func OneOf(funcs ...FieldsFunc) FieldsFunc {
	return func(args map[string]string) (out Arg, err error) {
		if args == nil {
			return out, fmt.Errorf("args is nil")
		}
		for _, fn := range funcs {
			if out, err = fn(args); err == nil {
				return out, nil
			}
		}
		return out, fmt.Errorf("not found")
	}
}

// Sanitize - no space for the command
// mc mb Sanitize(alias / bucketName) = mc mb alias/bucketName
func Sanitize(funcs ...FieldsFunc) FieldsFunc {
	return func(args map[string]string) (out Arg, err error) {
		if args == nil {
			return out, fmt.Errorf("args is nil")
		}
		commands := []string{}
		for _, func1 := range funcs {
			if out, err = func1(args); err != nil {
				return out, err
			}
			if out.Command == "" {
				return out, fmt.Errorf("command is empty")
			}
			commands = append(commands, out.Command)
		}
		return Arg{Command: strings.Join(commands, "")}, nil
	}
}

var prefixKeyForamt = func(pkey string, ignoreKeys ...string) FieldsFunc {
	return func(args map[string]string) (out Arg, err error) {
		if args == nil {
			return out, fmt.Errorf("args is nil")
		}
		igrnoreKeyMap := make(map[string]bool)
		for _, key := range ignoreKeys {
			if !strings.HasPrefix(key, pkey) {
				key = fmt.Sprintf("%s%s%s", pkey, pkey, key)
			}
			igrnoreKeyMap[key] = true
		}
		data := []string{}
		for key, val := range args {
			if strings.HasPrefix(key, pkey) && !igrnoreKeyMap[key] {
				if val == "" {
					data = append(data, key)
				} else {
					for _, singalVal := range strings.Split(val, ",") {
						if strings.TrimSpace(singalVal) != "" {
							data = append(data, fmt.Sprintf("%s=%s", key, singalVal))
						}
					}
				}
			}
		}
		// avoid flags change the order
		sort.Slice(data, func(i, j int) bool {
			return data[i] > data[j]
		})
		return Arg{Command: strings.Join(data, " ")}, nil
	}
}
