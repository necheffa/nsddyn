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
	"log"
	"net/http"
	"os"

	"cmd/internal/auth"
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
	var zoneFile string

	dynupdCmd.BoolVar(&printHelp, "help", false, "Print usage message and exit successfully.")
	dynupdCmd.BoolVar(&printHelp, "h", false, "Print usage message and exit successfully.")
	dynupdCmd.BoolVar(&printVersion, "version", false, "Print version information and exit successfully.")
	dynupdCmd.BoolVar(&printVersion, "v", false, "Print version information and exit successfully.")
	dynupdCmd.BoolVar(&debugFlag, "d", false, "Activate verbose messaging.")
	dynupdCmd.BoolVar(&debugFlag, "debug", false, "Activate verbose messaging.")
	dynupdCmd.StringVar(&zoneFile, "f", "", "Specify the zonefile to manage.")
	dynupdCmd.StringVar(&zoneFile, "zone-file", "", "Specify the zonefile to manage.")

	dynupdCmd.Parse(os.Args[1:])

	if printVersion {
		version.PrintVersion()
		return
	}

	if debugFlag {
		config.Debug = true
	}

	if printHelp {
		PrintUsage()
		return
	}

	if zoneFile == "" {
		PrintUsage()
		fmt.Fprintf(os.Stderr, "dynupd: error: missing required argument, --zone-file\n")
		return
	}

	if config.Debug {
		msg := "dynupd: listening on: " + host + uri + "\n"
		fmt.Fprintf(os.Stderr, msg)
	}

	nsddynHome, ok := os.LookupEnv("NSDDYN_HOME")

	if !ok {
		nsddynHome = "/usr/local"
	}

	fi, err := os.Stat(nsddynHome)
	if os.IsNotExist(err) {
		log.Fatal("dynupd: Error: $NSDDYN_HOME set to non-existent location.")
	}
	if !fi.IsDir() {
		log.Fatal("dynupd: Error: $NSDDYN_HOME is not set to a directory.")
	}

	d := new(DynUpd)
	// TODO: support multiple auth mechanisms here...
	passwdDb := new(auth.FlatFile)
	passwdDb.SetFilePath(nsddynHome + "/etc/nsddynpasswd")
	d.NewDynUpd(passwdDb, zoneFile)

	http.HandleFunc(uri, d.DynUpdHandler)
	http.ListenAndServe(host, nil)

	return
}
