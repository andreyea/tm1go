package main

import (
	"context"
	"fmt"
	"log"

	"github.com/andreyea/tm1go/pkg/tm1"
)

func main() {
	fmt.Println("Testing TM1 Session Keep-Alive and Session ID Retrieval")
	fmt.Println("=========================================================")

	// Create configuration with KeepAlive enabled
	cfg := tm1.Config{
		BaseURL:             "https://localhost:8882/api/v1",
		User:                "admin",
		Password:            "",
		SkipSSLVerification: true,
		KeepAlive:           true, // Keep session alive after Close()
	}

	// Create TM1 client
	client, err := tm1.NewTM1Service(cfg)
	if err != nil {
		log.Fatalf("Failed to create TM1 client: %v", err)
	}

	// Get TM1 version to establish session
	ctx := context.Background()
	version, err := client.Version(ctx)
	if err != nil {
		log.Fatalf("Failed to get TM1 version: %v", err)
	}

	fmt.Printf("✓ Connected to TM1 Version: %s\n", version)

	// Retrieve the session ID
	sessionID := client.SessionID()
	if sessionID == "" {
		log.Fatal("Failed to retrieve session ID")
	}

	fmt.Printf("✓ Session ID: %s\n", sessionID)

	// Close the client (with KeepAlive, this won't logout)
	if err := client.Close(); err != nil {
		log.Fatalf("Failed to close client: %v", err)
	}

	fmt.Println("\n✓ Client closed (session kept alive)")
	fmt.Println("\nNow testing session reuse with the retrieved session ID...")

	// Create a new client using the session ID we just retrieved
	cfg2 := tm1.Config{
		BaseURL:             "https://localhost:8882/api/v1",
		SessionID:           sessionID,
		SkipSSLVerification: true,
	}

	client2, err := tm1.NewTM1Service(cfg2)
	if err != nil {
		log.Fatalf("Failed to create second client: %v", err)
	}
	defer client2.Close()

	// Verify the session is still valid
	version2, err := client2.Version(ctx)
	if err != nil {
		log.Fatalf("Failed to get version with reused session: %v", err)
	}

	fmt.Printf("✓ Successfully reused session!\n")
	fmt.Printf("  Session ID: %s\n", sessionID)
	fmt.Printf("  TM1 Version: %s\n", version2)

	fmt.Println("\n=========================================================")
	fmt.Println("✓ Keep-Alive and Session ID Retrieval test completed!")
}
