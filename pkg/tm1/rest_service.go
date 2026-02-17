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
	"time"
)

// AuthenticationMode represents the authentication method being used
type AuthenticationMode int

const (
	AuthModeBasic            AuthenticationMode = iota + 1
	AuthModeWIA                                 // Windows Integrated Authentication
	AuthModeCAM                                 // CAM authentication
	AuthModeCAMSSO                              // CAM SSO
	AuthModeIBMCloudAPIKey                      // IBM Cloud API Key
	AuthModeServiceToService                    // Service-to-service authentication
	AuthModePAProxy                             // Planning Analytics Proxy
	AuthModeBasicAPIKey                         // Basic API Key
	AuthModeAccessToken                         // Access Token
)

// RestService manages HTTP interactions with the TM1 REST API.
type RestService struct {
	baseURL                     *url.URL
	authURL                     *url.URL
	client                      *http.Client
	headers                     http.Header
	auth                        AuthProvider
	logger                      Logger
	keepAlive                   bool
	version                     string
	authMode                    AuthenticationMode
	kwargs                      Config // Store original config for reconnection
	reConnectOnSessionTimeout   bool
	reConnectOnRemoteDisconnect bool
	asyncRequestsMode           bool
	cancelAtTimeout             bool
	timeout                     time.Duration
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
		baseURL:                     baseURL,
		client:                      client,
		headers:                     cfg.HTTPHeaders(),
		auth:                        nil,
		logger:                      nopLogger{},
		keepAlive:                   cfg.KeepAlive,
		kwargs:                      cfg,
		reConnectOnSessionTimeout:   cfg.ReConnectOnSessionTimeout,
		reConnectOnRemoteDisconnect: cfg.ReConnectOnRemoteDisconnect,
		asyncRequestsMode:           cfg.AsyncRequestsMode,
		cancelAtTimeout:             cfg.CancelAtTimeout,
		timeout:                     cfg.Timeout,
	}

	// Set default reconnection behavior if not specified
	if !cfg.ReConnectOnSessionTimeout {
		rs.reConnectOnSessionTimeout = true
	}
	if !cfg.ReConnectOnRemoteDisconnect {
		rs.reConnectOnRemoteDisconnect = true
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
	// Determine authentication mode
	rs.authMode = rs.determineAuthMode(cfg)

	// Session ID takes precedence - reuse existing session
	if cfg.SessionID != "" {
		rs.auth = SessionCookieAuth{
			Name:  "TM1SessionId",
			Value: cfg.SessionID,
		}
		return nil
	}

	// SaaS API Key authentication (v12) - basic auth with username='apikey'
	if rs.authMode == AuthModeBasicAPIKey {
		rs.auth = BasicAuth{
			Username: "apikey",
			Password: cfg.APIKey,
		}
		return nil
	}

	// IBM Cloud API Key authentication (v12) - requires IAM token generation
	if cfg.APIKey != "" && rs.authMode == AuthModeIBMCloudAPIKey {
		// Generate IBM Cloud IAM access token
		accessToken, err := rs.generateIBMIAMCloudAccessToken(cfg)
		if err != nil {
			return fmt.Errorf("failed to generate IBM Cloud access token: %w", err)
		}
		rs.auth = BearerToken(accessToken)
		return nil
	}

	// Service-to-Service authentication (v12)
	if rs.authMode == AuthModeServiceToService {
		if cfg.ApplicationClientID == "" || cfg.ApplicationClientSecret == "" {
			return fmt.Errorf("ApplicationClientID and ApplicationClientSecret required for service-to-service auth")
		}
		// Service-to-service uses special endpoint and client credentials
		rs.auth = BasicAuth{
			Username: cfg.ApplicationClientID,
			Password: cfg.ApplicationClientSecret,
		}
		return nil
	}

	// Access Token authentication
	if cfg.AccessToken != "" {
		rs.auth = BearerToken(cfg.AccessToken)
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

	// Windows Integrated Authentication
	if cfg.IntegratedLogin {
		// Note: Windows integrated auth requires platform-specific implementation
		// This is a placeholder - actual implementation would use SSPI on Windows
		rs.authMode = AuthModeWIA
		return fmt.Errorf("Windows Integrated Authentication not yet implemented in Go")
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
// If asyncRequestsMode is enabled, the request will use async mode automatically.
func (rs *RestService) Request(ctx context.Context, method, endpoint string, body io.Reader, opts ...RequestOption) (*http.Response, error) {
	// If asyncRequestsMode is enabled, use async request handling
	if rs.asyncRequestsMode {
		resp, _, err := rs.RequestAsync(ctx, method, endpoint, body, false, opts...)
		return resp, err
	}

	req, err := rs.buildRequest(ctx, method, endpoint, body, opts...)
	if err != nil {
		return nil, err
	}

	rs.logRequest(req, "")

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
// Note: This always uses synchronous mode, regardless of asyncRequestsMode setting.
func (rs *RestService) Logout(ctx context.Context) error {
	// If keepAlive is set, skip logout
	if rs.keepAlive {
		return nil
	}

	// Build request directly without going through async mode
	req, err := rs.buildRequest(ctx, http.MethodPost, "/ActiveSession/tm1.Close", nil)
	if err != nil {
		return fmt.Errorf("build logout request: %w", err)
	}

	rs.logRequest(req, "")

	resp, err := rs.client.Do(req)
	if err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}
	defer resp.Body.Close()

	// If logout endpoint doesn't exist (404), treat as success
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if httpErr := newHTTPError(resp); httpErr != nil {
		return httpErr
	}

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

// RequestAsync executes an HTTP request with async mode enabled.
// It adds the 'Prefer: respond-async' header and handles async response polling.
// Returns the async_id if returnAsyncID is true, otherwise polls until completion.
func (rs *RestService) RequestAsync(ctx context.Context, method, endpoint string, body io.Reader, returnAsyncID bool, opts ...RequestOption) (*http.Response, string, error) {
	// Add async header
	asyncOpts := append(opts, WithHeader("Prefer", "respond-async"))

	req, err := rs.buildRequest(ctx, method, endpoint, body, asyncOpts...)
	if err != nil {
		return nil, "", err
	}

	rs.logRequest(req, " (async)")

	resp, err := rs.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("tm1 async request failed: %w", err)
	}

	if httpErr := newHTTPError(resp); httpErr != nil {
		resp.Body.Close()
		return nil, "", httpErr
	}

	// Check for async response (Location header contains async_id)
	location := resp.Header.Get("Location")
	if location != "" && strings.Contains(location, "'") {
		// Extract async_id from Location header (format: /_async('async_id'))
		asyncID := extractAsyncID(location)

		// Close initial response
		resp.Body.Close()

		if returnAsyncID {
			// Return async_id immediately
			return nil, asyncID, nil
		}

		// Poll for completion
		return rs.pollAsyncResponse(ctx, asyncID, method, endpoint)
	}

	// Not an async response, return as-is
	return resp, "", nil
}

// pollAsyncResponse polls the /_async endpoint until the operation completes
func (rs *RestService) pollAsyncResponse(ctx context.Context, asyncID, method, endpoint string) (*http.Response, string, error) {
	timeout := rs.timeout
	if timeout == 0 {
		timeout = 300 * time.Second // Default 5 minutes
	}

	deadline := time.Now().Add(timeout)
	waitTimes := []time.Duration{100 * time.Millisecond, 300 * time.Millisecond, 600 * time.Millisecond}
	waitIndex := 0

	for {
		// Check timeout
		if time.Now().After(deadline) {
			if rs.cancelAtTimeout {
				_ = rs.CancelAsyncOperation(ctx, asyncID)
			}
			return nil, "", fmt.Errorf("async operation timed out after %v", timeout)
		}

		// Wait before polling
		var waitTime time.Duration
		if waitIndex < len(waitTimes) {
			waitTime = waitTimes[waitIndex]
			waitIndex++
		} else {
			waitTime = 1 * time.Second
		}

		select {
		case <-ctx.Done():
			_ = rs.CancelAsyncOperation(ctx, asyncID)
			return nil, "", ctx.Err()
		case <-time.After(waitTime):
		}

		// Poll for result
		resp, err := rs.RetrieveAsyncResponse(ctx, asyncID)
		if err != nil {
			continue // Keep polling on error
		}

		// Check if operation completed successfully
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			// Transform response if needed (for TM1 v11 compatibility)
			return rs.transformAsyncResponse(resp)
		}

		// If status is 202 Accepted, operation is still running - continue polling
		if resp.StatusCode == http.StatusAccepted {
			resp.Body.Close()
			continue
		}

		// Any other status code is an error - return it
		if httpErr := newHTTPError(resp); httpErr != nil {
			resp.Body.Close()
			return nil, "", httpErr
		}

		// Unexpected status code but no error
		resp.Body.Close()
	}
}

// transformAsyncResponse handles response transformation for TM1 version compatibility
func (rs *RestService) transformAsyncResponse(resp *http.Response) (*http.Response, string, error) {
	// Read first few bytes to check if response starts with "HTTP/"
	// This is necessary for TM1 v11 where async responses contain raw HTTP response
	buf := make([]byte, 5)
	n, err := resp.Body.Read(buf)
	if err != nil && err != io.EOF {
		resp.Body.Close()
		return nil, "", fmt.Errorf("read response: %w", err)
	}

	// Recreate body with the read bytes
	resp.Body = io.NopCloser(io.MultiReader(bytes.NewReader(buf[:n]), resp.Body))

	// Check if response starts with "HTTP/" (TM1 v11 format)
	if n >= 5 && string(buf[:5]) == "HTTP/" {
		// For TM1 v11, we need to parse the embedded HTTP response
		// For now, just log a warning and return as-is
		// TODO: Implement full HTTP response parsing if needed
		rs.logger.Printf("Warning: Received TM1 v11 async response format (starts with HTTP/)")
	}

	// Check for asyncresult header (TM1 v12)
	if asyncResult := resp.Header.Get("asyncresult"); asyncResult != "" {
		// Parse status code from asyncresult header
		// Format: "200 OK" or "201 Created"
		parts := strings.SplitN(asyncResult, " ", 2)
		if len(parts) >= 1 {
			if statusCode, err := fmt.Sscanf(parts[0], "%d", &resp.StatusCode); err == nil && statusCode == 1 {
				// Status code successfully parsed
			}
		}
	}

	return resp, "", nil
}

// extractAsyncID extracts the async_id from a Location header
// Format: /_async('async_id') or similar
func extractAsyncID(location string) string {
	// Find content between single quotes
	start := strings.Index(location, "'")
	if start == -1 {
		return ""
	}
	end := strings.Index(location[start+1:], "'")
	if end == -1 {
		return ""
	}
	return location[start+1 : start+1+end]
}

// RetrieveAsyncResponse retrieves the response from an async operation using the async_id.
// The async_id is typically returned in the Location header of an async operation.
func (rs *RestService) RetrieveAsyncResponse(ctx context.Context, asyncID string) (*http.Response, error) {
	// TM1 async operations return a Location header with the async result URL
	// Format: /_async('async_id')
	endpoint := fmt.Sprintf("/_async('%s')", asyncID)
	return rs.Get(ctx, endpoint)
}

// CancelAsyncOperation cancels an async operation by its async_id.
// Returns true if cancellation was successful.
func (rs *RestService) CancelAsyncOperation(ctx context.Context, asyncID string) error {
	endpoint := fmt.Sprintf("/_async('%s')", asyncID)
	resp, err := rs.Delete(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("cancel async operation: %w", err)
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

// determineAuthMode determines which authentication mode to use based on config
func (rs *RestService) determineAuthMode(cfg Config) AuthenticationMode {
	// Session ID reuse
	if cfg.SessionID != "" {
		return AuthModeBasic // treat as basic since we're reusing session
	}

	// SaaS API Key (v12) - uses basic auth with username='apikey'
	if cfg.APIKey != "" && strings.Contains(cfg.Address, "planninganalytics.saas.ibm.com") {
		return AuthModeBasicAPIKey
	}

	// IBM Cloud API Key (v12) - requires IAM token generation
	if cfg.APIKey != "" && (cfg.Tenant != "" || cfg.IAMUrl != "") {
		return AuthModeIBMCloudAPIKey
	}

	// Service-to-Service (v12)
	if cfg.ApplicationClientID != "" && cfg.ApplicationClientSecret != "" {
		return AuthModeServiceToService
	}

	// Access Token
	if cfg.AccessToken != "" {
		return AuthModeAccessToken
	}

	// CAM Passport
	if cfg.CAMPassport != "" {
		return AuthModeCAM
	}

	// CAM with namespace
	if cfg.Namespace != "" {
		return AuthModeCAM
	}

	// Windows Integrated Authentication
	if cfg.IntegratedLogin {
		return AuthModeWIA
	}

	// PA Proxy (for Planning Analytics Workspace)
	if cfg.CPDUrl != "" {
		return AuthModePAProxy
	}

	// Default to basic auth
	return AuthModeBasic
}

// generateIBMIAMCloudAccessToken generates an IBM Cloud IAM access token
func (rs *RestService) generateIBMIAMCloudAccessToken(cfg Config) (string, error) {
	iamURL := cfg.IAMUrl
	if iamURL == "" {
		iamURL = "https://iam.cloud.ibm.com"
	}

	// Create request to IBM IAM token endpoint
	tokenURL := fmt.Sprintf("%s/identity/token", iamURL)
	payload := fmt.Sprintf("grant_type=urn:ibm:params:oauth:grant-type:apikey&apikey=%s", cfg.APIKey)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Use a separate client for IAM requests
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("IBM IAM token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}

	return tokenResp.AccessToken, nil
}

// constructServiceRoot constructs the appropriate base URL and auth URL based on deployment type
func (rs *RestService) constructServiceRoot(cfg Config) (string, string, error) {
	// If base URL is explicitly provided, use it
	if cfg.BaseURL != "" {
		return rs.constructAllVersionServiceFromBaseURL(cfg)
	}

	// IBM Cloud (v12) - uses tenant and database
	if cfg.Tenant != "" && cfg.Database != "" {
		return rs.constructIBMCloudServiceAndAuthRoot(cfg)
	}

	// Service-to-Service (v12) - uses instance and database
	if cfg.Instance != "" && cfg.Database != "" {
		return rs.constructS2SServiceAndAuthRoot(cfg)
	}

	// PA Proxy (Planning Analytics Workspace)
	if cfg.CPDUrl != "" {
		return rs.constructPAProxyServiceAndAuthRoot(cfg)
	}

	// Traditional v11 TM1 Server
	return rs.constructV11ServiceAndAuthRoot(cfg)
}

func (rs *RestService) constructIBMCloudServiceAndAuthRoot(cfg Config) (string, string, error) {
	if cfg.Address == "" || cfg.Tenant == "" || cfg.Database == "" {
		return "", "", fmt.Errorf("Address, Tenant, and Database required for IBM Cloud deployment")
	}

	baseURL := fmt.Sprintf("https://%s/api/%s/v0/tm1/%s", cfg.Address, cfg.Tenant, cfg.Database)
	authURL := fmt.Sprintf("%s/Configuration/ProductVersion/$value", baseURL)

	return baseURL, authURL, nil
}

func (rs *RestService) constructS2SServiceAndAuthRoot(cfg Config) (string, string, error) {
	if cfg.Instance == "" || cfg.Database == "" {
		return "", "", fmt.Errorf("Instance and Database required for Service-to-Service deployment")
	}

	scheme := "https"
	if !cfg.SSL {
		scheme = "http"
	}

	host := "localhost"
	if cfg.Address != "" {
		host = cfg.Address
	}

	portStr := ""
	if cfg.Port > 0 {
		portStr = fmt.Sprintf(":%d", cfg.Port)
	}

	baseURL := fmt.Sprintf("%s://%s%s/%s/api/v1/Databases('%s')", scheme, host, portStr, cfg.Instance, cfg.Database)
	authURL := fmt.Sprintf("%s://%s%s/%s/auth/v1/session", scheme, host, portStr, cfg.Instance)

	return baseURL, authURL, nil
}

func (rs *RestService) constructPAProxyServiceAndAuthRoot(cfg Config) (string, string, error) {
	if cfg.Address == "" || cfg.Database == "" {
		return "", "", fmt.Errorf("Address and Database required for PA Proxy deployment")
	}

	scheme := "https"
	if !cfg.SSL {
		scheme = "http"
	}

	baseURL := fmt.Sprintf("%s://%s/tm1/%s/api/v1", scheme, cfg.Address, cfg.Database)
	authURL := fmt.Sprintf("%s://%s/login", scheme, cfg.Address)

	return baseURL, authURL, nil
}

func (rs *RestService) constructV11ServiceAndAuthRoot(cfg Config) (string, string, error) {
	if cfg.Address == "" {
		return "", "", fmt.Errorf("Address required for TM1 v11 deployment")
	}

	scheme := "https"
	if !cfg.SSL {
		scheme = "http"
	}

	host := cfg.Address
	portStr := ""
	if cfg.Port > 0 {
		portStr = fmt.Sprintf(":%d", cfg.Port)
	}

	baseURL := fmt.Sprintf("%s://%s%s/api/v1", scheme, host, portStr)
	authURL := fmt.Sprintf("%s/Configuration/ProductVersion/$value", baseURL)

	return baseURL, authURL, nil
}

func (rs *RestService) constructAllVersionServiceFromBaseURL(cfg Config) (string, string, error) {
	baseURL := cfg.BaseURL

	// Detect deployment type from base URL
	if strings.Contains(baseURL, "/api/") && strings.Contains(baseURL, "/v0/tm1/") {
		// IBM Cloud format
		authURL := baseURL + "/Configuration/ProductVersion/$value"
		return baseURL, authURL, nil
	} else if strings.Contains(baseURL, "/api/v1/Databases") {
		// Service-to-Service format
		parts := strings.Split(baseURL, "/api/v1/Databases")
		authURL := parts[0] + "/auth/v1/session"
		return baseURL, authURL, nil
	} else if baseURL == strings.TrimSuffix(baseURL, "/api/v1")+"/api/v1" {
		// Standard v11 format
		authURL := baseURL + "/Configuration/ProductVersion/$value"
		return baseURL, authURL, nil
	}

	// Default
	authURL := baseURL + "/Configuration/ProductVersion/$value"
	return baseURL, authURL, nil
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

func (rs *RestService) logRequest(req *http.Request, suffix string) {
	payload := requestPayloadForLog(req, 4096)
	if payload == "" {
		rs.logger.Printf("tm1go %s %s%s", req.Method, req.URL, suffix)
		return
	}

	rs.logger.Printf("tm1go %s %s%s payload=%s", req.Method, req.URL, suffix, payload)
}

func requestPayloadForLog(req *http.Request, maxBytes int) string {
	if req == nil || req.GetBody == nil {
		return ""
	}

	bodyCopy, err := req.GetBody()
	if err != nil {
		return fmt.Sprintf("<payload-read-error: %v>", err)
	}
	defer bodyCopy.Close()

	data, err := io.ReadAll(io.LimitReader(bodyCopy, int64(maxBytes+1)))
	if err != nil {
		return fmt.Sprintf("<payload-read-error: %v>", err)
	}

	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return ""
	}

	if len(data) > maxBytes {
		return strings.TrimSpace(string(data[:maxBytes])) + "...(truncated)"
	}

	return trimmed
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
