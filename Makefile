BUILD_ROOT:=$(CURDIR)
COVERLOG:="./coverage.out"
GOAMD64:=v3

all: nsddynum dynupd dynupd-broker nsddynd

nsddynum:
	cd cmd/nsddynum; GOAMD64=$(GOAMD64) go build -buildmode=pie

dynupd:
	cd cmd/dynupd; GOAMD64=$(GOAMD64) go build -buildmode=pie

dynupd-broker:
	cd cmd/dynupd-broker; GOAMD64=$(GOAMD64) go build -buildmode=pie

nsddynd:
	cd cmd/nsddynd; GOAMD64=$(GOAMD64) go build -buildmode=pie

shellcheck:
	@shellcheck client/nsddyncc || true
	@shellcheck podman/integrate/automated_integration.sh || true
	@shellcheck podman/integrate/nsd.Deploy || true
	@shellcheck podman/integrate/dynupd.Deploy || true
	@shellcheck podman/integrate/runtest.Deploy || true
	@shellcheck podman/integrate/podtest || true
	@shellcheck podman/nsddynum/runtest || true
	@shellcheck podman/nsddynum/nsddynum.Deploy || true
	@shellcheck podman/nsddynum/podtest || true

quality: shellcheck
	go vet ./... || true
	golangci-lint run --enable godox --enable mnd --enable gosec --enable errorlint --enable gofmt --enable unconvert --enable ginkgolinter ./...

fmt:
	go fmt ./...

integrationtest: all
	podman/integrate/podtest || echo "dynupd test failed!"
	podman/nsddynum/podtest || echo "nsddynum test failed!"

test: ginkgotestlong integrationtest

ginkgotest:
	@BUILD_ROOT=$(BUILD_ROOT) GOAMD64=$(GOAMD64) ginkgo run -v --label-filter='!slow' --race --trace --covermode=atomic --coverprofile=$(COVERLOG) ./...

ginkgotestlong:
	@BUILD_ROOT=$(BUILD_ROOT) GOAMD64=$(GOAMD64) ginkgo run -v --race --trace --covermode=atomic --coverprofile=$(COVERLOG) ./...

testcoverage: ginkgotest
	@$(shell go tool cover -html=$(COVERLOG))

linecount:
	find . -not \( -path ./vendor -prune \) -type f -iname '*.go' | xargs wc -l

debian: all
	scripts/package-deb

vulns:
	govulncheck -show verbose ./...

cleanpodman:
	podman/clean

clean:
	cd cmd/nsddynum; go clean
	cd cmd/dynupd; go clean
	cd cmd/dynupd-broker; go clean
	cd cmd/nsddynd; go clean
	rm -rf $(COVERLOG) *.deb
