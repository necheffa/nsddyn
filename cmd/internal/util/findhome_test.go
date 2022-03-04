/*
   Copyright (C) 2022 Alexander Necheff

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
	"testing"
)

const NHOME = "NSDDYN_HOME"

func TestDefaultHome(t *testing.T) {
	os.Unsetenv(NHOME)

	h, err := FindHome()
	if err != nil {
		t.Error(err)
	}

	if h != "/usr/local" {
		t.Error("Default $NSDDYN_HOME not /usr/local")
	}
}

func TestHomeNotDir(t *testing.T) {
	os.Unsetenv(NHOME)

	os.Setenv(NHOME, "/dev/zero")

	_, err := FindHome()
	if err == nil {
		t.Error("Expected FindHome to fail but err == nil")
	}

	if err.Error() != "FindHome: Error: $NSDDYN_HOME is not set to a directory." {
		t.Log(err)
		t.Error("Unexpected error when $NSDDYN_HOME not a directory")
	}
}

func TestHomeNotExist(t *testing.T) {
	os.Unsetenv(NHOME)

	os.Setenv(NHOME, "/does/not/exist")

	_, err := FindHome()
	if err == nil {
		t.Error("Expected FindHome to fail but err == nil")
	}

	if err.Error() != "FindHome: Error: $NSDDYN_HOME set to non-existent location." {
		t.Log(err)
		t.Error("Unexpected error when $NSDDYN_HOME not a directory")
	}
}
