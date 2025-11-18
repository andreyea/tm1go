package tm1

import (
	"net/http"
	"net/url"
	"testing"
)

func TestBasicAuth(t *testing.T) {
	auth := BasicAuth{Username: "testuser", Password: "testpass"}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	err := auth.Apply(req)
	if err != nil {
		t.Errorf("BasicAuth.Apply() error = %v", err)
	}

	username, password, ok := req.BasicAuth()
	if !ok {
		t.Error("BasicAuth credentials not set")
	}
	if username != "testuser" {
		t.Errorf("username = %v, want %v", username, "testuser")
	}
	if password != "testpass" {
		t.Errorf("password = %v, want %v", password, "testpass")
	}
}

func TestBearerToken(t *testing.T) {
	auth := BearerToken("test-token-123")

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	err := auth.Apply(req)
	if err != nil {
		t.Errorf("BearerToken.Apply() error = %v", err)
	}

	authHeader := req.Header.Get("Authorization")
	expected := "Bearer test-token-123"
	if authHeader != expected {
		t.Errorf("Authorization header = %v, want %v", authHeader, expected)
	}
}

func TestSessionCookieAuth(t *testing.T) {
	auth := SessionCookieAuth{Name: "TM1SessionId", Value: "test-session-id"}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	err := auth.Apply(req)
	if err != nil {
		t.Errorf("SessionCookieAuth.Apply() error = %v", err)
	}

	cookies := req.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Name != "TM1SessionId" {
		t.Errorf("Cookie name = %v, want %v", cookies[0].Name, "TM1SessionId")
	}
	if cookies[0].Value != "test-session-id" {
		t.Errorf("Cookie value = %v, want %v", cookies[0].Value, "test-session-id")
	}
}

func TestHeaderAuth(t *testing.T) {
	auth := HeaderAuth{"X-Custom-Header": "custom-value"}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	err := auth.Apply(req)
	if err != nil {
		t.Errorf("HeaderAuth.Apply() error = %v", err)
	}

	headerValue := req.Header.Get("X-Custom-Header")
	if headerValue != "custom-value" {
		t.Errorf("Header value = %v, want %v", headerValue, "custom-value")
	}
}

func TestAuthFunc(t *testing.T) {
	called := false
	auth := AuthFunc(func(r *http.Request) error {
		called = true
		r.Header.Set("X-Test", "test-value")
		return nil
	})

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	err := auth.Apply(req)
	if err != nil {
		t.Errorf("AuthFunc.Apply() error = %v", err)
	}

	if !called {
		t.Error("AuthFunc was not called")
	}

	if req.Header.Get("X-Test") != "test-value" {
		t.Error("AuthFunc did not set header")
	}
}

func TestWithHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	opt := WithHeader("X-Custom", "value")
	opt(req)

	if req.Header.Get("X-Custom") != "value" {
		t.Error("WithHeader did not set header")
	}
}

func TestWithQueryValue(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	opt := WithQueryValue("key", "value")
	opt(req)

	if req.URL.Query().Get("key") != "value" {
		t.Error("WithQueryValue did not set query parameter")
	}
}

func TestWithQueryValues(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	values := url.Values{
		"key1": {"value1"},
		"key2": {"value2"},
	}
	opt := WithQueryValues(values)
	opt(req)

	query := req.URL.Query()
	if query.Get("key1") != "value1" {
		t.Error("WithQueryValues did not set key1")
	}
	if query.Get("key2") != "value2" {
		t.Error("WithQueryValues did not set key2")
	}
}

func TestWithLogger(t *testing.T) {
	cfg := Config{
		Address: "localhost",
		Port:    8882,
		SSL:     true,
	}

	logger := &testLogger{}
	rs, err := NewRestService(cfg, WithLogger(logger))
	if err != nil {
		t.Fatalf("NewRestService() failed: %v", err)
	}

	if rs.logger != logger {
		t.Error("WithLogger did not set custom logger")
	}
}

func TestWithAuthProvider(t *testing.T) {
	cfg := Config{
		Address: "localhost",
		Port:    8882,
		SSL:     true,
	}

	auth := BasicAuth{Username: "user", Password: "pass"}
	rs, err := NewRestService(cfg, WithAuthProvider(auth))
	if err != nil {
		t.Fatalf("NewRestService() failed: %v", err)
	}

	if rs.auth == nil {
		t.Error("WithAuthProvider did not set auth provider")
	}
}

func TestWithAdditionalHeaders(t *testing.T) {
	cfg := Config{
		Address: "localhost",
		Port:    8882,
		SSL:     true,
	}

	headers := http.Header{
		"X-Custom-1": {"value1"},
		"X-Custom-2": {"value2"},
	}
	rs, err := NewRestService(cfg, WithAdditionalHeaders(headers))
	if err != nil {
		t.Fatalf("NewRestService() failed: %v", err)
	}

	if rs.headers.Get("X-Custom-1") != "value1" {
		t.Error("WithAdditionalHeaders did not set X-Custom-1")
	}
	if rs.headers.Get("X-Custom-2") != "value2" {
		t.Error("WithAdditionalHeaders did not set X-Custom-2")
	}
}
