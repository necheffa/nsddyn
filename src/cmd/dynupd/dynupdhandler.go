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
	//	"encoding/json"
	"fmt"
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
		fmt.Fprintf(w, "Posted: %v", msg.Username)
	}

	fmt.Fprintf(w, "Yes")

	// parse the reqest

	// if successful, open a local socket to communicate with nsddynd
}
