package tm1

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// RestService manages HTTP interactions with the TM1 REST API.
type RestService struct {
	baseURL   *url.URL
	client    *http.Client
	headers   http.Header
	auth      AuthProvider
	logger    Logger
	keepAlive bool
}

// NewRestService constructs a RestService using the provided configuration and options.
func NewRestService(cfg Config, opts ...RestOption) (*RestService, error) {
	baseURLStr, err := cfg.EffectiveBaseURL()
	if err != nil {
		return nil, fmt.Errorf("determine base url: %w", err)
	}

	baseURL, err := url.Parse(baseURLStr)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}

	client, err := cfg.HTTPClientOrDefault()
	if err != nil {
		return nil, fmt.Errorf("create http client: %w", err)
	}

	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path = strings.TrimRight(baseURL.Path, "/") + "/"
	}

	rs := &RestService{
		baseURL:   baseURL,
		client:    client,
		headers:   cfg.HTTPHeaders(),
		auth:      nil,
		logger:    nopLogger{},
		keepAlive: cfg.KeepAlive,
	}

	// Set up logging if enabled
	if cfg.Logging {
		rs.logger = &defaultLogger{}
	}

	// Set up authentication based on config
	if err := rs.setupAuthentication(cfg); err != nil {
		return nil, fmt.Errorf("setup authentication: %w", err)
	}

	// Add impersonation header if specified
	if cfg.Impersonate != "" {
		rs.headers.Set("TM1-Impersonate", cfg.Impersonate)
	}

	// Apply any additional options (can override the auth provider)
	for _, opt := range opts {
		if err := opt(rs); err != nil {
			return nil, err
		}
	}

	return rs, nil
}

// setupAuthentication configures authentication based on Config settings
func (rs *RestService) setupAuthentication(cfg Config) error {
	// Session ID takes precedence - reuse existing session
	if cfg.SessionID != "" {
		rs.auth = SessionCookieAuth{
			Name:  "TM1SessionId",
			Value: cfg.SessionID,
		}
		return nil
	}

	// CAM Passport authentication
	if cfg.CAMPassport != "" {
		rs.auth = AuthFunc(func(req *http.Request) error {
			req.Header.Set("Authorization", "CAMPassport "+cfg.CAMPassport)
			return nil
		})
		return nil
	}

	// Access Token authentication
	if cfg.AccessToken != "" {
		rs.auth = BearerToken(cfg.AccessToken)
		return nil
	}

	// CAM authentication with namespace
	if cfg.Namespace != "" {
		password := cfg.Password
		if cfg.DecodeBase64 {
			password = decodeBase64Password(password)
		}
		rs.auth = AuthFunc(func(req *http.Request) error {
			credentials := fmt.Sprintf("%s:%s:%s", cfg.User, password, cfg.Namespace)
			encoded := base64Encode(credentials)
			req.Header.Set("Authorization", "CAMNamespace "+encoded)
			return nil
		})
		return nil
	}

	// Basic authentication
	if cfg.User != "" {
		password := cfg.Password
		if cfg.DecodeBase64 {
			password = decodeBase64Password(password)
		}
		rs.auth = BasicAuth{
			Username: cfg.User,
			Password: password,
		}
		return nil
	}

	return nil
}

// BaseURL returns the effective REST root URL.
func (rs *RestService) BaseURL() string {
	return rs.baseURL.String()
}

// Request executes an HTTP request against the TM1 REST API.
// The caller is responsible for closing the returned response body.
func (rs *RestService) Request(ctx context.Context, method, endpoint string, body io.Reader, opts ...RequestOption) (*http.Response, error) {
	req, err := rs.buildRequest(ctx, method, endpoint, body, opts...)
	if err != nil {
		return nil, err
	}

	rs.logger.Printf("tm1go %s %s", req.Method, req.URL)

	resp, err := rs.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tm1 request failed: %w", err)
	}

	if httpErr := newHTTPError(resp); httpErr != nil {
		resp.Body.Close()
		return nil, httpErr
	}

	return resp, nil
}

