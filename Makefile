VERSION=$(shell cat VERSION)

BUILD_ROOT:=$(CURDIR)
CMD_DIR:=$(CURDIR)/cmd
BIN_DIR:=$(CURDIR)/bin

.PHONY: bin
bin: ## build the standard distribution
	@$(MAKE) BUILD_ROOT=$(BUILD_ROOT) VERSION=$(VERSION) CMD_DIR=$(CMD_DIR) -C $(CMD_DIR) bin

.PHONY: vet
vet:
	@$(MAKE) BUILD_ROOT=$(BUILD_ROOT) VERSION=$(VERSION) CMD_DIR=$(CMD_DIR) -C $(CMD_DIR) vet

.PHONY: golangci-lint
golangci-lint:
	@$(MAKE) BUILD_ROOT=$(BUILD_ROOT) VERSION=$(VERSION) CMD_DIR=$(CMD_DIR) -C $(CMD_DIR) golangci-lint

.PHONY: shellcheck
shellcheck:
	@scripts/build/shellcheck client/nsddyncc
	@scripts/build/shellcheck scripts/test/automated_integration.sh
	@scripts/build/shellcheck test/nsddynum/run.sh

.PHONY: quality
quality: vet shellcheck golangci-lint

.PHONY: fmt
fmt: ## run `go fmt` on all source files
	@$(MAKE) BUILD_ROOT=$(BUILD_ROOT) VERSION=$(VERSION) CMD_DIR=$(CMD_DIR) -C $(CMD_DIR) fmt

.PHONY: integrationtest
integrationtest: bin
	$(BUILD_ROOT)/scripts/test/integrate

.PHONY: test
test: unittestlong integrationtest

.PHONY: ginkgotest
ginkgotest:
	@ulimit -n 100000 && export BUILD_ROOT=$(BUILD_ROOT); ginkgo run --label-filter='!slow' --race --covermode=atomic --output-dir=$(BUILD_ROOT) ./...

.PHONY: unittest
unittest:
	@$(MAKE) BUILD_ROOT=$(BUILD_ROOT) VERSION=$(VERSION) CMD_DIR=$(CMD_DIR) -C $(CMD_DIR) unittest

.PHONY: unittestlong
unittestlong:
	@$(MAKE) BUILD_ROOT=$(BUILD_ROOT) VERSION=$(VERSION) CMD_DIR=$(CMD_DIR) -C $(CMD_DIR) unittestlong

.PHONY: package
package: bin
	scripts/package

.PHONY: testcoverage
testcoverage: unittest
	@$(shell go tool cover -html=coverage.out)

.PHONY: linecount
linecount:
	find . -not \( -path ./vendor -prune \) -type f -iname *.go | xargs wc -l

.PHONY: install
install:
	@scripts/install $(BUILD_ROOT)

.PHONY: clean
clean: ## remove old binaries
	rm -rf $(BUILD_ROOT)/bin coverage.out quality.log
	find $(BUILD_ROOT) -type f -iname coverprofile.out -print | xargs rm -f
