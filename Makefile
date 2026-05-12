BINARY_NAME=wt
BIN_DIR=bin

.PHONY: all build test vet clean

all: build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/ ./cmd/wt

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -rf $(BIN_DIR)
