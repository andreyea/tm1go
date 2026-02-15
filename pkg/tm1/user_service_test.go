package tm1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestUserServiceGetAll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Users" && strings.Contains(r.URL.RawQuery, "$expand=Groups") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"value":[{"Name":"admin","Type":"Admin","Groups":[{"Name":"Admin"}]},{"Name":"user1","Type":"User","Groups":[{"Name":"Sales"}]}]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	service := NewUserService(rest)
	users, err := service.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("GetAll() len = %d, want 2", len(users))
	}
	if users[0].Name != "admin" || users[1].Name != "user1" {
		t.Fatalf("GetAll() unexpected users: %#v", users)
	}
}

func TestUserServiceIsActive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/Users('ActiveUser')/IsActive":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"value":true}`))
		case "/Users('InactiveUser')/IsActive":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"value":false}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	service := NewUserService(rest)
	active, err := service.IsActive(context.Background(), "ActiveUser")
	if err != nil {
		t.Fatalf("IsActive(active) error = %v", err)
	}
	if !active {
		t.Fatal("IsActive(active) = false, want true")
	}

	active, err = service.IsActive(context.Background(), "InactiveUser")
	if err != nil {
		t.Fatalf("IsActive(inactive) error = %v", err)
	}
	if active {
		t.Fatal("IsActive(inactive) = true, want false")
	}
}

func TestUserServiceDisconnectAll(t *testing.T) {
	disconnectCalls := make([]string, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/ActiveUser":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"Name":"Current User","Type":"Admin","Groups":[{"Name":"Admin"}]}`))
		case r.Method == "GET" && r.URL.Path == "/Users" && strings.Contains(r.URL.RawQuery, "IsActive%20eq%20true"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"value":[{"Name":"CurrentUser","Type":"Admin","Groups":[{"Name":"Admin"}]},{"Name":"John Doe","Type":"User","Groups":[]},{"Name":"Jane","Type":"User","Groups":[]}]}`))
		case r.Method == "POST" && strings.HasPrefix(r.URL.Path, "/Users('") && strings.HasSuffix(r.URL.Path, "')/tm1.Disconnect"):
			disconnectCalls = append(disconnectCalls, r.URL.Path)
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

	service := NewUserService(rest)
	disconnected, err := service.DisconnectAll(context.Background())
	if err != nil {
		t.Fatalf("DisconnectAll() error = %v", err)
	}

	if len(disconnected) != 2 {
		t.Fatalf("DisconnectAll() disconnected len = %d, want 2", len(disconnected))
	}
	if disconnected[0] != "John Doe" || disconnected[1] != "Jane" {
		t.Fatalf("DisconnectAll() unexpected disconnected users: %#v", disconnected)
	}
	if len(disconnectCalls) != 2 {
		t.Fatalf("expected 2 disconnect calls, got %d", len(disconnectCalls))
	}
}

func TestUserServiceDisconnectAllRequiresAdmin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/ActiveUser" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"Name":"user","Type":"User","Groups":[]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	service := NewUserService(rest)
	_, err := service.DisconnectAll(context.Background())
	if err == nil {
		t.Fatal("DisconnectAll() expected error for non-admin user")
	}
}
