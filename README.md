# Distributed Key-Value Store

A production-ready distributed key-value store built with Go and gRPC. This project demonstrates modern microservice patterns including structured logging, graceful shutdown, health checks, and pub/sub functionality.

## Overview

This is a simple but complete implementation of a distributed KV store service that can be deployed as multiple independent instances. Each instance maintains its own in-memory storage and supports real-time subscriptions to key changes.

The project was built with cloud-native deployment in mind, following SRE and DevOps best practices.

## Features

- **gRPC API** with Protocol Buffers for efficient communication
- **Three RPC methods**: Get, Set, and Subscribe (server-side streaming)
- **Structured logging** using Go's `log/slog` package with JSON output
- **Graceful shutdown** handling for SIGINT and SIGTERM signals
- **HTTP health endpoints** for liveness and readiness checks
- **Thread-safe operations** using sync.Map for concurrent access
- **Pub/sub pattern** with prefix-based key pattern matching
- **Multi-stage Docker builds** for optimized container images
- **Docker Compose** setup for running multiple instances

## Prerequisites

- Go 1.23 or later
- Protocol Buffer compiler (protoc) version 3.19+
- Docker and Docker Compose (for containerized deployment)
- Make

### Installing protoc

On macOS:
```bash
brew install protobuf
```

On Ubuntu/Debian:
```bash
sudo apt install protobuf-compiler
```

### Installing Go protobuf plugins

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Make sure `$GOPATH/bin` is in your PATH.

## Getting Started

Clone the repository and install dependencies:

```bash
git clone <your-repo-url>
cd distributed-kv-store
go mod download
```

Generate protobuf code:

```bash
make proto
```

Build the server and client:

```bash
make build-all
```

## Running Locally

Start the server:

```bash
./bin/kvstore-server
```

The server will start with:
- gRPC listening on port 50051
- HTTP health checks on port 8080

In another terminal, use the client:

```bash
# Set a value
./bin/kvstore-client -op=set -key=user:123 -value="Alice Smith"

# Get a value
./bin/kvstore-client -op=get -key=user:123

# Subscribe to changes
./bin/kvstore-client -op=subscribe -pattern=user:
```

## Docker Deployment

Build the Docker images:

```bash
make docker-build
docker build -f Dockerfile.client -t kvstore-client:latest .
```

Run a single instance:

```bash
make docker-run
```

Test with the client:

```bash
./bin/kvstore-client -op=set -key=docker:test -value="Hello"
```

Stop the container:

```bash
make docker-stop
```

## Multi-Instance Setup with Docker Compose

Start two independent server instances:

```bash
make compose-up
```

This will start:
- Instance 1: gRPC on localhost:50051, health on localhost:8080
- Instance 2: gRPC on localhost:50052, health on localhost:8081

Test both instances:

```bash
# Set on instance 1
./bin/kvstore-client -server=localhost:50051 -op=set -key=user:alice -value="Alice"

# Set on instance 2
./bin/kvstore-client -server=localhost:50052 -op=set -key=user:bob -value="Bob"

# Get from instance 1 (only has alice)
./bin/kvstore-client -server=localhost:50051 -op=get -key=user:alice  # Found
./bin/kvstore-client -server=localhost:50051 -op=get -key=user:bob    # Not found
```

Each instance maintains independent storage, demonstrating the distributed nature of the system.

View logs from all services:

```bash
make compose-logs
```

Stop all services:

```bash
make compose-down
```

## Testing Pub/Sub

Open three terminals.

Terminal 1 - Subscribe to instance 1:
```bash
./bin/kvstore-client -server=localhost:50051 -op=subscribe -pattern=user:
```

Terminal 2 - Subscribe to instance 2:
```bash
./bin/kvstore-client -server=localhost:50052 -op=subscribe -pattern=user:
```

Terminal 3 - Make changes:
```bash
# This will notify Terminal 1 only
./bin/kvstore-client -server=localhost:50051 -op=set -key=user:charlie -value="Charlie"

# This will notify Terminal 2 only
./bin/kvstore-client -server=localhost:50052 -op=set -key=user:david -value="David"
```

Subscribers only receive events from their connected instance.

## Health Checks

The server exposes two HTTP endpoints on port 8080:

```bash
# Liveness - is the service running?
curl http://localhost:8080/health/live

# Readiness - is the service ready to accept traffic?
curl http://localhost:8080/health/ready
```

These endpoints are used by Kubernetes and other orchestrators for health monitoring.

## Project Structure

```
.
├── cmd/
│   ├── server/          # Server entry point
│   └── client/          # CLI client
├── internal/
│   └── service/         # KV store service implementation
├── proto/
│   ├── store.proto      # Protocol buffer definitions
│   ├── store.pb.go      # Generated code (not in git)
│   └── store_grpc.pb.go # Generated code (not in git)
├── Dockerfile           # Server container image
├── Dockerfile.client    # Client container image
├── docker-compose.yml   # Multi-instance orchestration
├── Makefile            # Build automation
└── go.mod              # Go dependencies
```

## Development Workflow

The Makefile provides common development tasks:

```bash
make help              # Show all available targets
make proto             # Regenerate protobuf code
make build             # Build server binary
make build-client      # Build client binary
make build-all         # Build both binaries
make test              # Run tests with coverage
make clean             # Remove generated files and binaries
```

## Configuration

Both server and client can be configured via environment variables:

Server:
- `GRPC_PORT` - gRPC server port (default: 50051)
- `HTTP_PORT` - HTTP health check port (default: 8080)

Client:
- Use the `-server` flag to specify server address

## Architecture Notes

This implementation uses an in-memory store with `sync.Map` for thread-safe concurrent access. In a production system, you would typically:

- Add persistent storage (e.g., Redis, etcd)
- Implement replication between instances
- Add authentication and TLS
- Include metrics and distributed tracing
- Add request rate limiting

The current implementation demonstrates the core patterns and infrastructure needed for a production gRPC service.
