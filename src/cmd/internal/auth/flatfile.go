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
	"io/ioutil"
	"os"

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

func (f *FlatFile) AuthRequest(passwd []byte, userName string, hosts []string) (err error) {
	return fmt.Errorf("FlatFile: not yet implemented.")
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

	fileBuf, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("AddUser: %v", err)
	}
	defer EraseBuf(fileBuf)

	ok := userExists(userName, fileBuf)
	if !ok {
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

	size := len(userName+":") + len(hashedPasswd) + hostSize(hosts) + len("\n")
	lineBuf := make([]byte, size)
	defer EraseBuf(lineBuf)

	// because we manually calculated enough storage for the lineBuf byte slice, we can
	// safely call append without worrying about it reallocating lineBuf behind our backs,
	// leaving bits of the new entry, particuarlly the hashed password, floating around
	// in memory.
	lineBuf = append(lineBuf, []byte(userName+":")...)
	lineBuf = append(lineBuf, hashedPasswd...)
	lineBuf = append(lineBuf, hostsToBytes(hosts)...)
	lineBuf = append(lineBuf, []byte("\n")...)

	// as long as we opened with os.O_APPEND we don't need to file.Seek() to EOF first.
	// file.Write() returns err != nil when we have a short write so we don't really need
	// to explicitly catch the returned number of bytes written.
	//
	// TODO: currently, we assume that file.Write() passes a pointer to our lineBuf all
	// the way down to the kernel write() call and by wiping lineBuf here we have really
	// wiped it. It would be nice to confirm this is the case.
	_, err = file.Write(lineBuf)
	if err != nil {
		return fmt.Errorf("AddUser: %v", err)
	}

	return nil
}

func (f *FlatFile) RemoveUser(userName string) (err error) {
	return fmt.Errorf("FlatFile: not yet implemented.")
}

// userExists returns ok == true if the user exists in the passwd store buffer,
// otherwise, false.
// Note that userExists assumes it is already working with a validated buffer,
// to pass an un-validated buffer into userExists may result in grave errors.
func userExists(userName string, fileBuf []byte) (ok bool) {
	ok = false

	if len(fileBuf) == 0 {
		// there are not any users in the password store yet.
		ok = false
		return ok
	}

	lines := bytes.Split(fileBuf, []byte("\n"))
	for _, line := range lines {
		fields := bytes.Split(line, []byte(":"))
		if bytes.Equal([]byte(userName), fields[0]) {
			ok = true
			return ok
		}
	}

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

	// the number of commas needed is n - 1 if the number of hosts is n
	// note that we cannot assume "," is a single byte.
	size += len(",") * (len(hosts) - 1)

	return size
}

// hostsToBytes converts the hosts list from a string slice to a
// flat, comma delimited byte slice suitable to be written to a file.
func hostsToBytes(hosts []string) (buf []byte) {
	buf = nil
	numHosts := len(hosts)
	for i, host := range hosts {
		buf = append([]byte(host))
		if i+1 < numHosts {
			buf = append([]byte(","))
		}
	}

	return buf
}
