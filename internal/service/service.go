package service

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/amillerrr/distributed-kv-store/proto"
)

type subscriber struct {
	pattern string
	stream pb.KeyValueStore_SubscribeServer
	events chan *pb.ChangeEvent
}

type KVStoreService struct {
	pb.UnimplementedKeyValueStoreServer
	store sync.Map
	mu sync.RWMutex
	subscribers map[string][]*subscriber
	subID int
}

func NewKVStoreService() *KVStoreService {
	slog.Info("initializing KV store service")
	return &KVStoreService{
		subscribers: make(map[string][]*subscriber),
	}
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

	// Create change event
	event := &pb.ChangeEvent{
		ChangeType: pb.ChangeEvent_SET,
		Key: req.Key,
		Value: req.Value,
		Timestamp: time.Now().UnixMilli(),
	}

	// Notify subscribers
	s.notifySubscribers(event)

	slog.Info("key stored successfully", "key", req.Key, "value_length", len(req.Value))

	return &pb.SetResponse{
		Success: true,
		Message: "key stored successfully",
	}, nil
}

// Stream changes for matching keys
func (s *KVStoreService) Subscribe(req *pb.SubscribeRequest, stream pb.KeyValueStore_SubscribeServer) error {
	if req.KeyPattern == "" {
		slog.Warn("subscribe request with empty pattern")
		return status.Error(codes.InvalidArgument, "key_pattern cannot be empty")
	}

	slog.Info("new subscriber", "pattern", req.KeyPattern)

	// Create subscriber
	sub := &subscriber{
		pattern: req.KeyPattern,
		stream: stream,
		events: make(chan *pb.ChangeEvent, 100),
	}

	// Register subscriber
	s.mu.Lock()
	s.subscribers[req.KeyPattern] = append(s.subscribers[req.KeyPattern], sub)
	subscriberCount := len(s.subscribers[req.KeyPattern])
	s.mu.Unlock()

	slog.Info("subscriber reistered", "pattern", req.KeyPattern, "total_subscribers", subscriberCount)

	// Clean up on exit
	defer func() {
		s.removeSubscriber(req.KeyPattern, sub)
		close(sub.events)
		slog.Info("subscriber unregistered", "pattern", req.KeyPattern)
	}()

	// Stream events to client
	for {
		select {
		case event := <-sub.events:
			if err := stream.Send(event); err != nil {
				slog.Error("failed to send event to subscriber", "pattern", req.KeyPattern, "error", err)
				return err
			}
			slog.Debug("event sent to subscriber", "pattern", req.KeyPattern, "key", event.Key)
		case <-stream.Context().Done():
			slog.Info("subscription stream closed by client", "pattern", req.KeyPattern)
			return nil
		}
	}
}

// Send change events to matching subscribers
func (s *KVStoreService) notifySubscribers(event *pb.ChangeEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	notifiedCount := 0
	for pattern, subs := range s.subscribers {
		if strings.HasPrefix(event.Key, pattern) {
			for _, sub := range subs {
				select {
				case sub.events <- event:
					notifiedCount++
				default:
					slog.Warn("subscriber channel full, skipping event", "pattern", pattern, "key", event.Key)
				}
			}
		}
	}

	if notifiedCount > 0 {
		slog.Info("notified subscribers", "key", event.Key, "subscriber_count", notifiedCount)
	}
}

// remove a subscriber from the list
func (s *KVStoreService) removeSubscriber(pattern string, sub *subscriber) {
	s.mu.Lock()
	defer s.mu.Unlock()

	subs := s.subscribers[pattern]
	for i, existingSub :=  range subs {
		if existingSub == sub {
			s.subscribers[pattern] = append(subs[:i], subs[i+1:]...)
			break
		} 
	}

	// Clean up empty pattern lists
	if len(s.subscribers[pattern]) == 0 {
		delete(s.subscribers, pattern)
	} 
}
