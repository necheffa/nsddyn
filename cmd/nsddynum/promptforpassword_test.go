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

package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"necheff.net/nsddyn/cmd/internal/auth"
)

func TestValidPassword(t *testing.T) {
	t.Skip("skipping promptForPasswd valid password test")
	file, _ := os.CreateTemp(os.TempDir(), "fakestdin")
	defer os.Remove(file.Name())

	_, err := io.WriteString(file, "thepassword\n")
	if err != nil {
		t.Error(err)
	}

	_, err = promptForPasswd(file)
	if err != nil {
		t.Error(err)
	}
}

func TestShortPassword(t *testing.T) {
	t.Skip("skipping promptForPasswd short password test")
	file, _ := os.CreateTemp(os.TempDir(), "fakestdin")
	defer os.Remove(file.Name())

	_, err := io.WriteString(file, "short\n")
	if err != nil {
		t.Error(err)
	}

	_, err = promptForPasswd(file)
	if err != fmt.Errorf("promptForPasswd: Error: password length less than minimum length of %v", auth.MinPasswd) {
		t.Error(err)
	}
}

func TestLongPassword(t *testing.T) {
	t.Skip("skipping promptForPasswd long password test")
	file, _ := os.CreateTemp(os.TempDir(), "fakestdin")
	defer os.Remove(file.Name())

	long := strings.Repeat("a", auth.MaxPasswd+1)

	_, err := io.WriteString(file, long+"\n")
	if err != nil {
		t.Error(err)
	}

	_, err = promptForPasswd(file)
	if err != fmt.Errorf("promptForPasswd: Error: password length exceeds max password length of %v", auth.MaxPasswd) {
		t.Error(err)
	}
}
