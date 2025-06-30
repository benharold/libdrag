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

## Build the application
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/libdrag

## Run tests
test:
	$(GOTEST) -v ./...

## Run tests with coverage
coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	$(GOCMD) tool cover -func=coverage.out

## Lint the code
lint:
	$(GOLINT) run

## Format the code
fmt:
	$(GOFMT) -s -w .

## Vet the code
vet:
	$(GOCMD) vet ./...

## Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

## Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

## Run all quality checks
check: fmt vet lint test

## Show help
help:
	@echo "Available targets:"
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/##//g'

## Install development dependencies
dev-deps:
	$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
