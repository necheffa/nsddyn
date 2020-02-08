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
	"os"

	"cmd/internal/version"
)

func main() {
	if len(os.Args) <= 1 {
		log.Fatal("nsddynum: Error: no sub-command specified.")
	}

	switch os.Args[1] {
	default:
		log.Fatal("nsddynum: Error: unknown sub-command.")
	case "version":
		version.PrintVersion()
	case "adduser":
		addUserCmd := flag.NewFlagSet("adduser", flag.ExitOnError)

		var userName string
		var fileName string
		var password string
		var hostnames string

		nsddynHome, ok := os.LookupEnv("NSDDYN_HOME")

		if !ok {
			nsddynHome = "/usr/local"
		}

		fi, err := os.Stat(nsddynHome)
		if os.IsNotExist(err) {
			log.Fatal("nsddynum: Error: $NSDDYN_HOME set to non-existent location.")
		}
		if !fi.IsDir() {
			log.Fatal("nsddynum: Error: $NSDDYN_HOME is not set to a directory.")
		}

		addUserCmd.StringVar(&userName, "user-name", "", "Username to add.")
		addUserCmd.StringVar(&userName, "u", "", "Username to add.")
		addUserCmd.StringVar(&fileName, "passwd-file", nsddynHome+"/etc/nsddynpasswd", "Path to nsddyn passwd file.")
		addUserCmd.StringVar(&fileName, "f", nsddynHome+"/etc/nsddynpasswd", "Path to nsddyn passwd file.")
		addUserCmd.StringVar(&password, "passwd", "", "Desired password.")
		addUserCmd.StringVar(&password, "p", "", "Desired password.")
		addUserCmd.StringVar(&hostnames, "hosts", "", "Comma separated list of permitted hostnames.")
		addUserCmd.StringVar(&hostnames, "h", "", "Comma separated list of permitted hostnames.")

		addUserCmd.Parse(os.Args[2:])

		err = addUser(userName, fileName, password, hostnames)
		if err != nil {
			log.Println(err)
			log.Fatal("nsddynum: Error: could not add user to passwd file.")
		}
	case "help":
		msg := "Usage: nsddynum SUB-COMMAND [OPTS]\n" +
			"  nsddynnum is the nsddyn User Manager utility.\n" +
			"\n" +
			"Available SUB-COMMANDs:\n" +
			"help [SUB-COMMAND]\tPrints this message if no argument is given, otherwise prints help text for specified SUB-COMMAND.\n" +
			"version\t\t\tPrints version information and exits.\n" +
			"adduser [OPTS]\t\tAdds a user to the passwd file.\n"

		fmt.Fprintf(os.Stderr, msg)
	}
}
