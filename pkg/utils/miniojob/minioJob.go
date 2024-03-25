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

// Arg - parse the arg result
type Arg struct {
	Command     string
	FileName    string
	FileExt     string
	FileContext string
}

// IsFile - if it is a file
func (arg Arg) IsFile() bool {
	return arg.FileName != ""
}

// FieldsFunc - alias function
type FieldsFunc func(args map[string]string) (Arg, error)

// Key - key=value
func Key(key string) FieldsFunc {
	return KeyForamt(key, "$0")
}

// FLAGS - --key=value|value1,value2,value3
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

// File - file key, ext
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
				return out, nil
			}
		}
		return out, fmt.Errorf("file %s not found", fName)
	}
}

// KeyForamt - key,outPut
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

// NoSpace - no space for the command
func NoSpace(funcs ...FieldsFunc) FieldsFunc {
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
