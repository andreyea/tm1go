package tm1

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

const (
	defaultTimeout        = 60 * time.Second
	defaultSessionContext = "tm1go"
)

// Config captures the connection inputs needed to communicate with a TM1 REST API instance.
type Config struct {
	// Core connection parameters
	Address  string // Hostname or IP of the TM1 server
	Port     int    // HTTPPortNumber defined in tm1s.cfg
	BaseURL  string // Optional full base URL, e.g. https://localhost:12354/api/v1
	SSL      bool   // Controls http vs https when BaseURL is not provided
	Instance string // Planning Analytics Engine (v12) instance name
	Database string // Planning Analytics Engine (v12) database name

	// Authentication parameters
	User                    string // Username for authentication
	Password                string // Password for authentication
	DecodeBase64            bool   // Whether password argument is base64 encoded
	Namespace               string // Optional namespace for LDAP or CAM authentication
	CAMPassport             string // The CAM passport
	SessionID               string // TM1SessionId for reusing existing session e.g. q7O6e1w49AixeuLVxJ1GZg
	ApplicationClientID     string // Planning Analytics Engine (v12) named application client ID
	ApplicationClientSecret string // Planning Analytics Engine (v12) named application secret
	APIKey                  string // Planning Analytics Engine (v12) API Key from IBM Cloud
	IAMUrl                  string // Planning Analytics Engine (v12) IBM Cloud IAM URL. Default: "https://iam.cloud.ibm.com"
	PAUrl                   string // Planning Analytics Engine (v12) PA URL e.g., "https://us-east-2.aws.planninganalytics.ibm.com"
	CPDUrl                  string // Cloud Pack for Data URL (aka ZEN) e.g., "https://cpd-zen.apps.cp4dpa-test11.cp.fyre.ibm.com"
	Tenant                  string // Planning Analytics Engine (v12) Tenant e.g., YC4B2M1AG2Y6
	Gateway                 string // Optional gateway for CAM authentication
	AuthURL                 string // Auth URL for Planning Analytics Engine (v12)
	AccessToken             string // Access token for token-based authentication

	// Integrated login (Windows authentication) parameters
	IntegratedLogin         bool   // True for IntegratedSecurityMode3
	IntegratedLoginDomain   string // NT Domain name. Default: '.' for local account
	IntegratedLoginService  string // Kerberos Service type for remote Service Principal Name. Default: 'HTTP'
	IntegratedLoginHost     string // Host name for Service Principal Name. Default: Extracted from request URI
	IntegratedLoginDelegate bool   // Indicates that the user's credentials are to be delegated to the server

	// Request behavior parameters
	Timeout             time.Duration // Per-request timeout applied to the HTTP client
	CancelAtTimeout     bool          // Abort operation in TM1 when timeout is reached
	AsyncRequestsMode   bool          // Changes internal REST execution mode to avoid 60s timeout on IBM cloud
	SessionContext      string        // Value for TM1-SessionContext header, surfaced in TM1top/Arc
	Impersonate         string        // Name of user to impersonate
	Verify              interface{}   // Path to .cer file or boolean for SSL verification
	SkipSSLVerification bool          // Allows self-signed certificates during development
	Logging             bool          // Switch on/off verbose http logging

	// Connection pool parameters
	ConnectionPoolSize int // Maximum number of connections to save in the pool (default: 10)
	PoolConnections    int // Number of connection pools to cache (default: 1 for a single TM1 instance)

	// Reconnection behavior
	ReConnectOnSessionTimeout   bool // Attempt to reconnect once if session is timed out (default: true)
	ReConnectOnRemoteDisconnect bool // Attempt to reconnect once if connection is aborted by remote end (default: true)

	// Session management
	KeepAlive bool // If true, Close() will not logout from TM1, keeping the session alive for reuse

	// HTTP client customization
	ProxyURL       string       // Optional HTTP proxy definition
	Proxies        interface{}  // Pass a map or JSON string with proxies e.g. {"http": "http://proxy:8080"}
	HTTPClient     *http.Client // Optional pre-configured HTTP client
	DefaultHeaders http.Header  // Optional additional default headers
	SSLContext     interface{}  // Pass a user defined ssl context
	Cert           interface{}  // If String, path to SSL client cert file (.pem). If Tuple, ('cert', 'key') pair
}

// Validate performs static checks on the configuration.
func (c Config) Validate() error {
	if c.BaseURL == "" && c.Address == "" {
		return errors.New("either BaseURL or Address must be provided")
	}

	if c.BaseURL != "" {
		if !strings.HasPrefix(c.BaseURL, "http://") && !strings.HasPrefix(c.BaseURL, "https://") {
			return errors.New("BaseURL must include http or https scheme")
		}
	}

	return nil
}

