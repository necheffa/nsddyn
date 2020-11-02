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

package util

import (
	"fmt"
	"os"
)

const (
	DefaultHome = "/usr/local"
)

func FindHome() (nsddynHome string, err error) {
	nsddynHome, ok := os.LookupEnv("NSDDYN_HOME")

	if !ok {
		nsddynHome = DefaultHome
	}

	fi, err := os.Stat(nsddynHome)
	if os.IsNotExist(err) {
		err = fmt.Errorf("FindHome: Error: $NSDDYN_HOME set to non-existent location.")
		nsddynHome = ""
	} else {
		if !fi.IsDir() {
			err = fmt.Errorf("FindHome: Error: $NSDDYN_HOME is not set to a directory.")
			nsddynHome = ""
		}
	}

	return nsddynHome, err
}
