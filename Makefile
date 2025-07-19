.PHONY: build run clean test lint deps build-all man view-man

# Build variables
BINARY_NAME=get-repo
BUILD_DIR=.
VERSION=$(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse --short HEAD)
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X get-repo/pkg/version.Version=${VERSION} -X get-repo/pkg/version.GitCommit=${COMMIT} -X get-repo/pkg/version.BuildDate=${BUILD_DATE}"

# Default target
all: build

# Build the binary
build:
	go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} cmd/get-repo/main.go

# Run the application
run:
	go run cmd/get-repo/main.go

# Clean build artifacts
clean:
	rm -f ${BUILD_DIR}/${BINARY_NAME}
	rm -f debug.log

# Run tests
test:
	go test -v ./...

# Run linter
lint:
	golangci-lint run

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build for multiple platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-darwin-amd64 cmd/get-repo/main.go
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-darwin-arm64 cmd/get-repo/main.go
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-linux-amd64 cmd/get-repo/main.go
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-linux-arm64 cmd/get-repo/main.go

# Generate man page
man:
	@echo "Generating man page..."
	@if command -v go-md2man >/dev/null 2>&1; then \
		go-md2man -in docs/get-repo.1.md -out docs/get-repo.1; \
		echo "Man page generated at docs/get-repo.1"; \
	else \
		echo "go-md2man not found. Install with: go install github.com/cpuguy83/go-md2man/v2@latest"; \
		exit 1; \
	fi

# View man page
view-man: man
	man ./docs/get-repo.1