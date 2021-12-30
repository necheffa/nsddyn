package main_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "necheff.net/nsddyn/cmd/dynupd-broker"
	. "necheff.net/nsddyn/cmd/dynupd-broker/config"

	"bytes"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Dynupdbroker", func() {
	var bc BrokerConfig
	var mux *http.ServeMux
	var writer *httptest.ResponseRecorder

	BeforeEach(func() {
		bc.Address = "localhost:8080"
		bc.Uri = "/api/dynupdbroker"
		bc.ZoneName = "dyn.example.com"

		mux = http.NewServeMux()
		mux.HandleFunc(bc.Uri, DynupdBroker)

		writer = httptest.NewRecorder()
	})

	Describe("HTTP method response", func() {
		Context("With an invalid GET", func() {
			It("should return a 405", func() {
				request, _ := http.NewRequest("GET", bc.Uri, nil)
				mux.ServeHTTP(writer, request)
				Expect(writer.Code).To(Equal(http.StatusMethodNotAllowed))
			})
		})

		Context("With POST and a valid URI", func() {

			It("should return a 200", func() {
				request, _ := http.NewRequest("POST", bc.Uri, bytes.NewReader([]byte("asdf")))
				mux.ServeHTTP(writer, request)
				Expect(writer.Code).To(Equal(http.StatusOK))
			})

			It("should return a zone reloaded message", func() {
				request, _ := http.NewRequest("POST", bc.Uri, bytes.NewReader([]byte("asdf")))
				mux.ServeHTTP(writer, request)
				Expect(writer.Body).To(Equal("Zone Reloaded"))
			})
		})
	})
})
