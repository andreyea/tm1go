package tm1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestJobServiceGetAllAndCancelAll(t *testing.T) {
	cancelCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/Jobs":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"value":[{"ID":"1","Type":"Process"},{"ID":"2","Type":"Process"}]}`))
		case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/tm1.Cancel"):
			cancelCalls++
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	rest.version = "12.0.0"
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	service := NewJobService(rest)
	jobs, err := service.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(jobs) != 2 {
		t.Fatalf("GetAll() len = %d, want 2", len(jobs))
	}

	canceled, err := service.CancelAll(context.Background())
	if err != nil {
		t.Fatalf("CancelAll() error = %v", err)
	}
	if len(canceled) != 2 || cancelCalls != 2 {
		t.Fatalf("CancelAll() canceled=%d calls=%d, want 2/2", len(canceled), cancelCalls)
	}
}

func TestJobServiceVersionGuard(t *testing.T) {
	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	rest.version = "11.8.0"

	service := NewJobService(rest)
	_, err := service.GetAll(context.Background())
	if err == nil {
		t.Fatal("GetAll() expected version error")
	}
}
