/*
   Copyright (C) 2021 Alexander Necheff

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

package util

import (
	"os"
)

const (
	NsddynpasswdPerms    = 0600
	NsddynpasswdPermsStr = "0600"
)

// CheckFilePerms sees if the Unix permission bits set on fileName match the settings given in perm.
// If the permissions are as expected ok == true, otherwise false. On error err != nil.
func CheckFilePerms(fileName string, perm os.FileMode) (ok bool, err error) {
	ok = false
	err = nil

	f, err := os.Open(fileName)
	if err != nil {
		return ok, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return ok, err
	}

	if perm == fi.Mode() {
		ok = true
	}

	return ok, err
}
