/*
   Copyright (C) 2019, 2020 Alexander Necheff

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
)

// ldflags are used to set these.
var (
	PlatformName = ""
	Version      = "unknown-version"
	GitCommit    = "unknown-commit"
	BuildTime    = "unknown-buildtime"
	GoVersion    = "unknown-goversion"
)

func PrintVersion() {
	fmt.Fprintf(os.Stderr, filepath.Base(os.Args[0])+" v"+Version+"\n")
	fmt.Fprintf(os.Stderr, "Copyright (C) 2019, 2020\n")
	fmt.Fprintf(os.Stderr, "Alexander Necheff\nnsddynum is licensed under the terms of the GPLv3.\n")
	fmt.Fprintf(os.Stderr, "Git Commit: "+GitCommit+"\n")
	fmt.Fprintf(os.Stderr, "Build on: "+BuildTime+" by Go toolchain version: "+GoVersion+"\n")
}
