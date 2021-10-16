/*
   Copyright (C) 2021 Alexander Necheff

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
	"log"
	"net/http"

	"necheff.net/nsddyn/cmd/dynupd-broker/config"
)

var brokerConfig config.BrokerConfig

func main() {
	brokerConfig, err := config.ReadBrokerConfig()
	if err != nil {
		log.Fatalln(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc(brokerConfig.Uri, dynupdBroker)

	server := &http.Server{
		Addr:    brokerConfig.Address,
		Handler: mux,
	}
	server.ListenAndServe()
}
