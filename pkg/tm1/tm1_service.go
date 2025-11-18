package tm1

import (
	"context"
	"io"
	"net/http"
	"strings"
)

// TM1Service exposes higher level helpers built on top of RestService.
type TM1Service struct {
	rest *RestService
}

// NewTM1Service constructs a TM1Service with the supplied configuration.
func NewTM1Service(cfg Config, opts ...RestOption) (*TM1Service, error) {
	rest, err := NewRestService(cfg, opts...)
	if err != nil {
		return nil, err
	}

	return &TM1Service{rest: rest}, nil
}

// Rest returns the underlying REST client so additional services can be composed.
func (s *TM1Service) Rest() *RestService {
	return s.rest
}

// Close releases resources held by the service and terminates the TM1 session.
func (s *TM1Service) Close() error {
	ctx := context.Background()
	// First, logout from TM1 to properly close the session
	if err := s.rest.Logout(ctx); err != nil {
		// Log the error but continue with cleanup
		// Don't fail Close() just because logout failed
	}
	// Then close idle connections
	s.rest.Close()
	return nil
}

// Ping verifies that the TM1 instance is reachable.
func (s *TM1Service) Ping(ctx context.Context) error {
	return s.rest.Ping(ctx)
}

// Logout terminates the TM1 session explicitly.
// This is called automatically by Close(), but can be called separately if needed.
func (s *TM1Service) Logout(ctx context.Context) error {
	return s.rest.Logout(ctx)
}

// SessionID retrieves the current TM1SessionId from the active session.
// Returns empty string if no session is active or session cookie is not found.
func (s *TM1Service) SessionID() string {
	return s.rest.SessionID()
}

// WhoAmI returns the current authenticated user.
func (s *TM1Service) WhoAmI(ctx context.Context) (map[string]interface{}, error) {
	var user map[string]interface{}
	err := s.rest.JSON(ctx, "GET", "/ActiveUser", nil, &user)
	return user, err
}

// IsAdmin checks if the current user has admin privileges.
func (s *TM1Service) IsAdmin(ctx context.Context) (bool, error) {
	user, err := s.WhoAmI(ctx)
	if err != nil {
		return false, err
	}

	// Check for IsAdmin key in response
	if isAdmin := user["Type"].(string) == "Admin"; isAdmin {
		return true, nil
	}
	return false, nil
}

// IsDataAdmin checks if the current user has data admin privileges.
func (s *TM1Service) IsDataAdmin(ctx context.Context) (bool, error) {
	user, err := s.WhoAmI(ctx)
	if err != nil {
		return false, err
	}

	if isDataAdmin, ok := user["IsDataAdmin"].(bool); ok {
		return isDataAdmin, nil
	}
	return false, nil
}

// IsSecurityAdmin checks if the current user has security admin privileges.
func (s *TM1Service) IsSecurityAdmin(ctx context.Context) (bool, error) {
	user, err := s.WhoAmI(ctx)
	if err != nil {
		return false, err
	}

	if isSecurityAdmin, ok := user["IsSecurityAdmin"].(bool); ok {
		return isSecurityAdmin, nil
	}
	return false, nil
}

// IsOpsAdmin checks if the current user has operations admin privileges.
func (s *TM1Service) IsOpsAdmin(ctx context.Context) (bool, error) {
	user, err := s.WhoAmI(ctx)
	if err != nil {
		return false, err
	}

	if isOpsAdmin, ok := user["IsOpsAdmin"].(bool); ok {
		return isOpsAdmin, nil
	}
	return false, nil
}

// SandboxingDisabled checks if sandboxing is disabled for the current user.
func (s *TM1Service) SandboxingDisabled(ctx context.Context) (bool, error) {
	user, err := s.WhoAmI(ctx)
	if err != nil {
		return false, err
	}

	if sandboxingDisabled, ok := user["SandboxingDisabled"].(bool); ok {
		return sandboxingDisabled, nil
	}
	return false, nil
}

// IsConnected checks if the connection to TM1 server is active.
func (s *TM1Service) IsConnected(ctx context.Context) bool {
	err := s.Ping(ctx)
	return err == nil
}

// Reconnect re-establishes the connection to TM1.
// This can be useful after a session timeout or network interruption.
func (s *TM1Service) Reconnect(cfg Config, opts ...RestOption) error {
	rest, err := NewRestService(cfg, opts...)
	if err != nil {
		return err
	}
	s.rest = rest
	return nil
}

// Version returns the TM1 server product version string.
func (s *TM1Service) Version(ctx context.Context) (string, error) {
	resp, err := s.rest.Get(ctx, "/Configuration/ProductVersion/$value")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

// Metadata retrieves the TM1 REST metadata document.
func (s *TM1Service) Metadata(ctx context.Context) ([]byte, error) {
	resp, err := s.rest.Get(ctx, "/$metadata")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// RetrieveAsyncResponse retrieves the response from an async operation using the async_id.
func (s *TM1Service) RetrieveAsyncResponse(ctx context.Context, asyncID string) (*http.Response, error) {
	return s.rest.RetrieveAsyncResponse(ctx, asyncID)
}

// CancelAsyncOperation cancels an async operation by its async_id.
func (s *TM1Service) CancelAsyncOperation(ctx context.Context, asyncID string) error {
	return s.rest.CancelAsyncOperation(ctx, asyncID)
}

// CancelRunningOperation cancels a currently running operation by thread ID.
func (s *TM1Service) CancelRunningOperation(ctx context.Context, threadID string) error {
	return s.rest.CancelRunningOperation(ctx, threadID)
}
