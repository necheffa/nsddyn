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

// The util package contains miscellaneous utility functions.
package util

import (
	"fmt"
	"os"
)

const (
	DefaultHome = "/usr/local"
)

// FindHome locates and returns the base configuration directory for nsddyn binaries.
// This is the parent directory of configuration files and other resource files.
// First FindHome will attempt to evaluate the environment variable $NSDDYN_HOME,
// if this environment variable was not exported or FindHome otherwise was unable to
// read it then FindHome will attempt to default to /usr/local/.
// If basic directory existance and is-a-directory checks fail, err != nil is returned.
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
