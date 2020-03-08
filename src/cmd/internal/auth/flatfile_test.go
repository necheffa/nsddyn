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
	"testing"
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
