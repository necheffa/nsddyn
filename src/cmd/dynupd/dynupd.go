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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"cmd/internal/config"
	"cmd/internal/dynreq"
	"cmd/internal/version"

	"github.com/bwesterb/go-zonefile"
)

// DynUpd status codes - similar to HTTP status codes,
// but not exactly the same.
const (
	zoneUpdateFail    = "500"
	zoneUpdateSuccess = "200"
)

type DynUpd struct {
	passwdFile string
	zoneFile   string
}

func (d *DynUpd) NewDynUpd(passwdFile string, zoneFile string) {
	d.passwdFile = passwdFile
	d.zoneFile = zoneFile
}

// UpdateZone updates the zonefile with the specified request.
// Returns a status code as a string to be sent to the requesting client.
func (d *DynUpd) UpdateZone(r dynreq.DynReq) string {
	file, err := os.OpenFile(d.zoneFile, os.O_RDWR, 0664)
	if err != nil {
		if config.Debug {
			fmt.Fprintf(os.Stderr, "dynupd: error: failed to open zonefile for update: "+d.zoneFile)
		}
		return zoneUpdateFail
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		if config.Debug {
			fmt.Fprintf(os.Stderr, "dynupd: error: failed to read zonefile for update: "+d.zoneFile)
		}
		return zoneUpdateFail
	}

	zf, err := zonefile.Load(buf)
	if err != nil {
		if config.Debug {
			fmt.Fprintf(os.Stderr, "dynupd: error: failed to parse zonefile for update: "+d.zoneFile)
		}
		return zoneUpdateFail
	}

	updateSerial := false
	for i := 0; i < len(r.Hostnames); i++ {
		// find r.Hostnames[i] in the zonefile
		foundHost := false
		for _, e := range zf.Entries() {
			if !bytes.Equal(e.Type(), []byte("A")) {
				continue
			}

			if !bytes.Equal(e.Domain(), []byte(r.Hostnames[i])) {
				continue
			}

			// we found a match, update it
			e.SetValue(0, []byte(r.Ipaddr))
			foundHost = true
			updateSerial = true
			break
		}

		// if it doesn't exist, add it
		if !foundHost {
			e, _ := zonefile.ParseEntry([]byte(r.Hostnames[i] + " IN A " + r.Ipaddr))
			zf.AddEntry(e)
			updateSerial = true
		}
	}

	// does zonefile automatically update the serial? if not update it here.
	if updateSerial {
		for _, e := range zf.Entries() {
			if !bytes.Equal(e.Type(), []byte("SOA")) {
				continue
			}

			vals := e.Values()
			if len(vals) != 7 {
				if config.Debug {
					fmt.Fprintf(os.Stderr, "dynupd: error: malformed SOA in zonefile\n")
				}
				return zoneUpdateFail
			}

			serial, _ := strconv.Atoi(string(vals[2]))
			e.SetValue(2, []byte(strconv.Itoa(serial+1)))
			break
		}
	} else {
		if config.Debug {
			fmt.Fprintf(os.Stderr, "dynupd: error: failed to update A record\n")
			return zoneUpdateFail
		}
	}

	// if we made it this far...things have gone well.

	newZoneFile := zf.Save()
	//TODO: handle possible write error getting returned here...
	file.Write(newZoneFile)

	//TODO: actually reload the zone here
	fmt.Fprintf(os.Stderr, "dynupd: info: need to ask NSD to reload the zone here...\n")

	return zoneUpdateSuccess
}

func (d *DynUpd) DynUpdHandler(w http.ResponseWriter, r *http.Request) {

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

		if config.Debug {
			fmt.Fprintf(os.Stderr, "Receaved request:\n")
			fmt.Fprintf(os.Stderr, "Posted username: %v\n", msg.Username)
			fmt.Fprintf(os.Stderr, "Posted password: %v\n", msg.Password)
			fmt.Fprintf(os.Stderr, "Posted ipaddr: %v\n", msg.Ipaddr)
			fmt.Fprintf(os.Stderr, "Posted client version: %v\n", msg.Version)
			fmt.Fprintf(os.Stderr, "Posted hosts:\n")
			for i := 0; i < len(msg.Hostnames); i++ {
				fmt.Fprintf(os.Stderr, "host: %v\n", msg.Hostnames[i])
			}
		}

		// sanitize input

		// authenticate user

		// if authentication successful, update zonefile
		if config.Debug {
			fmt.Fprintf(os.Stderr, "authentication successful, updating: "+d.zoneFile+"\n")
		}

		status := d.UpdateZone(msg)
		retObj := "{ \"version\": " + version.Version + ", \"code\": " + status + " }"
		fmt.Fprintf(w, retObj)
	}
}
