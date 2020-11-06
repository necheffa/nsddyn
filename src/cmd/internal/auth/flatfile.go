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
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

/*
  The flat-file auth mechanism is a simple text file with a format reminicent of /etc/passwd.
  That is, each line is a "record" and each record contains fields seporated by ":".

  This auth mechanism is a great way to get started, the implementation is easy to reason about
  and can be quickly debugged with a humble text editor and the appropriate group membership on the system.
  However, as the number of users increases, performance suffers. Particularlly because the backing file
  is opened and closed for each auth action in order to avoid buffering the hashed passwords
  for longer than is needed to update or consult with the contents of the file.

  This design may be improved in a future release, for now this design is good enough for the small
  user base that is expected initially.
*/

// FlatFile satisfies the AuthReadWriter interface.
type FlatFile struct {
	// The path to the file backing this authentication method
	filePath string
}

func (f *FlatFile) SetFilePath(filePath string) {
	f.filePath = filePath
}

// AuthRequest attempts to authenticate the passed in data against the password store.
// Returns err != nil on error.
//
// Note that it is the caller's responsibility to handle safely erasing passwd on exit.
// All internally allocated memory is properly sanitized to avoid leaking the passwd.
func (f *FlatFile) AuthRequest(passwd []byte, userName string, hosts []string) (err error) {
	file, err := os.OpenFile(f.filePath, os.O_RDONLY, 0640)
	if err != nil {
		return fmt.Errorf("AuthRequest: %v", err)
	}
	defer file.Close()

	return f.authRequest(passwd, userName, hosts, file)
}

func (f *FlatFile) authRequest(passwd []byte, userName string, hosts []string, file io.Reader) (err error) {
	fileBuf, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("AuthRequest: %v", err)
	}
	defer EraseBuf(fileBuf)

	ok := userExists(userName, fileBuf)
	if !ok {
		return fmt.Errorf("AuthRequest: user account does not exist with name: %v", userName)
	}

	storedHash, storedHosts, err := getHashAndHosts(userName, fileBuf)
	defer EraseBuf(storedHash)
	if err != nil {
		return fmt.Errorf("AuthRequest: failed to retrieve account information for: %v", userName)
	}
	err = bcrypt.CompareHashAndPassword(storedHash, passwd)
	if err != nil {
		// don't be deceaved, this could be a bad password or a more general fault
		return fmt.Errorf("AuthRequest: %v", err)
	}

	ok = hostsMatch(storedHosts, hosts)
	if !ok {
		return fmt.Errorf("AuthRequest: failed to match requested hosts for: %v", userName)
	}

	err = nil
	return
}

// Retreives the stored password hash and permitted hosts for the given user or err != nil.
func getHashAndHosts(userName string, fileBuf []byte) (storedHash []byte, storedHosts []string, err error) {
	lines := bytes.Split(fileBuf, []byte("\n"))
	for _, line := range lines {
		field := bytes.Split(line, []byte(":"))
		if bytes.Equal(field[0], []byte(userName)) {
			storedHash = field[1]
			storedHosts = strings.Split(string(field[2]), ",")
			err = nil
			return
		}
	}
	storedHash = nil
	storedHosts = nil
	err = fmt.Errorf("getHashAndHosts: failed to find account information for: %v", userName)
	return
}

// Confirms that the requested host updates are contained in the stored hosts on the passwd file.
func hostsMatch(storedHosts []string, hosts []string) (ok bool) {
	ok = false

	// This has a somewhat ugly time complexity but host lists should be small and besides this way
	// we can allow partial requests. Shouldn't be a problem.
	for _, rhost := range hosts {
		thisHostMatched := false
		for _, shost := range storedHosts {
			if shost == rhost {
				thisHostMatched = true
			}
		}
		if !thisHostMatched {
			// first mis-match fails
			return
		}
	}

	// if we made it this far, we are ok
	ok = true
	return
}

// AddUser attempts to add the given user account information to the password store.
// Returns err != nil on error, this includes the case where userName already exists in the store.
//
// Note that it is the caller's responsibility to handle safely erasing passwd on exit.
// All internally allocated memory is properly sanitized to avoid leaking the passwd.
func (f *FlatFile) AddUser(passwd []byte, userName string, hosts []string) (err error) {
	file, err := os.OpenFile(f.filePath, os.O_APPEND|os.O_RDWR, 0640)
	if err != nil {
		return fmt.Errorf("AddUser: %v", err)
	}
	defer file.Close()

	return f.addUser(passwd, userName, hosts, file)
}

