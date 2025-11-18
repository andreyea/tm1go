package main

import (
	"context"
	"fmt"
	"log"

	"github.com/andreyea/tm1go/pkg/tm1"
)

func main() {
	// Create configuration for TM1
	cfg := tm1.Config{
		BaseURL:             "https://localhost:8882/api/v1",
		User:                "admin",
		Password:            "",
		SkipSSLVerification: true, // Allow self-signed certificates
	}

	// Create TM1 client
	client, err := tm1.NewTM1Service(cfg)
	if err != nil {
		log.Fatalf("Failed to create TM1 client: %v", err)
	}
	defer client.Close()

	// Get TM1 version
	ctx := context.Background()

	exists, err := client.Dimensions.Exists(ctx, "measure123")
	if err != nil {
		log.Fatalf("Failed to check if dimension exists: %v", err)
	}
	fmt.Printf("Dimension 'measure' exists: %v\n", exists)
	fmt.Println("TM1 client operations completed successfully.")

}
