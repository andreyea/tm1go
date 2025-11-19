package tm1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRestServiceAuthModeDetection tests authentication mode detection logic
func TestRestServiceAuthModeDetection(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		expected AuthenticationMode
	}{
		{
			name: "IBM Cloud API Key",
			cfg: Config{
				Address: "us-east-2.planninganalytics.cloud.ibm.com",
				Tenant:  "YC4B2M1AG2Y6",
				APIKey:  "test-api-key",
			},
			expected: AuthModeIBMCloudAPIKey,
		},
		{
			name: "Service-to-Service",
			cfg: Config{
				ApplicationClientID:     "my-app",
				ApplicationClientSecret: "my-secret",
			},
			expected: AuthModeServiceToService,
		},
		{
			name: "Access Token",
			cfg: Config{
				AccessToken: "bearer-token-123",
			},
			expected: AuthModeAccessToken,
		},
		{
			name: "CAM Passport",
			cfg: Config{
				CAMPassport: "cam-passport-value",
			},
			expected: AuthModeCAM,
		},
		{
			name: "CAM Namespace",
			cfg: Config{
				Namespace: "LDAP",
				User:      "john.doe",
				Password:  "password",
			},
			expected: AuthModeCAM,
		},
		{
			name: "Windows Integrated Authentication",
			cfg: Config{
				IntegratedLogin: true,
			},
			expected: AuthModeWIA,
		},
		{
			name: "PA Proxy",
			cfg: Config{
				CPDUrl: "https://cpd-zen.apps.company.com",
			},
			expected: AuthModePAProxy,
		},
		{
			name: "Basic Authentication",
			cfg: Config{
				User:     "admin",
				Password: "apple",
			},
			expected: AuthModeBasic,
		},
		{
			name: "Session ID Reuse",
			cfg: Config{
				SessionID: "q7O6e1w49AixeuLVxJ1GZg",
			},
			expected: AuthModeBasic, // Session reuse treated as basic
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &RestService{}
			mode := rs.determineAuthMode(tt.cfg)
			if mode != tt.expected {
				t.Errorf("determineAuthMode() = %v, want %v", mode, tt.expected)
			}
		})
	}
}

// TestRestServiceIBMCloudURLConstruction tests IBM Cloud URL construction
func TestRestServiceIBMCloudURLConstruction(t *testing.T) {
	rs := &RestService{}
	cfg := Config{
		Address:  "us-east-2.planninganalytics.cloud.ibm.com",
		Tenant:   "YC4B2M1AG2Y6",
		Database: "Planning Sample",
	}

	baseURL, authURL, err := rs.constructIBMCloudServiceAndAuthRoot(cfg)
	if err != nil {
		t.Fatalf("constructIBMCloudServiceAndAuthRoot() error = %v", err)
	}

	expectedBase := "https://us-east-2.planninganalytics.cloud.ibm.com/api/YC4B2M1AG2Y6/v0/tm1/Planning Sample"
	expectedAuth := expectedBase + "/Configuration/ProductVersion/$value"

	if baseURL != expectedBase {
		t.Errorf("Base URL = %v, want %v", baseURL, expectedBase)
	}

	if authURL != expectedAuth {
		t.Errorf("Auth URL = %v, want %v", authURL, expectedAuth)
	}
}

// TestRestServiceS2SURLConstruction tests Service-to-Service URL construction
func TestRestServiceS2SURLConstruction(t *testing.T) {
	rs := &RestService{}
	cfg := Config{
		Address:  "localhost",
		Port:     8001,
		Instance: "tm1s1",
		Database: "Planning Sample",
		SSL:      true,
	}

	baseURL, authURL, err := rs.constructS2SServiceAndAuthRoot(cfg)
	if err != nil {
		t.Fatalf("constructS2SServiceAndAuthRoot() error = %v", err)
	}

	expectedBase := "https://localhost:8001/tm1s1/api/v1/Databases('Planning Sample')"
	expectedAuth := "https://localhost:8001/tm1s1/auth/v1/session"

	if baseURL != expectedBase {
		t.Errorf("Base URL = %v, want %v", baseURL, expectedBase)
	}

	if authURL != expectedAuth {
		t.Errorf("Auth URL = %v, want %v", authURL, expectedAuth)
	}
}

// TestRestServicePAProxyURLConstruction tests PA Proxy URL construction
func TestRestServicePAProxyURLConstruction(t *testing.T) {
	rs := &RestService{}
	cfg := Config{
		Address:  "pa-workspace.company.com",
		Database: "Planning Sample",
		SSL:      true,
	}

	baseURL, authURL, err := rs.constructPAProxyServiceAndAuthRoot(cfg)
	if err != nil {
		t.Fatalf("constructPAProxyServiceAndAuthRoot() error = %v", err)
	}

	expectedBase := "https://pa-workspace.company.com/tm1/Planning Sample/api/v1"
	expectedAuth := "https://pa-workspace.company.com/login"

	if baseURL != expectedBase {
		t.Errorf("Base URL = %v, want %v", baseURL, expectedBase)
	}

	if authURL != expectedAuth {
		t.Errorf("Auth URL = %v, want %v", authURL, expectedAuth)
	}
}

