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

package color

import (
	"fmt"

	"github.com/fatih/color"
)

// global colors.
var (
	// Check if we stderr, stdout are dumb terminals, we do not apply
	// ansi coloring on dumb terminals.
	IsTerminal = func() bool {
		return !color.NoColor
	}

	Bold = func() func(a ...interface{}) string {
		if IsTerminal() {
			return color.New(color.Bold).SprintFunc()
		}
		return fmt.Sprint
	}()

	FgRed = func() func(a ...interface{}) string {
		if IsTerminal() {
			return color.New(color.FgRed).SprintFunc()
		}
		return fmt.Sprint
	}()

	BgRed = func() func(format string, a ...interface{}) string {
		if IsTerminal() {
			return color.New(color.BgRed).SprintfFunc()
		}
		return fmt.Sprintf
	}()

	FgWhite = func() func(format string, a ...interface{}) string {
		if IsTerminal() {
			return color.New(color.FgWhite).SprintfFunc()
		}
		return fmt.Sprintf
	}()
)
