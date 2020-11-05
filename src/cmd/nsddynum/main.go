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
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"cmd/internal/auth"
	"cmd/internal/util"
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
	case "deluser":
		delUserCmd := flag.NewFlagSet("deluser", flag.ExitOnError)

		var userName string
		var fileName string

		nsddynHome, err := util.FindHome()
		if err != nil {
			log.Fatal(fmt.Errorf("nsddynum: %v", err))
		}

		delUserCmd.StringVar(&userName, "user-name", "", "Username to delete.")
		delUserCmd.StringVar(&userName, "u", "", "Username to delete.")
		delUserCmd.StringVar(&fileName, "passwd-file", nsddynHome+"/etc/nsddynpasswd", "Path to nsddyn passwd file.")
		delUserCmd.StringVar(&fileName, "p", nsddynHome+"/etc/nsddynpasswd", "Path to nsddyn passwd file.")

		delUserCmd.Parse(os.Args[2:])

		// TODO: support multiple auth methods here...
		passwdDb := new(auth.FlatFile)
		passwdDb.SetFilePath(fileName)
		err = passwdDb.DelUser(userName)
		if err != nil {
			log.Fatal(fmt.Errorf("nsddynum: %v", err))
		}
	case "moduser":
		modUserCmd := flag.NewFlagSet("moduser", flag.ExitOnError)

		var userName string
		var fileName string

		nsddynHome, err := util.FindHome()
		if err != nil {
			log.Fatal(fmt.Errorf("nsddynum: %v", err))
		}

		modUserCmd.StringVar(&userName, "user-name", "", "Username to modify.")
		modUserCmd.StringVar(&userName, "u", "", "Username to modify.")
		modUserCmd.StringVar(&fileName, "passwd-file", nsddynHome+"/etc/nsddynpasswd", "Path to nsddyn passwd file.")
		modUserCmd.StringVar(&fileName, "p", nsddynHome+"/etc/nsddynpasswd", "Path to nsddyn passwd file.")

		modUserCmd.Parse(os.Args[2:])

		reader := bufio.NewReader(os.Stdin)

		fmt.Fprintf(os.Stderr, "Update password? [y/N]: ")
		upPassAns, _ := reader.ReadString('\n')
		upPassAns = strings.Trim(upPassAns, "\n")
		if upPassAns == "y" || upPassAns == "Y" {
			passwd, err := promptForPasswd()
			if err != nil {
				log.Fatal(fmt.Errorf("nsddynum: %v", err))
			}
			defer auth.EraseBuf(passwd)
			fmt.Println("updating password")
		} else {
			fmt.Fprintf(os.Stderr, "Not updating password.\n")
		}

		fmt.Fprintf(os.Stderr, "Update authorized hosts? [y/N] ")
		upHostsAns, _ := reader.ReadString('\n')
		upHostsAns = strings.Trim(upHostsAns, "\n")
		if upHostsAns == "y" || upHostsAns == "Y" {
			fmt.Println("updating authorized hosts")
		} else {
			fmt.Fprintf(os.Stderr, "Not updating authorized hosts.\n")
		}
	case "adduser":
		addUserCmd := flag.NewFlagSet("adduser", flag.ExitOnError)

		var userName string
		var fileName string
		var hostNames []string

		nsddynHome, err := util.FindHome()
		if err != nil {
			log.Fatal(fmt.Errorf("nsddynum: %v", err))
		}

		addUserCmd.StringVar(&userName, "user-name", "", "Username to add.")
		addUserCmd.StringVar(&userName, "u", "", "Username to add.")
		addUserCmd.StringVar(&fileName, "passwd-file", nsddynHome+"/etc/nsddynpasswd", "Path to nsddyn passwd file.")
		addUserCmd.StringVar(&fileName, "p", nsddynHome+"/etc/nsddynpasswd", "Path to nsddyn passwd file.")

		addUserCmd.Parse(os.Args[2:])

		hostNames = addUserCmd.Args()

		passwd, err := promptForPasswd()
		if err != nil {
			log.Fatal(fmt.Errorf("nsddynum: %v", err))
		}
		defer auth.EraseBuf(passwd)

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
			"adduser [-p PASSWD_FILE ] -u USER HOST1 HOST2\t\tAdds a user to the passwd file.\n" +
			"moduser [-p PASSWD_FILE ] -u USER\t\t\tModifies the password or authorized hostnames for a user account.\n" +
			"deluser [-p PASSWD_FILE ] -u USER\t\t\tRemoves the specified user account from the passwd file.\n"

		fmt.Fprintf(os.Stderr, msg)
	}
}
