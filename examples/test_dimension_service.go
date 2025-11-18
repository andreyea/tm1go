package main

import (
	"context"
	"fmt"
	"log"

	"github.com/andreyea/tm1go/pkg/models"
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

	// Example 1: Get all dimension names
	fmt.Println("\n=== Example 1: Get All Dimension Names ===")
	dimNames, err := svc.Dimensions.GetAllNames(ctx, false)
	if err != nil {
		log.Printf("Error getting dimension names: %v", err)
	} else {
		fmt.Printf("Total dimensions: %d\n", len(dimNames))
		fmt.Printf("First 5 dimensions: %v\n", dimNames[:min(5, len(dimNames))])
	}

	// Example 2: Get dimension count
	fmt.Println("\n=== Example 2: Get Dimension Count ===")
	count, err := svc.Dimensions.GetNumberOfDimensions(ctx, false)
	if err != nil {
		log.Printf("Error getting dimension count: %v", err)
	} else {
		fmt.Printf("Number of dimensions: %d\n", count)
	}

	// Example 3: Check if dimension exists
	fmt.Println("\n=== Example 3: Check Dimension Existence ===")
	exists, err := svc.Dimensions.Exists(ctx, "}Clients")
	if err != nil {
		log.Printf("Error checking dimension: %v", err)
	} else {
		fmt.Printf("}Clients exists: %t\n", exists)
	}

	// Example 4: Create a simple dimension
	fmt.Println("\n=== Example 4: Create Simple Dimension ===")

	// Create dimension with elements
	dim := models.NewDimension("TestDimension")

	// Create default hierarchy
	hierarchy := models.NewHierarchy("TestDimension", "TestDimension")

	// Add elements
	hierarchy.AddElement(models.Element{
		Name: "Element1",
		Type: models.ElementTypeNumeric,
	})
	hierarchy.AddElement(models.Element{
		Name: "Element2",
		Type: models.ElementTypeNumeric,
	})
	hierarchy.AddElement(models.Element{
		Name: "Element3",
		Type: models.ElementTypeNumeric,
	})
	hierarchy.AddElement(models.Element{
		Name: "Total",
		Type: models.ElementTypeConsolidated,
	})

	// Add edges (parent-child relationships)
	hierarchy.AddEdge("Total", "Element1", 1.0)
	hierarchy.AddEdge("Total", "Element2", 1.0)
	hierarchy.AddEdge("Total", "Element3", 1.0)

	dim.AddHierarchy(*hierarchy)

	// Create dimension in TM1
	err = svc.Dimensions.Create(ctx, dim)
	if err != nil {
		log.Printf("Error creating dimension: %v", err)
	} else {
		fmt.Println("✓ Dimension 'TestDimension' created successfully")
	}

	// Example 5: Get a dimension
	fmt.Println("\n=== Example 5: Get Dimension ===")
	retrievedDim, err := svc.Dimensions.Get(ctx, "TestDimension")
	if err != nil {
		log.Printf("Error getting dimension: %v", err)
	} else {
		fmt.Printf("Retrieved dimension: %s\n", retrievedDim.Name)
		fmt.Printf("Number of hierarchies: %d\n", len(retrievedDim.Hierarchies))
		if len(retrievedDim.Hierarchies) > 0 {
			fmt.Printf("Number of elements in default hierarchy: %d\n",
				len(retrievedDim.Hierarchies[0].Elements))
		}
	}

	// Example 6: Check for alternate hierarchies
	fmt.Println("\n=== Example 6: Check Alternate Hierarchies ===")
	hasAlternate, err := svc.Dimensions.UsesAlternateHierarchies(ctx, "TestDimension")
	if err != nil {
		log.Printf("Error checking alternate hierarchies: %v", err)
	} else {
		fmt.Printf("Uses alternate hierarchies: %t\n", hasAlternate)
	}

	// Example 7: Delete dimension
	fmt.Println("\n=== Example 7: Delete Dimension ===")
	err = svc.Dimensions.Delete(ctx, "TestDimension")
	if err != nil {
		log.Printf("Error deleting dimension: %v", err)
	} else {
		fmt.Println("✓ Dimension 'TestDimension' deleted successfully")
	}

	// Example 8: Create or update dimension
	fmt.Println("\n=== Example 8: Update or Create Dimension ===")
	err = svc.Dimensions.UpdateOrCreate(ctx, dim)
	if err != nil {
		log.Printf("Error in UpdateOrCreate: %v", err)
	} else {
		fmt.Println("✓ Dimension created/updated via UpdateOrCreate")
	}

	// Clean up
	svc.Dimensions.Delete(ctx, "TestDimension")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
