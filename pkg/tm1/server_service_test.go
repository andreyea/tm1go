package tm1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestServerServiceGetServerName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/Configuration/ServerName/$value" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("TM1Server"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	rest.version = "11.8.0"
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	svc := NewServerService(rest)
	name, err := svc.GetServerName(context.Background())
	if err != nil {
		t.Fatalf("GetServerName() error = %v", err)
	}
	if name != "TM1Server" {
		t.Fatalf("GetServerName() = %s, want TM1Server", name)
	}
}

func TestServerServiceUpdateMessageLoggerLevel(t *testing.T) {
	var patchBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/ActiveUser":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"Name":"admin","Type":"Admin","Groups":[{"Name":"Admin"}]}`))
		case r.Method == "PATCH" && r.URL.Path == "/Loggers('TM1.Server')":
			_ = json.NewDecoder(r.Body).Decode(&patchBody)
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	rest.version = "11.8.0"
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	svc := NewServerService(rest)
	if err := svc.UpdateMessageLoggerLevel(context.Background(), "TM1.Server", "INFO"); err != nil {
		t.Fatalf("UpdateMessageLoggerLevel() error = %v", err)
	}
	if level, ok := patchBody["Level"].(float64); !ok || int(level) != 3 {
		t.Fatalf("expected Level=3 in PATCH body, got %#v", patchBody)
	}
}

func TestServerServiceDeltaRequests(t *testing.T) {
	deltaPath := "TransactionLogEntries/!delta('abc')"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/TailTransactionLog()":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"value":[],"@odata.deltaLink":"http://localhost/api/v1/TransactionLogEntries/!delta('abc')"}`))
		case r.Method == "GET" && r.URL.Path == "/"+deltaPath:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"value":[{"ID":1}],"@odata.deltaLink":"http://localhost/api/v1/TransactionLogEntries/!delta('def')"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	rest.version = "11.8.0"
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	svc := NewServerService(rest)
	if err := svc.InitializeTransactionLogDeltaRequests(context.Background(), ""); err != nil {
		t.Fatalf("InitializeTransactionLogDeltaRequests() error = %v", err)
	}
	entries, err := svc.ExecuteTransactionLogDeltaRequest(context.Background())
	if err != nil {
		t.Fatalf("ExecuteTransactionLogDeltaRequest() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestServerServiceSaveDataVersionGuard(t *testing.T) {
	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	rest.version = "12.0.0"

	svc := NewServerService(rest)
	err := svc.SaveData(context.Background())
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "deprecated") {
		t.Fatalf("expected deprecated version error, got %v", err)
	}
}
