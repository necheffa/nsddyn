/*
   Copyright (C) 2019-2025 Alexander Necheff

   This file is part of nsddyn.

   nsddyn is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, at version 3 of the License.

   nsddyn is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with nsddyn.  If not, see <https://www.gnu.org/licenses/>.
*/

// Package version provides versioning information.
package version

import (
	"fmt"
	"os"
	"path/filepath"

	"necheff.net/nsddyn"
)

// PrintVersion is used to display version, copyright, and licensing
// whenever an nsddyn binary is called with a version command-line flag.
func PrintVersion() {
	fmt.Fprintf(os.Stderr, "%s v%s\n", filepath.Base(os.Args[0]), nsddyn.Version())
	fmt.Fprintf(os.Stderr, "Copyright (C) 2019-2024\n")
	fmt.Fprintf(os.Stderr, "Alexander Necheff\nnsddyn is licensed under the terms of the GPLv3.\n")
}
