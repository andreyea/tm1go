package tm1

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RestService implements the Client interface for TM1 REST API communication
type RestService struct {
	config      *Config
	client      *http.Client
	baseURL     string
	authURL     string
	version     string
	headers     map[string]string
	sessionID   string
	authMode    AuthenticationMode
	isConnected bool
	mu          sync.RWMutex

	// Admin flags (cached after first check)
	isAdmin         *bool
	isDataAdmin     *bool
	isSecurityAdmin *bool
	isOpsAdmin      *bool
}

// NewRestService creates a new RestService instance
func NewRestService(config *Config) (*RestService, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Validate required configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	rs := &RestService{
		config:  config,
		headers: make(map[string]string),
	}

	// Copy default headers
	for k, v := range HTTPHeaders {
		rs.headers[k] = v
	}

	// Override session context if provided
	if config.SessionContext != "" {
		rs.headers["TM1-SessionContext"] = config.SessionContext
	}

	// Determine authentication mode
	rs.authMode = rs.determineAuthMode()

	// Construct URLs
	if err := rs.constructServiceAndAuthRoot(); err != nil {
		return nil, fmt.Errorf("failed to construct service URLs: %w", err)
	}

	// Initialize HTTP client
	rs.initHTTPClient()

	return rs, nil
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	if config.BaseURL == "" && config.Address == "" {
		return fmt.Errorf("either BaseURL or Address must be specified")
	}

	if config.BaseURL != "" && config.Address != "" {
		return fmt.Errorf("BaseURL and Address cannot both be specified")
	}

	return nil
}

// determineAuthMode determines the authentication mode based on config
func (rs *RestService) determineAuthMode() AuthenticationMode {
	if rs.config.APIKey != "" && rs.config.IAMURL != "" {
		return IBMCloudAPIKey
	}
	if rs.config.ApplicationClientID != "" && rs.config.ApplicationClientSecret != "" {
		return ServiceToService
	}
	if rs.config.APIKey != "" && rs.config.BaseURL != "" {
		return BasicAPIKey
	}
	if rs.config.CAMPassport != "" {
		if rs.config.User == "" && rs.config.Password == "" {
			return CAMSso
		}
		return CAM
	}
	if rs.config.IntegratedLogin {
		return WIA
	}

	return Basic
}

// constructServiceAndAuthRoot constructs the base and auth URLs
func (rs *RestService) constructServiceAndAuthRoot() error {
	if rs.authMode.UseV12Auth() {
		return rs.constructV12ServiceAndAuthRoot()
	}
	return rs.constructV11ServiceAndAuthRoot()
}

// constructV11ServiceAndAuthRoot constructs URLs for v11 and earlier
func (rs *RestService) constructV11ServiceAndAuthRoot() error {
	if rs.config.BaseURL != "" {
		rs.baseURL = rs.config.BaseURL
		if !strings.HasSuffix(rs.baseURL, "/api/v1") {
			rs.baseURL = strings.TrimSuffix(rs.baseURL, "/") + "/api/v1"
		}
		rs.authURL = rs.baseURL + "/Configuration/ProductVersion/$value"
	} else {
		protocol := "http"
		if rs.config.SSL {
			protocol = "https"
		}

		address := rs.config.Address
		if address == "" {
			address = "localhost"
		}

		port := ""
		if rs.config.Port > 0 {
			port = ":" + strconv.Itoa(rs.config.Port)
		}

		rs.baseURL = fmt.Sprintf("%s://%s%s/api/v1", protocol, address, port)
		rs.authURL = rs.baseURL + "/Configuration/ProductVersion/$value"
	}

	return nil
}

