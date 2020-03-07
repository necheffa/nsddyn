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

func (f *FlatFile) AddUser(passwd []byte, userName string, hosts []string) (err error) {
	return fmt.Errorf("FlatFile: not yet implemented.")
}

func (f *FlatFile) RemoveUser(userName string) (err error) {
	return fmt.Errorf("FlatFile: not yet implemented.")
}
