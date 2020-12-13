/*
   Copyright (C) 2020 Alexander Necheff

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

	"necheff.net/nsddyn/cmd/internal/auth"

	"golang.org/x/crypto/ssh/terminal"
)

// promptForPasswd reads a plaintext password from STDIN without echoing it on the terminal.
// Either the password is returned as a []byte or err == nil.
func promptForPasswd() (passwd []byte, err error) {
	fmt.Fprintf(os.Stderr, "new password: ")
	passwd, err = terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		err = fmt.Errorf("promptForPasswd: Error reading password: %v", err)
		return nil, err
	}

	// TODO: there has to be a better way to handle this since the whole point is to avoid filling memory
	// with a maliciously crafted password...
	if len(passwd) > auth.MaxPasswd {
		err = fmt.Errorf("promptForPasswd: Error: password length exceeds max password length of %v", auth.MaxPasswd)
		passwd = nil
	} else if len(passwd) < auth.MinPasswd {
		err = fmt.Errorf("promptForPasswd: Error: password length less than minimum length of %v", auth.MinPasswd)
		passwd = nil
	}

	// we need to advance the prompt because we successfully consumed the \n
	fmt.Fprintf(os.Stderr, "\n")
	return passwd, err
}