// JSON performs a request where the payload and response are JSON encoded.
func (rs *RestService) JSON(ctx context.Context, method, endpoint string, payload any, dest any, opts ...RequestOption) error {
	var body io.Reader

	if payload != nil {
		buffer := &bytes.Buffer{}
		encoder := json.NewEncoder(buffer)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(payload); err != nil {
			return fmt.Errorf("encode payload: %w", err)
		}
		body = buffer
	}

	resp, err := rs.Request(ctx, method, endpoint, body, opts...)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if dest == nil {
		_, err = io.Copy(io.Discard, resp.Body)
		return err
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(dest); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

// Get executes an HTTP GET.
func (rs *RestService) Get(ctx context.Context, endpoint string, opts ...RequestOption) (*http.Response, error) {
	return rs.Request(ctx, http.MethodGet, endpoint, nil, opts...)
}

// Post executes an HTTP POST.
func (rs *RestService) Post(ctx context.Context, endpoint string, body io.Reader, opts ...RequestOption) (*http.Response, error) {
	return rs.Request(ctx, http.MethodPost, endpoint, body, opts...)
}

// Patch executes an HTTP PATCH.
func (rs *RestService) Patch(ctx context.Context, endpoint string, body io.Reader, opts ...RequestOption) (*http.Response, error) {
	return rs.Request(ctx, http.MethodPatch, endpoint, body, opts...)
}

// Put executes an HTTP PUT.
func (rs *RestService) Put(ctx context.Context, endpoint string, body io.Reader, opts ...RequestOption) (*http.Response, error) {
	return rs.Request(ctx, http.MethodPut, endpoint, body, opts...)
}

// Delete executes an HTTP DELETE.
func (rs *RestService) Delete(ctx context.Context, endpoint string, opts ...RequestOption) (*http.Response, error) {
	return rs.Request(ctx, http.MethodDelete, endpoint, nil, opts...)
}

// Ping ensures the TM1 endpoint is reachable by fetching the product version.
func (rs *RestService) Ping(ctx context.Context) error {
	resp, err := rs.Get(ctx, "/Configuration/ProductVersion/$value")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

// Close releases idle connections held by the underlying HTTP client.
func (rs *RestService) Close() {
	if transport, ok := rs.client.Transport.(*http.Transport); ok && transport != nil {
		transport.CloseIdleConnections()
	}
}

// Logout terminates the TM1 session by calling the ActiveSession/tm1.Close endpoint.
// This properly closes the session on the TM1 server side.
func (rs *RestService) Logout(ctx context.Context) error {
	// If keepAlive is set, skip logout
	if rs.keepAlive {
		return nil
	}

	// Create request body with Connection: close header
	resp, err := rs.Post(ctx, "/ActiveSession/tm1.Close", nil)
	if err != nil {
		// If logout endpoint doesn't exist (404), treat as success
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == http.StatusNotFound {
			return nil
		}
		return fmt.Errorf("logout failed: %w", err)
	}
	defer resp.Body.Close()

	// Discard response body
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

// SessionID retrieves the current TM1SessionId cookie value.
// Returns empty string if no session cookie is found.
func (rs *RestService) SessionID() string {
	// Try to extract session ID from the HTTP client's cookie jar
	if rs.client.Jar != nil {
		cookies := rs.client.Jar.Cookies(rs.baseURL)
		for _, cookie := range cookies {
			if cookie.Name == "TM1SessionId" || cookie.Name == "paSession" {
				return cookie.Value
			}
		}
	}
	return ""
}

// IsConnected checks if the connection to TM1 server is active.
func (rs *RestService) IsConnected(ctx context.Context) bool {
	resp, err := rs.Get(ctx, "/Configuration/ServerName/$value")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return true
}

// Version returns the cached TM1 version string.
func (rs *RestService) Version() string {
	// Note: In TM1py, version is set during connect.
	// In Go, we don't cache it, so we'll need to call the API.
	// For now, return empty - users should use TM1Service.Version(ctx)
	return ""
}

// AddCompactJSONHeader modifies the Accept header to request compact JSON responses.
// Returns the original Accept header value for restoration if needed.
func (rs *RestService) AddCompactJSONHeader() string {
	original := rs.headers.Get("Accept")

	// Parse existing header
	parts := strings.Split(original, ";")

	// Insert compact format after application/json
	if len(parts) > 0 {
		result := []string{parts[0], "tm1.compact=v0"}
		result = append(result, parts[1:]...)
		modified := strings.Join(result, ";")
		rs.headers.Set("Accept", modified)
	}

	return original
}

// RetrieveAsyncResponse retrieves the response from an async operation using the async_id.
// The async_id is typically returned in the Location header of an async operation.
func (rs *RestService) RetrieveAsyncResponse(ctx context.Context, asyncID string) (*http.Response, error) {
	// TM1 async operations return a Location header with the async result URL
	// Format: /api/v1/AsyncResults('async_id')
	endpoint := fmt.Sprintf("/AsyncResults('%s')", asyncID)
	return rs.Get(ctx, endpoint)
}

// CancelAsyncOperation cancels an async operation by its async_id.
// Returns true if cancellation was successful.
func (rs *RestService) CancelAsyncOperation(ctx context.Context, asyncID string) error {
	endpoint := fmt.Sprintf("/AsyncResults('%s')", asyncID)
	resp, err := rs.Delete(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("cancel async operation: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

// CancelRunningOperation cancels a currently running operation (process, chore, etc.)
// by terminating the session thread.
func (rs *RestService) CancelRunningOperation(ctx context.Context, threadID string) error {
	// In TM1, you can cancel running operations by calling the Thread endpoint
	endpoint := fmt.Sprintf("/Threads(%s)/tm1.Cancel", threadID)
	resp, err := rs.Post(ctx, endpoint, nil)
	if err != nil {
		return fmt.Errorf("cancel running operation: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

// AddDefaultHeader appends a header that will be sent with every request.
func (rs *RestService) AddDefaultHeader(key, value string) {
	rs.headers.Add(key, value)
}

// RemoveDefaultHeader removes a default header.
func (rs *RestService) RemoveDefaultHeader(key string) {
	rs.headers.Del(key)
}

// SetBaseURL sets the base URL for the REST service (used for testing)
func (rs *RestService) SetBaseURL(baseURLStr string) error {
	baseURL, err := url.Parse(baseURLStr)
	if err != nil {
		return fmt.Errorf("parse base url: %w", err)
	}

	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path = strings.TrimRight(baseURL.Path, "/") + "/"
	}

	rs.baseURL = baseURL
	return nil
}

func (rs *RestService) buildRequest(ctx context.Context, method, endpoint string, body io.Reader, opts ...RequestOption) (*http.Request, error) {
	targetURL, err := rs.resolve(endpoint)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, targetURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header = cloneHeader(rs.headers)

	if rs.auth != nil {
		if err := rs.auth.Apply(req); err != nil {
			return nil, fmt.Errorf("apply auth: %w", err)
		}
	}

	for _, opt := range opts {
		opt(req)
	}

	return req, nil
}

func (rs *RestService) resolve(endpoint string) (*url.URL, error) {
	if endpoint == "" {
		clone := *rs.baseURL
		return &clone, nil
	}

	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return url.Parse(endpoint)
	}

	relPath := strings.TrimSpace(endpoint)
	relPath = strings.TrimLeft(relPath, "/")
	relURL, err := url.Parse(relPath)
	if err != nil {
		return nil, fmt.Errorf("parse endpoint: %w", err)
	}

	return rs.baseURL.ResolveReference(relURL), nil
}

func cloneHeader(in http.Header) http.Header {
	out := http.Header{}
	for key, values := range in {
		for _, value := range values {
			out.Add(key, value)
		}
	}
	return out
}

// defaultLogger implements basic logging to stdout
type defaultLogger struct{}

func (d *defaultLogger) Printf(format string, args ...any) {
	log.Printf(format, args...)
}

// decodeBase64Password decodes a base64-encoded password
func decodeBase64Password(encoded string) string {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return encoded // if decode fails, return original
	}
	return string(decoded)
}

// base64Encode encodes a string to base64
func base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
