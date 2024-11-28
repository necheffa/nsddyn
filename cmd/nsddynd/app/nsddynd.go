// Copyright (C) 2024 Alexander Necheff
// This program is licensed under the terms of the GPLv3.
// See the COPYING file that came packaged with this source code for the full terms.

package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path"
	"time"

	"github.com/miekg/dns"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

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

	// Only match HOST/example.com anything else is a 404.
	n.WebRouter.HandleFunc("POST /{zoneName}", n.ZoneUpdateWebHandle)
	n.WebRouter.HandleFunc("POST /{zoneName}/{somethingElse...}", n.NotFoundHandle)
	n.WebRouter.HandleFunc("/", n.NotFoundHandle)

	for _, zone := range n.Config.Zones {
		n.DnsRouter.HandleFunc(zone.CanonicalName(), n.ZoneUpdateDnsHandle)
	}

	return n
}

func (n *NsdDynd) Run() {
	readTimeout := 5 * time.Second
	writeTimeout := 10 * time.Second
	idleTimeout := 120 * time.Second

	defaultWebListen := ":8080"
	// NOTE: for some reason 5353 doesn't show up in nmap, but dig @localhost -p 5353 foo.example.com +tcp works.
	defaultDnsListen := ":5353"

	webServer := &http.Server{
		Addr:         defaultWebListen,
		Handler:      n.WebRouter,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	go func() {
		n.sugar.Debugw("starting up nsddynd web server", "address", defaultWebListen)
		if err := webServer.ListenAndServe(); err != nil {
			n.sugar.Debugw("nsddynd web server exited on error", "error", err)
		}
	}()

	secrets := n.Config.TsigSecrets()
	n.sugar.Debugw("Registering TSIG secrets", "secrets", secrets)

	dnsServer := &dns.Server{
		Addr:       defaultDnsListen,
		Net:        "tcp",
		Handler:    n.DnsRouter,
		TsigSecret: secrets,
	}

	go func() {
		n.sugar.Debugw("starting up nsddynd dns server", "address", defaultDnsListen)
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

func (n *NsdDynd) errorWebHandle(w http.ResponseWriter, r *http.Request, code int) {
	n.sugar.Debugw("Sending HTTP error in response", "status", code, "remote", r.RemoteAddr, "URL", r.URL.Path)
	w.WriteHeader(code)
}

func (n *NsdDynd) NotFoundHandle(w http.ResponseWriter, r *http.Request) {
	n.errorWebHandle(w, r, http.StatusNotFound)
}

func (n *NsdDynd) ZoneUpdateWebHandle(w http.ResponseWriter, r *http.Request) {
	name := path.Base(r.URL.Path)

	n.sugar.Debugw("Got a zone update request for", "zone", name)

	zone := n.Config.ZoneByName(name)
	if zone == nil {
		n.sugar.Debugw("Requested zone not located in managed zones", "zone", name)
		n.NotFoundHandle(w, r)
		return
	}
	zone.CacheRecords()

	msg := new(dns.Msg)
	msg.SetNotify(zone.CanonicalName())
	msg.RecursionDesired = false

	key := n.Config.KeyByZone(name)
	if key == nil {
		n.sugar.Debugw("No key associated with requeted zone", "zone", name)
		n.errorWebHandle(w, r, http.StatusInternalServerError)
		return
	}
	msg.SetTsig(key.Name, key.HmacAlgo(), 300, time.Now().Unix())

	client := &dns.Client{
		TsigSecret: n.Config.TsigSecrets(),
	}

	for _, sec := range n.Config.Secondaries {
		res, _, err := client.Exchange(msg, sec.Host())
		if err != nil {
			n.sugar.Debugw("Failed to send NOTIFY to secondary", "secondary", sec.Host(), "zone", zone.Name, "key", key.Name, "error", err)
		} else if res.Rcode != dns.RcodeSuccess {
			n.sugar.Debugw("NOTIFY to secondary was unsuccessful", "secondary", sec.Host(), "rcode", dns.RcodeToString[res.Rcode])
		}
	}

	return
}

func (n *NsdDynd) ZoneUpdateDnsHandle(w dns.ResponseWriter, r *dns.Msg) {
	n.sugar.Debugw("Receving DNS message...")
	// We only support AXFR.
	if r.Opcode == dns.OpcodeQuery && r.Question[0].Qtype == dns.TypeAXFR {
		n.sugar.Debugw("Receving an AXFR DNS message...")
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

		n.sugar.Debugw("AXFR message for key", "key", r.Question[0].Name)

		// TODO: look up the patterns and see if this host is even permitted to request AXFR from us.

		// TODO: the name may be compressed, will need to look in to how to decompress.
		name := r.Question[0].Name
		key := n.Config.KeyByName(name)
		msg.SetTsig(key.Name, key.HmacAlgo(), 300, time.Now().Unix())

		zone := n.Config.ZoneByName(name)
		zoneRecords := zone.CachedRecords()
		for _, rec := range zoneRecords {
			msg.Answer = append(msg.Answer, rec)
		}

		w.WriteMsg(msg)
	}
}
