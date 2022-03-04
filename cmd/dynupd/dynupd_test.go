/*
   Copyright (C) 2021, 2022 Alexander Necheff

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

package main_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "necheff.net/nsddyn/cmd/dynupd"
	"necheff.net/nsddyn/cmd/internal/auth"

	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
)

var passwd *os.File
var zonefile *os.File
var dynupd *DynUpd
var authFile *auth.FlatFile
var srv *httptest.Server

var _ = BeforeSuite(func() {
	passwd, _ = os.CreateTemp(os.TempDir(), "nsddynpasswd")
	zonefile, _ = os.CreateTemp(os.TempDir(), "zonefile")
	authFile = new(auth.FlatFile)
	authFile.SetFilePath(passwd.Name())

	_ = authFile.AddUser([]byte("password"), "alex", []string{"host1", "host2"})

	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// don't do anything, successfully
		fmt.Fprintf(w, "%v", r.Body)
	}))
})

var _ = AfterSuite(func() {
	os.Remove(passwd.Name())
	os.Remove(zonefile.Name())
})

var _ = Describe("Dynupd", func() {
	var mux *http.ServeMux
	var writer *httptest.ResponseRecorder
	var uri string
	type nsddynbody struct {
		Code string
	}
	var nb nsddynbody

	BeforeEach(func() {
		uri = "/api/dynupd"
		dynupd = new(DynUpd)
		dynupd.NewDynUpd(authFile, zonefile.Name(), "example.com", srv.URL)

		mux = http.NewServeMux()
		mux.HandleFunc(uri, dynupd.DynUpdHandler)

		writer = httptest.NewRecorder()

	})

	Describe("HTTP method response", func() {
		Context("With an invalid GET", func() {
			It("should return an HTTP 405", func() {
				request, _ := http.NewRequest("GET", uri, nil)
				mux.ServeHTTP(writer, request)
				Expect(writer.Code).To(Equal(http.StatusMethodNotAllowed))
			})
		})

		Context("With a bad password", func() {
			It("should return an HTTP 200 and nsddyn code 403", func() {
				var sr = strings.NewReader(`{"username": "alex", "password": "badpassword", "ipaddr": "192.0.2.4", "hostnames": [ "host1", "host2" ], "version": "0.1.0"}`)
				request, _ := http.NewRequest("POST", uri, sr)
				mux.ServeHTTP(writer, request)
				Expect(writer.Code).To(Equal(http.StatusOK))

				json.Unmarshal(writer.Body.Bytes(), &nb)
				Expect(nb.Code).To(Equal("403"))
			})
		})

		Context("With a good username", func() {
			It("should return nsddyn code 200", func() {
				var sr = strings.NewReader(`{"username": "alex", "password": "password", "ipaddr": "192.0.2.4", "hostnames": [ "host1", "host2" ], "version": "0.1.0"}`)
				request, _ := http.NewRequest("POST", uri, sr)
				mux.ServeHTTP(writer, request)

				json.Unmarshal(writer.Body.Bytes(), &nb)
				Expect(nb.Code).To(Equal("200"))
			})
		})

		Context("With a bad username", func() {
			It("should return nsddyn code 403", func() {
				var sr = strings.NewReader(`{"username": "badusername", "password": "password", "ipaddr": "192.0.2.4", "hostnames": [ "host1", "host2" ], "version": "0.1.0"}`)
				request, _ := http.NewRequest("POST", uri, sr)
				mux.ServeHTTP(writer, request)

				json.Unmarshal(writer.Body.Bytes(), &nb)
				Expect(nb.Code).To(Equal("403"))
			})
		})

		Context("With a bad host", func() {
			It("should return nsddyn code 418", func() {
				var sr = strings.NewReader(`{"username": "alex", "password": "password", "ipaddr": "192.0.2.4", "hostnames": [ "host1", "badhost" ], "version": "0.1.0"}`)
				request, _ := http.NewRequest("POST", uri, sr)
				mux.ServeHTTP(writer, request)

				json.Unmarshal(writer.Body.Bytes(), &nb)
				Expect(nb.Code).To(Equal("418"))
			})
		})

		Context("With a malformed request - missing password", func() {
			It("should return nsddyn code 400", func() {
				var sr = strings.NewReader(`{"username": "alex", "ipaddr": "192.0.2.4", "hostnames": [ "host1", "badhost" ], "version": "0.1.0"}`)
				request, _ := http.NewRequest("POST", uri, sr)
				mux.ServeHTTP(writer, request)

				json.Unmarshal(writer.Body.Bytes(), &nb)
				Expect(nb.Code).To(Equal("400"))
			})
		})

		Context("With a malformed request - totally bogus", func() {
			It("should return nsddyn code 400", func() {
				var sr = strings.NewReader(`{"username":ersio}`)
				request, _ := http.NewRequest("POST", uri, sr)
				mux.ServeHTTP(writer, request)

				json.Unmarshal(writer.Body.Bytes(), &nb)
				Expect(nb.Code).To(Equal("400"))
			})
		})
	})
})
