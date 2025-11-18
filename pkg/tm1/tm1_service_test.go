package tm1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestNewTM1Service(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewTM1Service(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTM1Service() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if svc == nil {
					t.Error("NewTM1Service() returned nil without error")
				}
				if svc.rest == nil {
					t.Error("TM1Service.rest is nil")
				}
			}
		})
	}
}

func TestTM1ServiceRest(t *testing.T) {
	cfg := Config{
		Address: "localhost",
		Port:    8882,
		SSL:     true,
	}
	svc, err := NewTM1Service(cfg)
	if err != nil {
		t.Fatalf("NewTM1Service() failed: %v", err)
	}

	rest := svc.Rest()
	if rest == nil {
		t.Error("Rest() returned nil")
	}
}

func TestTM1ServiceSessionID(t *testing.T) {
	cfg := Config{
		Address: "localhost",
		Port:    8882,
		SSL:     true,
	}
	svc, err := NewTM1Service(cfg)
	if err != nil {
		t.Fatalf("NewTM1Service() failed: %v", err)
	}

	// Initially should be empty
	sessionID := svc.SessionID()
	if sessionID != "" {
		t.Errorf("SessionID() = %v, want empty string", sessionID)
	}

	// Simulate setting a cookie
	cookie := &http.Cookie{
		Name:  "TM1SessionId",
		Value: "test-session-456",
	}
	svc.rest.client.Jar.SetCookies(svc.rest.baseURL, []*http.Cookie{cookie})

	// Now should return the session ID
	sessionID = svc.SessionID()
	if sessionID != "test-session-456" {
		t.Errorf("SessionID() = %v, want %v", sessionID, "test-session-456")
	}
}

func TestTM1ServiceWithMockServer(t *testing.T) {
	// Create a mock TM1 server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/Configuration/ProductVersion/$value":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("11.8.02500.3"))
		case "/$metadata":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0"?><edmx:Edmx Version="4.0"></edmx:Edmx>`))
		case "/ActiveUser":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"Name":"admin","Type":"Admin","IsAdmin":true,"IsDataAdmin":true}`))
		case "/Configuration/ServerName/$value":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("TM1Server"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create TM1Service pointing to mock server
	cfg := Config{
		Address: "localhost",
		Port:    8882,
		SSL:     false,
	}
	svc, err := NewTM1Service(cfg)
	if err != nil {
		t.Fatalf("NewTM1Service() failed: %v", err)
	}

	// Override baseURL to point to mock server
	mockBaseURL, _ := url.Parse(server.URL)
	svc.rest.baseURL = mockBaseURL

	ctx := context.Background()

	t.Run("Version", func(t *testing.T) {
		version, err := svc.Version(ctx)
		if err != nil {
			t.Errorf("Version() error = %v", err)
			return
		}
		if version != "11.8.02500.3" {
			t.Errorf("Version() = %v, want %v", version, "11.8.02500.3")
		}
	})

	t.Run("Metadata", func(t *testing.T) {
		metadata, err := svc.Metadata(ctx)
		if err != nil {
			t.Errorf("Metadata() error = %v", err)
			return
		}
		if len(metadata) == 0 {
			t.Error("Metadata() returned empty data")
		}
	})

	t.Run("Ping", func(t *testing.T) {
		err := svc.Ping(ctx)
		if err != nil {
			t.Errorf("Ping() error = %v", err)
		}
	})

	t.Run("IsConnected", func(t *testing.T) {
		connected := svc.IsConnected(ctx)
		if !connected {
			t.Error("IsConnected() = false, want true")
		}
	})

	t.Run("WhoAmI", func(t *testing.T) {
		user, err := svc.WhoAmI(ctx)
		if err != nil {
			t.Errorf("WhoAmI() error = %v", err)
			return
		}
		if user["Name"] != "admin" {
			t.Errorf("WhoAmI() Name = %v, want %v", user["Name"], "admin")
		}
	})

	t.Run("IsAdmin", func(t *testing.T) {
		isAdmin, err := svc.IsAdmin(ctx)
		if err != nil {
			t.Errorf("IsAdmin() error = %v", err)
			return
		}
		if !isAdmin {
			t.Error("IsAdmin() = false, want true")
		}
	})

	t.Run("IsDataAdmin", func(t *testing.T) {
		isDataAdmin, err := svc.IsDataAdmin(ctx)
		if err != nil {
			t.Errorf("IsDataAdmin() error = %v", err)
			return
		}
		if !isDataAdmin {
			t.Error("IsDataAdmin() = false, want true")
		}
	})
}

func TestTM1ServiceReconnect(t *testing.T) {
	cfg := Config{
		Address: "localhost",
		Port:    8882,
		SSL:     true,
	}
	svc, err := NewTM1Service(cfg)
	if err != nil {
		t.Fatalf("NewTM1Service() failed: %v", err)
	}

	// Get original rest service
	originalRest := svc.rest

	// Reconnect with new config
	newCfg := Config{
		Address: "localhost",
		Port:    8883,
		SSL:     true,
	}
	err = svc.Reconnect(newCfg)
	if err != nil {
		t.Errorf("Reconnect() error = %v", err)
		return
	}

	// Should have new rest service
	if svc.rest == originalRest {
		t.Error("Reconnect() did not create new RestService")
	}
}

func TestTM1ServiceClose(t *testing.T) {
	cfg := Config{
		Address: "localhost",
		Port:    8882,
		SSL:     true,
	}
	svc, err := NewTM1Service(cfg)
	if err != nil {
		t.Fatalf("NewTM1Service() failed: %v", err)
	}

	err = svc.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestTM1ServiceKeepAlive(t *testing.T) {
	cfg := Config{
		Address:   "localhost",
		Port:      8882,
		SSL:       true,
		KeepAlive: true,
	}
	svc, err := NewTM1Service(cfg)
	if err != nil {
		t.Fatalf("NewTM1Service() failed: %v", err)
	}

	if !svc.rest.keepAlive {
		t.Error("KeepAlive flag not set in RestService")
	}
}
