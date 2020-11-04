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
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// MockFile wraps a byte slice and implements the Seeker, Reader, and Writer interfaces.
// This allows it to MockFiles for unit testing without acutally invoking the filesystem.
type MockFile struct {
	buf    []byte // slice backing the MockFile
	curPos int    // current position into slice b, effectively a file offset
}

// Returns the contents of the MockFile as a []byte.
// Similar to bytes.Buffer.Bytes() for bytes.Buffer
func (mf *MockFile) Bytes() []byte {
	return mf.buf
}

func (mf *MockFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	default:
		err := fmt.Errorf("MockFile: invalid whence mode specified: %v", whence)
		return int64(mf.curPos), err
	case io.SeekStart:
		// from start of file
		if offset < 0 {
			err := fmt.Errorf("MockFile: can't seek backwards beyond start of file, offset was: %v", offset)
			return int64(mf.curPos), err
		}

		// any positive offset is valid, but beyond the end of the file is implementation defined.
		// so, we won't stop callers from causing a segmentation fault, that is just "implementation defined behavior" ;-)
		mf.curPos = int(offset)
		return int64(mf.curPos), nil
	case io.SeekCurrent:
		// from current mf.curPos offset

		if mf.curPos+int(offset) < 0 {
			err := fmt.Errorf("MockFile: can't seek backwards beyond start of file, offset was: %v", offset)
			return int64(mf.curPos), err
		}
		// any positive offset is valid, but beyond the end of the file is implementation defined.
		// so, we won't stop callers from causing a segmentation fault, that is just "implementation defined behavior" ;-)
		mf.curPos += int(offset)
		return int64(mf.curPos), nil
	case io.SeekEnd:
		// from end of file
		if len(mf.buf)+int(offset) < 0 {
			err := fmt.Errorf("MockFile: can't seek backwards beyond start of file, offset was: %v", offset)
			return int64(mf.curPos), err
		}
		// any positive offset is valid, but beyond the end of the file is implementation defined.
		// so, we won't stop callers from causing a segmentation fault, that is just "implementation defined behavior" ;-)
		mf.curPos = len(mf.buf) + int(offset)
		return int64(mf.curPos), nil
	}
	// default case above handles return "here"
}

func (mf *MockFile) Read(p []byte) (n int, err error) {
	if len(p)+mf.curPos > len(mf.buf) {
		// only read what is left in mf.buf
		n = copy(p, mf.buf[mf.curPos:])
		mf.curPos = len(mf.buf)
		return n, io.EOF
	}

	if mf.curPos >= len(mf.buf) {
		// we are already at EOF, can't read any more
		return 0, io.EOF
	}

	// just do the read then, assume len(p) < len(mf.buf) and mf.curPos
	n = copy(p, mf.buf[mf.curPos:])
	mf.curPos += n
	return n, nil
}

func (mf *MockFile) Write(p []byte) (n int, err error) {
	// TODO: technically, Write() should return a non-nil err when n != len(p)
	// but for now that is above and beyond what MockFile needs to do.

	mf.buf = append(mf.buf, p...)
	mf.curPos += len(p)
	return len(p), nil
}

func (mf *MockFile) Rewind() {
	// ultra cheap mf.Seek(0, io.SeekStart) :-D
	mf.curPos = 0
}

// TestUserExists() mocks a nsddynpasswd file to isolate testing to the userExists() function.
func TestUserExists(t *testing.T) {
	var emptyBuf []byte
	if userExists("alex", emptyBuf) {
		t.Error("userExists reported a user exists in an empty passwd store buffer.")
	}

	singleBuf := []byte("alex:hash:hosts")
	if userExists("joe", singleBuf) {
		t.Error("userExists reported a user exists when it does not.")
	}

	if !userExists("alex", singleBuf) {
		t.Error("userExists did not find a user when it should have.")
	}

	multiBuf := []byte("joe:hash:host\nalex:hash:host")
	if userExists("fred", multiBuf) {
		t.Error("userExists reported a user exists when it does not.")
	}

	if !userExists("alex", multiBuf) {
		t.Error("userExists did not find a user when it should have.")
	}
}

