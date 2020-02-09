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

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"cmd/internal/version"
)

func main() {
	nsddyndCmd := flag.NewFlagSet("nsddynd", flag.ExitOnError)

	var printHelp bool
	var printVersion bool

	nsddyndCmd.BoolVar(&printHelp, "help", false, "Print usage message and exit successfully.")
	nsddyndCmd.BoolVar(&printHelp, "h", false, "Print usage message and exit successfully.")
	nsddyndCmd.BoolVar(&printVersion, "version", false, "Print version information and exit successfully")
	nsddyndCmd.BoolVar(&printVersion, "v", false, "Print version information and exit successfully")

	nsddyndCmd.Parse(os.Args[1:])

	if printVersion {
		version.PrintVersion()
		return
	}

	if printHelp {
		msg := "Usage: nsddynd [OPTS]\n" +
			"  nsddynd is the DNS zone update component of nsddyn.\n" +
			"\n" +
			"  -v,--version\t\tPrint version information and exit successfully.\n" +
			"  -h,--help\t\tPrint usage message and exit successfully.\n"
		fmt.Fprintf(os.Stderr, msg)
		return
	}

	http.HandleFunc("/api/nsddynd", nsddyndHandler)
	http.ListenAndServe("localhost:1337", nil)

	return
}
