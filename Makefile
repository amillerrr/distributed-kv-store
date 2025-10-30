.PHONY: proto clean build run test help

# Variables
PROTO_DIR := proto
PROTO_FILES := $(wildcard $(PROTO_DIR)/*.proto)

# Default
help:
	@echo "Available targets:"
	@echo "  proto         - Generate Go code from .proto files"
	@echo "  build-server  - Build the server binary"
	@echo "  build-client  - Build the server binary"
	@echo "  build-all     - Build the server binary"
	@echo "  run           - Run the server locally"
	@echo "  test          - Run tests"
	@echo "  clean         - Remove generated files and binaries"

# Generate Go code from proto files
proto:
	@echo "Generating gRPC code from proto files"
	@protoc \
		--proto_path=. \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)
	@echo "Proto generation complete"

# Build server binary
build-server: proto
	@echo "Building server"
	go build -o bin/kvstore-server ./cmd/server
	@echo "Complete: bin/kvstore-server"

# Build the client binary
build-client: proto
	@echo "Building client"
	@go build -o bin/kvstore-client ./cmd/client
	@echo "Complete: bin/kvstore-client"

# Build both server and client
build-all: proto
	@echo "Building all binaries"
	@go build -o bin/kvstore-server ./cmd/server
	@go build -o bin/kvstore-client ./cmd/client
	@echo "Complete: bin/kvstore-server, bin/kvstore-client"

# Run the server
run: build
	@./bin/kvstore-server

# Run tests
test:
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

# Clean up generated files and binaries
clean:
	@echo "Cleaning generated files"
	@rm -f $(PROTO_DIR)/*.pb.go
	@rm -rf bin/
	@rm -f coverage.out
	@echo "Clean complete"

# Download dependencies
deps:
	@echo "Downloading dependencies"
	@go mod download
	@go mod tidy
	@echo "Dependencies ready"
