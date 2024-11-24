/*
   Copyright (C) 2020-2022, 2024 Alexander Necheff

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
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"necheff.net/nsddyn"
	"necheff.net/nsddyn/internal/auth"
	"necheff.net/nsddyn/internal/config"
	"necheff.net/nsddyn/internal/dynreq"
	"necheff.net/nsddyn/internal/util"

	"github.com/bwesterb/go-zonefile"
)

// DynUpd status codes - similar to HTTP status codes,
// but not exactly the same.
const (
	zoneUpdateFail     = "500"
	zoneUpdateSuccess  = "200"
	zoneNotModified    = "304"
	malformedRequest   = "400"
	authenticationFail = "403"
	badMethod          = "405"
	hostsFail          = "418" // it's tea time
)

const zonefileMode os.FileMode = 0664

// The number of data fields in a SOA record.
const SOANumFields = 7

// The position of the serial in the SOA record.
const SOASerialPos = 2

// The position of the IP in an A record value.
const IPPos = 0

type DynUpd struct {
	passwdDb   auth.AuthReader
	zoneFile   string
	domainName string
	brokerUrl  string
	lock       sync.Mutex
}

// NewDynUpd creates and initalizes an instance of DynUpd.
func NewDynUpd(passwdDb auth.AuthReader, zoneFile string, domainName string, brokerUrl string) (d *DynUpd) {
	d = new(DynUpd)

	d.passwdDb = passwdDb
	d.zoneFile = zoneFile
	d.domainName = domainName
	d.brokerUrl = brokerUrl

	return d
}

// NewDynUpd initalizes a new instance of DynUpd that has already been allocated.
//
// Deprecated: This method is preserved for legacy compatability. The NewDynUpd function should be called
// instead in new code.
func (d *DynUpd) NewDynUpd(passwdDb auth.AuthReader, zoneFile string, domainName string, brokerUrl string) {
	d.passwdDb = passwdDb
	d.zoneFile = zoneFile
	d.domainName = domainName
	d.brokerUrl = brokerUrl
}

// UpdateZone updates the zonefile with the specified request.
// Returns a status code as a string to be sent to the requesting client.
func (d *DynUpd) UpdateZone(r dynreq.DynReq) string {
	d.lock.Lock()
	defer d.lock.Unlock()

	file, err := os.OpenFile(d.zoneFile, os.O_RDWR, zonefileMode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", "dynupd: error: failed to open zonefile for update: "+d.zoneFile)
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return zoneUpdateFail
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", "dynupd: error: failed to read zonefile for update: "+d.zoneFile)
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return zoneUpdateFail
	}

	zf, err := zonefile.Load(buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", "dynupd: error: failed to parse zonefile for update: "+d.zoneFile)
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return zoneUpdateFail
	}

	zoneReloadNotNeeded := 0
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

			vals := e.Values()
			if bytes.Equal(vals[IPPos], []byte(r.Ipaddr)) {
				// the existing A record already has the IP in the request.
				// no need to update the zone.
				if config.Debug {
					fmt.Fprintf(os.Stderr, "dynupd: A record for %v already has IP %v so not updating.\n", r.Hostnames[i], r.Ipaddr)
				}
				// can't return here because there could be other hostnames in the request that could need a reload.
				zoneReloadNotNeeded += 1
			}

			// we found a match, update it
			err = e.SetValue(IPPos, []byte(r.Ipaddr))
			if err != nil {
				fmt.Fprintf(os.Stderr, "dynupd: error: updating A record: %v\n", err.Error())
				return zoneUpdateFail
			}
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

	if zoneReloadNotNeeded == len(r.Hostnames) {
		// none of the hosts required a record update.
		return zoneNotModified
	}

	// does zonefile automatically update the serial? if not update it here.
	if updateSerial {
		for _, e := range zf.Entries() {
			if !bytes.Equal(e.Type(), []byte("SOA")) {
				continue
			}

			vals := e.Values()
			if len(vals) != SOANumFields {
				fmt.Fprintf(os.Stderr, "dynupd: error: malformed SOA in zonefile\n")
				return zoneUpdateFail
			}

			serial, _ := strconv.Atoi(string(vals[SOASerialPos]))
			err = e.SetValue(SOASerialPos, []byte(strconv.Itoa(serial+1)))
			if err != nil {
				fmt.Fprintf(os.Stderr, "dynupd: error: unable to increment SOA serial: %v\n", err.Error())
				return zoneUpdateFail
			}
			break
		}
	} else {
		fmt.Fprintf(os.Stderr, "dynupd: error: failed to update A record\n")
		return zoneUpdateFail
	}

	// if we made it this far...things have gone well.

	newZoneFile := zf.Save()
	if config.Debug {
		fmt.Fprintf(os.Stderr, "dynupd: info: updating zonefile with: %s\n", newZoneFile)
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dynupd: error: updating zonefile: %v\n", err.Error())
		return zoneUpdateFail
	}
	_, err = file.Write(newZoneFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dynupd: error: updating zonefile: %v\n", err.Error())
		return zoneUpdateFail
	}
	if newZoneFile[len(newZoneFile)-3] != '\n' {
		// make sure we always have a new-line after the final A record, otherwise NSD doesn't like zonefile
		_, err = file.Write([]byte("\n"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "dynupd: error: updating zonefile: %v\n", err.Error())
			return zoneUpdateFail
		}
	}

	nsddynHome, err := util.FindHome()
	if err != nil {
		fmt.Fprintf(os.Stderr, "dynupd: error: %v\n", err)
		return zoneUpdateFail
	}

	_, err = http.Post(d.brokerUrl, "text/plain", strings.NewReader(nsddynHome))
	if err != nil {
		fmt.Fprintf(os.Stderr, "dynupd: error: broker call failed, error was: "+err.Error())
		return zoneUpdateFail
	}
	// SEE: https://gitlab.com/necheffa/nsddyn/-/issues/102

	return zoneUpdateSuccess
}

// DynUpdHandler is a handler for incomming zone update requests
// and is meant to be registered with http.HandleFunc.
func (d *DynUpd) DynUpdHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	default:
		fmt.Fprintf(os.Stderr, "Bad method requested.\n")
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "%v", craftResponse(badMethod))
		return
	case http.MethodPost:
		var msg dynreq.DynReq
		// parse the reqest
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read request: %v\n", err)
			// we don't acutally know if the request was well formed or not, failing here
			// is likely an internal server error...
			fmt.Fprintf(w, "%v", craftResponse(zoneUpdateFail))
			return
		}
		err = json.Unmarshal(body, &msg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to unmarshal JSON request: %v\n", err)
			fmt.Fprintf(w, "%v", craftResponse(malformedRequest))
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
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			fmt.Fprintf(w, "%v", craftResponse(malformedRequest))
			return
		}

		err = d.passwdDb.AuthRequest(passwd, msg.Username, msg.Hostnames)
		if err != nil {
			if strings.HasPrefix(err.Error(), "AuthRequest: failed to match requested hosts for:") {
				fmt.Fprintf(w, "%v", craftResponse(hostsFail))
			} else if strings.HasPrefix(err.Error(), "AuthRequest: user account does not exist with name:") ||
				strings.HasPrefix(err.Error(), "AuthRequest: crypto/bcrypt: hashedPassword is not the hash of the given password") {
				fmt.Fprintf(w, "%v", craftResponse(authenticationFail))
			} else {
				fmt.Fprintf(w, "%v", craftResponse(zoneUpdateFail))
			}

			if config.Debug {
				fmt.Fprintf(os.Stderr, "Authentication failed: %v\n", err)
			}

			return
		}

		if config.Debug {
			fmt.Fprintf(os.Stderr, "authentication successful, updating: "+d.zoneFile+"\n")
		}

		status := d.UpdateZone(msg)
		fmt.Fprintf(w, "%v", craftResponse(status))
	}
}

// craftResponse crafts a JSON responce object conforming to the nsddyn protocol using
// the given status code as a string.
func craftResponse(status string) string {
	return "{ \"version\": \"" + nsddyn.Version() + "\", \"code\": \"" + status + "\" }"
}
