package tm1

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Logger is the minimal interface required for debug logging inside the clients.
type Logger interface {
	Printf(format string, args ...any)
}

type nopLogger struct{}

func (nopLogger) Printf(string, ...any) {}

// AuthProvider supplies authentication material for outgoing requests.
type AuthProvider interface {
	Apply(req *http.Request) error
}

// AuthFunc adapts a function to the AuthProvider interface.
type AuthFunc func(req *http.Request) error

// Apply implements AuthProvider.
func (f AuthFunc) Apply(req *http.Request) error {
	return f(req)
}

// BasicAuth provides HTTP Basic authentication.
type BasicAuth struct {
	Username string
	Password string
}

// Apply implements AuthProvider.
func (b BasicAuth) Apply(req *http.Request) error {
	req.SetBasicAuth(b.Username, b.Password)
	return nil
}

// BearerToken supplies static bearer token authentication.
type BearerToken string

// Apply implements AuthProvider.
func (b BearerToken) Apply(req *http.Request) error {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", string(b)))
	return nil
}

// HeaderAuth injects arbitrary headers required by upstream gateways.
type HeaderAuth map[string]string

// Apply implements AuthProvider.
func (h HeaderAuth) Apply(req *http.Request) error {
	for key, value := range h {
		req.Header.Set(key, value)
	}
	return nil
}

// SessionCookieAuth reuses an existing TM1 session by attaching the TM1SessionId cookie.
type SessionCookieAuth struct {
	Name  string
	Value string
}

// Apply implements AuthProvider.
func (s SessionCookieAuth) Apply(req *http.Request) error {
	value := strings.TrimSpace(s.Value)
	if value == "" {
		return errors.New("tm1: session cookie value is empty")
	}

	name := strings.TrimSpace(s.Name)
	if name == "" {
		name = "TM1SessionId"
	}

	req.AddCookie(&http.Cookie{Name: name, Value: value})
	return nil
}

// RestOption configures a RestService instance.
// These options allow runtime customization beyond what Config provides.
type RestOption func(*RestService) error

// WithAuthProvider overrides the authentication provider determined from Config.
// Use this for custom authentication schemes not supported by Config.
func WithAuthProvider(provider AuthProvider) RestOption {
	return func(rs *RestService) error {
		if provider == nil {
			return errors.New("tm1: auth provider cannot be nil")
		}
		rs.auth = provider
		return nil
	}
}

// WithLogger sets a custom logger for debugging HTTP requests.
func WithLogger(logger Logger) RestOption {
	return func(rs *RestService) error {
		if logger == nil {
			rs.logger = nopLogger{}
		} else {
			rs.logger = logger
		}
		return nil
	}
}

// WithAdditionalHeaders extends the default headers applied to every request.
// Use this to add gateway or proxy headers not in Config.
func WithAdditionalHeaders(headers http.Header) RestOption {
	return func(rs *RestService) error {
		for key, values := range headers {
			for _, value := range values {
				rs.headers.Add(key, value)
			}
		}
		return nil
	}
}

// RequestOption mutates individual HTTP requests before they are sent.
// Use these for per-request customization (query params, headers, etc).
type RequestOption func(*http.Request)

// WithHeader sets a request-scoped header value.
func WithHeader(key, value string) RequestOption {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

// WithQueryValues appends the provided query parameters to the request.
func WithQueryValues(values url.Values) RequestOption {
	return func(req *http.Request) {
		query := req.URL.Query()
		for key, vals := range values {
			for _, value := range vals {
				query.Add(key, value)
			}
		}
		req.URL.RawQuery = query.Encode()
	}
}

// WithQueryValue sets a single query parameter on the request.
func WithQueryValue(key, value string) RequestOption {
	return WithQueryValues(url.Values{key: {value}})
}