// addUser is a private helper funcion which allows the mocking of the passwd file for unit testing.
func (f *FlatFile) addUser(passwd []byte, userName string, hosts []string, file io.ReadWriter) (err error) {
	fileBuf, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("AddUser: %v", err)
	}
	defer EraseBuf(fileBuf)

	ok := userExists(userName, fileBuf)
	if ok {
		return fmt.Errorf("AddUser: user account already exists with name: %v", userName)
	}

	// Note: we go through a lot of extra hassle to use a []byte instead of bytes.Buffer here,
	// this is intentional, in order to retain control over the allocation so that _all_ the
	// memory can be wiped when we are done.

	// as of 2020-03-08 a cost of 8 seems to be an acceptable compromise
	// between performance and security.
	hashedPasswd, err := bcrypt.GenerateFromPassword(passwd, 8)
	if err != nil {
		return fmt.Errorf("AddUser: %v", err)
	}
	defer EraseBuf(hashedPasswd)

    size := len(userName+":") + len(hashedPasswd) + len(":") + hostSize(hosts) + len("\n")
	lineBuf := make([]byte, 0, size)
	defer EraseBuf(lineBuf)

	// because we manually calculated enough storage for the lineBuf byte slice, we can
	// safely call append without worrying about it reallocating lineBuf behind our backs,
	// leaving bits of the new entry, particuarlly the hashed password, floating around
	// in memory.
	lineBuf = append(lineBuf, []byte(userName+":")...)
	lineBuf = append(lineBuf, hashedPasswd...)
	lineBuf = append(lineBuf, []byte(":")...)
	lineBuf = append(lineBuf, hostsToBytes(hosts)...)
	lineBuf = append(lineBuf, []byte("\n")...)

	// as long as we opened with os.O_APPEND we don't need to file.Seek() to EOF first.
	// file.Write() returns err != nil when we have a short write so we don't really need
	// to explicitly catch the returned number of bytes written.
	//
	// TODO: currently, we assume that file.Write() passes a pointer to our lineBuf all
	// the way down to the kernel write() call and by wiping lineBuf here we have really
	// wiped it. It would be nice to confirm this is the case.
	/*
	   _, err = file.Seek(0, io.SeekEnd) // needed for unit tests that mock a real file as a buffer.
	   if err != nil {
	       return fmt.Errorf("Adduser: %v", err)
	   }
	*/
	_, err = file.Write(lineBuf)
	if err != nil {
		return fmt.Errorf("AddUser: %v", err)
	}

	return nil
}

// DelUser attempts to remove the given user account information from the password store.
// Returns err != nil on err, this includes the case where userName does not exist in the store.
func (f *FlatFile) DelUser(userName string) (err error) {
	file, err := os.OpenFile(f.filePath, os.O_RDWR, 0640)
	if err != nil {
		return fmt.Errorf("DelUser: %v", err)
	}
	defer file.Close()

	size, err := f.delUser(userName, file)
	if err == nil {
		file.Truncate(size)
		file.Sync()
	}

	return err
}

// delUser is a private helper function which allows the mocking of the passwd file for unit testing.
func (f *FlatFile) delUser(userName string, file io.ReadWriteSeeker) (newSize int64, err error) {
	fileBuf, err := ioutil.ReadAll(file)
	if err != nil {
		return 0, fmt.Errorf("DelUser: %v", err)
	}
	defer EraseBuf(fileBuf)

	ok := userExists(userName, fileBuf)
	if !ok {
		return 0, fmt.Errorf("DelUser: user account with name %v does not exist", userName)
	}

	// note that userExists() already validated if we have an empty password store or not
	newFileBuf := removeUserRecord(userName, fileBuf)
	defer EraseBuf(newFileBuf)

	file.Seek(0, io.SeekStart)
	file.Write(newFileBuf)

	return int64(len(newFileBuf)), nil
}

// ModUser attempts to modify the password or authorized hosts list for the desired user account.
// Returns err != nil on error.
func (f *FlatFile) ModUser(userName string, passwd []byte, modPasswd bool, hostNames []string, modHostNames bool) (err error) {
	file, err := os.OpenFile(f.filePath, os.O_RDWR, 0640)
	if err != nil {
		return fmt.Errorf("ModUser: %v", err)
	}
	defer file.Close()

	size, err := f.modUser(userName, passwd, modPasswd, hostNames, modHostNames, file)
	if err == nil {
		file.Truncate(size)
		file.Sync()
	}

	return err
}

