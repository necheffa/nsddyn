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
	"bytes"
	"testing"
)

func TestEmptyPasswd(t *testing.T) {
	var buf []byte
	_, err := passwdParse(buf)
	if err == nil {
		t.Error("Expected empty password to fail but it did not.")
	}
}

func TestJustQuotes(t *testing.T) {
	_, err := passwdParse([]byte(`""`))
	if err == nil {
		t.Error("Expected just quotes to fail but it did not.")
	}
}

func TestValidPasswd(t *testing.T) {
	p, err := passwdParse([]byte(`"password"`))
	if err != nil {
		t.Log(err)
		t.Error("Expected well formed password to succeed but it did not.")
	}

	if !bytes.Equal(p, []byte("password")) {
		t.Error("Expected parsed password to be 'password' but got: " + string(p))
	}
}
