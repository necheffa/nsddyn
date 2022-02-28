
all: bin

.PHONY: bin
bin: ## build the binary proper
	@$(BUILD_ROOT)/scripts/build/bin
	@mv dynupd $(BUILD_ROOT)/bin/

.PHONY: vet
vet:
	@go vet

.PHONY: golangci-lint
golangci-lint:
	@$(BUILD_ROOT)/scripts/build/golangci-lint ./...