// constructV12ServiceAndAuthRoot constructs URLs for v12
func (rs *RestService) constructV12ServiceAndAuthRoot() error {
	if rs.authMode == IBMCloudAPIKey {
		return rs.constructIBMCloudServiceAndAuthRoot()
	}

	if rs.config.BaseURL == "" {
		return fmt.Errorf("BaseURL is required for v12 authentication")
	}

	rs.baseURL = rs.config.BaseURL
	if rs.config.AuthURL == "" {
		return fmt.Errorf("AuthURL is required for v12 authentication")
	}
	rs.authURL = rs.config.AuthURL

	return nil
}

// constructIBMCloudServiceAndAuthRoot constructs URLs for IBM Cloud
func (rs *RestService) constructIBMCloudServiceAndAuthRoot() error {
	if rs.config.PAURL == "" || rs.config.Tenant == "" {
		return fmt.Errorf("PAURL and Tenant are required for IBM Cloud authentication")
	}

	rs.baseURL = fmt.Sprintf("%s/api/v1/Databases('%s')",
		strings.TrimSuffix(rs.config.PAURL, "/"), rs.config.Tenant)
	rs.authURL = strings.TrimSuffix(rs.config.IAMURL, "/") + "/identity/token"

	return nil
}

// initHTTPClient initializes the HTTP client with proper configuration
func (rs *RestService) initHTTPClient() {
	timeout := 30 * time.Second
	if rs.config.Timeout != nil {
		timeout = *rs.config.Timeout
	}

	transport := &http.Transport{
		MaxIdleConns:        rs.config.ConnectionPoolSize,
		MaxIdleConnsPerHost: rs.config.ConnectionPoolSize,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	// Handle SSL verification
	if !rs.shouldVerifySSL() {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// Handle proxies
	if proxies, err := ParseProxies(rs.config.Proxies); err == nil && len(proxies) > 0 {
		transport.Proxy = func(req *http.Request) (*url.URL, error) {
			scheme := req.URL.Scheme
			if proxyURL, ok := proxies[scheme]; ok {
				return url.Parse(proxyURL)
			}
			return nil, nil
		}
	}

	rs.client = &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// shouldVerifySSL determines if SSL verification should be enabled
func (rs *RestService) shouldVerifySSL() bool {
	if rs.config.Verify == nil {
		// Default SSL verification in v12 is true
		return rs.authMode.UseV12Auth()
	}

	switch v := rs.config.Verify.(type) {
	case bool:
		return v
	case string:
		upper := strings.ToUpper(v)
		if upper == "FALSE" {
			return false
		}
		if upper == "TRUE" {
			return true
		}
		// Assume it's a certificate path
		return true
	default:
		return false
	}
}

// Connect establishes connection to TM1 server
func (rs *RestService) Connect() error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.config.Logging {
		log.Printf("Connecting to TM1 server: %s", rs.baseURL)
	}

	switch rs.authMode {
	case Basic:
		return rs.connectBasic()
	case WIA:
		return rs.connectWIA()
	case CAM:
		return rs.connectCAM()
	case CAMSso:
		return rs.connectCAMSso()
	case IBMCloudAPIKey:
		return rs.connectIBMCloud()
	case ServiceToService:
		return rs.connectServiceToService()
	case BasicAPIKey:
		return rs.connectBasicAPIKey()
	default:
		return fmt.Errorf("unsupported authentication mode: %d", rs.authMode)
	}
}

// connectBasic performs basic authentication
func (rs *RestService) connectBasic() error {
	// For basic auth, we'll authenticate on first request
	rs.isConnected = true
	// Set version after marking as connected to avoid deadlock
	return rs.setVersionDirect()
}

// connectWIA performs Windows Integrated Authentication
func (rs *RestService) connectWIA() error {
	// TODO: Implement WIA authentication
	return fmt.Errorf("WIA authentication not yet implemented")
}

// connectCAM performs CAM authentication
func (rs *RestService) connectCAM() error {
	// TODO: Implement CAM authentication
	return fmt.Errorf("CAM authentication not yet implemented")
}

// connectCAMSso performs CAM SSO authentication
func (rs *RestService) connectCAMSso() error {
	// TODO: Implement CAM SSO authentication
	return fmt.Errorf("CAM SSO authentication not yet implemented")
}

// connectIBMCloud performs IBM Cloud authentication
func (rs *RestService) connectIBMCloud() error {
	// TODO: Implement IBM Cloud authentication
	return fmt.Errorf("IBM Cloud authentication not yet implemented")
}

// connectServiceToService performs service-to-service authentication
func (rs *RestService) connectServiceToService() error {
	// TODO: Implement service-to-service authentication
	return fmt.Errorf("Service-to-service authentication not yet implemented")
}

// connectBasicAPIKey performs basic API key authentication
func (rs *RestService) connectBasicAPIKey() error {
	// TODO: Implement basic API key authentication
	return fmt.Errorf("Basic API key authentication not yet implemented")
}

// setVersion retrieves and sets the TM1 server version
func (rs *RestService) setVersion() error {
	resp, err := rs.GET("/Configuration/ProductVersion/$value", nil)
	if err != nil {
		return fmt.Errorf("failed to retrieve version: %w", err)
	}

	rs.version = strings.Trim(string(resp.Body), `"`)
	if rs.config.Logging {
		log.Printf("Connected to TM1 server version: %s", rs.version)
	}

	return nil
}

// setVersionDirect retrieves version without going through the request method (to avoid deadlock)
func (rs *RestService) setVersionDirect() error {
	// Build URL directly
	versionURL := rs.baseURL + "/Configuration/ProductVersion/$value"

	// Create HTTP request directly
	req, err := http.NewRequest("GET", versionURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create version request: %w", err)
	}

	// Add headers
	for k, v := range rs.headers {
		req.Header.Set(k, v)
	}

	// Add authentication
	if err := rs.addAuthentication(req); err != nil {
		return fmt.Errorf("failed to add authentication: %w", err)
	}

	// Perform request
	resp, err := rs.client.Do(req)
	if err != nil {
		return fmt.Errorf("version request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read version response: %w", err)
	}

	// Check for errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("version request failed with status %d: %s", resp.StatusCode, string(body))
	}

	rs.version = strings.Trim(string(body), `"`)
	if rs.config.Logging {
		log.Printf("Connected to TM1 server version: %s", rs.version)
	}

	return nil
}

// Disconnect closes the connection to TM1 server
func (rs *RestService) Disconnect() error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if !rs.isConnected {
		return nil
	}

	// Attempt logout
	if rs.sessionID != "" {
		_, _ = rs.POST("/ActiveSession/tm1.Close", nil, nil)
	}

	rs.isConnected = false
	rs.sessionID = ""

	if rs.config.Logging {
		log.Printf("Disconnected from TM1 server")
	}

	return nil
}

