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
	"regexp"
)

// passwdParse parses a json.RawMessage to confirm it is a welformed password.
// A welformed password is of the form: "password".
// On success, a password without enclosing quotes is returned along with err == nil,
// otherwise, err != nil.
func passwdParse(buf []byte) (b []byte, err error) {
	// We require at least a single character...
	matched, err := regexp.Match(`".+"`, buf)
	if err != nil {
		return nil, fmt.Errorf("passwdParse: %v", err)
	}

	if !matched {
		return nil, fmt.Errorf("passwdParse: invalid password field")
	}

	// trim enclosing quotes but as a derived slice not a copy.
	// this way, a defered erase of buf by the caller erases b too.
	b = buf[1 : len(buf)-1]

	return b, nil
}
