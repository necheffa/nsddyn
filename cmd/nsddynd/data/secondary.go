// Copyright (C) 2024 Alexander Necheff
// This program is licensed under the terms of the GPLv3.
// See the COPYING file that came packaged with this source code for the full terms.

package data

type Secondary struct {
	Name string
}

func (s *Secondary) Host() string {
	// TODO: may need to do some footwork here if I choose to support @port syntax in the config file.
	return s.Name
}
