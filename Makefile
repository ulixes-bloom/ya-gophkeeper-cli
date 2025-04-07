# Build Configuration
APP_NAME := gophkeeper-cli
MAIN_PKG := ./cmd/gophkeeper-cli
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
LD_FLAGS := -X 'github.com/ulixes-bloom/ya-gophkeeper-cli/cli.Version=$(VERSION)' \
            -X 'github.com/ulixes-bloom/ya-gophkeeper-cli/cli.BuildTime=$(BUILD_TIME)' \
            -X 'github.com/ulixes-bloom/ya-gophkeeper-cli/cli.GitCommit=$(GIT_COMMIT)'

# Build Targets
.PHONY: build
build: ## Build for current platform
	go build -ldflags "$(LD_FLAGS)" -o bin/$(APP_NAME) $(MAIN_PKG)

.PHONY: build-all
build-all: ## Build for all platforms
	@$(MAKE) build-linux
	@$(MAKE) build-mac
	@$(MAKE) build-windows

build-linux: ## Build Linux binaries
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LD_FLAGS)" -o bin/$(APP_NAME)-linux-amd64 $(MAIN_PKG)
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LD_FLAGS)" -o bin/$(APP_NAME)-linux-arm64 $(MAIN_PKG)

build-mac: ## Build macOS binaries
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LD_FLAGS)" -o bin/$(APP_NAME)-darwin-amd64 $(MAIN_PKG)
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LD_FLAGS)" -o bin/$(APP_NAME)-darwin-arm64 $(MAIN_PKG)

build-windows: ## Build Windows binaries
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LD_FLAGS)" -o bin/$(APP_NAME)-windows-amd64.exe $(MAIN_PKG)

# Release Packaging
.PHONY: release
release: clean build-all ## Create release packages
	mkdir -p release
	zip -j release/$(APP_NAME)-$(VERSION)-linux-amd64.zip bin/$(APP_NAME)-linux-amd64
	zip -j release/$(APP_NAME)-$(VERSION)-linux-arm64.zip bin/$(APP_NAME)-linux-arm64
	zip -j release/$(APP_NAME)-$(VERSION)-darwin-amd64.zip bin/$(APP_NAME)-darwin-amd64
	zip -j release/$(APP_NAME)-$(VERSION)-darwin-arm64.zip bin/$(APP_NAME)-darwin-arm64
	zip -j release/$(APP_NAME)-$(VERSION)-windows-amd64.zip bin/$(APP_NAME)-windows-amd64.exe

# Utilities
.PHONY: clean
clean: ## Remove build artifacts
	rm -rf bin/ release/

.PHONY: test
test: ## Run tests
	go test ./...

.PHONY: help
help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help