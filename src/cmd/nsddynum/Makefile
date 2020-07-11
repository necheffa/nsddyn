
all: bin

.PHONY: bin
bin: ## build the binary proper
	@$(BUILD_ROOT)/scripts/build/bin
	@mv nsddynum $(BUILD_ROOT)/bin/

.PHONY: vet
vet:
	@go vet

.PHONY: gosec
gosec:
	@$(BUILD_ROOT)/scripts/build/gosec .
