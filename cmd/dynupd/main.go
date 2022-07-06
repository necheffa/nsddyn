/*
   Copyright (C) 2019, 2020, 2021, 2022 Alexander Necheff

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

	"necheff.net/nsddyn/cmd/internal/auth"
	"necheff.net/nsddyn/cmd/internal/config"
	"necheff.net/nsddyn/cmd/internal/util"
	"necheff.net/nsddyn/cmd/internal/version"
)

const (
	defaultUri       = "/api/dynupd"
	defaultHost      = "localhost:8080"
	defaultBrokerUri = "http://localhost:8081/api/dynupd-broker"
)

func main() {

	var host string
	var uri string

	dynupdCmd := flag.NewFlagSet("dynupd", flag.ExitOnError)

	var printHelp bool
	var printVersion bool
	var debugFlag bool
	var zoneFile string
	var passwdFile string
	var listenAddr string
	var listenUri string
	var domainName string
	var brokerUri string

	dynupdCmd.BoolVar(&printHelp, "help", false, "Print usage message and exit successfully.")
	dynupdCmd.BoolVar(&printHelp, "h", false, "Print usage message and exit successfully.")
	dynupdCmd.BoolVar(&printVersion, "version", false, "Print version information and exit successfully.")
	dynupdCmd.BoolVar(&printVersion, "v", false, "Print version information and exit successfully.")
	dynupdCmd.BoolVar(&debugFlag, "d", false, "Activate verbose messaging.")
	dynupdCmd.BoolVar(&debugFlag, "debug", false, "Activate verbose messaging.")
	dynupdCmd.StringVar(&zoneFile, "z", "", "Specify the zonefile to manage.")
	dynupdCmd.StringVar(&zoneFile, "zone-file", "", "Specify the zonefile to manage.")
	dynupdCmd.StringVar(&passwdFile, "p", "", "Override the location of the passwd store.")
	dynupdCmd.StringVar(&passwdFile, "passwd-file", "", "Override the location of the passwd store.")
	dynupdCmd.StringVar(&listenAddr, "a", "", "Override the default listen address and optionally port.")
	dynupdCmd.StringVar(&listenAddr, "addr", "", "Override the default listen address and optionally port.")
	dynupdCmd.StringVar(&listenUri, "u", "", "Override the ddefault listen URI.")
	dynupdCmd.StringVar(&listenUri, "uri", "", "Override the ddefault listen URI.")
	dynupdCmd.StringVar(&domainName, "n", "", "Specify the domain name associated with the zonefile.")
	dynupdCmd.StringVar(&domainName, "name", "", "Specify the domain name associated with the zonefile.")
	dynupdCmd.StringVar(&brokerUri, "b", "", "Specify the URI of dynupd-broker.")
	dynupdCmd.StringVar(&brokerUri, "broker", "", "Specify the URI of dynupd-broker.")

	err := dynupdCmd.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "dynupd: error: %v\n", err.Error())
	}

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
		fmt.Fprintf(os.Stderr, "%s\n", "dynupd: error: missing required argument, --zone-file")
		return
	}

	if domainName == "" {
		// A future enhancement might just default to using the domain name of the host system.
		PrintUsage()
		fmt.Fprintf(os.Stderr, "%s\n", "dynupd: error: missing required argument, --name")
		return
	}

	if listenAddr == "" {
		host = defaultHost
	} else {
		host = listenAddr
	}

	if listenUri == "" {
		uri = defaultUri
	} else {
		uri = listenUri
	}

	if brokerUri == "" {
		brokerUri = defaultBrokerUri
	}

	if config.Debug {
		msg := "dynupd: listening on: " + host + uri
		fmt.Fprintf(os.Stderr, "%s\n", msg)
	}

	nsddynHome, err := util.FindHome()
	if err != nil {
		log.Fatal(err)
	}

	if passwdFile != "" {
		fi, err := os.Stat(passwdFile)
		if os.IsNotExist(err) {
			log.Fatal("dynupd: Error: file specified by -p does not exist.")
		}
		if !fi.Mode().IsRegular() {
			log.Fatal("dynupd: Error: file specified by -p is not a regular file.")
		}
	}

	d := new(DynUpd)
	// NOTE: support multiple auth mechanisms here...
	passwdDb := new(auth.FlatFile)

	fileName := passwdFile
	if passwdFile == "" {
		fileName = nsddynHome + "/etc/nsddynpasswd"
	}

	ok, err := util.CheckFilePerms(fileName, util.NsddynpasswdPerms)
	if err != nil {
		log.Fatal(fmt.Errorf("dynupd: Error: %w", err))
	}
	if !ok {
		log.Print("dynupd: Warn: permissions on nsddynpasswd too permissive, recommend chmod " + util.NsddynpasswdPermsStr + ".")
	}

	passwdDb.SetFilePath(fileName)
	d.NewDynUpd(passwdDb, zoneFile, domainName, brokerUri)

	http.HandleFunc(uri, d.DynUpdHandler)
	err = http.ListenAndServe(host, nil)
	if err != nil {
		log.Fatal(fmt.Errorf("dynupd: Error: %w", err))
	}
}