// modUser works by atomically deleting the existing user record and
// adding a new one with the updated information or
// fails and no file modifications are made.
func (f *FlatFile) modUser(userName string, passwd []byte, modPasswd bool, hostNames []string, modHostNames bool, file io.ReadWriteSeeker) (newFileSize int64, err error) {
	fileBuf, err := ioutil.ReadAll(file)
	if err != nil {
		return 0, fmt.Errorf("ModUser: %v", err)
	}
	defer EraseBuf(fileBuf)

	userRecord, ok := getUserRecord(userName, fileBuf)
	if !ok {
		err = fmt.Errorf("ModUser: user account with name %v does not exist.", userName)
		return -1, err
	}

	var hostList []byte
	var passwdHash []byte
	defer EraseBuf(passwdHash)

	fields := bytes.Split(userRecord, []byte(":"))

	// we are either generating a new password hash, or keeping the existing one.
	if modPasswd {
		passwdHash, err = bcrypt.GenerateFromPassword(passwd, 8)
		if err != nil {
			return -1, fmt.Errorf("ModUser: %v", err)
		}
	} else {
		passwdHash = fields[1]
	}

	// we are either updating the authorized hosts list, or keeping the existing one.
	if modHostNames {
		hostList = hostsToBytes(hostNames)
	} else {
		hostList = fields[2]
	}

	// to ensure an atomic update, we'll work with a copy of the file buffer before updating the disk.
	newFileBuf := make([]byte, len(fileBuf))
	defer EraseBuf(newFileBuf)
	n := copy(newFileBuf, fileBuf)
	if n != len(fileBuf) {
		return -1, fmt.Errorf("ModUser: unable to atomically update password store.")
	}

	// getUserRecord() above already validated that userName exists
	removedAccountFileBuf := removeUserRecord(userName, newFileBuf)
	defer EraseBuf(removedAccountFileBuf)

	newRecordSize := len(userName) + len(":") + len(passwdHash) + len(":") + hostSize(hostNames) + len("\n")
	newUserRecord := make([]byte, 0, newRecordSize)
	defer EraseBuf(newUserRecord)
	newUserRecord = append(newUserRecord, []byte(userName+":")...)
	newUserRecord = append(newUserRecord, passwdHash...)
	newUserRecord = append(newUserRecord, []byte(":")...)
	newUserRecord = append(newUserRecord, hostList...)
	newUserRecord = append(newUserRecord, []byte("\n")...)

	finalFileBuf := make([]byte, 0, len(removedAccountFileBuf)+newRecordSize)
	defer EraseBuf(finalFileBuf)
	n = copy(finalFileBuf, removedAccountFileBuf)
	if n != len(removedAccountFileBuf) {
		return -1, fmt.Errorf("ModUser: unable to atomically update password store.")
	}
	finalFileBuf = append(finalFileBuf, newUserRecord...)

	file.Seek(0, io.SeekStart)
	file.Write(finalFileBuf)

	return int64(len(finalFileBuf)), nil
}

// removeUserRecord returns a copy of fileBuf with the account record for userName removed.
// The caller is responsible for:
//  1) validating that userName exists in the fileBuf before calling removeUserRecord.
//  2) erasing the returned []byte which contains sensitive account material.
func removeUserRecord(userName string, fileBuf []byte) (newFileBuf []byte) {
	lines := bytes.Split(fileBuf, []byte("\n"))
	runningSize := 0
	removeSize := 0
	for _, line := range lines {
		fields := bytes.Split(line, []byte(":"))
		if bytes.Equal([]byte(userName), fields[0]) {
			removeSize = len(line)
			break
		}
		runningSize += len(line)
	}

	size := len(fileBuf) - removeSize
	newFileBuf = make([]byte, 0, size)

	// this check prevents removing the very first user account from leaving a \n in the file
	if runningSize > 0 {
		newFileBuf = append(newFileBuf, fileBuf[:runningSize]...)
		newFileBuf = append(newFileBuf, fileBuf[runningSize+removeSize:]...)
	} else {
		newFileBuf = append(newFileBuf, fileBuf[removeSize+1:]...)
	}

	return newFileBuf
}

// getUserRecord attempts to return the password file record for the specified user as a byte slice.
// If getUserRecord is unable to locate a record, ok == false.
func getUserRecord(userName string, fileBuf []byte) (userRecord []byte, ok bool) {
	ok = false
	userRecord = []byte{}

	if len(fileBuf) == 0 {
		// there are not any users in the password store yet.
		ok = false
		userRecord = []byte{}
		return userRecord, ok
	}

	lines := bytes.Split(fileBuf, []byte("\n"))
	for _, line := range lines {
		fields := bytes.Split(line, []byte(":"))
		if bytes.Equal([]byte(userName), fields[0]) {
			ok = true
			userRecord = line
			return userRecord, ok
		}
	}

	return userRecord, ok
}

// userExists returns ok == true if the user exists in the passwd store buffer,
// otherwise, false.
// Note that userExists assumes it is already working with a validated buffer,
// to pass an un-validated buffer into userExists may result in grave errors.
func userExists(userName string, fileBuf []byte) (ok bool) {
	_, ok = getUserRecord(userName, fileBuf)
	return ok
}

// hostSize returns the length in bytes required to store a comma delimited
// list of hosts in the []string slice.
// Note that hostSize assumes hosts has already be somewhat validated. While a
// run-time error may not occure within hostSize as the result of passing in a
// zero-length hosts list, callers of hostSize may not get a >0 return value
// they expected.
func hostSize(hosts []string) (size int) {
	size = 0

	if len(hosts) == 0 {
		return size
	}

	for _, host := range hosts {
		size += len(host)
	}

	if size > 0 {
		// the number of commas needed is n - 1 if the number of hosts is n
		// note that we cannot assume "," is a single byte.
		size += len(",") * (len(hosts) - 1)
	}

	return size
}

// hostsToBytes converts the hosts list from a string slice to a
// flat, comma delimited byte slice suitable to be written to a file.
func hostsToBytes(hosts []string) (buf []byte) {
	buf = nil
	numHosts := len(hosts)
	for i, host := range hosts {
		buf = append(buf, []byte(host)...)
		if i+1 < numHosts {
			buf = append(buf, []byte(",")...)
		}
	}

	return buf
}
