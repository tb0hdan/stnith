.PHONY: build clean

all: lint build

lint:
	@golangci-lint run ./...

build:
	@go build -o build/stnith cmd/stnith/main.go

clean:
	@rm -rf build/
