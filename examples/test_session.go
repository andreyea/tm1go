package main

import (
	"context"
	"fmt"
	"log"

	"github.com/andreyea/tm1go/pkg/tm1"
)

func main() {
	fmt.Println("Testing TM1 Connection with Existing Session ID")
	fmt.Println("=================================================")

	// Create configuration using existing session ID
	cfg := tm1.Config{
		BaseURL:             "https://localhost:8882/api/v1",
		User:                "admin",
		Password:            "",
		SkipSSLVerification: true,
	}

	// Create TM1 client with existing session
	client, err := tm1.NewTM1Service(cfg)
	if err != nil {
		log.Fatalf("Failed to create TM1 client: %v", err)
	}
	defer client.Close()

	// Get TM1 version
	ctx := context.Background()
	version, err := client.Version(ctx)
	if err != nil {
		log.Fatalf("Failed to get TM1 version: %v", err)
	}

	fmt.Printf("✓ Successfully connected using existing session!\n")
	fmt.Printf("  Session ID: %s\n", cfg.SessionID)
	fmt.Printf("  TM1 Version: %s\n", version)

	// Test another API call to verify session is working
	metadata, err := client.Metadata(ctx)
	if err != nil {
		log.Fatalf("Failed to get metadata: %v", err)
	}

	fmt.Printf("✓ Metadata retrieved successfully (%d bytes)\n", len(metadata))
	fmt.Println("\nSession reuse test completed successfully!")
}
