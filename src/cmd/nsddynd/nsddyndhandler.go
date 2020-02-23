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
	"fmt"
	"net/http"
	//    "encoding/json"
	//    "io/ioutil"
	"os"
	//    "cmd/internal/dynreq"
)

func nsddyndHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	default:
		fmt.Fprintf(w, "Bad method: only POST is allowed.")
		return
	case "POST":
		fmt.Fprintln(os.Stderr, "nsddynd: handling authentication request")
	}
}
