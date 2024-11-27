// Copyright (C) 2024 Alexander Necheff
// This program is licensed under the terms of the GPLv3.
// See the COPYING file that came packaged with this source code for the full terms.

package data

import (
	"time"

	"github.com/spf13/viper"

	"necheff.net/nsddyn/internal/util"
)

type Key struct {
	Name   string
	Algo   string
	Secret string
}

type Pattern struct {
	Name       string
	NotifyTo   string
	NotifyFrom string
	XfrTo      string
	XfrFrom    string
	Key        string
}

type Zone struct {
	Name    string
	File    string
	Pattern string
}

type Configuration struct {
	ListenAddrWeb string
	ListenAddrDns string
	ExitWait      time.Duration
	LogLevel      string
	Keys          []Key
	Patterns      []Pattern
	Zones         []Zone
}

func (c *Configuration) TsigSecrets() map[string]string {
	// TODO: get these from the config file, base64 encode the secrets in this method.
	m := make(map[string]string)

	// there is probably a miekg/dns function for canonicalizing the zone names...
	m["example.com."] = "pretendThisIsABase64Secret"

	return m
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
