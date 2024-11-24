/*
   Copyright (C) 2021, 2022, 2024 Alexander Necheff

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
	"necheff.net/nsddyn/internal/auth"

	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
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
	_ = authFile.AddUser([]byte("password"), "alex0", []string{"host3", "host4"})

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
		dynupd = NewDynUpd(authFile, zonefile.Name(), "example.com", srv.URL)

		mux = http.NewServeMux()
		mux.HandleFunc(uri, dynupd.DynUpdHandler)

		writer = httptest.NewRecorder()

	})

	Describe("HTTP method response", func() {
		Context("With an invalid GET", func() {
			It("should return an HTTP 405", func() {
				request, _ := http.NewRequest(http.MethodGet, uri, nil)
				mux.ServeHTTP(writer, request)
				Expect(writer.Code).To(Equal(http.StatusMethodNotAllowed))
			})
		})
	})

	Describe("nsddyn client request", func() {

		Context("With a good request", func() {
			It("should return an HTTP 200 and nsddyn code 200", func() {
				var sr = strings.NewReader(`{"username": "alex", "password": "password", "ipaddr": "192.0.2.4", "hostnames": [ "host1", "host2" ], "version": "0.1.0"}`)
				request, _ := http.NewRequest(http.MethodPost, uri, sr)
				mux.ServeHTTP(writer, request)
				Expect(writer.Code).To(Equal(http.StatusOK))

				err := json.Unmarshal(writer.Body.Bytes(), &nb)
				Expect(err).NotTo(HaveOccurred())
				Expect(nb.Code).To(Equal("200"))
			})
		})

		Context("With a good request for multiple hosts where only one host has a new IP", func() {
			It("should return an HTTP 200 and nsddyn code 200", func() {
				var sr = strings.NewReader(`{"username": "alex0", "password": "password", "ipaddr": "192.0.2.10", "hostnames": [ "host3" ], "version": "0.1.0"}`)
				request, _ := http.NewRequest(http.MethodPost, uri, sr)
				mux.ServeHTTP(writer, request)

				// host3 already has an A record with IP 192.0.2.10 but host4 is new so we should still reload the zone
				sr = strings.NewReader(`{"username": "alex0", "password": "password", "ipaddr": "192.0.2.10", "hostnames": [ "host3", "host4" ], "version": "0.1.0"}`)
				request, _ = http.NewRequest(http.MethodPost, uri, sr)
				writer = httptest.NewRecorder()
				mux.ServeHTTP(writer, request)
				Expect(writer.Code).To(Equal(http.StatusOK))

				err := json.Unmarshal(writer.Body.Bytes(), &nb)
				Expect(err).NotTo(HaveOccurred())
				Expect(nb.Code).To(Equal("200"))
			})
		})

		Context("With a good request for multiple hosts where none need updated", func() {
			It("should return an HTTP 200 and nsddyn code 304", func() {
				var sr = strings.NewReader(`{"username": "alex0", "password": "password", "ipaddr": "192.0.2.10", "hostnames": [ "host3", "host4" ], "version": "0.1.0"}`)
				request, _ := http.NewRequest(http.MethodPost, uri, sr)
				mux.ServeHTTP(writer, request)
				Expect(writer.Code).To(Equal(http.StatusOK))

				err := json.Unmarshal(writer.Body.Bytes(), &nb)
				Expect(err).NotTo(HaveOccurred())
				Expect(nb.Code).To(Equal("304"))
			})
		})

		Context("With a bad password", func() {
			It("should return an HTTP 200 and nsddyn code 403", func() {
				var sr = strings.NewReader(`{"username": "alex", "password": "badpassword", "ipaddr": "192.0.2.4", "hostnames": [ "host1", "host2" ], "version": "0.1.0"}`)
				request, _ := http.NewRequest(http.MethodPost, uri, sr)
				mux.ServeHTTP(writer, request)
				Expect(writer.Code).To(Equal(http.StatusOK))

				err := json.Unmarshal(writer.Body.Bytes(), &nb)
				Expect(err).NotTo(HaveOccurred())
				Expect(nb.Code).To(Equal("403"))
			})
		})

		Context("With a good username", func() {
			It("should return nsddyn code 200", func() {
				var sr = strings.NewReader(`{"username": "alex", "password": "password", "ipaddr": "192.0.2.5", "hostnames": [ "host1", "host2" ], "version": "0.1.0"}`)
				request, _ := http.NewRequest(http.MethodPost, uri, sr)
				mux.ServeHTTP(writer, request)

				err := json.Unmarshal(writer.Body.Bytes(), &nb)
				Expect(err).NotTo(HaveOccurred())
				Expect(nb.Code).To(Equal("200"))
			})
		})

		Context("With a bad username", func() {
			It("should return nsddyn code 403", func() {
				var sr = strings.NewReader(`{"username": "badusername", "password": "password", "ipaddr": "192.0.2.4", "hostnames": [ "host1", "host2" ], "version": "0.1.0"}`)
				request, _ := http.NewRequest(http.MethodPost, uri, sr)
				mux.ServeHTTP(writer, request)

				err := json.Unmarshal(writer.Body.Bytes(), &nb)
				Expect(err).NotTo(HaveOccurred())
				Expect(nb.Code).To(Equal("403"))
			})
		})

		Context("With a bad host", func() {
			It("should return nsddyn code 418", func() {
				var sr = strings.NewReader(`{"username": "alex", "password": "password", "ipaddr": "192.0.2.4", "hostnames": [ "host1", "badhost" ], "version": "0.1.0"}`)
				request, _ := http.NewRequest(http.MethodPost, uri, sr)
				mux.ServeHTTP(writer, request)

				err := json.Unmarshal(writer.Body.Bytes(), &nb)
				Expect(err).NotTo(HaveOccurred())
				Expect(nb.Code).To(Equal("418"))
			})
		})

		Context("With a malformed request - missing password", func() {
			It("should return nsddyn code 400", func() {
				var sr = strings.NewReader(`{"username": "alex", "ipaddr": "192.0.2.4", "hostnames": [ "host1", "badhost" ], "version": "0.1.0"}`)
				request, _ := http.NewRequest(http.MethodPost, uri, sr)
				mux.ServeHTTP(writer, request)

				err := json.Unmarshal(writer.Body.Bytes(), &nb)
				Expect(err).NotTo(HaveOccurred())
				Expect(nb.Code).To(Equal("400"))
			})
		})

		Context("With a malformed request - totally bogus", func() {
			It("should return nsddyn code 400", func() {
				var sr = strings.NewReader(`{"username":ersio}`)
				request, _ := http.NewRequest(http.MethodPost, uri, sr)
				mux.ServeHTTP(writer, request)

				err := json.Unmarshal(writer.Body.Bytes(), &nb)
				Expect(err).NotTo(HaveOccurred())
				Expect(nb.Code).To(Equal("400"))
			})
		})

		Context("With many active clients", func() {
			It("should not corrupt the zone file", Label("slow"), func() {
				if testing.Short() {
					Skip("skipping in short mode...")
				}
				// Use a statistical approch to try and coax a race condition to occur;
				// i.e. use a large number of goroutines to hammer dynupd and see if it breaks.
				//
				// Run with `go test -race` for best results.
				// If the zonefile ends up getting malformed, dynupd will kick back an error from go-zonefile.
				var wg sync.WaitGroup
				for i := 0; i <= 1024; i++ {
					wg.Add(1)
					go func(addr int) {
						defer wg.Done()
						defer GinkgoRecover()

						var nnb nsddynbody
						w := httptest.NewRecorder()

						sr := strings.NewReader(`{"username":"alex","password":"password","ipaddr":"192.0.2.` + strconv.Itoa(addr) + `","hostnames":["host1"],"version":"0.1.0"}`)
						request, _ := http.NewRequest(http.MethodPost, uri, sr)
						mux.ServeHTTP(w, request)

						err := json.Unmarshal(w.Body.Bytes(), &nnb)
						Expect(err).NotTo(HaveOccurred())
						Expect(nnb.Code).To(Equal("200"))
					}(i)

					wg.Add(1)
					go func(addr int) {
						defer wg.Done()
						defer GinkgoRecover()

						var nnb nsddynbody
						w := httptest.NewRecorder()

						sr := strings.NewReader(`{"username":"alex","password":"password","ipaddr":"192.0.2.` + strconv.Itoa(addr) + `","hostnames":["host2"],"version":"0.1.0"}`)
						request, _ := http.NewRequest(http.MethodPost, uri, sr)
						mux.ServeHTTP(w, request)

						err := json.Unmarshal(w.Body.Bytes(), &nnb)
						Expect(err).NotTo(HaveOccurred())
						Expect(nnb.Code).To(Equal("200"))
					}(i)
				}
				wg.Wait()
			})
		})
	})
})
