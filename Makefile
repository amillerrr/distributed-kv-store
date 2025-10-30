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
	@echo "  deps          - Download dependencies"
	@echo ""
	@echo "Docker targets:"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run server in Docker container"
	@echo "  docker-stop   - Stop and remove Docker container"
	@echo "  docker-logs   - View container logs"
	@echo "  docker-shell  - Open shell in running container"

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

# Docker image name
DOCKER_IMAGE := kvstore-server
DOCKER_TAG := latest

# Build Docker image
docker-build:
	@echo "Building Docker image"
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

# Run Docker container
docker-run:
	@echo "Running Docker container"
	@docker run -d \
		--name kvstore-server \
		-p 50051:50051 \
		-p 8080:8080 \
		$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "  Container started: kvstore-server"
	@echo "  gRPC: localhost:50051"
	@echo "  Health: http://localhost:8080/health/live"

# Stop and remove Docker container
docker-stop:
	@echo "Stopping Docker container"
	@docker stop kvstore-server || true
	@docker rm kvstore-server || true
	@echo "Container stopped and removed"

# View Docker logs
docker-logs:
	@docker logs -f kvstore-server

# Docker shell (for debugging)
docker-shell:
	@docker exec -it kvstore-server sh
