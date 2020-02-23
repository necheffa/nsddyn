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

	"cmd/internal/config"
	"cmd/internal/version"
)

const (
	uri  = "/api/dynupd"
	host = "localhost:8080"
)

func main() {
	dynupdCmd := flag.NewFlagSet("dynupd", flag.ExitOnError)

	var printHelp bool
	var printVersion bool
	var debugFlag bool

	dynupdCmd.BoolVar(&printHelp, "help", false, "Print usage message and exit successfully.")
	dynupdCmd.BoolVar(&printHelp, "h", false, "Print usage message and exit successfully.")
	dynupdCmd.BoolVar(&printVersion, "version", false, "Print version information and exit successfully.")
	dynupdCmd.BoolVar(&printVersion, "v", false, "Print version information and exit successfully.")
	dynupdCmd.BoolVar(&debugFlag, "d", false, "Activate verbose messaging.")
	dynupdCmd.BoolVar(&debugFlag, "debug", false, "Activate verbose messaging.")

	dynupdCmd.Parse(os.Args[1:])

	if printVersion {
		version.PrintVersion()
		return
	}

	if debugFlag {
		config.Debug = true
	}

	if printHelp {
		msg := "Usage: dynupd [OPTS]\n" +
			"  dynupd is the HTTP API host component of nsddyn.\n" +
			"\n" +
			"  -v,--version\t\tPrint version information and exit successfully.\n" +
			"  -h,--help\t\tPrint usage message and exit successfully.\n" +
			"  -d,--debug\t\tActivate verbose messaging.\n"
		fmt.Fprintf(os.Stderr, msg)
		return
	}

	if config.Debug {
		msg := "dynupd: listening on: " + host + uri + "\n"
		fmt.Fprintf(os.Stderr, msg)
	}

	http.HandleFunc(uri, dynupdHandler)
	http.ListenAndServe(host, nil)

	return
}
