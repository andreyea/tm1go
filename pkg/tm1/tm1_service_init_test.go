package tm1

import (
	"testing"
)

func TestTM1Service_ServicesInitialized(t *testing.T) {
	cfg := Config{
		Address:  "localhost",
		Port:     8882,
		SSL:      false,
		User:     "admin",
		Password: "apple",
	}

	service, err := NewTM1Service(cfg)
	if err != nil {
		t.Fatalf("NewTM1Service failed: %v", err)
	}
	defer service.Close()

	// Test that all services are initialized
	if service.Dimensions == nil {
		t.Error("Dimensions service is nil")
	}

	if service.Processes == nil {
		t.Error("Processes service is nil")
	}

	if service.Hierarchies == nil {
		t.Error("Hierarchies service is nil")
	}

	if service.Elements == nil {
		t.Error("Elements service is nil")
	}

	// Test that Rest() returns the underlying rest service
	if service.Rest() == nil {
		t.Error("Rest() returned nil")
	}
}
