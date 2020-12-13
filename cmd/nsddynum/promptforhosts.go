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

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// promptForHosts prompts the user for a comma-seporated list of hostnames to be read on STDIN.
// The hostnames are then returned or err == nil.
func promptForHosts() (hostNames []string, err error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Fprintf(os.Stderr, "Enter hosts as a comma-seporated list (e.g. host1, host2, host3): ")
	hostEntry, _ := reader.ReadString('\n')

	hostEntries := strings.Split(hostEntry, ",")
	for _, host := range hostEntries {
		hostNames = append(hostNames, strings.Trim(host, " \n"))
	}

	return hostNames, nil
}
