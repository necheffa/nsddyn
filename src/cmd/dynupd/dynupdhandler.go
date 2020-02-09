/*
   Copyright (C) 2019, 2020 Alexander Necheff

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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"cmd/internal/dynreq"
)

func dynupdHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	default:
		fmt.Fprintf(w, "Bad method: only POST is allowed.")
		return
	case "POST":
		var msg dynreq.DynReq
		// parse the reqest
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintf(w, "Unknown read error")
			return
		}
		err = json.Unmarshal(body, &msg)
		if err != nil {
			fmt.Fprintf(w, "Unknown parse error")
			return
		}
		fmt.Println("Debug: " + string(body))

		// TODO: if successful, open a local socket to communicate with nsddynd

		//fmt.Println(r.Body)
		fmt.Fprintf(w, "Posted username: %v\n", msg.Username)
		fmt.Fprintf(w, "Posted password: %v\n", msg.Password)
		fmt.Fprintf(w, "Posted ipaddr: %v\n", msg.Ipaddr)
		fmt.Fprintf(w, "Posted client version: %v\n", msg.Version)
		fmt.Fprintf(w, "Posted hosts:\n")
		for i := 0; i < len(msg.Hostnames); i++ {
			fmt.Fprintf(w, "host: %v\n", msg.Hostnames[i])
		}
	}
}
