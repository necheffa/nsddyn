// Copyright (C) 2024 Alexander Necheff
// This program is licensed under the terms of the GPLv3.
// See the COPYING file that came packaged with this source code for the full terms.

package main

import (
	"log"

	"necheff.net/nsddyn/cmd/nsddynd/app"
	"necheff.net/nsddyn/cmd/nsddynd/data"
)

func main() {
	config, err := data.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	n := app.NewNsdDynd(config)
	n.Run()
}
