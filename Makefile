BUILD_ROOT:=$(CURDIR)
COVERLOG:="./coverage.out"
GOAMD64:=v3

all: nsddynum dynupd dynupd-broker

nsddynum:
	cd cmd/nsddynum; GOAMD64=$(GOAMD64) go build -buildmode=pie

dynupd:
	cd cmd/dynupd; GOAMD64=$(GOAMD64) go build -buildmode=pie

dynupd-broker:
	cd cmd/dynupd-broker; GOAMD64=$(GOAMD64) go build -buildmode=pie

shellcheck:
	@shellcheck client/nsddyncc || true
	@shellcheck scripts/test/automated_integration.sh || true
	@shellcheck test/nsddynum/run.sh || true

quality: shellcheck
	go vet ./... || true
	golangci-lint run --enable godox --enable mnd --enable gosec --enable errorlint --enable gofmt --enable unconvert --enable ginkgolinter ./...

fmt:
	go fmt ./...

integrationtest: all
	$(BUILD_ROOT)/scripts/test/integrate

test: ginkgotestlong integrationtest

ginkgotest:
	@ulimit -n 100000 && BUILD_ROOT=$(BUILD_ROOT) GOAMD64=$(GOAMD64) ginkgo run -v --label-filter='!slow' --race --trace --covermode=atomic --coverprofile=$(COVERLOG) ./...

ginkgotestlong:
	@ulimit -n 100000 && BUILD_ROOT=$(BUILD_ROOT) GOAMD64=$(GOAMD64) ginkgo run -v --race --trace --covermode=atomic --coverprofile=$(COVERLOG) ./...

testcoverage: ginkgotest
	@$(shell go tool cover -html=$(COVERLOG))

linecount:
	find . -not \( -path ./vendor -prune \) -type f -iname '*.go' | xargs wc -l

debian: all
	scripts/package-deb

vulns:
	govulncheck -show verbose ./...

clean:
	cd cmd/nsddynum; go clean
	cd cmd/dynupd; go clean
	cd cmd/dynupd-broker; go clean
	rm -rf $(COVERLOG) *.deb