// IsConnected returns the connection status
func (rs *RestService) IsConnected() bool {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.isConnected
}

// Config returns the configuration
func (rs *RestService) Config() *Config {
	return rs.config
}

// Version returns the TM1 server version
func (rs *RestService) Version() string {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.version
}

// GET performs an HTTP GET request
func (rs *RestService) GET(url string, opts *RequestOptions) (*Response, error) {
	return rs.request("GET", url, nil, opts)
}

// POST performs an HTTP POST request
func (rs *RestService) POST(url string, data []byte, opts *RequestOptions) (*Response, error) {
	return rs.request("POST", url, data, opts)
}

// PATCH performs an HTTP PATCH request
func (rs *RestService) PATCH(url string, data []byte, opts *RequestOptions) (*Response, error) {
	return rs.request("PATCH", url, data, opts)
}

// PUT performs an HTTP PUT request
func (rs *RestService) PUT(url string, data []byte, opts *RequestOptions) (*Response, error) {
	return rs.request("PUT", url, data, opts)
}

// DELETE performs an HTTP DELETE request
func (rs *RestService) DELETE(url string, opts *RequestOptions) (*Response, error) {
	return rs.request("DELETE", url, nil, opts)
}

// request performs the actual HTTP request
func (rs *RestService) request(method, urlPath string, data []byte, opts *RequestOptions) (*Response, error) {
	// Check if we need to connect (avoid deadlock by checking without lock first)
	needsConnection := false
	rs.mu.RLock()
	// Don't try to connect if we're already trying to get the version during connection
	isVersionCall := strings.Contains(urlPath, "Configuration/ProductVersion")
	needsConnection = !rs.isConnected && !isVersionCall
	rs.mu.RUnlock()

	if needsConnection {
		if err := rs.Connect(); err != nil {
			return nil, fmt.Errorf("failed to connect: %w", err)
		}
	}

	// Prepare URL
	fullURL := rs.baseURL + urlPath
	if !strings.HasPrefix(urlPath, "/") {
		fullURL = rs.baseURL + "/" + urlPath
	}

	// Create request
	ctx := context.Background()
	if opts != nil && opts.Context != nil {
		ctx = opts.Context
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	headers := rs.headers
	if opts != nil && opts.Headers != nil {
		headers = make(map[string]string)
		for k, v := range rs.headers {
			headers[k] = v
		}
		for k, v := range opts.Headers {
			headers[k] = v
		}
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Add authentication
	if err := rs.addAuthentication(req); err != nil {
		return nil, fmt.Errorf("failed to add authentication: %w", err)
	}

	// Handle async mode
	if opts != nil && opts.AsyncMode != nil && *opts.AsyncMode {
		req.Header.Set("Prefer", "respond-async")
	}

	// Set timeout
	timeout := rs.config.Timeout
	if opts != nil && opts.Timeout != nil {
		timeout = opts.Timeout
	}

	if timeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
		req = req.WithContext(ctx)
	}

	// Perform request
	resp, err := rs.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle async response
	if opts != nil && opts.ReturnAsyncID && resp.Header.Get("Location") != "" {
		asyncID, err := ExtractAsyncID(resp.Header.Get("Location"))
		if err != nil {
			return nil, fmt.Errorf("failed to extract async ID: %w", err)
		}
		return &Response{
			StatusCode: resp.StatusCode,
			Body:       []byte(asyncID),
			Headers:    resp.Header,
		}, nil
	}

	// Check for errors
	if err := rs.verifyResponse(resp, body, method, fullURL); err != nil {
		return nil, err
	}

	// Handle session timeout and retry
	if resp.StatusCode == 401 && rs.config.ReconnectOnSessionTimeout {
		if err := rs.Connect(); err == nil {
			// Retry the request once
			return rs.request(method, urlPath, data, opts)
		}
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       body,
		Headers:    resp.Header,
	}, nil
}

// addAuthentication adds authentication to the request
func (rs *RestService) addAuthentication(req *http.Request) error {
	switch rs.authMode {
	case Basic:
		if rs.config.User != "" {
			password := rs.config.Password
			if rs.config.DecodeB64 && password != "" {
				decoded, err := Base64Decode(password)
				if err != nil {
					return fmt.Errorf("failed to decode base64 password: %w", err)
				}
				password = decoded
			}
			req.SetBasicAuth(rs.config.User, password)
		}
	}

	return nil
}

// verifyResponse checks the response for errors
func (rs *RestService) verifyResponse(resp *http.Response, body []byte, method, url string) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	// Try to parse error response
	var errorResp struct {
		Error struct {
			Code    string        `json:"code"`
			Message string        `json:"message"`
			Details []ErrorDetail `json:"details"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Code != "" {
		return NewTM1RestException(method, url, errorResp.Error.Code,
			errorResp.Error.Message, resp.StatusCode)
	}

	// Generic error
	return NewTM1RestException(method, url, "HTTP_ERROR",
		string(body), resp.StatusCode)
}
