/*
   Copyright (C) 2019-2022, 2024 Alexander Necheff

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
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"necheff.net/nsddyn/internal/auth"
	"necheff.net/nsddyn/internal/util"
	"necheff.net/nsddyn/internal/version"
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
			log.Fatal(fmt.Errorf("nsddynum: %w", err))
		}

		delUserCmd.StringVar(&userName, "user-name", "", "Username to delete.")
		delUserCmd.StringVar(&userName, "u", "", "Username to delete.")
		delUserCmd.StringVar(&fileName, "passwd-file", nsddynHome+"/etc/nsddynpasswd", "Path to nsddyn passwd file.")
		delUserCmd.StringVar(&fileName, "p", nsddynHome+"/etc/nsddynpasswd", "Path to nsddyn passwd file.")

		err = delUserCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(fmt.Errorf("nsddynum: %w", err))
		}

		if userName == "" {
			log.Fatal(fmt.Errorf("nsddynum: required argument `-u` missing."))
		}

		ok, err := util.CheckFilePerms(fileName, util.NsddynpasswdPerms)
		if err != nil {
			log.Fatal(fmt.Errorf("nsddynum: %w", err))
		}
		if !ok {
			log.Print(fmt.Errorf("nsddynum: permissions on nsddynpasswd too permissive, recommend chmod " + util.NsddynpasswdPermsStr + "."))
		}

		// NOTE: support multiple auth methods here...
		passwdDb := new(auth.FlatFile)
		passwdDb.SetFilePath(fileName)
		err = passwdDb.DelUser(userName)
		if err != nil {
			log.Fatal(fmt.Errorf("nsddynum: %w", err))
		}
	case "moduser":
		modUserCmd := flag.NewFlagSet("moduser", flag.ExitOnError)

		var userName string
		var fileName string

		nsddynHome, err := util.FindHome()
		if err != nil {
			log.Fatal(fmt.Errorf("nsddynum: %w", err))
		}

		modUserCmd.StringVar(&userName, "user-name", "", "Username to modify.")
		modUserCmd.StringVar(&userName, "u", "", "Username to modify.")
		modUserCmd.StringVar(&fileName, "passwd-file", nsddynHome+"/etc/nsddynpasswd", "Path to nsddyn passwd file.")
		modUserCmd.StringVar(&fileName, "p", nsddynHome+"/etc/nsddynpasswd", "Path to nsddyn passwd file.")

		err = modUserCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(fmt.Errorf("nsddynum: %w", err))
		}

		if userName == "" {
			log.Fatal(fmt.Errorf("nsddynum: required argument `-u` missing."))
		}

		ok, err := util.CheckFilePerms(fileName, util.NsddynpasswdPerms)
		if err != nil {
			log.Fatal(fmt.Errorf("nsddynum: %w", err))
		}
		if !ok {
			log.Print(fmt.Errorf("nsddynum: permissions on nsddynpasswd too permissive, recommend chmod " + util.NsddynpasswdPermsStr + "."))
		}

		reader := bufio.NewReader(os.Stdin)
		passwdDb := new(auth.FlatFile)
		passwdDb.SetFilePath(fileName)

		passwd := []byte{}
		hostNames := []string{}
		modPasswd := false
		modHostNames := false
		defer auth.EraseBuf(passwd)

		fmt.Fprintf(os.Stderr, "Update password? [y/N]: ")
		upPassAns, _ := reader.ReadString('\n')
		upPassAns = strings.Trim(upPassAns, "\n")
		if upPassAns == "y" || upPassAns == "Y" {
			modPasswd = true
			passwd, err = promptForPasswd(os.Stdin)
			if err != nil {
				log.Fatal(fmt.Errorf("nsddynum: %w", err))
			}
		} else {
			fmt.Fprintf(os.Stderr, "Not updating password.\n")
		}

		fmt.Fprintf(os.Stderr, "Update authorized hosts? [y/N] ")
		upHostsAns, _ := reader.ReadString('\n')
		upHostsAns = strings.Trim(upHostsAns, "\n")
		if upHostsAns == "y" || upHostsAns == "Y" {
			hostNames, err = promptForHosts()
			modHostNames = true
			if err != nil {
				log.Fatal(fmt.Errorf("nsddynum: %w", err))
			}
		} else {
			fmt.Fprintf(os.Stderr, "Not updating authorized hosts.\n")
		}

		if !modPasswd && !modHostNames {
			return
		}

		err = passwdDb.ModUser(userName, passwd, modPasswd, hostNames, modHostNames)
		if err != nil {
			log.Fatal(fmt.Errorf("nsddynum: %w", err))
		}

	case "adduser":
		addUserCmd := flag.NewFlagSet("adduser", flag.ExitOnError)

		var userName string
		var fileName string
		var hostNames []string

		nsddynHome, err := util.FindHome()
		if err != nil {
			log.Fatal(fmt.Errorf("nsddynum: %w", err))
		}

		addUserCmd.StringVar(&userName, "user-name", "", "Username to add.")
		addUserCmd.StringVar(&userName, "u", "", "Username to add.")
		addUserCmd.StringVar(&fileName, "passwd-file", nsddynHome+"/etc/nsddynpasswd", "Path to nsddyn passwd file.")
		addUserCmd.StringVar(&fileName, "p", nsddynHome+"/etc/nsddynpasswd", "Path to nsddyn passwd file.")

		err = addUserCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(fmt.Errorf("nsddynum: %w", err))
		}

		if userName == "" {
			log.Fatal(fmt.Errorf("nsddynum: required argument `-u` missing."))
		}

		hostNames = addUserCmd.Args()

		if len(hostNames) < 1 {
			log.Fatal(fmt.Errorf("nsddynum: hostname arugment required when adding user."))
		}

		passwd, err := promptForPasswd(os.Stdin)
		if err != nil {
			log.Fatal(fmt.Errorf("nsddynum: %w", err))
		}
		defer auth.EraseBuf(passwd)

		if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
			log.Print(fmt.Errorf("nsddynum: creating nsddynpasswd file at: " + fileName))

			f, err := os.Create(fileName)
			if err != nil {
				log.Fatal(fmt.Errorf("nsddynum: unable to create non-existant nsddynpasswd file. error was: %w", err))
			}
			f.Close()

			err = os.Chmod(fileName, util.NsddynpasswdPerms)
			if err != nil {
				log.Print(fmt.Errorf("nsddynum: unable to modify nsddynpasswd permissions: %w", err))
			}
		}

		ok, err := util.CheckFilePerms(fileName, util.NsddynpasswdPerms)
		if err != nil {
			log.Fatal(fmt.Errorf("nsddynum: %w", err))
		}
		if !ok {
			log.Print(fmt.Errorf("nsddynum: permissions on nsddynpasswd too permissive, recommend chmod " + util.NsddynpasswdPermsStr + "."))
		}

		// NOTE: support multiple auth methods here...
		passwdDb := new(auth.FlatFile)
		passwdDb.SetFilePath(fileName)
		err = passwdDb.AddUser(passwd, userName, hostNames)
		if err != nil {
			log.Fatal(fmt.Errorf("nsddynum: %w", err))
		}
	case "help":

		var msg string

		if len(os.Args) <= 2 {
			// NOTE: need to print the sub-command specific usage message if a sub-command is given as an argument
			// SEE: https://gitlab.com/necheffa/nsddyn/-/issues/100
			msg = "Usage: nsddynum SUB-COMMAND [OPTS]\n" +
				"  nsddynnum is the nsddyn User Manager utility.\n" +
				"\n" +
				"Available SUB-COMMANDs:\n" +
				"help [SUB-COMMAND]\t\t\t\t\tPrints this message if no argument is given, otherwise prints help text for specified SUB-COMMAND.\n" +
				"version\t\t\t\t\t\t\tPrints version information and exits.\n" +
				"adduser [-p PASSWD_FILE ] -u USER HOST1 HOST2\t\tAdds a user to the passwd file.\n" +
				"moduser [-p PASSWD_FILE ] -u USER\t\t\tModifies the password or authorized hostnames for a user account.\n" +
				"deluser [-p PASSWD_FILE ] -u USER\t\t\tRemoves the specified user account from the passwd file.\n"
		} else {
			switch os.Args[2] {
			default:
				log.Fatal(errors.New("nsddynum: error: unrecognized sub-command: " + os.Args[2]))

			case "version":
				msg = "Usage: nsddynum version\n" +
					"\n" +
					"Prints version information and exits.\n"
			case "adduser":
				msg = "Usage: nsddynum adduser [-p PASSWD_FILE ] -u USER HOST1 HOST2\n" +
					"\n" +
					"Adds a user to the passwd file.\n" +
					"\t-p  Optionally, specify an alternative passwd file to edit. See README for default passwd locations.\n" +
					"\t-u Specify the name of the user account to add.\n" +
					"\n" +
					"Following the argument flags, a list of one or more hostnames associated with the user are specified.\n" +
					"These hostnames are A records the user will be permitted to update.\n"
			case "moduser":
				msg = "Usage: nsddynum moduser [-p PASSWD_FILE ] -u USER\n" +
					"\n" +
					"Modifies attributes of the specified user account.\n" +
					"\t-p  Optionally, specify an alternative passwd file to edit. See README for default passwd locations.\n" +
					"\t-u Specify the name of the user account to modify.\n"
			case "deluser":
				msg = "Usage: nsddynum deluser [-p PASSWD_FILE ] -u USER\n" +
					"\n" +
					"Deletes the specified user account.\n" +
					"\t-p  Optionally, specify an alternative passwd file to edit. See README for default passwd locations.\n" +
					"\t-u Specify the name of the user account to delete.\n"
			}
		}
		fmt.Fprintf(os.Stderr, "%s", msg)
	}
}
