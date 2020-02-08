VERSION=$(shell cat VERSION)

BUILD_ROOT:=$(CURDIR)
CMD_DIR:=$(CURDIR)/cmd
#VER_DIR:=$(CURDIR)/version
BIN_DIR:=$(CURDIR)/bin
OLD_GOPATH:=$(GOPATH)
GOPATH="$(shell dirname $(CURDIR)):$(HOME)/go:$(OLD_GOPATH)"

.PHONY: bin
bin: ## build the standard distribution
	@$(MAKE) BUILD_ROOT=$(BUILD_ROOT) VERSION=$(VERSION) CMD_DIR=$(CMD_DIR) GOPATH=$(GOPATH) -C $(CMD_DIR) bin

.PHONY: fmt
fmt: ## run `go fmt` on all source files
	@$(MAKE) BUILD_ROOT=$(BUILD_ROOT) VERSION=$(VERSION) CMD_DIR=$(CMD_DIR) GOPATH=$(GOPATH) -C $(CMD_DIR) fmt