// TestRestServiceV11URLConstruction tests traditional v11 URL construction
func TestRestServiceV11URLConstruction(t *testing.T) {
	rs := &RestService{}
	cfg := Config{
		Address: "localhost",
		Port:    12354,
		SSL:     true,
	}

	baseURL, authURL, err := rs.constructV11ServiceAndAuthRoot(cfg)
	if err != nil {
		t.Fatalf("constructV11ServiceAndAuthRoot() error = %v", err)
	}

	expectedBase := "https://localhost:12354/api/v1"
	expectedAuth := expectedBase + "/Configuration/ProductVersion/$value"

	if baseURL != expectedBase {
		t.Errorf("Base URL = %v, want %v", baseURL, expectedBase)
	}

	if authURL != expectedAuth {
		t.Errorf("Auth URL = %v, want %v", authURL, expectedAuth)
	}
}

// TestRestServiceIBMIAMTokenGeneration tests IBM Cloud IAM token generation
func TestRestServiceIBMIAMTokenGeneration(t *testing.T) {
	// Create a mock IAM server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/identity/token" {
			t.Errorf("Expected path /identity/token, got %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/x-www-form-urlencoded" {
			t.Errorf("Expected Content-Type application/x-www-form-urlencoded, got %s", contentType)
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			t.Fatalf("Failed to parse form: %v", err)
		}

		grantType := r.FormValue("grant_type")
		if grantType != "urn:ibm:params:oauth:grant-type:apikey" {
			t.Errorf("Expected grant_type urn:ibm:params:oauth:grant-type:apikey, got %s", grantType)
		}

		apiKey := r.FormValue("apikey")
		if apiKey != "test-api-key" {
			t.Errorf("Expected apikey test-api-key, got %s", apiKey)
		}

		// Return mock token response
		resp := map[string]interface{}{
			"access_token": "mock-iam-access-token-12345",
			"token_type":   "Bearer",
			"expires_in":   3600,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	rs := &RestService{}
	cfg := Config{
		APIKey: "test-api-key",
		IAMUrl: server.URL, // Use mock server
	}

	token, err := rs.generateIBMIAMCloudAccessToken(cfg)
	if err != nil {
		t.Fatalf("generateIBMIAMCloudAccessToken() error = %v", err)
	}

	expectedToken := "mock-iam-access-token-12345"
	if token != expectedToken {
		t.Errorf("Token = %v, want %v", token, expectedToken)
	}
}

// TestRestServiceAuthenticationModes tests all authentication modes
func TestRestServiceAuthenticationModes(t *testing.T) {
	tests := []struct {
		name           string
		cfg            Config
		expectedHeader string
		expectedAuth   bool
	}{
		{
			name: "Basic Authentication",
			cfg: Config{
				User:     "admin",
				Password: "apple",
			},
			expectedHeader: "Authorization",
			expectedAuth:   true,
		},
		{
			name: "CAM Namespace",
			cfg: Config{
				User:      "john.doe",
				Password:  "password",
				Namespace: "LDAP",
			},
			expectedHeader: "Authorization",
			expectedAuth:   true,
		},
		{
			name: "CAM Passport",
			cfg: Config{
				CAMPassport: "test-passport",
			},
			expectedHeader: "Authorization",
			expectedAuth:   true,
		},
		{
			name: "Access Token",
			cfg: Config{
				AccessToken: "bearer-token-123",
			},
			expectedHeader: "Authorization",
			expectedAuth:   true,
		},
		{
			name: "Session ID Reuse",
			cfg: Config{
				SessionID: "q7O6e1w49AixeuLVxJ1GZg",
			},
			expectedHeader: "Cookie",
			expectedAuth:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal RestService for testing
			rs := &RestService{
				headers: http.Header{},
			}

			err := rs.setupAuthentication(tt.cfg)
			if err != nil && tt.expectedAuth {
				t.Fatalf("setupAuthentication() error = %v", err)
			}

			if tt.expectedAuth && rs.auth == nil {
				t.Error("Expected auth to be set, but it's nil")
			}
		})
	}
}

// TestRestServiceBase64Password tests base64 password decoding
func TestRestServiceBase64Password(t *testing.T) {
	rs := &RestService{
		headers: http.Header{},
	}

	cfg := Config{
		User:         "admin",
		Password:     "YXBwbGU=", // "apple" in base64
		DecodeBase64: true,
	}

	err := rs.setupAuthentication(cfg)
	if err != nil {
		t.Fatalf("setupAuthentication() error = %v", err)
	}

	if rs.auth == nil {
		t.Fatal("Expected auth to be set")
	}

	// Verify that the password was decoded correctly
	basicAuth, ok := rs.auth.(BasicAuth)
	if !ok {
		t.Fatal("Expected BasicAuth type")
	}

	if basicAuth.Password != "apple" {
		t.Errorf("Password = %v, want %v", basicAuth.Password, "apple")
	}
}

// TestRestServiceServiceToServiceAuth tests service-to-service authentication
func TestRestServiceServiceToServiceAuth(t *testing.T) {
	rs := &RestService{
		headers: http.Header{},
	}

	cfg := Config{
		ApplicationClientID:     "my-application",
		ApplicationClientSecret: "my-secret",
	}

	// Set auth mode
	rs.authMode = AuthModeServiceToService

	err := rs.setupAuthentication(cfg)
	if err != nil {
		t.Fatalf("setupAuthentication() error = %v", err)
	}

	if rs.auth == nil {
		t.Fatal("Expected auth to be set")
	}

	basicAuth, ok := rs.auth.(BasicAuth)
	if !ok {
		t.Fatal("Expected BasicAuth type for S2S")
	}

	if basicAuth.Username != "my-application" {
		t.Errorf("Username = %v, want my-application", basicAuth.Username)
	}

	if basicAuth.Password != "my-secret" {
		t.Errorf("Password = %v, want my-secret", basicAuth.Password)
	}
}
