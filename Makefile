.PHONY: build test lint fmt clean vet tidy check run help

# Variables
BINARY := naeos
MODULE := github.com/NAEOS-foundation/naeos
CMD := ./cmd/naeos
VERSION := $(shell cat VERSION 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X 'github.com/NAEOS-foundation/naeos/internal/version.Version=$(VERSION)' \
           -X 'github.com/NAEOS-foundation/naeos/internal/version.GitCommit=$(GIT_COMMIT)' \
           -X 'github.com/NAEOS-foundation/naeos/internal/version.BuildDate=$(BUILD_DATE)'

# Default target
all: check build

## build: Build the binary
build:
	@echo "Building $(BINARY) $(VERSION)..."
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(CMD)

## test: Run tests
test:
	@echo "Running tests..."
	go test -v -race -count=1 ./...

## test-cover: Run tests with coverage
test-cover:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo "HTML report: go tool cover -html=coverage.out"

## lint: Run golangci-lint
lint:
	@echo "Running linter..."
	golangci-lint run ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	gofmt -s -w .
	goimports -w -local $(MODULE) .

## vet: Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

## tidy: Run go mod tidy
tidy:
	@echo "Running go mod tidy..."
	go mod tidy

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY)
	rm -f coverage.out

## version: Show current version
version:
	@echo $(VERSION)

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test

## run: Build and run
run: build
	./$(BINARY)

## init: Initialize project (for new users)
init: tidy build
	@echo "Project initialized. Run './$(BINARY) --help' to get started."

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'
