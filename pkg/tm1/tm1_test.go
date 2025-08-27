package tm1

import (
	"net/http"
	"testing"
	"time"
)

// MockClient implements the Client interface for testing
type MockClient struct {
	responses map[string]*Response
	errors    map[string]error
	connected bool
	config    *Config
	version   string
}

// NewMockClient creates a new MockClient for testing
func NewMockClient() *MockClient {
	return &MockClient{
		responses: make(map[string]*Response),
		errors:    make(map[string]error),
		connected: true,
		config:    DefaultConfig(),
		version:   "11.8.00100.12",
	}
}

// SetResponse sets a mock response for a given URL
func (m *MockClient) SetResponse(url string, response *Response) {
	m.responses[url] = response
}

// SetError sets a mock error for a given URL
func (m *MockClient) SetError(url string, err error) {
	m.errors[url] = err
}

// Client interface implementation
func (m *MockClient) GET(url string, opts *RequestOptions) (*Response, error) {
	return m.getResponse(url)
}

func (m *MockClient) POST(url string, data []byte, opts *RequestOptions) (*Response, error) {
	return m.getResponse(url)
}

func (m *MockClient) PATCH(url string, data []byte, opts *RequestOptions) (*Response, error) {
	return m.getResponse(url)
}

func (m *MockClient) PUT(url string, data []byte, opts *RequestOptions) (*Response, error) {
	return m.getResponse(url)
}

func (m *MockClient) DELETE(url string, opts *RequestOptions) (*Response, error) {
	return m.getResponse(url)
}

func (m *MockClient) Connect() error {
	m.connected = true
	return nil
}

func (m *MockClient) Disconnect() error {
	m.connected = false
	return nil
}

func (m *MockClient) IsConnected() bool {
	return m.connected
}

func (m *MockClient) Config() *Config {
	return m.config
}

func (m *MockClient) Version() string {
	return m.version
}

func (m *MockClient) getResponse(url string) (*Response, error) {
	if err, exists := m.errors[url]; exists {
		return nil, err
	}
	if resp, exists := m.responses[url]; exists {
		return resp, nil
	}
	// Default response
	return &Response{
		StatusCode: 200,
		Body:       []byte(`{"value": []}`),
		Headers:    make(map[string][]string),
	}, nil
}

// Test DefaultConfig
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if !config.SSL {
		t.Error("Expected SSL to be true by default")
	}

	if config.ConnectionPoolSize != 10 {
		t.Errorf("Expected ConnectionPoolSize to be 10, got %d", config.ConnectionPoolSize)
	}

	if !config.ReconnectOnSessionTimeout {
		t.Error("Expected ReconnectOnSessionTimeout to be true by default")
	}

	if config.SessionContext != "TM1go" {
		t.Errorf("Expected SessionContext to be 'TM1go', got '%s'", config.SessionContext)
	}
}

// Test TranslateToBool utility function
func TestTranslateToBool(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected bool
	}{
		{true, true},
		{false, false},
		{"true", true},
		{"false", false},
		{"TRUE", true},
		{"FALSE", false},
		{"1", true},
		{"0", false},
		{"yes", true},
		{"no", false},
		{1, true},
		{0, false},
		{42, true},
		{-1, true},
		{1.0, true},
		{0.0, false},
		{"invalid", false},
		{nil, false},
	}

	for _, test := range tests {
		result := TranslateToBool(test.input)
		if result != test.expected {
			t.Errorf("TranslateToBool(%v) = %t, expected %t", test.input, result, test.expected)
		}
	}
}

// Test CaseAndSpaceInsensitiveEquals utility function
func TestCaseAndSpaceInsensitiveEquals(t *testing.T) {
	tests := []struct {
		s1, s2   string
		expected bool
	}{
		{"test", "test", true},
		{"TEST", "test", true},
		{"Test", "TEST", true},
		{"test test", "testtest", true},
		{"Test Test", "TESTTEST", true},
		{" test ", "test", true},
		{"different", "other", false},
		{"", "", true},
	}

	for _, test := range tests {
		result := CaseAndSpaceInsensitiveEquals(test.s1, test.s2)
		if result != test.expected {
			t.Errorf("CaseAndSpaceInsensitiveEquals(%q, %q) = %t, expected %t",
				test.s1, test.s2, result, test.expected)
		}
	}
}

