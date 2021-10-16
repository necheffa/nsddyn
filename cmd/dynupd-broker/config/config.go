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

package config

import (
	"encoding/json"
	"fmt"
	"os"

	"necheff.net/nsddyn/cmd/internal/util"
)

type BrokerConfig struct {
	Address      string
	ReadTimeout  int64
	WriteTimeout int64
	Uri          string
	ZoneName     string
}

func ReadBrokerConfig() (bc BrokerConfig, err error) {
	bc = BrokerConfig{}

	nsddynHome, err := util.FindHome()
	if err != nil {
		return bc, fmt.Errorf("dynupd-broker: %v", err)
	}

	file, err := os.Open(nsddynHome + "/etc/dynupd-broker.json")
	if err != nil {
		return bc, fmt.Errorf("dynupd-broker: unable to read configuration file: %v", err)
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&bc)
	if err != nil {
		return bc, fmt.Errorf("dynupd-broker: parse error reading configuration file: %v", err)
	}

	return bc, nil
}
