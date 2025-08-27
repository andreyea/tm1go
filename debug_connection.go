package main

import (
	"fmt"
	"log"

	"github.com/andreyea/tm1go/pkg/tm1"
)

func main() {
	fmt.Println("=== TM1go Debug Test ===")

	// Test configuration - adjust these values for your TM1 server
	config := tm1.DefaultConfig()
	config.Address = "localhost"
	config.Port = 8882
	config.User = "admin"
	config.Password = "apple" // Try with a common default password
	config.SSL = true
	config.SessionContext = "TM1go-Debug"
	config.Logging = true // Enable logging to see detailed info

	fmt.Printf("Attempting to connect to: %s:%d (SSL: %t)\n", config.Address, config.Port, config.SSL)
	fmt.Printf("User: %s\n", config.User)

	// Create TM1 service
	tm1Service, err := tm1.NewTM1Service(config)
	if err != nil {
		log.Printf("❌ Failed to create TM1 service: %v", err)

		// Try to provide more specific debugging info
		if restErr, ok := err.(*tm1.TM1RestException); ok {
			fmt.Printf("   Status Code: %d\n", restErr.StatusCode)
			fmt.Printf("   Method: %s\n", restErr.Method)
			fmt.Printf("   URL: %s\n", restErr.URL)

			if restErr.StatusCode == 401 {
				fmt.Println("   🚨 Authentication failed! Possible issues:")
				fmt.Println("      - Wrong password")
				fmt.Println("      - Wrong username")
				fmt.Println("      - Server requires different authentication mode")
				fmt.Println("   💡 Try setting the correct password in the config.Password field")
			}
		}

		// Test with different common passwords
		fmt.Println("\n--- Trying with empty password ---")
		config.Password = ""
		tm1Service2, err2 := tm1.NewTM1Service(config)
		if err2 != nil {
			log.Printf("❌ Also failed with empty password: %v", err2)
		} else {
			fmt.Printf("✅ Connected successfully with empty password!\n")
			fmt.Printf("Version: %s\n", tm1Service2.Version())
			tm1Service2.Close()
			return
		}

		return
	}

	defer tm1Service.Close()

	// Success!
	fmt.Printf("✅ Connected successfully!\n")
	fmt.Printf("Version: %s\n", tm1Service.Version())
	fmt.Printf("Connection status: %t\n", tm1Service.IsConnected())

	// Try to get some basic info
	if metadata, err := tm1Service.Metadata(); err == nil {
		fmt.Printf("Metadata available with %d keys\n", len(metadata))
	} else {
		log.Printf("Failed to get metadata: %v", err)
	}
}