// Test IsAdmin utility function
func TestIsAdmin(t *testing.T) {
	tests := []struct {
		username string
		expected bool
	}{
		{"admin", true},
		{"ADMIN", true},
		{"Admin", true},
		{"ad min", true},
		{"AD MIN", true},
		{"user", false},
		{"administrator", false},
		{"", false},
	}

	for _, test := range tests {
		result := IsAdmin(test.username)
		if result != test.expected {
			t.Errorf("IsAdmin(%q) = %t, expected %t", test.username, result, test.expected)
		}
	}
}

// Test ConstructURL utility function
func TestConstructURL(t *testing.T) {
	tests := []struct {
		base     string
		parts    []string
		expected string
	}{
		{"http://localhost", []string{"api", "v1"}, "http://localhost/api/v1"},
		{"http://localhost/", []string{"api", "v1"}, "http://localhost/api/v1"},
		{"http://localhost", []string{"/api/", "/v1/"}, "http://localhost/api/v1"},
		{"http://localhost/api", []string{"v1", "test"}, "http://localhost/api/v1/test"},
		{"http://localhost", []string{}, "http://localhost/"},
	}

	for _, test := range tests {
		result := ConstructURL(test.base, test.parts...)
		if result != test.expected {
			t.Errorf("ConstructURL(%q, %v) = %q, expected %q",
				test.base, test.parts, result, test.expected)
		}
	}
}

// Test ParseTimeout utility function
func TestParseTimeout(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected *time.Duration
		hasError bool
	}{
		{nil, nil, false},
		{time.Second * 30, durationPtr(time.Second * 30), false},
		{30.0, durationPtr(time.Second * 30), false},
		{30, durationPtr(time.Second * 30), false},
		{"30", durationPtr(time.Second * 30), false},
		{"invalid", nil, true},
		{[]string{"invalid"}, nil, true},
	}

	for _, test := range tests {
		result, err := ParseTimeout(test.input)

		if test.hasError {
			if err == nil {
				t.Errorf("ParseTimeout(%v) expected error but got none", test.input)
			}
			continue
		}

		if err != nil {
			t.Errorf("ParseTimeout(%v) returned unexpected error: %v", test.input, err)
			continue
		}

		if test.expected == nil && result != nil {
			t.Errorf("ParseTimeout(%v) = %v, expected nil", test.input, result)
		} else if test.expected != nil && result != nil && *result != *test.expected {
			t.Errorf("ParseTimeout(%v) = %v, expected %v", test.input, *result, *test.expected)
		}
	}
}

// Test ElementService with mock client
func TestElementService(t *testing.T) {
	mockClient := NewMockClient()
	elementService := NewElementServiceImpl(mockClient)

	// Test GetElementNames
	mockClient.SetResponse("/Dimensions('Product')/Hierarchies('Product')/Elements?$select=Name,Type",
		&Response{
			StatusCode: 200,
			Body: []byte(`{
				"value": [
					{"Name": "Product1", "Type": 2},
					{"Name": "Product2", "Type": 2}
				]
			}`),
		})

	names, err := elementService.GetElementNames("Product", "Product")
	if err != nil {
		t.Fatalf("GetElementNames failed: %v", err)
	}

	if len(names) != 2 {
		t.Errorf("Expected 2 element names, got %d", len(names))
	}

	if names[0] != "Product1" || names[1] != "Product2" {
		t.Errorf("Unexpected element names: %v", names)
	}
}

// Test error handling
func TestTM1RestException(t *testing.T) {
	err := NewTM1RestException("GET", "/test", "ERROR_CODE", "Test error message", http.StatusBadRequest)

	expectedMsg := "TM1 REST Exception [GET /test]: TM1 Error [400]: Test error message (Code: ERROR_CODE)"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}

	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, err.StatusCode)
	}
}

// Helper function for tests
func durationPtr(d time.Duration) *time.Duration {
	return &d
}
