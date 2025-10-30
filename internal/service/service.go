package service

import (
	"context"
	"log/slog"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/amillerrr/distributed-kv-store/proto"
)

type KVStoreService struct {
	pb.UnimplementedKeyValueStoreServer
	mu sync.RWMutex
	store sync.Map // In-memory storage (thread-safe)
}

func NewKVStoreService() *KVStoreService {
	slog.Info("initializing KV store service")
	return &KVStoreService{}
}

// Retrieve value by key
func (s *KVStoreService) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	if req.Key == "" {
		slog.Warn("get request with empty key")
		return nil, status.Error(codes.InvalidArgument, "key cannot be empty")
	}

	slog.Info("get request", "key", req.Key)

	value, found := s.store.Load(req.Key)
	if !found {
		slog.Info("key not found", "key", req.Key)
		return &pb.GetResponse{
			Value: "",
			Found: false,
		}, nil
	}

	valueStr, ok := value.(string)
	if !ok {
		slog.Error("stored value is not a string", "key", req.Key)
		return nil, status.Error(codes.Internal, "internal storage error")
	}

	slog.Info("kkey retrieved successfully", "key", req.Key)
	return &pb.GetResponse{
		Value: valueStr,
		Found: true,
	}, nil
}

// Store k/v pair
func (s *KVStoreService) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	if req.Key == "" {
		slog.Warn("set request with empty key")
		return nil, status.Error(codes.InvalidArgument, "key cannot be empty")
	} 

	slog.Info("set request", "key", req.Key)

	// Store the value
	s.store.Store(req.Key, req.Value)

	slog.Info("key stored successfully", "key", req.Key, "value_length", len(req.Value))

	return &pb.SetResponse{
		Success: true,
		Message: "key stored successfully",
	}, nil
}

// Stream changes for matching keys
func (s *KVStoreService) Subscribe(req *pb.SubscribeRequest, stream pb.KeyValueStore_SubscribeServer) error {
	// TODO: need to implement this
	return nil
}
