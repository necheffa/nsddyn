NSDDYNUM_DIR:=$(CURDIR)/nsddynum
DYNUPD_DIR:=$(CURDIR)/dynupd
DYNUPD-BROKER_DIR:=$(CURDIR)/dynupd-broker

all: bin

.PHONY: bin
bin: ## build the binary proper
	@$(MAKE) BUILD_ROOT=$(BUILD_ROOT) VERSION=$(VERSION) CMD_DIR=$(CMD_DIR) GOPATH=$(GOPATH) -C $(NSDDYNUM_DIR) bin
	@$(MAKE) BUILD_ROOT=$(BUILD_ROOT) VERSION=$(VERSION) CMD_DIR=$(CMD_DIR) GOPATH=$(GOPATH) -C $(DYNUPD_DIR) bin
	@$(MAKE) BUILD_ROOT=$(BUILD_ROOT) VERSION=$(VERSION) CMD_DIR=$(CMD_DIR) GOPATH=$(GOPATH) -C $(DYNUPD-BROKER_DIR) bin

.PHONY: unittest
unittest:
	@ulimit -n 100000 && go test -short -race -covermode=atomic -coverprofile=../coverage.out ./...

# specifically omits the -short flag to run all tests, even the long running ones.
.PHONY: unittestlong
unittestlong:
	@ulimit -n 100000 && go test -race -covermode=atomic -coverprofile=../coverage.out ./...

.PHONY: fmt
fmt: ## run `go fmt` on all source files
	@go list -f {{.Dir}} ./... | xargs gofmt -w -s -d

.PHONY: vet
vet:
	@$(MAKE) BUILD_ROOT=$(BUILD_ROOT) VERSION=$(VERSION) CMD_DIR=$(CMD_DIR) GOPATH=$(GOPATH) -C $(NSDDYNUM_DIR) vet
	@$(MAKE) BUILD_ROOT=$(BUILD_ROOT) VERSION=$(VERSION) CMD_DIR=$(CMD_DIR) GOPATH=$(GOPATH) -C $(DYNUPD_DIR) vet
	@$(MAKE) BUILD_ROOT=$(BUILD_ROOT) VERSION=$(VERSION) CMD_DIR=$(CMD_DIR) GOPATH=$(GOPATH) -C $(DYNUPD-BROKER_DIR) vet

.PHONY: golangci-lint
golangci-lint:
	@$(MAKE) BUILD_ROOT=$(BUILD_ROOT) VERSION=$(VERSION) CMD_DIR=$(CMD_DIR) GOPATH=$(GOPATH) -C $(NSDDYNUM_DIR) golangci-lint
	@$(MAKE) BUILD_ROOT=$(BUILD_ROOT) VERSION=$(VERSION) CMD_DIR=$(CMD_DIR) GOPATH=$(GOPATH) -C $(DYNUPD_DIR) golangci-lint
	@$(MAKE) BUILD_ROOT=$(BUILD_ROOT) VERSION=$(VERSION) CMD_DIR=$(CMD_DIR) GOPATH=$(GOPATH) -C $(DYNUPD-BROKER_DIR) golangci-lint