// TestHostSize tests the hostSize() function.
func TestHostSize(t *testing.T) {
	var noHost []string
	if hostSize(noHost) != 0 {
		t.Error("hostSize failed on an empty host list.")
	}

	oneHost := []string{"host1"}
	if hostSize(oneHost) != 5 {
		t.Error("hostSize failed on single entry list.")
	}

	twoHost := []string{"host1", "host2"}
	if hostSize(twoHost) != 11 {
		t.Error("hostSize failed on two entry list.")
	}

	listOfEmpty := []string{"", ""}
	if hostSize(listOfEmpty) != 0 {
		t.Error("hostSize failed on a list of empty strings.")
	}
}

func TestHostsToBytes(t *testing.T) {
	emptyList := []string{}
	buf := hostsToBytes(emptyList)

	if !bytes.Equal(buf, []byte("")) {
		t.Errorf("Expected empty buf but got: %v", string(buf))
	}

	oneList := []string{"host1"}
	buf = hostsToBytes(oneList)
	if !bytes.Equal(buf, []byte("host1")) {
		t.Errorf("Expected: %v but got: %v", "host1", string(buf))
	}

	twoList := []string{"host1", "host2"}
	buf = hostsToBytes(twoList)
	if !bytes.Equal(buf, []byte("host1,host2")) {
		t.Errorf("Expected: %v but got: %v", "host1,host2", string(buf))
	}
}

// TestDelUser mocks a nsddynpasswd file in order to test the flatfile.DelUser method.
// We can't test DelUser directly, but we can test the internal delUser Method.
func TestDelUser(t *testing.T) {
	passwd := new(FlatFile)
	emptyBuf := new(MockFile)

	// test deleting a user from an empty passwd file
	emptySize, err := passwd.delUser("username", emptyBuf)
	if err == nil {
		t.Error(err)
	}
	expected1 := "DelUser: user account with name username does not exist"
	if err.Error() != expected1 {
		t.Errorf("Expected: [%s] but got: [%v]", expected1, err)
	}
	if emptySize != 0 {
		t.Errorf("Expected: [%v] but got: [%v]", 0, emptySize)
	}

	ok := userExists("username", emptyBuf.Bytes())
	if ok {
		t.Errorf("Expected: false but got: true")
	}

	// test deleting the only user in a passwd file
	oneBuf := new(MockFile)
	oneBuf.buf = make([]byte, 0, 8)
	passwd.addUser([]byte("password"), "alex", []string{"host1"}, oneBuf)
	oneBuf.Rewind()
	ok = userExists("alex", oneBuf.Bytes())
	if !ok {
		t.Errorf("Expected: true but got: false")
	}
	oneSize, err := passwd.delUser("alex", oneBuf)
	if err != nil {
		t.Error(err)
	}
	if oneSize != 0 {
		t.Errorf("Expected: [%v] but got: [%v]", 0, oneSize)
	}
	ok = userExists("alex", oneBuf.Bytes())
	if !ok {
		t.Errorf("Expected: false but got: true")
	}

	// test deleting the user from a passwd file with more than one account in it
	twoBuf := new(MockFile)
	twoBuf.buf = make([]byte, 0, 8)
	passwd.addUser([]byte("password"), "alex", []string{"host1"}, twoBuf)
	passwd.addUser([]byte("password"), "yolo", []string{"host1"}, twoBuf)
	twoBuf.Rewind()
	ok = userExists("alex", twoBuf.Bytes())
	if !ok {
		t.Errorf("Expected: true but got: false")
	}
	twoSize, err := passwd.delUser("alex", twoBuf)
	if err != nil {
		t.Error(err)
	}
	if twoSize != 1 {
		t.Errorf("Expected: [%v] but got: [%v]", 1, twoSize)
	}
	ok = userExists("alex", twoBuf.Bytes())
	if !ok {
		t.Errorf("Expected: false but got: true")
	}
	ok = userExists("yolo", twoBuf.Bytes())
	if ok {
		t.Errorf("Expected: true but got: false")
	}

	// test deleting the last user from a passwd file with more than one account in it

	// test deleting a middle user from a passwd file with more than two accounts in it
}

