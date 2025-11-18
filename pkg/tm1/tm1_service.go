package tm1

import (
	"context"
	"io"
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
