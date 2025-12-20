/*
   Copyright (C) 2020, 2025 Alexander Necheff

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
	"io"
	"testing"
)

const (
	Hello           = "Hello World!"
	HelloHello      = "Hello World!Hello World!"
	HelloHelloHello = "Hello World!Hello World!Hello World!"
	Spaces12        = "            "
)

func TestMockSeek(t *testing.T) {
	mf := NewMockFile()

	buf12 := make([]byte, 12)
	_, _ = mf.Write([]byte(HelloHello))
	// HelloHello
	mf.Seek(0, io.SeekStart)
	_, _ = mf.Read(buf12)
	mf.Seek(0, io.SeekEnd)
	_, _ = mf.Write([]byte(Hello))
	// HelloHelloHello
	mf.Seek(-12, io.SeekCurrent)
	_, _ = mf.Write([]byte(HelloHelloHello))
	// HelloHelloHelloHelloHello
	_, _ = mf.Write([]byte(Hello))
	// HelloHelloHelloHelloHelloHello
	mf.Seek(0, io.SeekStart)
	buf24 := make([]byte, 24)
	_, _ = mf.Read(buf24)
	mf.Seek(12, io.SeekCurrent)
	_, _ = mf.Write([]byte(Spaces12))
	// HelloHelloHello     HelloHello

	expected := []byte("Hello World!Hello World!Hello World!            Hello World!Hello World!")
	if !bytes.Equal(expected, mf.Bytes()) {
		t.Errorf("After MockFile.Seek(), expected \"%v\" but got \"%v\"", string(expected), string(mf.Bytes()))
	}

	_, err := mf.Seek(0, 99)
	if err == nil {
		t.Error("Expected non-nill error for invalid whence during MockFile.Seek()")
	}
	_, err = mf.Seek(0, -99)
	if err == nil {
		t.Error("Expected non-nill error for invalid whence during MockFile.Seek()")
	}

	_, err = mf.Seek(-99, io.SeekEnd)
	if err == nil {
		t.Errorf("Expected error to be raised when attempting to MockFile.Seek() backwards past the start of the MockFile.")
	}
	_, err = mf.Seek(-99, io.SeekStart)
	if err == nil {
		t.Errorf("Expected error to be raised when attempting to MockFile.Seek() backwards past the start of the MockFile.")
	}
	_, err = mf.Seek(-99, io.SeekCurrent)
	if err == nil {
		t.Errorf("Expected error to be raised when attempting to MockFile.Seek() backwards past the start of the MockFile.")
	}
	_, err = mf.Seek(99, io.SeekEnd)
	if err != nil {
		t.Errorf("Expected to be permitted to MockFile.Seek() past the end of the MockFile without raising an error.")
	}
}

func TestMockRead(t *testing.T) {
	mf := NewMockFile()

	_, _ = mf.Write([]byte(HelloHello))

	buf := make([]byte, 12)
	n, err := mf.Read(buf)
	if err == nil {
		t.Errorf("Expected io.EOF error but got nil")
	}
	if err != io.EOF {
		t.Errorf("Expected io.EOF error but got %v", err)
	}
	if n != 0 {
		t.Errorf("Expected to read 0 bytes but return value indicated %v", n)
	}

	// intentional duplication of above code to handle reading again once we are already at EOF
	n, err = mf.Read(buf)
	if err == nil {
		t.Errorf("Expected io.EOF error but got nil")
	}
	if err != io.EOF {
		t.Errorf("Expected io.EOF error but got %v", err)
	}
	if n != 0 {
		t.Errorf("Expected to read 0 bytes but return value indicated %v", n)
	}

	mf.Seek(0, io.SeekStart)
	n, err = mf.Read(buf)
	if err != nil {
		t.Errorf("Expected success but got non-nil err %v", err)
	}
	if n != 12 {
		t.Errorf("Expected to read 12 bytes but return value indicated %v", n)
	}
	if !bytes.Equal(buf, []byte(Hello)) {
		t.Errorf("Expected to read \"%v\" but got \"%v\"", Hello, string(buf))
	}
}

func TestMockWrite(t *testing.T) {
	mf := NewMockFile()

	n, err := mf.Write([]byte(Hello))
	if err != nil {
		t.Errorf("%s",err.Error())
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
		t.Errorf("%s", err.Error())
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
