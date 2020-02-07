VERSION=$(shell cat VERSION)

BUILD_ROOT:=$(CURDIR)
CMD_DIR:=$(CURDIR)/cmd
#VER_DIR:=$(CURDIR)/version
BIN_DIR:=$(CURDIR)/bin
GOPATH=$(shell dirname $(CURDIR))

.PHONY: bin
bin: ## build the standard distribution
	$(MAKE) BUILD_ROOT=$(BUILD_ROOT) VERSION=$(VERSION) CMD_DIR=$(CMD_DIR) GOPATH=$(GOPATH) -C $(CMD_DIR) bin

