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
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func addUser(userName, fileName, password, hostnames string) error {
	fmt.Println("calling addUser with:")
	fmt.Println("username: " + userName)
	fmt.Println("filename: " + fileName)
	fmt.Println("password: " + password)
	fmt.Println("hostnames: " + hostnames)

	// TODO: need to check if the user already exists in the passwd file or not

	// TODO: validate the format of the hostnames, check for conflicts...

	// TODO: wipe memory before exiting this routine, particularly `password`, `hashedPassword`, and `lineBuf`

	passwdFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0640)
	if err != nil {
		return fmt.Errorf("AddUser opening passwd file: %v", err)
	}

	// 8 seems to be the currently recommended computation time, be careful with setting this too high or too low
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 8)

	// allocate our own slice to make life easier.
	// ultimately, we want to retain references to all data so memory can be wiped when we are done.
	size := len([]byte(userName+":")) + len(hashedPassword) + len([]byte(":"+hostnames+"\n"))
	lineBuf := make([]byte, size)
	lineBuf = append(lineBuf, []byte(userName+":")...)
	lineBuf = append(lineBuf, hashedPassword...)
	lineBuf = append(lineBuf, []byte(":"+hostnames+"\n")...)

	// write a line of the form: userName:hashedPassword:permittedHostnames
	_, err = passwdFile.Write(lineBuf)
	if err != nil {
		return fmt.Errorf("AddUser updating passwd file: %v", err)
	}

	return nil
}
