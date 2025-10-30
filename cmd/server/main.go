package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Create HTTP server for health checks
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health/live", livenessHandler)
	healthMux.HandleFunc("/health/ready", readinessHandler(kvStore))

	httpServer := &http.Server{
		Addr: fmt.Sprintf(":%s", httpPort),
		Handler: healthMux,
	}

	// Channel to listen for errors
	serverErrors := make(chan error, 1)
	
	go func() {
		slog.Info("gRPC server listening", "address", lis.Addr().String())
		serverErrors <- grpcServer.Serve(lis)
	}()

	go func() {
		slog.Info("HTTP health server listening", "address", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Block until signal or error received
	select {
	case err := <-serverErrors:
		slog.Error("server error", "error", err)
	case sig := <-sigChan:
		slog.Info("received shutdown signal", "signal", sig.String())
	}

	slog.Info("initiating graceful shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
	} else {
		slog.Info("HTTP server stopped gracefully")
	}

	grpcServer.GracefulStop()
	slog.Info("gRPC server stopped gracefully")
	slog.Info("shutdown complete")
}

// Log incoming gRPC requests
func loggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	slog.Info("gRPC request", "method", info.FullMethod)

	resp, err := handler(ctx, req)

	if err != nil {
		slog.Error("gRPC request failed", "method", info.FullMethod, "error", err)
	} else {
		slog.Info("gRPC request completed", "method", info.FullMethod)
	}

	return resp, err
}

// Indicate whether the service is running
func livenessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"alive"}`))
}

// Indicate if the service is ready
func readinessHandler(kvStore *service.KVStoreService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// In production, might check db connections, dependant service availability, or resource availability
		// Report ready since using in-memory for test
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	}
}

// Retrieve environment variable or use default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
