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
	"io"
	"io/ioutil"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

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
