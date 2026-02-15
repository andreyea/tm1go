package tm1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestChoreServiceExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/Chores('ExistingChore')":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"Name":"ExistingChore"}`))
		case "/Chores('MissingChore')":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	service := NewChoreService(rest)
	ctx := context.Background()

	exists, err := service.Exists(ctx, "ExistingChore")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Fatal("Exists() = false, want true")
	}

	exists, err = service.Exists(ctx, "MissingChore")
	if err != nil {
		t.Fatalf("Exists() missing error = %v", err)
	}
	if exists {
		t.Fatal("Exists() = true, want false")
	}
}

func TestChoreServiceGetAllNames(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Chores" && strings.Contains(r.URL.RawQuery, "$select=Name") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"value":[{"Name":"Daily"},{"Name":"Hourly"}]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	service := NewChoreService(rest)
	names, err := service.GetAllNames(context.Background())
	if err != nil {
		t.Fatalf("GetAllNames() error = %v", err)
	}
	if len(names) != 2 || names[0] != "Daily" || names[1] != "Hourly" {
		t.Fatalf("GetAllNames() = %#v, want [Daily Hourly]", names)
	}
}

func TestChoreServiceSetLocalStartTimeReactivates(t *testing.T) {
	calls := make([]string, 0, 4)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		switch {
		case r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/Chores('Nightly')"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"Name":"Nightly","Active":true,"Tasks":[]}`))
		case r.Method == "POST" && r.URL.Path == "/Chores('Nightly')/tm1.Deactivate":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == "POST" && r.URL.Path == "/Chores('Nightly')/tm1.SetServerLocalStartTime":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == "POST" && r.URL.Path == "/Chores('Nightly')/tm1.Activate":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	service := NewChoreService(rest)
	err := service.SetLocalStartTime(context.Background(), "Nightly", time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("SetLocalStartTime() error = %v", err)
	}

	if len(calls) != 4 {
		t.Fatalf("unexpected call count: %d, calls=%v", len(calls), calls)
	}
	if calls[1] != "POST /Chores('Nightly')/tm1.Deactivate" ||
		calls[2] != "POST /Chores('Nightly')/tm1.SetServerLocalStartTime" ||
		calls[3] != "POST /Chores('Nightly')/tm1.Activate" {
		t.Fatalf("unexpected call order: %v", calls)
	}
}
