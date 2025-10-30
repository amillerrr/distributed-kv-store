# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git make protobuf-dev

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

RUN make proto

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o kvstore-server \
    ./cmd/server

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates && \
    addgroup -g 1000 kvstore && \
    adduser -D -u 1000 -G kvstore kvstore

WORKDIR /app

COPY --from=builder /build/kvstore-server .

RUN chown -R kvstore:kvstore /app

USER kvstore

# 50051 for gRPC
# 8080 for HTTP health checks
EXPOSE 50051 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health/live || exit 1

LABEL maintainer="andrew@mill3r.la" \
      version="1.0.0" \
      description="Distributed Key-Value Store gRPC Service"

ENTRYPOINT ["./kvstore-server"]
