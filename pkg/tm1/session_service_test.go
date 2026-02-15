package tm1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestSessionServiceCloseAll(t *testing.T) {
	closeCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/ActiveUser":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"Name":"Admin User","Type":"Admin","Groups":[{"Name":"Admin"}]}`))
		case r.Method == "GET" && r.URL.Path == "/Sessions":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"value":[{"ID":"s1","User":{"Name":"adminuser"}},{"ID":"s2","User":{"Name":"Bob"}},{"ID":"s3"},{"ID":"s4","User":null}]}`))
		case r.Method == "POST" && strings.HasPrefix(r.URL.Path, "/Sessions('") && strings.HasSuffix(r.URL.Path, "')/tm1.Close"):
			closeCalls++
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

	service := NewSessionService(rest)
	closed, err := service.CloseAll(context.Background())
	if err != nil {
		t.Fatalf("CloseAll() error = %v", err)
	}
	if len(closed) != 1 || closeCalls != 1 {
		t.Fatalf("CloseAll() closed=%d calls=%d, want 1/1", len(closed), closeCalls)
	}
}

func TestSessionServiceGetThreadsForCurrent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/ActiveSession/Threads" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"value":[{"ID":1}]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	service := NewSessionService(rest)
	threads, err := service.GetThreadsForCurrent(context.Background(), true)
	if err != nil {
		t.Fatalf("GetThreadsForCurrent() error = %v", err)
	}
	if len(threads) != 1 {
		t.Fatalf("GetThreadsForCurrent() len=%d, want 1", len(threads))
	}
}
