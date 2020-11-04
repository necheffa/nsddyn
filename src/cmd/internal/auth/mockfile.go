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
	"fmt"
	"io"
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

// Truncate changes the size of the file. It does not change the I/O offset.
func (mf *MockFile) Truncate(size int64) error {
	mf.buf = mf.buf[:int(size)]
	return nil
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

	if len(p)+mf.curPos > cap(mf.buf) {
		// we need to reallocate mf.buf before doing the write.
		newSize := (cap(mf.buf) + len(p)) * 2 // weird growth rate, but it works...
		newFileBuf := make([]byte, 0, newSize)
		_ = copy(newFileBuf, mf.buf)
		mf.buf = newFileBuf
	}

	n = copy(mf.buf[mf.curPos:], p)
	mf.curPos += n
	return n, nil
}

func (mf *MockFile) Rewind() {
	// ultra cheap mf.Seek(0, io.SeekStart) :-D
	mf.curPos = 0
}
