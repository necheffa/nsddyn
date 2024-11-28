// Copyright (C) 2024 Alexander Necheff
// This program is licensed under the terms of the GPLv3.
// See the COPYING file that came packaged with this source code for the full terms.

package data

import (
	"time"

	"github.com/miekg/dns"
	"github.com/spf13/viper"

	"necheff.net/nsddyn/internal/util"
)

type Pattern struct {
	Name       string
	NotifyTo   string
	NotifyFrom string
	XfrTo      string
	XfrFrom    string
	Key        string
}

type Configuration struct {
	ListenAddrWeb string
	ListenAddrDns string
	ExitWait      time.Duration
	LogLevel      string
	Keys          []Key
	Patterns      []Pattern
	Zones         []Zone
	Secondaries   []Secondary
}

func (c *Configuration) TsigSecrets() map[string]string {
	m := make(map[string]string)
	for _, k := range c.Keys {
		m[dns.CanonicalName(k.Name)] = k.Base64()
	}

	return m
}

func (c *Configuration) KeyByName(name string) *Key {
	for _, key := range c.Keys {
		if key.Name == name {
			return &key
		}
	}

	return nil
}

func (c *Configuration) KeyByZone(zone string) *Key {
	z := c.ZoneByName(zone)
	if z == nil {
		return nil
	}

	return c.KeyByName(z.Key)
}

func (c *Configuration) ZoneByName(name string) *Zone {
	for _, zone := range c.Zones {
		if zone.Name == name {
			return &zone
		}
	}

	return nil
}

func locateConfig() (string, error) {
	nsdHome, err := util.FindHome()
	if err != nil {
		return "", err
	}

	return nsdHome + "/etc", nil
}

func NewConfig() (*Configuration, error) {
	config := &Configuration{}
	v := viper.New()

	v.SetConfigName("nsddynd.toml")
	v.SetConfigType("toml")

	confPath, err := locateConfig()
	if err != nil {
		return config, err
	}
	v.AddConfigPath(confPath)
	v.AddConfigPath(".") // for development only.

	v.SetDefault("ListenAddrWeb", "127.0.0.1:8080")
	v.SetDefault("ListenAddrDns", "127.0.0.1:5353")
	v.SetDefault("ExitWait", 15)
	v.SetDefault("LogLevel", "info")

	err = v.ReadInConfig()
	if err != nil {
		return config, err
	}

	err = v.Unmarshal(config)
	if err != nil {
		return config, err
	}

	return config, err
}
