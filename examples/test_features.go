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

	// Test 4: WhoAmI
	fmt.Println("\n=== Test 4: WhoAmI ===")
	user, err := svc.WhoAmI(ctx)
	if err != nil {
		log.Printf("WhoAmI error: %v", err)
	} else {
		fmt.Printf("User Info: %+v\n", user)
	}

	// Test 5: IsAdmin
	fmt.Println("\n=== Test 5: IsAdmin ===")
	isAdmin, err := svc.IsAdmin(ctx)
	if err != nil {
		log.Printf("IsAdmin error: %v", err)
	} else {
		fmt.Printf("Is Admin: %t\n", isAdmin)
	}
}
