BINARY_NAME=wt
BIN_DIR=bin

.PHONY: all build clean format check-format lint test 

all: build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/ ./cmd/wt

clean:
	rm -rf $(BIN_DIR)

format:
	go fmt ./...

check-format:
	@test -z "$$(gofmt -s -l .)" || (echo "The following files need formatting:" && gofmt -s -l . && exit 1)

lint:
	go vet ./...
	golangci-lint run ./...

test:
	go test ./...


