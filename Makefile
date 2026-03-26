.PHONY: build test clean install lint fmt deps

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

build:
	go build $(LDFLAGS) -o bin/redash-cli ./cmd/redash-cli

build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/redash-cli-linux-amd64 ./cmd/redash-cli
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/redash-cli-linux-arm64 ./cmd/redash-cli
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/redash-cli-darwin-amd64 ./cmd/redash-cli
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/redash-cli-darwin-arm64 ./cmd/redash-cli
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/redash-cli-windows-amd64.exe ./cmd/redash-cli

test:
	go test -v -race ./...

test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run

fmt:
	go fmt ./...

deps:
	go mod tidy

install:
	go install $(LDFLAGS) ./cmd/redash-cli

clean:
	rm -rf bin/ coverage.out coverage.html
