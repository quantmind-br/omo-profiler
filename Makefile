# omo-profiler Makefile

BINARY_NAME := omo-profiler
INSTALL_PATH := $(HOME)/.local/bin
GO := go
GOFLAGS := -v

.PHONY: all build install uninstall test lint clean help

all: build

## Build the binary
build:
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) ./cmd/omo-profiler

## Install binary to ~/.local/bin
install: build
	@mkdir -p $(INSTALL_PATH)
	@cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to $(INSTALL_PATH)"

## Remove installed binary
uninstall:
	@rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Uninstalled $(BINARY_NAME) from $(INSTALL_PATH)"

## Run all tests
test:
	$(GO) test $(GOFLAGS) ./...

## Run linter (requires golangci-lint)
lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not found. Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; exit 1; }
	golangci-lint run ./...

## Remove build artifacts
clean:
	@rm -f $(BINARY_NAME)
	$(GO) clean

## Show help
help:
	@echo "Available targets:"
	@echo "  build     - Build the binary"
	@echo "  install   - Install binary to $(INSTALL_PATH)"
	@echo "  uninstall - Remove installed binary"
	@echo "  test      - Run all tests"
	@echo "  lint      - Run golangci-lint"
	@echo "  clean     - Remove build artifacts"
	@echo "  help      - Show this help"
