package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/amillerrr/distributed-kv-store/proto"
)

const (
	defaultServerAddr = "localhost:50051"
	defaultTimeout    = 5 * time.Second
)

func main() {
	// Define command-line flags
	serverAddr := flag.String("server", defaultServerAddr, "Server address (host:port)")
	operation := flag.String("op", "", "Operation: get, set, or subscribe")
	key := flag.String("key", "", "Key for get/set operations")
	value := flag.String("value", "", "Value for set operation")
	pattern := flag.String("pattern", "", "Key pattern for subscribe operation")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Set a value\n")
		fmt.Fprintf(os.Stderr, "  %s -op=set -key=user:123 -value=\"John Doe\"\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Get a value\n")
		fmt.Fprintf(os.Stderr, "  %s -op=get -key=user:123\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Subscribe to changes\n")
		fmt.Fprintf(os.Stderr, "  %s -op=subscribe -pattern=user:\n\n", os.Args[0])
	}

	flag.Parse()

	// Validate operation
	if *operation == "" {
		fmt.Fprintf(os.Stderr, "Error: -op flag is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Create gRPC connection
	conn, err := grpc.NewClient(*serverAddr,grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	fmt.Printf("Connected to server: %s\n", *serverAddr)

	// Create client
	client := pb.NewKeyValueStoreClient(conn)

	// Execute operation
	switch *operation {
	case "get":
		executeGet(client, *key)
	case "set":
		executeSet(client, *key, *value)
	case "subscribe":
		executeSubscribe(client, *pattern)
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid operation '%s'. Must be: get, set, or subscribe\n", *operation)
		os.Exit(1)
	}
}

func executeGet(client pb.KeyValueStoreClient, key string) {
	if key == "" {
		log.Fatal("Error: -key flag is required for get operation")
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	resp, err := client.Get(ctx, &pb.GetRequest{Key: key})
	if err != nil {
		log.Fatalf("Get failed: %v", err)
	}

	if resp.Found {
		fmt.Printf("Key found\n")
		fmt.Printf("  Key:   %s\n", key)
		fmt.Printf("  Value: %s\n", resp.Value)
	} else {
		fmt.Printf("Key not found: %s\n", key)
	}
}

func executeSet(client pb.KeyValueStoreClient, key, value string) {
	if key == "" {
		log.Fatal("Error: -key flag is required for set operation")
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	resp, err := client.Set(ctx, &pb.SetRequest{
		Key:   key,
		Value: value,
	})
	if err != nil {
		log.Fatalf("Set failed: %v", err)
	}

	if resp.Success {
		fmt.Printf("Key stored successfully\n")
		fmt.Printf("  Key:   %s\n", key)
		fmt.Printf("  Value: %s\n", value)
		fmt.Printf("  Message: %s\n", resp.Message)
	} else {
		fmt.Printf("Set failed: %s\n", resp.Message)
	}
}

func executeSubscribe(client pb.KeyValueStoreClient, pattern string) {
	if pattern == "" {
		log.Fatal("Error: -pattern flag is required for subscribe operation")
	}

	ctx := context.Background()

	stream, err := client.Subscribe(ctx, &pb.SubscribeRequest{
		KeyPattern: pattern,
	})
	if err != nil {
		log.Fatalf("Subscribe failed: %v", err)
	}

	fmt.Printf("Subscribed to pattern: %s\n", pattern)
	fmt.Printf("Listening for changes (Ctrl+C to exit)\n\n")

	// Receive events
	for {
		event, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("Stream closed by server")
			break
		}
		if err != nil {
			log.Fatalf("Error receiving event: %v", err)
		}

		// Format timestamp
		timestamp := time.UnixMilli(event.Timestamp).Format(time.RFC3339)

		// Print event
		fmt.Printf("─────────────────────────────────────────\n")
		fmt.Printf("Event: %s\n", event.ChangeType)
		fmt.Printf("  Key:       %s\n", event.Key)
		fmt.Printf("  Value:     %s\n", event.Value)
		fmt.Printf("  Timestamp: %s\n", timestamp)
		fmt.Printf("\n")
	}
}
