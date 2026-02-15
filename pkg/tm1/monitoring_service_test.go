package tm1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestMonitoringServiceGetCurrentUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/ActiveUser" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"Name":"admin","Type":"Admin","Groups":[{"Name":"Admin"}]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	service := NewMonitoringService(rest)
	user, err := service.GetCurrentUser(context.Background())
	if err != nil {
		t.Fatalf("GetCurrentUser() error = %v", err)
	}
	if user.Name != "admin" {
		t.Fatalf("GetCurrentUser().Name=%s, want admin", user.Name)
	}
}
