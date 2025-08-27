package main

import (
	"fmt"
	"log"

	"github.com/andreyea/tm1go/pkg/tm1"
)

func main() {
	// Create configuration
	config := tm1.DefaultConfig()
	config.Address = "localhost"
	config.Port = 8882
	config.User = "admin"
	config.Password = ""
	config.SSL = true
	config.SessionContext = "TM1go-Debug"
	config.Logging = true

	// Create TM1 service
	tm1Service, err := tm1.NewTM1Service(config)
	if err != nil {
		log.Fatalf("Failed to create TM1 service: %v", err)
	}
	defer tm1Service.Close()

	fmt.Printf("Connected to TM1 version: %s\n", tm1Service.Version())

	elements := tm1Service.Elements()

	// First, let's get available dimensions
	fmt.Println("\n=== Testing Raw Element Response ===")

	// Make a raw request to see what's actually returned
	// Let's first check what dimensions are available
	fmt.Println("\n=== Available Dimensions ===")

	// Try to get a single element from Product dimension with more detailed output
	dimensionName := "Product"
	hierarchyName := "Product"

	fmt.Printf("Testing single element from %s dimension...\n", dimensionName)

	// Get the first element to examine its structure
	elementNames, err := elements.GetElementNames(dimensionName, hierarchyName)
	if err != nil {
		log.Fatalf("Failed to get element names: %v", err)
	}

	if len(elementNames) > 0 {
		fmt.Printf("First element name: %s\n", elementNames[0])

		// Try to get this specific element
		element, err := elements.Get(dimensionName, hierarchyName, elementNames[0])
		if err != nil {
			fmt.Printf("Failed to get single element: %v\n", err)
		} else {
			fmt.Printf("Successfully retrieved element: %+v\n", element)
		}
	}
}
