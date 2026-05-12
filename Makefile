BINARY_NAME=wt
BIN_DIR=bin

.PHONY: all build clean format lint test 

all: build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/ ./cmd/wt

clean:
	rm -rf $(BIN_DIR)

format:
	go fmt ./...

lint:
	go vet ./...
	golangci-lint run ./...

test:
	go test ./...

