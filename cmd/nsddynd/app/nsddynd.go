// Copyright (C) 2024 Alexander Necheff
// This program is licensed under the terms of the GPLv3.
// See the COPYING file that came packaged with this source code for the full terms.

package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/miekg/dns"

	"necheff.net/nsddyn/cmd/nsddynd/data"
)

type NsdDynd struct {
	Config *data.Configuration

	logger *zap.Logger
	sugar  *zap.SugaredLogger

	WebRouter *http.ServeMux
	DnsRouter *dns.ServeMux
}

func NewNsdDynd(config *data.Configuration) *NsdDynd {
	n := new(NsdDynd)

	n.Config = config

	n.logger = func() *zap.Logger {
		encConfig := zap.NewProductionEncoderConfig()
		encConfig.TimeKey = "timestamp"
		encConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		logLevel := zap.InfoLevel
		if config.LogLevel == "debug" {
			logLevel = zap.DebugLevel
		}

		cfg := zap.Config{
			Level:             zap.NewAtomicLevelAt(logLevel),
			Development:       false,
			DisableCaller:     false,
			DisableStacktrace: false,
			Sampling:          nil,
			Encoding:          "json",
			EncoderConfig:     encConfig,
			OutputPaths:       []string{"stderr"},
			ErrorOutputPaths:  []string{"stderr"},
			InitialFields:     map[string]any{"pid": os.Getpid()},
		}

		return zap.Must(cfg.Build())
	}()

	n.sugar = n.logger.Sugar()

	n.WebRouter = http.NewServeMux()
	n.DnsRouter = dns.NewServeMux()

	n.WebRouter.HandleFunc("POST /{$}", n.ZoneUpdateWebHandle)
	n.WebRouter.HandleFunc("/", n.NotFoundHandle)

	// TODO: add one for each zone in the config
	n.DnsRouter.HandleFunc("example.com.", n.ZoneUpdateDnsHandle)

	return n
}

func (n *NsdDynd) Run() {
	readTimeout := 5 * time.Second
	writeTimeout := 10 * time.Second
	idleTimeout := 120 * time.Second

	webServer := &http.Server{
		Addr:         ":8080",
		Handler:      n.WebRouter,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	go func() {
		if err := webServer.ListenAndServe(); err != nil {
			n.sugar.Debugw("nsddynd web server exited on error", "error", err)
		}
	}()

	dnsServer := &dns.Server{
		// NOTE: for some reason 5353 doesn't show up in nmap, but dig @localhost -p 5353 foo.example.com +tcp works.
		Addr:       ":5353",
		Net:        "tcp",
		Handler:    n.DnsRouter,
		TsigSecret: n.Config.TsigSecrets(),
	}

	go func() {
		if err := dnsServer.ListenAndServe(); err != nil {
			n.sugar.Debugw("nsddyn dns server exited on error", "error", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	wait := time.Second * n.Config.ExitWait
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	webServer.Shutdown(ctx)
	dnsServer.Shutdown()
	os.Exit(0)
}

func (n *NsdDynd) errorWebHandler(w http.ResponseWriter, r *http.Request, code int) {
	w.WriteHeader(code)
}

func (n *NsdDynd) NotFoundHandle(w http.ResponseWriter, r *http.Request) {
	n.errorWebHandler(w, r, http.StatusNotFound)
}

func (n *NsdDynd) ZoneUpdateWebHandle(w http.ResponseWriter, r *http.Request) {
	// TODO: Send the DNS NOTIFY from here to configured secondaries.
	// Force the in-memory record cache to upate.
	// dynupd will POST the zone that is updating, we just need to extract it from the HTTP.
	return
}

func (n *NsdDynd) ZoneUpdateDnsHandle(w dns.ResponseWriter, r *dns.Msg) {
	// We only support AXFR.
	if r.Opcode == dns.OpcodeQuery && r.Question[0].Qtype == dns.TypeAXFR {
		if r.IsTsig() == nil {
			n.sugar.Debugw("TSIG verification failed")
			return
		}

		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true
		status := w.TsigStatus()
		if status != nil {
			n.sugar.Debugw("TSIG invalid status", "status", status)
			return
		}
		// TODO: how to actually sign this message with TSIG?
		// server_test.go from miekg/dns looks like a potential start.
		// The key name and HMAC need to come from the config file and we need to look it up based on zone.
		msg.SetTsig("ProbablyTheKeyName", dns.HmacSHA512, 300, time.Now().Unix())

		// TODO: pull zoneRecords out of an in-memory cache, keyed by domain name assocated with the query.
		zoneRecords := []dns.RR{}
		for _, rec := range zoneRecords {
			msg.Answer = append(msg.Answer, rec)
		}

		w.WriteMsg(msg)
	}
}
