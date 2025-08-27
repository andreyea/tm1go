// Package main demonstrates the usage of TM1go library
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/andreyea/tm1go/pkg/tm1"
)

func main() {
	// Example 1: Basic connection
	fmt.Println("=== TM1go Example Application ===")
	basicConnectionExample()

	// Example 2: Configuration from file
	fmt.Println("\n=== Configuration from File ===")
	configFileExample()

	// Example 3: Working with elements
	fmt.Println("\n=== Working with Elements ===")
	elementsExample()

	// Example 4: Error handling
	fmt.Println("\n=== Error Handling ===")
	errorHandlingExample()

	// Example 5: Advanced configuration
	fmt.Println("\n=== Advanced Configuration ===")
	advancedConfigExample()
}

// basicConnectionExample demonstrates basic connection setup
func basicConnectionExample() {
	// Create configuration with default values
	config := tm1.DefaultConfig()
	config.Address = "localhost"
	config.Port = 8882
	config.User = "admin"
	config.Password = ""
	config.SSL = true
	config.SessionContext = "TM1go-Example"

	// Create TM1 service
	tm1Service, err := tm1.NewTM1Service(config)
	if err != nil {
		log.Printf("Failed to create TM1 service: %v", err)
		return
	}
	defer tm1Service.Close()

	// Get server information
	fmt.Printf("Connected to TM1 version: %s\n", tm1Service.Version())
	fmt.Printf("Connection status: %t\n", tm1Service.IsConnected())

	// Get metadata (if available)
	if metadata, err := tm1Service.Metadata(); err == nil {
		fmt.Printf("Metadata keys: %d\n", len(metadata))
	}
}

// configFileExample demonstrates loading configuration from file
func configFileExample() {
	// Create example configuration file
	config := &tm1.Config{
		Address:                   "localhost",
		Port:                      8882,
		User:                      "admin",
		Password:                  "",
		SSL:                       true,
		SessionContext:            "TM1go-ConfigFile",
		Timeout:                   durationPtr(30 * time.Second),
		ConnectionPoolSize:        5,
		ReconnectOnSessionTimeout: true,
		Logging:                   true,
	}

	// Save configuration to file
	if err := tm1.SaveConfigToFile(config, "example_config.json"); err != nil {
		log.Printf("Failed to save config: %v", err)
		return
	}
	defer os.Remove("example_config.json") // Cleanup

	// Load configuration from file
	tm1Service, err := tm1.NewTM1ServiceFromConfig("example_config.json")
	if err != nil {
		log.Printf("Failed to load config from file: %v", err)
		return
	}
	defer tm1Service.Close()

	fmt.Printf("Loaded config - Address: %s, Port: %d\n",
		tm1Service.Config().Address, tm1Service.Config().Port)
}

// elementsExample demonstrates working with elements
func elementsExample() {
	config := createTestConfig()
	tm1Service, err := tm1.NewTM1Service(config)
	if err != nil {
		log.Printf("Failed to create TM1 service: %v", err)
		return
	}
	defer tm1Service.Close()

	elements := tm1Service.Elements()
	dimensionName := "Product"
	hierarchyName := "Product"

	// Get element count
	count, err := elements.GetNumberOfElements(dimensionName, hierarchyName)
	if err != nil {
		log.Printf("Failed to get element count: %v", err)
		return
	}
	fmt.Printf("Total elements in %s: %d\n", dimensionName, count)

	// Get element names
	elementNames, err := elements.GetElementNames(dimensionName, hierarchyName)
	if err != nil {
		log.Printf("Failed to get element names: %v", err)
		return
	}
	fmt.Printf("First 5 elements: %v\n", limitSlice(elementNames, 5))

	// Get leaf elements
	leafElements, err := elements.GetLeafElements(dimensionName, hierarchyName)
	if err != nil {
		log.Printf("Failed to get leaf elements: %v", err)
		return
	}
	fmt.Printf("Leaf elements count: %d\n", len(leafElements))

	// Get consolidated elements
	consolidatedElements, err := elements.GetConsolidatedElements(dimensionName, hierarchyName)
	if err != nil {
		log.Printf("Failed to get consolidated elements: %v", err)
		return
	}
	fmt.Printf("Consolidated elements count: %d\n", len(consolidatedElements))

	// Execute MDX query
	mdxParams := tm1.MDXExecuteParams{
		MDX:              fmt.Sprintf("{[%s].[%s].Members}", dimensionName, hierarchyName),
		MemberProperties: []string{"Name"},
		TopRecords:       intPtr(10),
	}

	result, err := elements.ExecuteSetMDX(mdxParams)
	if err != nil {
		log.Printf("Failed to execute MDX: %v", err)
		return
	}
	fmt.Printf("MDX returned %d tuples\n", len(result.Tuples))

	// Check if specific element exists
	exists, err := elements.Exists(dimensionName, hierarchyName, "Total Product")
	if err != nil {
		log.Printf("Failed to check element existence: %v", err)
		return
	}
	fmt.Printf("Element 'Total Product' exists: %t\n", exists)
}

