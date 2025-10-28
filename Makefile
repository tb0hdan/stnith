.PHONY: build clean

LINTER_VERSION ?= v2.5.0

all: lint test build

lint:
	@golangci-lint run ./...

build:
	@go build -o build/stnith cmd/stnith/main.go

clean:
	@rm -rf build/

test:
	@go test -race -v ./...

tools:
    @curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(shell go env GOPATH)/bin $(LINTER_VERSION)
