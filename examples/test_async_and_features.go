package main

import (
	"context"
	"fmt"
	"log"

	"github.com/andreyea/tm1go/pkg/tm1"
)

func main() {
	// Create TM1 config
	cfg := tm1.Config{
		Address:  "localhost",
		Port:     8882,
		User:     "admin",
		Password: "",
		SSL:      true,
		Logging:  true,
	}

	// Create TM1Service
	svc, err := tm1.NewTM1Service(cfg)
	if err != nil {
		log.Fatalf("Failed to create TM1 service: %v", err)
	}
	defer svc.Close()

	ctx := context.Background()

	// Demonstrate async operation API (conceptual - requires actual async operation to test)
	fmt.Println("\n=== Async Operations API Demo ===")
	fmt.Println("Available methods:")
	fmt.Println("  • RetrieveAsyncResponse(ctx, asyncID) - Retrieves async operation result")
	fmt.Println("  • CancelAsyncOperation(ctx, asyncID) - Cancels an async operation")
	fmt.Println("  • CancelRunningOperation(ctx, threadID) - Cancels a running operation")

	// Example: How to use async operations
	fmt.Println("\nExample usage:")
	fmt.Println("  1. Start async operation (e.g., long-running process)")
	fmt.Println("  2. Extract async_id from Location header")
	fmt.Println("  3. Poll result: resp, err := svc.RetrieveAsyncResponse(ctx, asyncID)")
	fmt.Println("  4. Or cancel: err := svc.CancelAsyncOperation(ctx, asyncID)")

	// Note: Actual async operations require a process or operation that runs async
	// This is just demonstrating the API is available

	fmt.Println("\n=== Compact JSON Header Demo ===")
	rest := svc.Rest()

	// Get current Accept header
	fmt.Println("\nDefault Accept header (before):")
	fmt.Println("  application/json;odata.metadata=none,text/plain")

	// Add compact JSON
	originalHeader := rest.AddCompactJSONHeader()
	fmt.Printf("\nOriginal header saved: %s\n", originalHeader)
	fmt.Println("✓ Compact JSON header added (tm1.compact=v0)")
	fmt.Println("  Now: application/json;tm1.compact=v0;odata.metadata=none,text/plain")

	// Test with actual request
	fmt.Println("\n=== Testing with actual request ===")
	version, err := svc.Version(ctx)
	if err != nil {
		log.Printf("Version error: %v", err)
	} else {
		fmt.Printf("✓ Version retrieved with compact JSON: %s\n", version)
	}

	fmt.Println("\n=== Admin Properties Demo ===")

	// Get all admin properties in one call
	isAdmin, _ := svc.IsAdmin(ctx)
	isDataAdmin, _ := svc.IsDataAdmin(ctx)
	isSecurityAdmin, _ := svc.IsSecurityAdmin(ctx)
	isOpsAdmin, _ := svc.IsOpsAdmin(ctx)
	sandboxingDisabled, _ := svc.SandboxingDisabled(ctx)

	fmt.Printf("Admin Properties:\n")
	fmt.Printf("  Is Admin:              %t\n", isAdmin)
	fmt.Printf("  Is Data Admin:         %t\n", isDataAdmin)
	fmt.Printf("  Is Security Admin:     %t\n", isSecurityAdmin)
	fmt.Printf("  Is Ops Admin:          %t\n", isOpsAdmin)
	fmt.Printf("  Sandboxing Disabled:   %t\n", sandboxingDisabled)

	fmt.Println("\n=== All features demonstrated ===")
}
