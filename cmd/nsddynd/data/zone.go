// Copyright (C) 2024 Alexander Necheff
// This program is licensed under the terms of the GPLv3.
// See the COPYING file that came packaged with this source code for the full terms.

package data

import (
	"bytes"
	"io"
	"os"

	"golang.org/x/sys/unix"

	"github.com/bwesterb/go-zonefile"
	"github.com/miekg/dns"
)

type Zone struct {
	Name    string
	File    string
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
	for _, e := range zf.Entries() {
		rr, err := dns.NewRR(e.String())
		if err != nil {
			// TODO: don't just bail on error here.
			// may need to work on a copy of the cache first so we can revert back if an error occurs.
			return err
		}
		z.cache = append(z.cache, rr)
	}

	return nil
}

func (z *Zone) CachedRecords() []dns.RR {
	return z.cache
}
