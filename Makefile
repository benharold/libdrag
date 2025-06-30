# Build variables
BINARY_NAME=libdrag
VERSION?=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

.PHONY: all build clean test coverage lint fmt vet deps help

# Default target - show help when no arguments provided
all: help

## Show this help message
help:
	@echo ""
	@echo "🏁 libdrag - Drag Racing Timing System"
	@echo "====================================="
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /' | column -t -s ':'
	@echo ""
	@echo "Examples:"
	@echo "  make build     - Build the application"
	@echo "  make test      - Run all tests"
	@echo "  make coverage  - Run tests with coverage report"
	@echo "  make clean     - Clean build artifacts"
	@echo ""

## Build the application
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/libdrag

## Run all tests
test:
	$(GOTEST) -v ./...

## Run tests with coverage report
coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	$(GOCMD) tool cover -func=coverage.out

## Lint the code using golangci-lint
lint:
	$(GOLINT) run

## Format the code using gofmt
fmt:
	$(GOFMT) -s -w .

## Vet the code for potential issues
vet:
	$(GOCMD) vet ./...

## Download and tidy dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

## Clean build artifacts and temporary files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

## Run all checks (fmt, vet, lint, test)
check: fmt vet lint test

## Build and run the application
run: build
	./$(BINARY_NAME)

## Install development dependencies
dev-deps:
	$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
