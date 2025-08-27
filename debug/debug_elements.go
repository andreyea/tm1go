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

	// Make a raw request to see what's actually returned
	// Let's first check what dimensions are available
	fmt.Println("\n=== Available Dimensions ===")

	// Try to get a single element from Product dimension with more detailed output
	dimensionName := "Product"
	hierarchyName := "Product"

	fmt.Printf("Testing single element from %s dimension...\n", dimensionName)

	// Get all elements to examine their structure
	elementList, err := elements.GetElements(dimensionName, hierarchyName)
	if err != nil {
		log.Fatalf("Failed to get elements: %v", err)
	}

	if len(elementList) > 0 {
		fmt.Printf("First element details: Name='%s', Type=%d, Level=%d, Index=%d\n",
			elementList[0].Name, int(elementList[0].Type), elementList[0].Level, elementList[0].Index)

		// Also show a few more elements
		count := len(elementList)
		if count > 5 {
			count = 5
		}
		fmt.Printf("First %d elements:\n", count)
		for i := 0; i < count; i++ {
			fmt.Printf("  %d: %s (Type: %d)\n", i+1, elementList[i].Name, int(elementList[i].Type))
		}

		// Try to get this specific element by name
		// Let's try with a simpler element name that might not have encoding issues
		var testElement *tm1.Element
		var testErr error

		// Try different elements until we find one that works
		for i := 0; i < len(elementList) && i < 10; i++ {
			testElement, testErr = elements.Get(dimensionName, hierarchyName, elementList[i].Name)
			if testErr == nil {
				fmt.Printf("Successfully retrieved element: Name='%s', Type=%d\n", testElement.Name, int(testElement.Type))
				break
			} else {
				fmt.Printf("Failed element %d '%s': %v\n", i+1, elementList[i].Name, testErr)
			}
		}

		if testErr != nil {
			fmt.Printf("All element retrieval attempts failed\n")
		}
	}
}
