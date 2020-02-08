NSDDYNUM_DIR:=$(CURDIR)/nsddynum
DYNUPD_DIR:=$(CURDIR)/dynupd

all: bin

.PHONY: bin
bin: ## build the binary proper
	@$(MAKE) BUILD_ROOT=$(BUILD_ROOT) VERSION=$(VERSION) CMD_DIR=$(CMD_DIR) GOPATH=$(GOPATH) -C $(NSDDYNUM_DIR) bin
	@$(MAKE) BUILD_ROOT=$(BUILD_ROOT) VERSION=$(VERSION) CMD_DIR=$(CMD_DIR) GOPATH=$(GOPATH) -C $(DYNUPD_DIR) bin

.PHONY: fmt
fmt: ## run `go fmt` on all source files
	@go list -f {{.Dir}} ./... | xargs gofmt -w -s -d
