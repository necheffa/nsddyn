NSDDYNUM_DIR:=$(CURDIR)/nsddynum

all: bin

.PHONY: bin
bin: ## build the binary proper
	$(MAKE) BUILD_ROOT=$(BUILD_ROOT) VERSION=$(VERSION) CMD_DIR=$(CMD_DIR) GOPATH=$(GOPATH) -C $(NSDDYNUM_DIR) bin

.PHONY: fmt
fmt: ## run `go fmt` on all source files
	go list -f {{.Dir}} ./... | xargs gofmt -w -s -d
