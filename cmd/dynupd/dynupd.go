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
	"os/exec"
	"strconv"
	"strings"

	"necheff.net/nsddyn/cmd/internal/auth"
	"necheff.net/nsddyn/cmd/internal/config"
	"necheff.net/nsddyn/cmd/internal/dynreq"
	"necheff.net/nsddyn/cmd/internal/version"

	"github.com/bwesterb/go-zonefile"
)

// DynUpd status codes - similar to HTTP status codes,
// but not exactly the same.
const (
	zoneUpdateFail     = "500"
	zoneUpdateSuccess  = "200"
	malformedRequest   = "400"
	authenticationFail = "403"
	badMethod          = "405"
	hostsFail          = "418" // it's tea time
)

type DynUpd struct {
	passwdDb   auth.AuthReader
	zoneFile   string
	domainName string
}

// NewDynUpd initalizes a new instance of DynUpd that has already been allocated.
func (d *DynUpd) NewDynUpd(passwdDb auth.AuthReader, zoneFile string, domainName string) {
	d.passwdDb = passwdDb
	d.zoneFile = zoneFile
	d.domainName = domainName
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
	if config.Debug {
		fmt.Fprintf(os.Stderr, "dynupd: info: updating zonefile with: %s\n", newZoneFile)
	}
	//TODO: handle possible write error getting returned here...
	file.Seek(0, os.SEEK_SET)
	file.Write(newZoneFile)
	if newZoneFile[len(newZoneFile)-3] != '\n' {
		// make sure we always have a new-line after the final A record, otherwise NSD doesn't like zonefile
		file.Write([]byte("\n"))
	}

	// TODO: there has to be a better way to ask NSD to reload the zone...
	cmd := exec.Command("nsd-control", "reload", d.domainName)
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "dynupd: error: failed to reload zone for "+d.domainName+" error was %v\n", err)
		return zoneUpdateFail
	}

	return zoneUpdateSuccess
}

// DynUpdHandler is a handler for incomming zone update requests
// and is meant to be registered with http.HandleFunc.
func (d *DynUpd) DynUpdHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	default:
		if config.Debug {
			fmt.Fprintf(os.Stderr, "Bad method requested.\n")
		}
		fmt.Fprintf(w, craftResponse(badMethod))
		return
	case "POST":
		var msg dynreq.DynReq
		// parse the reqest
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			if config.Debug {
				fmt.Fprintf(os.Stderr, "Failed to read request: %v\n", err)
			}
			// we don't acutally know if the request was well formed or not, failing here
			// is likely an internal server error...
			fmt.Fprintf(w, craftResponse(zoneUpdateFail))
			return
		}
		err = json.Unmarshal(body, &msg)
		if err != nil {
			if config.Debug {
				fmt.Fprintf(os.Stderr, "Failed to unmarshal JSON request: %v\n", err)
			}
			fmt.Fprintf(w, craftResponse(malformedRequest))
			return
		}
		defer auth.EraseBuf(msg.Password)

		if config.Debug {
			fmt.Fprintf(os.Stderr, "Receaved request:\n")
			fmt.Fprintf(os.Stderr, "Posted username: %v\n", msg.Username)
			fmt.Fprintf(os.Stderr, "Posted ipaddr: %v\n", msg.Ipaddr)
			fmt.Fprintf(os.Stderr, "Posted client version: %v\n", msg.Version)
			fmt.Fprintf(os.Stderr, "Posted hosts:\n")
			for i := 0; i < len(msg.Hostnames); i++ {
				fmt.Fprintf(os.Stderr, "host: %v\n", msg.Hostnames[i])
			}
		}

		passwd, err := passwdParse([]byte(msg.Password))
		if err != nil {
			if config.Debug {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			}
			fmt.Fprintf(w, craftResponse(malformedRequest))
			return
		}

		err = d.passwdDb.AuthRequest(passwd, msg.Username, msg.Hostnames)
		if err != nil {
			if strings.HasPrefix(err.Error(), "AuthRequest: failed to match requested hosts for:") {
				fmt.Fprintf(w, craftResponse(hostsFail))
			} else if strings.HasPrefix(err.Error(), "AuthRequest: user account does not exist with name:") ||
				strings.HasPrefix(err.Error(), "AuthRequest: crypto/bcrypt: hashedPassword is not the hash of the given password") {
				fmt.Fprintf(w, craftResponse(authenticationFail))
			} else {
				fmt.Fprintf(w, craftResponse(zoneUpdateFail))
			}
			if config.Debug {
				fmt.Fprintf(os.Stderr, "Authentication failed: %v\n", err)
			}
			return
		}

		// if authentication successful, update zonefile
		if config.Debug {
			fmt.Fprintf(os.Stderr, "authentication successful, updating: "+d.zoneFile+"\n")
		}

		status := d.UpdateZone(msg)
		fmt.Fprintf(w, craftResponse(status))
	}
}

// craftResponse crafts a JSON responce object conforming to the nsddyn protocol using
// the given status code as a string.
func craftResponse(status string) string {
	return "{ \"version\": \"" + version.Version + "\", \"code\": \"" + status + "\" }"
}
