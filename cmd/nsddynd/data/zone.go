// Copyright (C) 2024 Alexander Necheff
// This program is licensed under the terms of the GPLv3.
// See the COPYING file that came packaged with this source code for the full terms.

package data

import (
	"github.com/miekg/dns"
)

type Zone struct {
	Name    string
	File    string
	Pattern string
}

func (z *Zone) CanonicalName() string {
	return dns.CanonicalName(z.Name)
}
