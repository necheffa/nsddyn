// Copyright (C) 2024 Alexander Necheff
// This program is licensed under the terms of the GPLv3.
// See the COPYING file that came packaged with this source code for the full terms.

package data

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"slices"

	"golang.org/x/sys/unix"

	"github.com/bwesterb/go-zonefile"
	"github.com/miekg/dns"
)

type Zone struct {
	Name    string
	File    string
	Key     string
	Pattern string

	cache []dns.RR
}

func (z *Zone) CanonicalName() string {
	return dns.CanonicalName(z.Name)
}

func (z *Zone) CacheRecords() error {
	// TODO: assume an absolute path exists in the Zone.File string for now
	file, err := os.OpenFile(z.File, os.O_RDONLY, 0640)
	if err != nil {
		return err
	}
	defer file.Close()

	if err = unix.Flock(int(file.Fd()), unix.LOCK_SH); err != nil {
		return err
	}

	defer func() {
		if err = unix.Flock(int(file.Fd()), unix.LOCK_UN); err != nil {
			//return err // TODO: need to make sure we are not clobbering an existing error...
		}
	}()

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, file)
	if err != nil {
		// NOTE: may need to add something here to handle EOF.
		return err
	}

	zf, err := zonefile.Load(buf.Bytes())
	if err != nil {
		return err
	}

	// TODO: probably delete the old cache first
	var soa dns.RR
	for _, e := range zf.Entries() {
		if bytes.Equal(e.Command(), []byte("$ORIGIN")) || bytes.Equal(e.Command(), []byte("$TTL")) {
			continue
		}
		rr, err := dns.NewRR(convertToString(e))
		if err != nil {
			// TODO: don't just bail on error here.
			// may need to work on a copy of the cache first so we can revert back if an error occurs.
			return fmt.Errorf("zone: %s: type: %s: %w", e.String(), string(e.Type()), err)
		}
		if bytes.Equal(e.Type(), []byte("SOA")) {
			z.cache = slices.Insert(z.cache, 0, rr)
			soa = rr
		} else {
			z.cache = append(z.cache, rr)
		}
	}

	// The envelope appears to also want the SOA record duplicated at the end.
	z.cache = append(z.cache, soa)

	return nil
}

func (z *Zone) Envelope() (error, []dns.RR) {
	err := z.CacheRecords()
	if err != nil {
		return err, []dns.RR{}
	}
	return nil, z.cache
}

func (z *Zone) CachedRecords() []dns.RR {
	return z.cache
}

func convertToString(e zonefile.Entry) string {
	// TODO: use a string builder or some shit here to be more efficient.
	// TODO: deal with the TTL, I don't need it right this second and bwesterb is gonna make me convert a *int to string...
	if bytes.Equal(e.Command(), []byte("$ORIGIN")) {
		s := "$ORIGIN "
		for _, val := range e.Values() {
			s = s + string(val) + " "
		}

		return s
	}

	if bytes.Equal(e.Command(), []byte("$TTL")) {
		s := "$TTL "
		for _, val := range e.Values() {
			s = s + string(val) + " "
		}

		return s
	}

	// TODO: fudge a TTL of 2
	s := string(e.Domain()) + " 2 " + string(e.Class()) + " " + string(e.Type()) + " "
	for _, val := range e.Values() {
		s = s + string(val) + " "
	}

	return s
}