// EffectiveBaseURL returns the normalized REST root URL.
func (c Config) EffectiveBaseURL() (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}

	if c.BaseURL != "" {
		return strings.TrimRight(c.BaseURL, "/"), nil
	}

	// IBM Cloud / SaaS (v12) - uses tenant and database
	if c.Tenant != "" && c.Database != "" {
		scheme := "https"
		if !c.SSL {
			scheme = "http"
		}
		return fmt.Sprintf("%s://%s/api/%s/v0/tm1/%s", scheme, c.Address, c.Tenant, c.Database), nil
	}

	// Service-to-Service (v12) - uses instance and database
	if c.Instance != "" && c.Database != "" {
		scheme := "https"
		if !c.SSL {
			scheme = "http"
		}
		host := c.Address
		if host == "" {
			host = "localhost"
		}
		portStr := ""
		if c.Port > 0 {
			portStr = fmt.Sprintf(":%d", c.Port)
		}
		return fmt.Sprintf("%s://%s%s/%s/api/v1/Databases('%s')", scheme, host, portStr, c.Instance, c.Database), nil
	}

	// Traditional v11
	scheme := "https"
	if !c.SSL {
		scheme = "http"
	}

	host := c.Address
	port := ""
	if c.Port > 0 {
		port = fmt.Sprintf(":%d", c.Port)
	}

	return fmt.Sprintf("%s://%s%s/api/v1", scheme, host, port), nil
}

// SessionContextValue resolves the session context header to send with each request.
func (c Config) SessionContextValue() string {
	if strings.TrimSpace(c.SessionContext) != "" {
		return c.SessionContext
	}

	return defaultSessionContext
}

// HTTPHeaders merges default TM1 headers with caller overrides.
func (c Config) HTTPHeaders() http.Header {
	headers := http.Header{}
	headers.Set("Connection", "keep-alive")
	headers.Set("User-Agent", "tm1go")
	headers.Set("Content-Type", "application/json; odata.streaming=true; charset=utf-8")
	headers.Set("Accept", "application/json;odata.metadata=none,text/plain")
	headers.Set("TM1-SessionContext", c.SessionContextValue())

	for key, values := range c.DefaultHeaders {
		for _, value := range values {
			headers.Add(key, value)
		}
	}

	return headers
}

// HTTPClientOrDefault returns the caller-supplied HTTP client or a safe default.
func (c Config) HTTPClientOrDefault() (*http.Client, error) {
	if c.HTTPClient != nil {
		return c.HTTPClient, nil
	}

	// Determine SSL verification settings
	verify := c.determineVerify()

	tlsConfig := &tls.Config{
		InsecureSkipVerify: c.SkipSSLVerification || !verify, // #nosec G402: explicitly controlled via config
	}

	// Handle custom SSL context or certificate paths
	if c.SSLContext != nil {
		if tlsCtx, ok := c.SSLContext.(*tls.Config); ok {
			tlsConfig = tlsCtx
		}
	}

	// Handle client certificates
	if c.Cert != nil {
		if certPath, ok := c.Cert.(string); ok && certPath != "" {
			cert, err := tls.LoadX509KeyPair(certPath, certPath)
			if err != nil {
				return nil, fmt.Errorf("load client cert: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
	}

	// Set connection pool size
	maxIdleConns := c.ConnectionPoolSize
	if maxIdleConns <= 0 {
		maxIdleConns = 10 // default from TM1py
	}

	maxIdleConnsPerHost := c.PoolConnections
	if maxIdleConnsPerHost <= 0 {
		maxIdleConnsPerHost = 1 // default from TM1py
	}

	transport := &http.Transport{
		Proxy:               c.getProxyFunc(),
		TLSClientConfig:     tlsConfig,
		MaxIdleConns:        maxIdleConns,
		MaxIdleConnsPerHost: maxIdleConnsPerHost,
	}

	timeout := c.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	// Create cookie jar to store session cookies
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("create cookie jar: %w", err)
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
		Jar:       jar,
	}, nil
}

// determineVerify resolves SSL verification settings
func (c Config) determineVerify() bool {
	if c.Verify == nil {
		return false // default for v11
	}

	if b, ok := c.Verify.(bool); ok {
		return b
	}

	if s, ok := c.Verify.(string); ok {
		if strings.ToUpper(s) == "FALSE" {
			return false
		}
		if strings.ToUpper(s) == "TRUE" {
			return true
		}
		// If it's a path to a .cer file, assume verification is enabled
		return true
	}

	return false
}

// getProxyFunc returns the proxy function based on config
func (c Config) getProxyFunc() func(*http.Request) (*url.URL, error) {
	// If ProxyURL is set, use it
	if strings.TrimSpace(c.ProxyURL) != "" {
		proxyParsed, err := url.Parse(c.ProxyURL)
		if err == nil {
			return http.ProxyURL(proxyParsed)
		}
	}

	// If Proxies is set as a map, handle it
	if c.Proxies != nil {
		if proxiesMap, ok := c.Proxies.(map[string]string); ok {
			return func(req *http.Request) (*url.URL, error) {
				scheme := req.URL.Scheme
				if proxyURL, ok := proxiesMap[scheme]; ok {
					return url.Parse(proxyURL)
				}
				return http.ProxyFromEnvironment(req)
			}
		}
	}

	// Default to environment proxy
	return http.ProxyFromEnvironment
}
