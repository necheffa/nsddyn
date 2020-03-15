/*
   Copyright (C) 2020 Alexander Necheff

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

package main

import (
	"fmt"
	"os"
)

func PrintUsage() {
	msg := "Usage: dynupd [OPTS]\n" +
		"  dynupd is the HTTP API host component of nsddyn.\n" +
		"\n" +
		"  -v,--version\t\tPrint version information and exit successfully.\n" +
		"  -h,--help\t\tPrint usage message and exit successfully.\n" +
		"  -d,--debug\t\tActivate verbose messaging.\n" +
		"  -f,--zone-file\tSpecify the zonefile to manage.\n" +
		"  -p,--passwd-file\tOverride the location of the password store.\n"
	fmt.Fprintf(os.Stderr, msg)
}
