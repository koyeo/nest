# Nest - Makefile
# ============================================================================

APP_NAME    := nest
MODULE      := github.com/koyeo/nest
GO          := go
GOFLAGS     ?=
GO_LDFLAGS  :=

# Build output
BUILD_DIR   := build
BINARY      := $(BUILD_DIR)/$(APP_NAME)

# Version info (override via: make build VERSION=1.2.3)
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT      := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME  := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GO_LDFLAGS  += -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)

# Cross-compilation targets
PLATFORMS   := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# ============================================================================
# Development
# ============================================================================

.PHONY: all
all: tidy lint build ## Default: tidy, lint, build

.PHONY: build
build: ## Build binary for current platform
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(GO_LDFLAGS)" -o $(BINARY) .

.PHONY: install
install: ## Install binary to $GOPATH/bin
	$(GO) install $(GOFLAGS) -ldflags "$(GO_LDFLAGS)" .

.PHONY: run
run: build ## Build and run
	$(BINARY)

.PHONY: dev
dev: ## Run without building binary (go run)
	$(GO) run $(GOFLAGS) . $(ARGS)

# ============================================================================
# Code Quality
# ============================================================================

.PHONY: fmt
fmt: ## Format code
	$(GO) fmt ./...

.PHONY: vet
vet: ## Run go vet
	$(GO) vet ./...

.PHONY: lint
lint: vet ## Run linters (go vet)

.PHONY: tidy
tidy: ## Tidy and verify module dependencies
	$(GO) mod tidy
	$(GO) mod verify

# ============================================================================
# Testing
# ============================================================================

.PHONY: test
test: ## Run tests
	$(GO) test $(GOFLAGS) ./...

.PHONY: test-verbose
test-verbose: ## Run tests with verbose output
	$(GO) test $(GOFLAGS) -v ./...

.PHONY: test-cover
test-cover: ## Run tests with coverage report
	@mkdir -p $(BUILD_DIR)
	$(GO) test $(GOFLAGS) -coverprofile=$(BUILD_DIR)/coverage.out ./...
	$(GO) tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "Coverage report: $(BUILD_DIR)/coverage.html"

# ============================================================================
# Cross-Compilation & Release
# ============================================================================

.PHONY: build-all
build-all: ## Build for all platforms
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		output=$(BUILD_DIR)/$(APP_NAME)-$${os}-$${arch}; \
		if [ "$${os}" = "windows" ]; then output=$${output}.exe; fi; \
		echo "Building $${os}/$${arch} -> $${output}"; \
		GOOS=$${os} GOARCH=$${arch} $(GO) build $(GOFLAGS) -ldflags "$(GO_LDFLAGS)" -o $${output} . || exit 1; \
	done

.PHONY: checksums
checksums: build-all ## Generate SHA256 checksums for all builds
	@cd $(BUILD_DIR) && shasum -a 256 $(APP_NAME)-* > checksums.txt
	@echo "Checksums: $(BUILD_DIR)/checksums.txt"

# ============================================================================
# Cleanup
# ============================================================================

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)
	$(GO) clean

# ============================================================================
# Help
# ============================================================================

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
