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
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"

	"cmd/internal/auth"
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

		// we are intentionally using lower level, unbuffered routines for reading the new password.
		// this way, we control the buffer and can wipe it when we are done.
		// also, we need to read the password as a []byte not a string to make the wipe effective.
		// we allocate +1 to detect if the entered password was too big/store \n if auth.MaxPasswd is entered.
		passwd := make([]byte, auth.MaxPasswd+1)
		defer auth.EraseBuf(passwd)
		fmt.Fprintf(os.Stderr, "new password: ")
		_, err = os.Stdin.Read(passwd)
		if err != nil {
			log.Fatal(fmt.Errorf("nsddynum: Error reading password: %v", err))
		}

		newLine := bytes.IndexRune(passwd, '\n')
		if newLine < 0 {
			log.Fatal(fmt.Errorf("nsddynum: Error: password length exceeds max password length of %d", auth.MaxPasswd))
		} else if newLine < auth.MinPasswd {
			log.Fatal("nsddynum: Error: password length too short.")
		} else {
			// we need to advance the prompt because we successfully consumed the \n
			fmt.Fprintf(os.Stderr, "\n")
		}

		// TODO: support multiple auth methods here...
		passwdDb := new(auth.FlatFile)
		passwdDb.SetFilePath(fileName)
		err = passwdDb.AddUser(passwd[:newLine], userName, hostNames)
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
