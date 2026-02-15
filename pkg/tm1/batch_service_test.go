package tm1

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/andreyea/tm1go/pkg/models"
)

func TestBatchServiceBatchV11PrefixesAPIv1(t *testing.T) {
	var captured struct {
		Requests []models.BatchRequest `json:"requests"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/$batch" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"responses":[{"id":"1","status":200}]}`))
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())
	rest.version = "11.8.0"

	service := NewBatchService(rest)
	ctx := context.Background()

	_, err := service.Batch(ctx, []models.BatchRequest{{Method: "GET", URL: "/Cubes", ID: "1"}})
	if err != nil {
		t.Fatalf("Batch() error = %v", err)
	}

	if len(captured.Requests) != 1 {
		t.Fatalf("captured requests = %d, want 1", len(captured.Requests))
	}
	if captured.Requests[0].URL != "/api/v1/Cubes" {
		t.Errorf("request URL = %s, want /api/v1/Cubes", captured.Requests[0].URL)
	}
}

func TestBatchServiceBatchV12DoesNotPrefixAPIv1(t *testing.T) {
	var captured struct {
		Requests []models.BatchRequest `json:"requests"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/$batch" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"responses":[{"id":"1","status":200}]}`))
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())
	rest.version = "12.0.1"

	service := NewBatchService(rest)
	ctx := context.Background()

	_, err := service.Batch(ctx, []models.BatchRequest{{Method: "GET", URL: "/Cubes", ID: "1"}})
	if err != nil {
		t.Fatalf("Batch() error = %v", err)
	}

	if len(captured.Requests) != 1 {
		t.Fatalf("captured requests = %d, want 1", len(captured.Requests))
	}
	if captured.Requests[0].URL != "/Cubes" {
		t.Errorf("request URL = %s, want /Cubes", captured.Requests[0].URL)
	}
}
