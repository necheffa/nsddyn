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

import (
	"bytes"
	"testing"
)

func TestEraseBuf(t *testing.T) {
	buf := make([]byte, 5)
	fiveAs := "aaaaa"
	spaces := "     "

	for i := range buf {
		buf[i] = 'a'
	}

	if !bytes.Equal(buf, []byte(fiveAs)) {
		t.Errorf("buf(%q) == %q = false", string(buf), fiveAs)
	}

	EraseBuf(buf)
	if !bytes.Equal(buf, []byte(spaces)) {
		t.Errorf("EraseBuf(%q) is not all spaces", string(buf))
	}
}
