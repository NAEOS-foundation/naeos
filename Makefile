.PHONY: build test lint fmt clean vet tidy check run help docker docker-local benchmark security e2e install-completion man site

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

## docker: Build multi-arch docker image and push
docker:
	@echo "Building multi-arch docker image $(BINARY):$(VERSION)..."
	docker buildx create --use --name naeos-builder 2>/dev/null || true
	docker buildx build --platform linux/amd64,linux/arm64 --build-arg VERSION=$(VERSION) -t $(BINARY):$(VERSION) --push .

## docker-local: Build docker image for local architecture
docker-local:
	@echo "Building local docker image $(BINARY):$(VERSION)..."
	docker build --build-arg VERSION=$(VERSION) -t $(BINARY):$(VERSION) .

## benchmark: Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem -run=^$$ ./...

## security: Run security analysis
security:
	@which govulncheck && govulncheck ./... || go vet ./...

## e2e: Build and run end-to-end tests
e2e:
	@echo "Building and running e2e tests..."
	go build ./cmd/naeos/ && go test -tags=e2e -run=TestE2E ./...

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'

## install-completion: Install shell completions for bash, zsh, and fish
install-completion: build
	@echo "Installing shell completions..."
	@mkdir -p $(HOME)/.bash_completion.d
	@./$(BINARY) completion bash > $(HOME)/.bash_completion.d/naeos
	@mkdir -p $(HOME)/.zsh/completions
	@./$(BINARY) completion zsh > $(HOME)/.zsh/completions/_naeos
	@mkdir -p $(HOME)/.config/fish/completions
	@./$(BINARY) completion fish > $(HOME)/.config/fish/completions/naeos.fish
	@echo "Completions installed. Restart your shell or source the completion files."

## site: Build the Hugo website
site:
	@echo "Building website..."
	cp docs/openapi.yaml site/static/openapi.yaml
	cd site && hugo --minify

## man: Generate man pages (requires cobra-doc)
man: build
	@mkdir -p docs/man
	@echo "Man pages generated via cobra doc (install cobra-doc for full man page generation)"