// errorHandlingExample demonstrates error handling patterns
func errorHandlingExample() {
	config := createTestConfig()
	tm1Service, err := tm1.NewTM1Service(config)
	if err != nil {
		log.Printf("Failed to create TM1 service: %v", err)
		return
	}
	defer tm1Service.Close()

	elements := tm1Service.Elements()

	// Try to get a non-existent element to demonstrate error handling
	_, err = elements.Get("NonExistentDimension", "NonExistentHierarchy", "NonExistentElement")
	if err != nil {
		handleTM1Error(err)
	}

	// Try with invalid MDX to show different error type
	mdxParams := tm1.MDXExecuteParams{
		MDX: "INVALID MDX SYNTAX",
	}
	_, err = elements.ExecuteSetMDX(mdxParams)
	if err != nil {
		handleTM1Error(err)
	}
}

// advancedConfigExample demonstrates advanced configuration options
func advancedConfigExample() {
	config := tm1.DefaultConfig()
	config.Address = "localhost"
	config.Port = 8882
	config.User = "admin"
	config.Password = ""
	config.SSL = true

	// Advanced settings
	config.Timeout = durationPtr(60 * time.Second)
	config.ConnectionPoolSize = 20
	config.AsyncRequestsMode = true
	config.CancelAtTimeout = true
	config.ReconnectOnSessionTimeout = true
	config.Logging = true

	// Proxy settings (example)
	config.Proxies = map[string]string{
		"http":  "http://proxy.company.com:8080",
		"https": "https://proxy.company.com:8080",
	}

	fmt.Printf("Advanced config created with timeout: %v\n", config.Timeout)
	fmt.Printf("Connection pool size: %d\n", config.ConnectionPoolSize)
	fmt.Printf("Async mode: %t\n", config.AsyncRequestsMode)

	// Show configuration as JSON
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err == nil {
		fmt.Printf("Configuration as JSON:\n%s\n", string(configJSON))
	}
}

// Helper functions

func createTestConfig() *tm1.Config {
	config := tm1.DefaultConfig()
	config.Address = "localhost"
	config.Port = 8882
	config.User = "admin"
	config.Password = ""
	config.SSL = true
	return config
}

func handleTM1Error(err error) {
	switch e := err.(type) {
	case *tm1.TM1RestException:
		fmt.Printf("TM1 REST Error: %s (HTTP %d)\n", e.Message, e.StatusCode)
		fmt.Printf("  Method: %s, URL: %s\n", e.Method, e.URL)
		if len(e.Details) > 0 {
			fmt.Printf("  Details: %s\n", e.Details[0].Message)
		}
	case *tm1.TM1TimeoutException:
		fmt.Printf("TM1 Timeout Error: %s (%.2fs)\n", e.Message, e.Timeout)
	case *tm1.TM1VersionDeprecationException:
		fmt.Printf("TM1 Version Deprecation: %s (v%s)\n", e.Message, e.Version)
	default:
		fmt.Printf("General Error: %v\n", err)
	}
}

func limitSlice(slice []string, limit int) []string {
	if len(slice) <= limit {
		return slice
	}
	return slice[:limit]
}

func durationPtr(d time.Duration) *time.Duration {
	return &d
}

func intPtr(i int) *int {
	return &i
}
