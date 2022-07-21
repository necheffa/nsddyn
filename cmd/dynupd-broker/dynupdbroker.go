/*
   Copyright (C) 2021, 2022 Alexander Necheff

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
	"fmt"
	"log"
	"net/http"
	"os/exec"
)

func DynupdBroker(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	default:
		log.Print("dynupd-broker: error: bad connect method: " + r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	case http.MethodPost:
		// #nosec G204 -- The ZoneName is passed by a configuration file.
		cmd := exec.Command("sh", "-c", "nsd-control reload "+brokerConfig.ZoneName)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Print("dynupd-broker: error: unable to execute nsd-control, error was: " + err.Error())
			log.Print("dynupd-broker: info: nsd-control output was: " + string(out))
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Zone Not Reloaded")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Zone Reloaded")
	}
}
