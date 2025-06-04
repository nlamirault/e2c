# SPDX-FileCopyrightText: Copyright (C) Nicolas Lamirault <nicolas.lamirault@gmail.com>
# SPDX-License-Identifier: Apache-2.0

.PHONY: all build clean test lint fmt help

BINARY_NAME=e2c
MAIN_PACKAGE=./cmd/e2c
BUILD_DIR=build
GO_FILES=$(shell find . -name "*.go" -type f -not -path "./vendor/*")
GOARCH=$(shell go env GOARCH)
GOOS=$(shell go env GOOS)
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X github.com/nlamirault/e2c/internal/version.Version=$(VERSION)"

# Colors for terminal output
COLOR_RESET=\033[0m
COLOR_BLUE=\033[34m
COLOR_GREEN=\033[32m

help: ## Display this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_BLUE)%-20s$(COLOR_RESET) %s\n", $$1, $$2}'

all: lint test build ## Run lint, test, and build

build: ## Build the binary
	@echo "$(COLOR_GREEN)Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(COLOR_GREEN)Binary built at $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)"

clean: ## Remove build artifacts
	@echo "$(COLOR_GREEN)Cleaning up...$(COLOR_RESET)"
	@rm -rf $(BUILD_DIR)
	@go clean

test: ## Run tests
	@echo "$(COLOR_GREEN)Running tests...$(COLOR_RESET)"
	@go test -v ./...

coverage: ## Run tests with coverage
	@echo "$(COLOR_GREEN)Running tests with coverage...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@go test -coverprofile=$(BUILD_DIR)/coverage.out ./...
	@go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "$(COLOR_GREEN)Coverage report generated at $(BUILD_DIR)/coverage.html$(COLOR_RESET)"

lint: ## Run linters
	@echo "$(COLOR_GREEN)Running linters...$(COLOR_RESET)"
	@golangci-lint run

fmt: ## Format the code
	@echo "$(COLOR_GREEN)Formatting code...$(COLOR_RESET)"
	@gofmt -w -s $(GO_FILES)

install: build ## Install the binary
	@echo "$(COLOR_GREEN)Installing $(BINARY_NAME)...$(COLOR_RESET)"
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

run: ## Run the application
	@echo "$(COLOR_GREEN)Running $(BINARY_NAME)...$(COLOR_RESET)"
	@go run $(MAIN_PACKAGE)

release: ## Build for all platforms
	@echo "$(COLOR_GREEN)Building release binaries...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)/release
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)
	@echo "$(COLOR_GREEN)Release binaries built in $(BUILD_DIR)/release/$(COLOR_RESET)"
