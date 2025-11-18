package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/andreyea/tm1go/pkg/tm1"
)

// CustomLogger implements the tm1.Logger interface with custom formatting
type CustomLogger struct {
	prefix string
	logger *log.Logger
}

func (c *CustomLogger) Printf(format string, args ...any) {
	c.logger.Printf(c.prefix+" "+format, args...)
}

func testDefaultLogging() {
	fmt.Println("\n=== Test 1: Default Logging (Config.Logging = true) ===")

	cfg := tm1.Config{
		BaseURL:             "https://localhost:8882/api/v1",
		User:                "admin",
		Password:            "",
		SkipSSLVerification: true,
		Logging:             true, // Enable default logging
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

	fmt.Printf("✓ Connected with default logging, Version: %s\n", version)
}

func testCustomLogger() {
	fmt.Println("\n=== Test 2: Custom Logger (WithLogger option) ===")

	// Create custom logger with prefix
	customLogger := &CustomLogger{
		prefix: "[TM1-CUSTOM]",
		logger: log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds),
	}

	cfg := tm1.Config{
		BaseURL:             "https://localhost:8882/api/v1",
		User:                "admin",
		Password:            "",
		SkipSSLVerification: true,
		// Note: Logging is false, but we override with custom logger
	}

	// Use WithLogger option to set custom logger
	client, err := tm1.NewTM1Service(cfg, tm1.WithLogger(customLogger))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	version, err := client.Version(ctx)
	if err != nil {
		log.Fatalf("Failed to get version: %v", err)
	}

	fmt.Printf("✓ Connected with custom logger, Version: %s\n", version)
}

func testNoLogging() {
	fmt.Println("\n=== Test 3: No Logging (default behavior) ===")

	cfg := tm1.Config{
		BaseURL:             "https://localhost:8882/api/v1",
		User:                "admin",
		Password:            "",
		SkipSSLVerification: true,
		// Logging is false (default)
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

	fmt.Printf("✓ Connected with no logging, Version: %s (no HTTP logs above)\n", version)
}

func testMultipleRequests() {
	fmt.Println("\n=== Test 4: Multiple Requests with Logging ===")

	cfg := tm1.Config{
		BaseURL:             "https://localhost:8882/api/v1",
		User:                "admin",
		Password:            "",
		SkipSSLVerification: true,
		Logging:             true,
	}

	client, err := tm1.NewTM1Service(cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Make multiple API calls to see logging
	fmt.Println("\nMaking multiple API calls (watch for HTTP logs):")

	version, _ := client.Version(ctx)
	fmt.Printf("  1. Version: %s\n", version)

	metadata, _ := client.Metadata(ctx)
	fmt.Printf("  2. Metadata: %d bytes\n", len(metadata))

	err = client.Ping(ctx)
	if err == nil {
		fmt.Printf("  3. Ping: OK\n")
	}

	fmt.Println("\n✓ All requests completed")
}

func testFileLogger() {
	fmt.Println("\n=== Test 5: File Logger ===")

	// Create a file logger
	logFile, err := os.Create("tm1_http.log")
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	defer logFile.Close()

	fileLogger := &CustomLogger{
		prefix: "[TM1-FILE]",
		logger: log.New(logFile, "", log.Ldate|log.Ltime),
	}

	cfg := tm1.Config{
		BaseURL:             "https://localhost:8882/api/v1",
		User:                "admin",
		Password:            "",
		SkipSSLVerification: true,
	}

	client, err := tm1.NewTM1Service(cfg, tm1.WithLogger(fileLogger))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	version, err := client.Version(ctx)
	if err != nil {
		log.Fatalf("Failed to get version: %v", err)
	}

	fmt.Printf("✓ Connected with file logger, Version: %s\n", version)
	fmt.Println("✓ HTTP logs written to: tm1_http.log")
}

func main() {
	fmt.Println("Testing tm1go Logger Functionality")
	fmt.Println("====================================")

	testNoLogging()        // First test with no logging (clean output)
	testDefaultLogging()   // Test with default logging
	testCustomLogger()     // Test with custom logger
	testMultipleRequests() // Test multiple requests with logging
	testFileLogger()       // Test file-based logging

	fmt.Println("\n====================================")
	fmt.Println("✓ All logger tests completed!")
}
