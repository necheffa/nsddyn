/*
   Copyright (C) 2020, 2022 Alexander Necheff

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

// PrintUsage is called to display a useage message either when the user has explicitly requested one or
// the user has not passed in sensable arguments.
func PrintUsage() {
	msg := "Usage: dynupd [-vhd] [-p PASSWD_FILE] [-a ADDRESS[:PORT]] [-u URI] [-b URI] -z ZONE_FILE -n DOMAIN_NAME\n" +
		"  dynupd is the HTTP API host component of nsddyn.\n" +
		"\n" +
		"  -v,--version\t\tPrint version information and exit successfully.\n" +
		"  -h,--help\t\tPrint usage message and exit successfully.\n" +
		"  -d,--debug\t\tActivate verbose messaging.\n" +
		"  -z,--zone-file\tSpecify the zonefile to manage.\n" +
		"  -n,--name\t\tSpecify the domain name associated with the zonefile.\n" +
		"  -p,--passwd-file\tOverride the location of the password store.\n" +
		"  -a,--addr\t\tOverride the default listen host.\n" +
		"  -u,--uri\t\tOverride the default listen URI.\n" +
		"  -b,--broker\t\tOverride the default dynupd-broker URI.\n"
	fmt.Fprintf(os.Stderr, "%v", msg)
}