// TestAddUser mocks a nsddynpasswd file in order to test the flatfile.AddUser method.
// We can't test Adduser directly, but we can test the internal addUser method.
func TestAddUser(t *testing.T) {
	passwd := new(FlatFile)
	emptyBuf := bytes.NewBufferString("")

	err := passwd.addUser([]byte("password"), "alex", []string{"host1"}, emptyBuf)
	if err != nil {
		t.Error(err)
	}

	oneBuf := bytes.NewBufferString("alex:hash:host1")
	err = passwd.addUser([]byte("password"), "alex", []string{"host1"}, oneBuf)
	expected1 := "AddUser: user account already exists with name: alex"
	if err.Error() != expected1 {
		t.Errorf("Expected: [%s], but got: [%v]", expected1, err)
	}
	// TODO test that oneBuf is unchanged

	err = passwd.addUser([]byte("password"), "joe", []string{"host"}, oneBuf)
	if err != nil {
		t.Error(err)
	}
	// TODO test that oneBuf was updated with a new entry for joe

	twoBuf := bytes.NewBufferString("dave:hash:host1,host2\nalex:hash:host3,host4")
	err = passwd.addUser([]byte("password"), "alex", []string{"host1", "host2"}, twoBuf)
	if err.Error() != expected1 {
		t.Errorf("Expected: [%s], but got: [%v]", expected1, err)
	}

	err = passwd.addUser([]byte("password"), "joe", []string{"host"}, twoBuf)
	if err != nil {
		t.Error(err)
	}
}

// TestAuthRequest mocks a nsddynpasswd file to test the flatfile.AuthRequest method.
func TestAuthRequest(t *testing.T) {
	passwd := new(FlatFile)
	buf := bytes.NewBufferString("")

	// need to use bytes.Reader to get a Seekable type, otherwise if buf is left as
	// just a bytes.Buffer we can only read from it once.
	bufReader := bytes.NewReader(buf.Bytes())

	// test with empty passwd file
	err := passwd.authRequest([]byte("password"), "alex", []string{"host1"}, bufReader)
	expected := "AuthRequest: user account does not exist with name: alex"
	if err.Error() != expected {
		t.Errorf("Expected: [%s], but got: [%v]", expected, err)
	}

	// test with single entry passwd file
	err = passwd.addUser([]byte("password"), "alex", []string{"host1", "host2"}, buf)
	if err != nil {
		t.Error(err)
	}

	bufReader = bytes.NewReader(buf.Bytes())
	bufReader.Seek(io.SeekStart, 0)

	// account exists, password matches, host good
	err = passwd.authRequest([]byte("password"), "alex", []string{"host1"}, bufReader)
	if err != nil {
		t.Error(err)
	}

	bufReader.Seek(io.SeekStart, 0)

	// account exists, password bad, host good
	err = passwd.authRequest([]byte("badpassword"), "alex", []string{"host1"}, bufReader)
	expected = "AuthRequest: " + bcrypt.ErrMismatchedHashAndPassword.Error()
	if err.Error() != expected {
		t.Errorf("Expected: [%s], but got: [%v]", expected, err)
	}

	bufReader.Seek(io.SeekStart, 0)

	// account exists, password good, host bad
	err = passwd.authRequest([]byte("password"), "alex", []string{"badhost1"}, bufReader)
	expected = "AuthRequest: failed to match requested hosts for: alex"
	if err.Error() != expected {
		t.Errorf("Expected: [%s], but got: [%v]", expected, err)
	}

	bufReader.Seek(io.SeekStart, 0)

	// account does not exist
	err = passwd.authRequest([]byte("password"), "baduser", []string{"host1"}, bufReader)
	expected = "AuthRequest: user account does not exist with name: baduser"
	if err.Error() != expected {
		t.Errorf("Expected: [%s], but got: [%v]", expected, err)
	}

	// TODO: test with two entry passwd file
}

func TestGetHashAndHosts(t *testing.T) {
	passwd := new(FlatFile)
	buf := bytes.NewBufferString("")

	err := passwd.addUser([]byte("password"), "alex", []string{"host1", "host2", "host3"}, buf)
	if err != nil {
		t.Error(err)
	}

	bufReader := bytes.NewReader(buf.Bytes())
	fileBuf, err := ioutil.ReadAll(bufReader)
	if err != nil {
		t.Error(err)
	}

	// test user account doesn't exist
	hash, host, err := getHashAndHosts("joe", fileBuf)
	expected := "getHashAndHosts: failed to find account information for: joe"
	if err.Error() != expected {
		t.Errorf("Expected [%s], but got: [%v]", expected, err)
	}
	if hash != nil {
		t.Errorf("Expected hash to be nil but got: %v", hash)
	}
	if host != nil {
		t.Errorf("Expected host to be nil but got: %v", host)
	}
}
