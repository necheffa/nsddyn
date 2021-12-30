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

package main_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "necheff.net/nsddyn/cmd/dynupd"

	"net/http"
	"net/http/httptest"
)

var _ = Describe("Dynupd", func() {
	var mux *http.ServeMux
	var writer *httptest.ResponseRecorder
	var uri string
	var dynupd *DynUpd

	BeforeEach(func() {
		uri = "/api/dynupd"
		dynupd = new(DynUpd)

		mux = http.NewServeMux()
		mux.HandleFunc(uri, dynupd.DynUpdHandler)

		writer = httptest.NewRecorder()

	})

	Describe("HTTP method response", func() {
		Context("With an invalid GET", func() {
			It("should return a 405", func() {
				request, _ := http.NewRequest("GET", uri, nil)
				mux.ServeHTTP(writer, request)
				Expect(writer.Code).To(Equal(http.StatusMethodNotAllowed))
			})
		})
	})
})
