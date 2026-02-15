package tm1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestThreadServiceCancelAllRunning(t *testing.T) {
	cancelCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/Threads":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"value":[{"ID":1,"State":"Idle","Type":"User","Name":"A","Function":"X"},{"ID":2,"State":"Busy","Type":"System","Name":"A","Function":"X"},{"ID":3,"State":"Busy","Type":"User","Name":"Pseudo","Function":"X"},{"ID":4,"State":"Busy","Type":"User","Name":"A","Function":"GET /Threads"},{"ID":5,"State":"Busy","Type":"User","Name":"A","Function":"Calc"}]}`))
		case r.Method == "POST" && strings.HasPrefix(r.URL.Path, "/Threads('") && strings.HasSuffix(r.URL.Path, "')/tm1.CancelOperation"):
			cancelCalls++
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	rest.version = "11.7.0"
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	service := NewThreadService(rest)
	canceled, err := service.CancelAllRunning(context.Background())
	if err != nil {
		t.Fatalf("CancelAllRunning() error = %v", err)
	}
	if len(canceled) != 1 || cancelCalls != 1 {
		t.Fatalf("CancelAllRunning() canceled=%d calls=%d, want 1/1", len(canceled), cancelCalls)
	}
}

func TestThreadServiceVersionGuard(t *testing.T) {
	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	rest.version = "12.0.0"

	service := NewThreadService(rest)
	_, err := service.GetAll(context.Background())
	if err == nil {
		t.Fatal("GetAll() expected version error on v12")
	}
}
