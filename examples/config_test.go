package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/andreyea/tm1go/pkg/tm1"
)

func testBasicAuth() {
	fmt.Println("\n=== Testing Basic Authentication ===")
	cfg := tm1.Config{
		BaseURL:             "https://localhost:8882/api/v1",
		User:                "admin",
		Password:            "",
		SkipSSLVerification: true,
	}

	client, err := tm1.NewTM1Service(cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	version, err := client.Version(ctx)
	if err != nil {
		log.Fatalf("Failed to get version: %v", err)
	}
	fmt.Printf("✓ Basic Auth: Connected, TM1 Version: %s\n", version)
}

func testBase64DecodedPassword() {
	fmt.Println("\n=== Testing Base64 Decoded Password ===")
	// Encode empty password for testing
	encodedPassword := base64.StdEncoding.EncodeToString([]byte(""))

	cfg := tm1.Config{
		BaseURL:             "https://localhost:8882/api/v1",
		User:                "admin",
		Password:            encodedPassword,
		DecodeBase64:        true,
		SkipSSLVerification: true,
	}

	client, err := tm1.NewTM1Service(cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	version, err := client.Version(ctx)
	if err != nil {
		log.Fatalf("Failed to get version: %v", err)
	}
	fmt.Printf("✓ Base64 Password: Connected, TM1 Version: %s\n", version)
}

func testConnectionPooling() {
	fmt.Println("\n=== Testing Connection Pooling ===")
	cfg := tm1.Config{
		BaseURL:             "https://localhost:8882/api/v1",
		User:                "admin",
		Password:            "",
		SkipSSLVerification: true,
		ConnectionPoolSize:  20,
		PoolConnections:     5,
	}

	client, err := tm1.NewTM1Service(cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	version, err := client.Version(ctx)
	if err != nil {
		log.Fatalf("Failed to get version: %v", err)
	}
	fmt.Printf("✓ Connection Pooling: Connected with custom pool settings, Version: %s\n", version)
}

func testCustomSessionContext() {
	fmt.Println("\n=== Testing Custom Session Context ===")
	cfg := tm1.Config{
		BaseURL:             "https://localhost:8882/api/v1",
		User:                "admin",
		Password:            "",
		SkipSSLVerification: true,
		SessionContext:      "MyCustomApp",
	}

	client, err := tm1.NewTM1Service(cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	version, err := client.Version(ctx)
	if err != nil {
		log.Fatalf("Failed to get version: %v", err)
	}
	fmt.Printf("✓ Custom Session Context: Connected with context 'MyCustomApp', Version: %s\n", version)
}

func testAddressAndPort() {
	fmt.Println("\n=== Testing Address + Port (instead of BaseURL) ===")
	cfg := tm1.Config{
		Address:             "localhost",
		Port:                8882,
		SSL:                 true,
		User:                "admin",
		Password:            "",
		SkipSSLVerification: true,
	}

	client, err := tm1.NewTM1Service(cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	version, err := client.Version(ctx)
	if err != nil {
		log.Fatalf("Failed to get version: %v", err)
	}
	fmt.Printf("✓ Address + Port: Connected to localhost:8882, Version: %s\n", version)
}

func testVerifyOptions() {
	fmt.Println("\n=== Testing Verify Option ===")
	// Test with Verify as boolean
	cfg := tm1.Config{
		BaseURL:             "https://localhost:8882/api/v1",
		User:                "admin",
		Password:            "",
		Verify:              false, // Explicitly disable SSL verification
		SkipSSLVerification: true,  // Also set this for compatibility
	}

	client, err := tm1.NewTM1Service(cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	version, err := client.Version(ctx)
	if err != nil {
		log.Fatalf("Failed to get version: %v", err)
	}
	fmt.Printf("✓ Verify=false: Connected with SSL verification disabled, Version: %s\n", version)
}

func main() {
	fmt.Println("Testing tm1go Config Options Compatibility with TM1py")
	fmt.Println("======================================================")

	testBasicAuth()
	testBase64DecodedPassword()
	testConnectionPooling()
	testCustomSessionContext()
	testAddressAndPort()
	testVerifyOptions()

	fmt.Println("\n======================================================")
	fmt.Println("✓ All config option tests passed!")
}
