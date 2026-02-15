package tm1

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTM1Service_ServicesInitialized(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/Configuration/ProductVersion/$value":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("12.0.0"))
		case "/api/v1/ActiveSession/tm1.Close":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := Config{
		BaseURL:             server.URL + "/api/v1",
		SkipSSLVerification: true,
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

	if service.Batches == nil {
		t.Error("Batches service is nil")
	}

	if service.Sandboxes == nil {
		t.Error("Sandboxes service is nil")
	}

	if service.Hierarchies == nil {
		t.Error("Hierarchies service is nil")
	}

	if service.Elements == nil {
		t.Error("Elements service is nil")
	}

	if service.Users == nil {
		t.Error("Users service is nil")
	}

	if service.Security == nil {
		t.Error("Security service is nil")
	}

	if service.Jobs == nil {
		t.Error("Jobs service is nil")
	}

	if service.Threads == nil {
		t.Error("Threads service is nil")
	}

	if service.Sessions == nil {
		t.Error("Sessions service is nil")
	}

	if service.Monitoring == nil {
		t.Error("Monitoring service is nil")
	}

	if service.Server == nil {
		t.Error("Server service is nil")
	}

	if service.Configuration == nil {
		t.Error("Configuration service is nil")
	}

	// Test that Rest() returns the underlying rest service
	if service.Rest() == nil {
		t.Error("Rest() returned nil")
	}
}
