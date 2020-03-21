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

	"cmd/internal/auth"
	"cmd/internal/version"

	"golang.org/x/crypto/ssh/terminal"
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
	case "removeuser":
		passwdDb := new(auth.FlatFile)
		log.Fatal(passwdDb.RemoveUser("user"))
	case "adduser":
		addUserCmd := flag.NewFlagSet("adduser", flag.ExitOnError)

		var userName string
		var fileName string
		var hostNames []string

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

		addUserCmd.Parse(os.Args[2:])

		hostNames = addUserCmd.Args()

		fmt.Fprintf(os.Stderr, "new password: ")
		passwd, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatal(fmt.Errorf("nsddynum: Error reading password: %v", err))
		}
		defer auth.EraseBuf(passwd)

		// TODO: there has to be a better way to handle this since the whole point is to avoid filling memory
		// with a maliciously crafted password...
		if len(passwd) > auth.MaxPasswd {
			log.Fatal("nsddynum: Error: password length exceeds max password length of ", auth.MaxPasswd)
		} else if len(passwd) < auth.MinPasswd {
			log.Fatal("nsddynum: Error: password length less than minimum length of ", auth.MinPasswd)
		} else {
			// we need to advance the prompt because we successfully consumed the \n
			fmt.Fprintf(os.Stderr, "\n")
		}

		// TODO: support multiple auth methods here...
		passwdDb := new(auth.FlatFile)
		passwdDb.SetFilePath(fileName)
		err = passwdDb.AddUser(passwd, userName, hostNames)
		if err != nil {
			log.Fatal(fmt.Errorf("nsddynum: %v", err))
		}
	case "help":
		// TODO: print the sub-command specific usage message if a sub-command is given as an argument
		msg := "Usage: nsddynum SUB-COMMAND [OPTS]\n" +
			"  nsddynnum is the nsddyn User Manager utility.\n" +
			"\n" +
			"Available SUB-COMMANDs:\n" +
			"help [SUB-COMMAND]\t\t\t\t\tPrints this message if no argument is given, otherwise prints help text for specified SUB-COMMAND.\n" +
			"version\t\t\t\t\t\t\tPrints version information and exits.\n" +
			"adduser [-f PASSWD_FILE ] -u USER HOST1 HOST2\t\tAdds a user to the passwd file.\n"

		fmt.Fprintf(os.Stderr, msg)
	}
}
