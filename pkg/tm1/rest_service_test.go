package tm1

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewRestService(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Address: "localhost",
				Port:    8882,
				SSL:     true,
			},
			wantErr: false,
		},
		{
			name: "with authentication",
			config: Config{
				Address:  "localhost",
				Port:     8882,
				User:     "admin",
				Password: "password",
				SSL:      true,
			},
			wantErr: false,
		},
		{
			name: "with session ID",
			config: Config{
				Address:   "localhost",
				Port:      8882,
				SessionID: "test-session-id",
				SSL:       true,
			},
			wantErr: false,
		},
		{
			name: "with CAM namespace",
			config: Config{
				Address:   "localhost",
				Port:      8882,
				Namespace: "LDAP",
				User:      "user",
				Password:  "password",
				SSL:       true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs, err := NewRestService(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRestService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && rs == nil {
				t.Error("NewRestService() returned nil without error")
			}
			if rs != nil {
				if rs.baseURL == nil {
					t.Error("baseURL is nil")
				}
				if rs.client == nil {
					t.Error("client is nil")
				}
				if rs.headers == nil {
					t.Error("headers is nil")
				}
			}
		})
	}
}

func TestRestServiceResolve(t *testing.T) {
	cfg := Config{
		Address: "localhost",
		Port:    8882,
		SSL:     true,
	}
	rs, err := NewRestService(cfg)
	if err != nil {
		t.Fatalf("NewRestService() failed: %v", err)
	}

	tests := []struct {
		name     string
		endpoint string
		want     string
	}{
		{
			name:     "simple path",
			endpoint: "/Cubes",
			want:     "https://localhost:8882/api/v1/Cubes",
		},
		{
			name:     "path with leading slash removed",
			endpoint: "Cubes",
			want:     "https://localhost:8882/api/v1/Cubes",
		},
		{
			name:     "metadata endpoint",
			endpoint: "/$metadata",
			want:     "https://localhost:8882/api/v1/$metadata",
		},
		{
			name:     "empty endpoint",
			endpoint: "",
			want:     "https://localhost:8882/api/v1/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := rs.resolve(tt.endpoint)
			if err != nil {
				t.Errorf("resolve() error = %v", err)
				return
			}
			if url.String() != tt.want {
				t.Errorf("resolve() = %v, want %v", url.String(), tt.want)
			}
		})
	}
}

func TestRestServiceSessionID(t *testing.T) {
	cfg := Config{
		Address: "localhost",
		Port:    8882,
		SSL:     true,
	}
	rs, err := NewRestService(cfg)
	if err != nil {
		t.Fatalf("NewRestService() failed: %v", err)
	}

	// Initially should be empty
	sessionID := rs.SessionID()
	if sessionID != "" {
		t.Errorf("SessionID() = %v, want empty string", sessionID)
	}

	// Simulate setting a cookie
	cookie := &http.Cookie{
		Name:  "TM1SessionId",
		Value: "test-session-123",
	}
	rs.client.Jar.SetCookies(rs.baseURL, []*http.Cookie{cookie})

	// Now should return the session ID
	sessionID = rs.SessionID()
	if sessionID != "test-session-123" {
		t.Errorf("SessionID() = %v, want %v", sessionID, "test-session-123")
	}
}

func TestRestServiceAddCompactJSONHeader(t *testing.T) {
	cfg := Config{
		Address: "localhost",
		Port:    8882,
		SSL:     true,
	}
	rs, err := NewRestService(cfg)
	if err != nil {
		t.Fatalf("NewRestService() failed: %v", err)
	}

	originalHeader := rs.AddCompactJSONHeader()
	if originalHeader == "" {
		t.Error("AddCompactJSONHeader() returned empty original header")
	}

	// Check that header was modified
	newHeader := rs.headers.Get("Accept")
	if !strings.Contains(newHeader, "tm1.compact=v0") {
		t.Errorf("Header does not contain compact format: %v", newHeader)
	}
}

func TestRestServiceWithLogger(t *testing.T) {
	cfg := Config{
		Address: "localhost",
		Port:    8882,
		SSL:     true,
	}

	customLogger := &testLogger{}
	rs, err := NewRestService(cfg, WithLogger(customLogger))
	if err != nil {
		t.Fatalf("NewRestService() failed: %v", err)
	}

	if rs.logger != customLogger {
		t.Error("Custom logger was not set")
	}
}

func TestRestServiceLogsPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	cfg := Config{
		Address: "localhost",
		Port:    8882,
		SSL:     false,
	}

	logger := &testLogger{}
	rs, err := NewRestService(cfg, WithLogger(logger))
	if err != nil {
		t.Fatalf("NewRestService() failed: %v", err)
	}

	testURL := server.URL + "/api/v1"
	rs.baseURL, _ = rs.baseURL.Parse(testURL)

	ctx := context.Background()
	resp, err := rs.Post(ctx, "/test", strings.NewReader(`{"a":1}`))
	if err != nil {
		t.Fatalf("Post() failed: %v", err)
	}
	defer resp.Body.Close()

	if len(logger.messages) == 0 {
		t.Fatal("expected at least one log message")
	}

	last := logger.messages[len(logger.messages)-1]
	if !strings.Contains(last, `payload={"a":1}`) {
		t.Fatalf("expected payload in log message, got: %s", last)
	}
}

func TestRestServiceWithAuthProvider(t *testing.T) {
	cfg := Config{
		Address: "localhost",
		Port:    8882,
		SSL:     true,
	}

	auth := BasicAuth{Username: "testuser", Password: "testpass"}
	rs, err := NewRestService(cfg, WithAuthProvider(auth))
	if err != nil {
		t.Fatalf("NewRestService() failed: %v", err)
	}

	if rs.auth == nil {
		t.Error("Auth provider was not set")
	}
}

func TestRestServiceHTTPMethods(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	// Create RestService pointing to test server
	cfg := Config{
		Address: "localhost",
		Port:    8882,
		SSL:     false,
	}
	rs, err := NewRestService(cfg)
	if err != nil {
		t.Fatalf("NewRestService() failed: %v", err)
	}

	// Override baseURL to point to test server
	testURL := server.URL + "/api/v1"
	rs.baseURL, _ = rs.baseURL.Parse(testURL)

	ctx := context.Background()

	tests := []struct {
		name   string
		method func() (*http.Response, error)
	}{
		{
			name: "GET",
			method: func() (*http.Response, error) {
				return rs.Get(ctx, "/test")
			},
		},
		{
			name: "POST",
			method: func() (*http.Response, error) {
				return rs.Post(ctx, "/test", strings.NewReader("{}"))
			},
		},
		{
			name: "PUT",
			method: func() (*http.Response, error) {
				return rs.Put(ctx, "/test", strings.NewReader("{}"))
			},
		},
		{
			name: "PATCH",
			method: func() (*http.Response, error) {
				return rs.Patch(ctx, "/test", strings.NewReader("{}"))
			},
		},
		{
			name: "DELETE",
			method: func() (*http.Response, error) {
				return rs.Delete(ctx, "/test")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.method()
			if err != nil {
				t.Errorf("%s failed: %v", tt.name, err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}
		})
	}
}

// Helper type for testing
type testLogger struct {
	messages []string
}

func (l *testLogger) Printf(format string, args ...any) {
	l.messages = append(l.messages, fmt.Sprintf(format, args...))
}
