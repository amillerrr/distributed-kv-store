package service

import (
	"context"
	"sync"

	pb "github.com/amillerrr/distributed-kv-store/proto"
)

type KVStoreService struct {
	pb.UnimplementedKeyValueStoreServer
	mu sync.RWMutex
	store sync.Map // In-memory storage
}

func NewKVStoreService() *KVStoreService {
	return &KVStoreService{}
}

// Retrieve value by key
func (s *KVStoreService) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	// TODO: need to implement this
	return &pb.GetResponse{
		Value: "",
		Found: false,
	}, nil
}

// Store k/v pair
func (s *KVStoreService) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	// TODO: need to implement this
	return &pb.SetResponse{
		Success: true,
		Message: "not yet implemented",
	}, nil
}

// Stream changes for matching keys
func (s *KVStoreService) Subscribe(req *pb.SubscribeRequest, stream pb.KeyValueStore_SubscribeServer) error {
	// TODO: need to implement this
	return nil
}
