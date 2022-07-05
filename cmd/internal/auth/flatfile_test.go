/*
   Copyright (C) 2020, 2022 Alexander Necheff

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
	"os"
	"sync"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestAuthFileLocks tests to make sure concurrent access to the auth file
// does not result in file corruption.
func TestAuthFileLocks(t *testing.T) {
	passwd, _ := os.CreateTemp(os.TempDir(), "nsddynpasswd")
	defer os.Remove(passwd.Name())

	authFile0 := new(FlatFile)
	authFile0.SetFilePath(passwd.Name())

	authFile1 := new(FlatFile)
	authFile1.SetFilePath(passwd.Name())

	authFile0.AddUser([]byte("password"), "alex0", []string{"host0"})
	authFile1.AddUser([]byte("password"), "alex1", []string{"host1"})

	var wg sync.WaitGroup
	for i := 0; i < 512; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := authFile0.ModUser("alex0", []byte("password"), false, []string{"host00"}, true)
			if err != nil {
				t.Error(err)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := authFile1.ModUser("alex1", []byte("password"), false, []string{"host11"}, true)
			if err != nil {
				t.Error(err)
			}
		}()
	}

	wg.Wait()

	err := authFile0.AuthRequest([]byte("password"), "alex0", []string{"host00"})
	if err != nil {
		t.Error(err)
	}

	err = authFile0.AuthRequest([]byte("password"), "alex1", []string{"host11"})
	if err != nil {
		t.Error(err)
	}
}

// TestModUserPos tests modifying user accounts from multiple mositions in
// the passwd file to ensure the records are not malformed.
func TestModUserPos(t *testing.T) {
	passwd, _ := os.CreateTemp(os.TempDir(), "nsddynpasswd")
	defer os.Remove(passwd.Name())

	authFile := new(FlatFile)
	authFile.SetFilePath(passwd.Name())

	// mod account on empty file
	err := authFile.ModUser("alex", []byte("password"), false, []string{"host1"}, true)
	if err == nil {
		t.Error(err)
	}
	if err.Error() != "ModUser: user account with name alex does not exist." {
		t.Error("Unexpected error: " + err.Error())
	}

	// mod only account in file
	authFile.AddUser([]byte("password"), "alex0", []string{"host1", "host2"})
	err = authFile.ModUser("alex0", []byte("password1"), false, []string{"host11", "host22"}, false)
	if err != nil {
		t.Error(err)
	}
	err = authFile.AuthRequest([]byte("password"), "alex0", []string{"host1"})
	if err != nil {
		t.Error(err)
	}
	err = authFile.ModUser("alex0", []byte("password1"), true, []string{"host11"}, true)
	if err != nil {
		t.Error(err)
	}
	err = authFile.AuthRequest([]byte("password1"), "alex0", []string{"host11"})
	if err != nil {
		t.Error(err)
	}
	err = authFile.AuthRequest([]byte("password1"), "alex0", []string{"host1", "host2", "host11"})
	if err == nil {
		t.Error("Expected the auth request to fail but it did not.")
	}
	if err.Error() != "AuthRequest: failed to match requested hosts for: alex0" {
		t.Error("Unexpected error: " + err.Error())
	}
	err = authFile.AuthRequest([]byte("password"), "alex0", []string{"host11"})
	if err == nil {
		t.Error("Expected the auth request to fail but it did not.")
	}
	if err.Error() != "AuthRequest: crypto/bcrypt: hashedPassword is not the hash of the given password" {
		t.Error("Unexpected error: " + err.Error())
	}

	// mod account from last position

	// mod middle account position
	authFile.DelUser("alex0")
	authFile.DelUser("alex1")
	authFile.AddUser([]byte("password"), "alex0", []string{"host1", "host2"})
	authFile.AddUser([]byte("password"), "alex1", []string{"host1", "host2"})
	authFile.AddUser([]byte("password"), "alex2", []string{"host1", "host2"})
	err = authFile.ModUser("alex1", []byte("password1"), false, []string{"host11", "host22"}, false)
	if err != nil {
		t.Error(err)
	}
	err = authFile.AuthRequest([]byte("password"), "alex1", []string{"host1"})
	if err != nil {
		t.Error(err)
	}
	err = authFile.ModUser("alex1", []byte("password1"), true, []string{"host11"}, true)
	if err != nil {
		t.Error(err)
	}
	err = authFile.AuthRequest([]byte("password1"), "alex1", []string{"host11"})
	if err != nil {
		t.Error(err)
	}
	err = authFile.AuthRequest([]byte("password1"), "alex1", []string{"host1", "host2", "host11"})
	if err == nil {
		t.Error("Expected the auth request to fail but it did not.")
	}
	if err.Error() != "AuthRequest: failed to match requested hosts for: alex1" {
		t.Error("Unexpected error: " + err.Error())
	}
	err = authFile.AuthRequest([]byte("password"), "alex1", []string{"host11"})
	if err == nil {
		t.Error("Expected the auth request to fail but it did not.")
	}
	if err.Error() != "AuthRequest: crypto/bcrypt: hashedPassword is not the hash of the given password" {
		t.Error("Unexpected error: " + err.Error())
	}
	err = authFile.AuthRequest([]byte("password"), "alex0", []string{"host1"})
	if err != nil {
		t.Error(err)
	}
	err = authFile.AuthRequest([]byte("password"), "alex2", []string{"host1"})
	if err != nil {
		t.Error(err)
	}
}

// TestDelUserPos tests deleting user accounts from multiple positions in
// the passwd file to ensure the remaining records are not malformed.
func TestDelUserPos(t *testing.T) {
	passwd, _ := os.CreateTemp(os.TempDir(), "nsddynpasswd")
	defer os.Remove(passwd.Name())

	authFile := new(FlatFile)
	authFile.SetFilePath(passwd.Name())

	authFile.AddUser([]byte("password"), "alex0", []string{"host1", "host2"})
	authFile.AddUser([]byte("password"), "alex1", []string{"host1", "host2"})
	authFile.AddUser([]byte("password"), "alex2", []string{"host1", "host2"})

	// delete account from middle position
	err := authFile.DelUser("alex1")
	if err != nil {
		t.Error(err)
	}
	err = authFile.AuthRequest([]byte("password"), "alex0", []string{"host1"})
	if err != nil {
		t.Error(err)
	}
	err = authFile.AuthRequest([]byte("password"), "alex2", []string{"host1"})
	if err != nil {
		t.Error(err)
	}
	err = authFile.AuthRequest([]byte("password"), "alex1", []string{"host1"})
	if err == nil {
		t.Error("Expected user alex1 to no longer exist but it does.")
	}
	if err.Error() != "AuthRequest: user account does not exist with name: alex1" {
		t.Error("Unexpected error: " + err.Error())
	}

	// delete account from last position
	err = authFile.DelUser("alex2")
	if err != nil {
		t.Error(err)
	}
	err = authFile.AuthRequest([]byte("password"), "alex0", []string{"host1"})
	if err != nil {
		t.Error(err)
	}
	err = authFile.AuthRequest([]byte("password"), "alex2", []string{"host1"})
	if err == nil {
		t.Error("Expected user alex2 to no longer exist but it does.")
	}
	if err.Error() != "AuthRequest: user account does not exist with name: alex2" {
		t.Error("Unexpected error: " + err.Error())
	}

	// delete only account in file
	err = authFile.DelUser("alex0")
	if err != nil {
		t.Error(err)
	}
	err = authFile.AuthRequest([]byte("password"), "alex0", []string{"host1"})
	if err == nil {
		t.Error("Expected user alex0 to no longer exist but it does.")
	}
	if err.Error() != "AuthRequest: user account does not exist with name: alex0" {
		t.Error("Unexpected error: " + err.Error())
	}

	authFile.AddUser([]byte("password"), "alex0", []string{"host1", "host2"})
	authFile.AddUser([]byte("password"), "alex1", []string{"host1", "host2"})
	authFile.AddUser([]byte("password"), "alex2", []string{"host1", "host2"})

	// delete account from first position
	err = authFile.DelUser("alex0")
	if err != nil {
		t.Error(err)
	}
	err = authFile.AuthRequest([]byte("password"), "alex0", []string{"host1"})
	if err == nil {
		t.Error("Expected user alex0 to no longer exist but it does.")
	}
	if err.Error() != "AuthRequest: user account does not exist with name: alex0" {
		t.Error("Unexpected error: " + err.Error())
	}
	err = authFile.AuthRequest([]byte("password"), "alex1", []string{"host1"})
	if err != nil {
		t.Error(err)
	}
	err = authFile.AuthRequest([]byte("password"), "alex2", []string{"host1"})
	if err != nil {
		t.Error(err)
	}
}

func TestUserMod(t *testing.T) {
	passwd, _ := os.CreateTemp(os.TempDir(), "nsddynpasswd")
	defer os.Remove(passwd.Name())

	authFile := new(FlatFile)
	authFile.SetFilePath(passwd.Name())

	err := authFile.AddUser([]byte("password"), "alex", []string{"host1", "host2"})
	if err != nil {
		t.Error(err)
	}

	err = authFile.ModUser("alex", []byte("password"), false, []string{"host3"}, true)
	if err != nil {
		t.Error(err)
	}

	err = authFile.AuthRequest([]byte("password"), "alex", []string{"host3"})
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	err = authFile.ModUser("alex", []byte("newpassword"), true, []string{"host3"}, false)
	if err != nil {
		t.Error(err)
	}

	err = authFile.AuthRequest([]byte("newpassword"), "alex", []string{"host3"})
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	err = authFile.DelUser("alex")
	if err != nil {
		t.Error(err)
	}

	err = authFile.ModUser("alex", []byte("newpassword"), true, []string{"host3"}, false)
	if err == nil {
		t.Errorf("Expected error trying to modify non-existant user but did not recieve one.")
	}
	if err.Error() != "ModUser: user account with name alex does not exist." {
		t.Errorf("Did not get expected error when trying to modify non-existant user. Got: %v", err.Error())
	}
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
	oneBuf.Truncate(oneSize)
	ok = userExists("alex", oneBuf.Bytes())
	if ok {
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

	// TODO: programmatically determine expected size
	if twoSize != 72 {
		t.Errorf("Expected: [%v] but got: [%v]", 72, twoSize)
	}
	twoBuf.Truncate(twoSize)
	ok = userExists("alex", twoBuf.Bytes())
	if ok {
		t.Errorf("Expected: false but got: true")
		t.Errorf("\n%s", string(twoBuf.Bytes()))
	}
	ok = userExists("yolo", twoBuf.Bytes())
	if !ok {
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
