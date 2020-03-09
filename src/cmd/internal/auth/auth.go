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

// Package auth provides the authentication interface to nsddyn.
// Byte slices are used throughout this package for handling things that
// naively should be handled by strings (e.g. passwords). This is because
// strings are immutable and are therefore unable to be erased when no longer
// needed. Care should be taken by callers to not just cast a string to []byte
// while using the facilities provided by this package, favor raw []byte which
// can be overwritten before releasing the memory back to the system.
package auth

const (
	MaxPasswd = 256 // max (inclusive) password length in bytes
	MinPasswd = 8   // min (inclusive) password length in bytes
)

// AuthReader provides read-only access to the password store.
// This allows better enforcement of role seperation.
type AuthReader interface {
	// AuthRequest authenticates that the given user is permitted to update an A record for
	// the given hosts.
	// If successful, nil is returned, otherwise a specific non-nil error is returned indicating
	// what the reason for authentication failure is.
	AuthRequest(passwd []byte, userName string, hosts []string) (err error)
}

// AuthWriter provides write access to the password store.
// Access is not purely write-only as some reads are needed for validation,
// e.g. don't add a user account which already exists,
// but to the extent possible these types of reads should be encapsulated,
// not exposed by this interface.
// This allows better enforcement of role seperation.
type AuthWriter interface {
	// AddUser adds the specified user to the password store provided the user
	// does not already exist. On success err == nil otherwise err != nil.
	AddUser(passwd []byte, userName string, hosts []string) (err error)

	// RemoveUser removes the specified user from the password store provided the
	// user account exists already. On success err == nil otherwise err != nil.
	RemoveUser(userName string) (err error)

	// ModifyUser will likely need to be made up of 3 seporate functions,
	// one to update the name, one to update the password, and one to update the allowed hosts.
	//ModifyUser()
}

// AuthReadWriter combines the AuthReader and AuthWriter interfaces to
// provide read-write access to the password store.
type AuthReadWriter interface {
	AuthReader
	AuthWriter
}

// EraseBuf fills len(buf) with space characters.
// if buf is nil or zero length, EraseBuf takes no action.
func EraseBuf(buf []byte) {
	for i := 0; i < len(buf); i++ {
		buf[i] = ' '
	}
}
