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

package auth

import (
	"bytes"
	"testing"
)

const (
	Hello      = "Hello World!"
	HelloHello = "Hello World!Hello World!"
)

func TestMockWrite(t *testing.T) {
	mf := NewMockFile()

	n, err := mf.Write([]byte(Hello))
	if err != nil {
		t.Errorf(err.Error())
	}
	if n != len([]byte(Hello)) {
		t.Errorf("Expected write length of \"%s\" to be %v but got %v", Hello, len([]byte(Hello)), n)
	}
	if mf.curPos != len([]byte(Hello)) {
		t.Errorf("Expected current offset to be %v after writing \"%s\" but got %v", len([]byte(Hello)), Hello, mf.curPos)
	}
	if !bytes.Equal(mf.Bytes(), []byte(Hello)) {
		t.Errorf("Not all bytes for string \"%s\" found in MockFile buffer.", Hello)
	}

	n, err = mf.Write([]byte(Hello))
	if err != nil {
		t.Errorf(err.Error())
	}
	if n != len([]byte(Hello)) {
		t.Errorf("Expected write length of \"%s\" to be %v but got %v", Hello, len([]byte(Hello)), n)
	}
	if mf.curPos != len([]byte(HelloHello)) {
		t.Errorf("Expected current offset to be %v after writing \"%s\" but got %v", len([]byte(HelloHello)), HelloHello, mf.curPos)
	}
	if !bytes.Equal(mf.Bytes(), []byte(HelloHello)) {
		t.Errorf("Not all bytes for string \"%s\" found in MockFile buffer.", HelloHello)
	}
}
