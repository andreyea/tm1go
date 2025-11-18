package tm1

import (
	"net/http"
	"testing"
)

func TestConfigHTTPClientOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		wantNil  bool
		checkSSL bool
	}{
		{
			name: "default client with SSL verification",
			config: Config{
				Verify: true,
			},
			wantNil:  false,
			checkSSL: true,
		},
		{
			name: "client with SSL verification disabled",
			config: Config{
				Verify: false,
			},
			wantNil:  false,
			checkSSL: false,
		},
		{
			name: "custom client provided",
			config: Config{
				HTTPClient: &http.Client{},
			},
			wantNil: false,
		},
		{
			name: "connection pool size set",
			config: Config{
				ConnectionPoolSize: 20,
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := tt.config.HTTPClientOrDefault()
			if err != nil {
				t.Fatalf("HTTPClientOrDefault() error = %v", err)
			}
			if client == nil {
				t.Fatal("HTTPClientOrDefault() returned nil")
			}

			// Check transport if custom client not provided
			if tt.config.HTTPClient == nil {
				// Check that cookie jar is set
				if client.Jar == nil {
					t.Error("Expected cookie jar to be set")
				}

				transport, ok := client.Transport.(*http.Transport)
				if !ok {
					t.Fatal("Expected *http.Transport")
				}

				if transport.TLSClientConfig == nil && tt.checkSSL {
					t.Error("Expected TLSClientConfig to be set")
				}
			}
		})
	}
}

func TestConfigDetermineVerify(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   bool
	}{
		{
			name:   "verify explicitly false",
			config: Config{Verify: false},
			want:   false,
		},
		{
			name:   "verify explicitly true",
			config: Config{Verify: true},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.determineVerify()
			if got != tt.want {
				t.Errorf("determineVerify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeBase64Password(t *testing.T) {
	tests := []struct {
		name     string
		encoded  string
		expected string
	}{
		{
			name:     "valid base64",
			encoded:  "cGFzc3dvcmQ=", // "password"
			expected: "password",
		},
		{
			name:     "invalid base64 returns original",
			encoded:  "not-base64!@#",
			expected: "not-base64!@#",
		},
		{
			name:     "empty string",
			encoded:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decodeBase64Password(tt.encoded)
			if got != tt.expected {
				t.Errorf("decodeBase64Password() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBase64Encode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "password",
			expected: "cGFzc3dvcmQ=",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "special characters",
			input:    "p@ssw0rd!",
			expected: "cEBzc3cwcmQh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := base64Encode(tt.input)
			if got != tt.expected {
				t.Errorf("base64Encode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConfigGetProxyFunc(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "no proxy",
			config: Config{
				ProxyURL: "",
			},
		},
		{
			name: "proxy set",
			config: Config{
				ProxyURL: "http://proxy:8080",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxyFunc := tt.config.getProxyFunc()
			if proxyFunc == nil {
				t.Error("getProxyFunc() returned nil")
			}
		})
	}
}
