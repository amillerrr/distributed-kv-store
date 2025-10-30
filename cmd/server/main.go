package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/amillerrr/distributed-kv-store/proto"
	"github.com/amillerrr/distributed-kv-store/internal/service"
)

const (
	defaultGRPCPort = "50051"
	defaultHTTPPort = "8080"
)

func main() {
	// Initialize JSON logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Get environment config
	grpcPort := getEnv("GRPC_PORT", defaultGRPCPort)
	httpPort := getEnv("HTTP_PORT", defaultHTTPPort)

	slog.Info("starting distributed KV store server", "grpc_port", grpcPort, "http_port", httpPort)

	// Create TCP listener for gRPC
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		slog.Error("failed to listen", "error", err, "port", grpcPort)
		os.Exit(1)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(loggingInterceptor))

	// Register the KV store service
	kvStore := service.NewKVStoreService()
	pb.RegisterKeyValueStoreServer(grpcServer, kvStore)

	// Register reflection service
	reflection.Register(grpcServer)

	slog.Info("gRPC server listening", "address", lis.Addr().String())

	// Start serving
	if err := grpcServer.Serve(lis); err != nil {
		slog.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}

// Log incoming gRPC requests
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	slog.Info("gRPC request", "method", info.FullMethod)

	resp, err := handler(ctx, req)

	if err != nil {
		slog.Error("gRPC request failed", "method", info.FullMethod, "error", err)
	} else {
		slog.Info("gRPC request completed", "method", info.FullMethod)
	}

	return resp, err
}

// Retrieve environment variable or use default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
