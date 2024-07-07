package tm1go

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Groups struct {
	Value []struct {
		Name string `json:"Name"`
	} `json:"value"`
}

// Nullable bool
type NullableBool int

const (
	TRUE NullableBool = iota
	FALSE
	NULL
)

// AuthenticationMode is an enum for the different authentication modes
type AuthenticationMode int

// Default HTTP headers
var Headers = map[string]string{
	"Connection":         "keep-alive",
	"User-Agent":         "TM1go",
	"Content-Type":       "application/json; odata.streaming=true; charset=utf-8",
	"Accept":             "application/json;odata.metadata=none,text/plain",
	"TM1-SessionContext": "TM1go",
}

const DefaultConnectionPoolSize = 10

// Authentication modes
const (
	BASIC AuthenticationMode = iota + 1
	WIA
	CAM
	CAM_SSO
	IBM_CLOUD_API_KEY
	SERVICE_TO_SERVICE
)

// TCP socket options
const (
	TCPTCPKeepIdle  = 30
	TCPTCPKeepIntvl = 15
	TCPTCPKeepCnt   = 60
)

// Determines whether the authentication mode uses v12 authentication
func (mode AuthenticationMode) UseV12Auth() bool {
	return mode >= IBM_CLOUD_API_KEY
}

// RestService is a struct that contains the configuration for a TM1 REST API connection
type RestService struct {
	Address                   string            // The address of the target server or service
	Port                      int               // The HTTP port number for communication
	SSL                       bool              // Indicates whether SSL/TLS encryption is enabled
	Instance                  string            // The name of the specific instance or system
	Database                  string            // The name of the database or data source
	BaseURL                   string            // The base URL for API endpoints
	AuthURL                   string            // The URL for authentication services
	User                      string            // The username or user identifier
	Password                  string            // The user's password
	DecodeB64                 bool              // Specifies if the password is base64 encoded
	Namespace                 string            // An optional namespace for access control
	CAMPassport               string            // The CAM (Cognos Access Manager) passport
	SessionID                 map[string]string // The unique session identifier
	ApplicationClientID       string            // The client ID for application integration
	ApplicationClientSecret   string            // The client secret for application integration
	APIKey                    string            // The API Key for authentication
	IAMURL                    string            // The IBM Cloud IAM (Identity and Access Management) URL
	PAURL                     string            // The URL for the Planning Analytics Engine
	Tenant                    string            // The tenant identifier
	SessionContext            string            // The name of the application context
	Verify                    bool              //
	Logging                   bool              // Specifies whether verbose HTTP logging is enabled
	Timeout                   float64           // The maximum time to wait for the first byte in seconds
	CancelAtTimeout           bool              // Indicates whether operations should be aborted on timeout
	AsyncRequestsMode         bool              // Enables a mode to avoid 60s timeouts on IBM Cloud
	TCPKeepAlive              bool              // Maintains the TCP connection continuously
	ConnectionPoolSize        int               // Size of the connection pool in a multi-threaded environment
	IntegratedLogin           bool              // Enables IntegratedSecurityMode3
	IntegratedLoginDomain     string            // The NT Domain name for integrated login
	IntegratedLoginService    string            // The Kerberos Service type for remote Service Principal Name
	IntegratedLoginHost       string            // The host name for Service Principal Name
	IntegratedLoginDelegate   bool              // Indicates whether user credentials are delegated to the server
	Impersonate               string            // The name of the user to impersonate
	ReconnectOnSessionTimeout bool              // Attempts to reconnect once if the session times out
	Proxies                   map[string]string // A dictionary of proxy settings
	MaxRetryAttempts          int               // The maximum number of times to retry a request
	Gateway                   string
	headers                   map[string]string
	httpClient                *http.Client
	AuthMode                  AuthenticationMode // The authentication mode for the connection
	VerifyCertPath            string             // The path to a certificate file for verification
	version                   string
	isOpsAdmin                NullableBool // Indicates whether the user is an administrator
	isDataAdmin               NullableBool // Indicates whether the user is a data administrator
}

// Constructor for RestService
func NewRestClient(config TM1ServiceConfig) *RestService {
	var authMode = determineAuthMode(config.AuthURL, config.Instance, config.Database, config.APIKey, config.IAMURL, config.PAURL, config.Tenant, config.Namespace, config.Gateway, config.IntegratedLogin)
	var headers = Headers
	if config.SessionContext != "" {
		headers["TM1-SessionContext"] = config.SessionContext
	}

	var rs = &RestService{
		BaseURL:                   config.BaseURL,
		AuthURL:                   config.AuthURL,
		Address:                   config.Address,
		Port:                      config.Port,
		SSL:                       config.SSL,
		Instance:                  config.Instance,
		Database:                  config.Database,
		User:                      config.User,
		Password:                  config.Password,
		DecodeB64:                 config.DecodeB64,
		Namespace:                 config.Namespace,
		CAMPassport:               config.CAMPassport,
		SessionID:                 config.SessionID,
		ApplicationClientID:       config.ApplicationClientID,
		ApplicationClientSecret:   config.ApplicationClientSecret,
		APIKey:                    config.APIKey,
		IAMURL:                    config.IAMURL,
		PAURL:                     config.PAURL,
		Tenant:                    config.Tenant,
		SessionContext:            config.SessionContext,
		Verify:                    determineVerify(config.Verify, authMode),
		VerifyCertPath:            config.VerifyCertPath,
		Logging:                   config.Logging,
		Timeout:                   config.Timeout,
		CancelAtTimeout:           config.CancelAtTimeout,
		AsyncRequestsMode:         config.AsyncRequestsMode,
		TCPKeepAlive:              determineTCPOption(config.TCPKeepAlive, false),
		ConnectionPoolSize:        setConnectionPoolSize(config.ConnectionPoolSize),
		IntegratedLogin:           config.IntegratedLogin,
		IntegratedLoginDomain:     config.IntegratedLoginDomain,
		IntegratedLoginService:    config.IntegratedLoginService,
		IntegratedLoginHost:       config.IntegratedLoginHost,
		IntegratedLoginDelegate:   config.IntegratedLoginDelegate,
		Impersonate:               config.Impersonate,
		ReconnectOnSessionTimeout: config.ReconnectOnSessionTimeout,
		Proxies:                   config.Proxies,
		Gateway:                   config.Gateway,
		MaxRetryAttempts:          1,
		headers:                   headers,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
				MaxIdleConns:        DefaultConnectionPoolSize,
				MaxIdleConnsPerHost: DefaultConnectionPoolSize,
			},
		},
		AuthMode: authMode,
	}
	baseUrl, authUrl, err := rs.constructServiceAndAuthRoot()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	rs.BaseURL = baseUrl
	rs.AuthURL = authUrl
	return rs
}

func (rs *RestService) generateIBMCloudAccessToken() (string, error) {
	if rs.APIKey == "" {
		return "", fmt.Errorf("'iamURL' and 'apiKey' must be provided to generate access token from IBM Cloud")
	}

	payload := fmt.Sprintf("grant_type=urn%%3Aibm%%3Aparams%%3Aoauth%%3Agrant-type%%3Aapikey&apikey=%s", rs.APIKey)
	req, err := http.NewRequest("POST", rs.IAMURL, bytes.NewBufferString(payload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to generate access token from URL: '%s'", rs.IAMURL)
	}

	var responseMap map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(body, &responseMap); err != nil {
		return "", err
	}

	accessToken, ok := responseMap["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("access token not found in the response")
	}

	return accessToken, nil

}

// Determine whether the user is an administrator
func (rs *RestService) determineIsAdmin(user string) (bool, error) {
	if user == "" {
		return false, nil
	}
	return strings.EqualFold(user, "ADMIN"), nil
}

// Add additional headers to the request
func (rs *RestService) addHeaders(additionalHeaders map[string]string) {
	for key, value := range additionalHeaders {
		rs.headers[key] = value
	}
}

// Determine the authentication mode based on the provided parameters
func (rs *RestService) determineAuthMode() AuthenticationMode {
	if rs.AuthURL == "" && rs.Instance == "" && rs.Database == "" && rs.APIKey == "" && rs.IAMURL == "" && rs.PAURL == "" && rs.Tenant == "" {
		// v11
		if rs.Namespace == "" && rs.Gateway == "" {
			return BASIC
		}

		if rs.Gateway != "" {
			return CAM_SSO
		}

		if rs.IntegratedLogin {
			return WIA
		}

		return CAM
	}

	// v12
	if rs.APIKey != "" {
		return IBM_CLOUD_API_KEY
	}

	return SERVICE_TO_SERVICE
}

// Construct the base URL and auth URL for v12
func (rs *RestService) constructAllVersionServiceAndAuthRootFromBaseURL() (string, string, error) {
	if rs.Address != "" {
		return "", "", fmt.Errorf("base URL and Address cannot be specified at the same time")
	}

	// v12 requires an auth URL be provided if a base URL is specified
	if strings.Contains(rs.BaseURL, "api/v1/Databases") {
		if rs.AuthURL == "" {
			return "", "", fmt.Errorf("AuthURL is missing when connecting to planning analytics engine and using the base_url; you must specify a corresponding auth url")
		}
	} else if strings.HasSuffix(rs.BaseURL, "/api/v1") {
		rs.AuthURL = rs.BaseURL + "/Configuration/ProductVersion/$value"
	} else {
		rs.BaseURL += "/api/v1"
		rs.AuthURL = rs.BaseURL + "/Configuration/ProductVersion/$value"
	}

	return rs.BaseURL, rs.AuthURL, nil
}

// Logout from TM1
func (rs *RestService) logout() error {
	url := "/ActiveSession/tm1.Close"
	var asyncRequest = false
	_, err := rs.POST(url, "", nil, 0, &asyncRequest)
	if err != nil {
		return err
	}
	return nil
}

// Connect to TM1
func (rs *RestService) connect() bool {
	if rs.SessionID != nil {
		// Set cookie
		if _, ok := rs.SessionID["TM1SessionId"]; ok {
			rs.headers["Cookie"] = "TM1SessionId=" + rs.SessionID["TM1SessionId"]
			return true
		} else if _, ok := rs.SessionID["paSession"]; ok {
			rs.headers["Cookie"] = "paSession=" + rs.SessionID["paSession"]
			return true
		}
		return false
	} else {
		// Attempt to start a session
		if err := rs.startSession(rs.User, rs.Password, rs.DecodeB64, rs.Namespace, rs.Gateway, rs.CAMPassport, rs.IntegratedLogin, rs.IntegratedLoginDomain, rs.IntegratedLoginService, rs.IntegratedLoginHost, rs.IntegratedLoginDelegate, rs.Impersonate, rs.ApplicationClientID, rs.ApplicationClientSecret, rs.AuthMode); err != nil {
			return false
		}
		return true
	}
}

func (rs *RestService) GET(url string, customHeaders map[string]string, timeout time.Duration, asyncRequest *bool) (*http.Response, error) {
	return rs.request("GET", url, "", customHeaders, timeout, rs.MaxRetryAttempts, asyncRequest)
}

func (rs *RestService) POST(url string, data string, customHeaders map[string]string, timeout time.Duration, asyncRequest *bool) (*http.Response, error) {
	return rs.request("POST", url, data, customHeaders, timeout, rs.MaxRetryAttempts, asyncRequest)
}

func (rs *RestService) PUT(url string, data string, customHeaders map[string]string, timeout time.Duration, asyncRequest *bool) (*http.Response, error) {
	return rs.request("PUT", url, data, customHeaders, timeout, rs.MaxRetryAttempts, asyncRequest)
}

func (rs *RestService) DELETE(url string, customHeaders map[string]string, timeout time.Duration, asyncRequest *bool) (*http.Response, error) {
	return rs.request("DELETE", url, "", customHeaders, timeout, rs.MaxRetryAttempts, asyncRequest)
}

func (rs *RestService) PATCH(url string, data string, customHeaders map[string]string, timeout time.Duration, asyncRequest *bool) (*http.Response, error) {
	return rs.request("PATCH", url, data, customHeaders, timeout, rs.MaxRetryAttempts, asyncRequest)
}

// Creates a new HTTP request
func (rs *RestService) request(method string, urlEndPoint string, data string, customHeaders map[string]string, timeout time.Duration, retryCount int, asyncRequest *bool) (*http.Response, error) {

	if retryCount <= 0 {
		return nil, fmt.Errorf("max retry attempts reached")
	}

	var fullURL = rs.BaseURL + urlEndPoint
	fullURL = strings.Replace(fullURL, " ", "%20", -1)
	var body = bytes.NewBufferString(data)
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, err
	}

	// Attach headers
	for key, value := range rs.headers {
		req.Header.Add(key, value)
	}

	// Attach custom headers
	for key, value := range customHeaders {
		req.Header.Add(key, value)
	}

	if timeout > 0 {
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
		defer cancel()
		req = req.WithContext(ctx)
	}

	rs.setAsyncRequestMode(&asyncRequest)
	if *asyncRequest {
		req.Header.Set("Prefer", "respond-async")
		// Send the request
		rsp, err := rs.httpClient.Do(req)
		if err != nil {
			return rsp, err
		}

		// Check if unauthorized
		if rsp.StatusCode == http.StatusUnauthorized {
			// Attempt to reconnect
			if !rs.connect() {
				return nil, fmt.Errorf("authentication failed")
			}

			// Use exponential backoff or a fixed wait time before retrying
			time.Sleep(time.Duration(2^(3-retryCount)) * time.Second)

			// Retry request with decremented retryCount
			return rs.request(method, urlEndPoint, data, customHeaders, timeout, retryCount-1, asyncRequest)
		}

		err = verifyResponse(rsp)
		if err != nil {
			return rsp, err
		}

		location := rsp.Header.Get("Location")
		if location == "" || !strings.Contains(location, "'") {
			return rsp, fmt.Errorf("failed to retrieve async_id from request %s '%s'", method, urlEndPoint)
		}
		asyncID := strings.Split(location, "'")[1]

		waitChannel := make(chan time.Duration)

		// Start the wait time generator in a goroutine.
		go waitTimeGenerator(int(timeout), waitChannel)

		// Use the wait times from the channel to attempt retrieving the async response.
		for waitTime := range waitChannel {
			rsp, err = rs.retrieveAsyncResponse(asyncID)
			if err != nil {
				break
			}
			if rsp.StatusCode == 200 || rsp.StatusCode == 201 {
				// Exit the loop on successful response.
				break
			}
			// Wait for the specified duration before the next attempt.
			time.Sleep(waitTime)
		}

		// All wait times consumed and still no 200
		if rsp.StatusCode != 200 && rsp.StatusCode != 201 {
			if timeout > 0 || rs.CancelAtTimeout {
				rs.cancelAsyncOperation(asyncID)
			}
			return rsp, fmt.Errorf("failed to retrieve async response for request %s '%s'", method, urlEndPoint)
		}

		bodyBytes, err := io.ReadAll(rsp.Body)

		if err != nil {
			return rsp, err
		}
		// Restore the body for future use.
		rsp.Body = io.NopCloser(io.Reader(bytes.NewBuffer(bodyBytes)))
		if bytes.HasPrefix(bodyBytes, []byte("HTTP/")) {
			return buildResponseFromBinaryData(bodyBytes), nil
		}
		asyncResult := rsp.Header.Get("asyncresult")
		if asyncResult != "" {
			statusCode, err := strconv.Atoi(asyncResult[:3])
			if err != nil {
				return rsp, fmt.Errorf("invalid asyncresult header value: %v", err)
			}
			rsp.StatusCode = statusCode
		}
		err = verifyResponse(rsp)
		if err != nil {
			return nil, err
		}
		return rsp, nil
	} else {
		// Send the request
		rsp, err := rs.httpClient.Do(req)
		if err != nil {
			return rsp, err
		}

		// Check if unauthorized
		if rsp.StatusCode == http.StatusUnauthorized {
			// Attempt to reconnect
			if !rs.connect() {
				return rsp, fmt.Errorf("authentication failed")
			}

			// Use exponential backoff or a fixed wait time before retrying
			time.Sleep(time.Duration(2^(3-retryCount)) * time.Second)

			// Retry request with decremented retryCount
			return rs.request(method, urlEndPoint, data, customHeaders, timeout, retryCount-1, asyncRequest)
		}

		err = verifyResponse(rsp)
		if err != nil {
			return rsp, err
		}
		return rsp, nil
	}
}

// Set the async request mode
func (rs *RestService) setAsyncRequestMode(asyncRequest **bool) {
	if *asyncRequest == nil {
		*asyncRequest = &rs.AsyncRequestsMode
	}
}

// Start a session with TM1
func (rs *RestService) startSession(
	user string,
	password string,
	decodeB64 bool,
	namespace string,
	gateway string,
	camPassport string,
	integratedLogin bool,
	integratedLoginDomain string,
	integratedLoginService string,
	integratedLoginHost string,
	integratedLoginDelegate bool,
	impersonate string,
	applicationClientID string,
	applicationClientSecret string,
	authMode AuthenticationMode) error {

	var token string
	req, err := http.NewRequest("GET", rs.AuthURL, nil)
	if err != nil {
		return err
	}
	// Authorization with integrated_login
	switch authMode {
	case WIA:
		// Handle integrated_login authorization
		// ...
	case SERVICE_TO_SERVICE:
		// Handle SERVICE_TO_SERVICE authorization
		// ...
	case IBM_CLOUD_API_KEY:
		// Handle IBM_CLOUD_API_KEY authorization
		if rs.IAMURL == "" && rs.APIKey != "" {
			token = buildAuthorizationTokenBasic("apikey", rs.APIKey)
			req.Header.Add("Authorization", token)
		} else {
			token, err = rs.generateIBMCloudAccessToken()
			req.Header.Add("Authorization", "Bearer "+token)
		}

		if err != nil {
			return err
		}

	default:
		// Handle other authorization modes (BASIC, CAM)

		if err != nil {
			return err
		}

		var token, _ = buildAuthorizationToken(rs.User, rs.Password, rs.Namespace, rs.Gateway, rs.CAMPassport, false)
		req.Header.Add("Authorization", token)

	}

	for key, value := range rs.headers {
		req.Header.Add(key, value)
	}

	rsp, err := rs.httpClient.Do(req)

	if err != nil {
		return err
	}

	defer rsp.Body.Close()
	// Extract product version from the response
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return err
	}
	rs.version = string(body)

	var sessionName, sessionID = extractSetCookieHeader(rsp.Header)
	rs.SessionID = make(map[string]string)
	rs.SessionID[sessionName] = sessionID

	// Loop rs.SessionID and set the cookie
	for key, value := range rs.SessionID {
		rs.headers["Cookie"] = key + "=" + value
	}

	// Process additional headers
	if impersonate != "" {
		if authMode.UseV12Auth() {
			return fmt.Errorf("User Impersonation is not supported in TM1 v12")
		} else {
			// Add the TM1-Impersonate header
			rs.headers["TM1-Impersonate"] = impersonate
		}
	}

	// After obtaining the session cookie, drop the Authorization Header
	delete(rs.headers, "Authorization")

	return nil
}

func (rs *RestService) generateIBMIAMCloudAccessToken() (string, error) {
	if rs.IAMURL == "" || rs.APIKey == "" {
		return "", fmt.Errorf("'IAMURL' and 'APIKey' must be provided to generate access token from IBM Cloud")
	}

	data := url.Values{}
	data.Set("grant_type", "urn:ibm:params:oauth:grant-type:apikey")
	data.Set("apikey", rs.APIKey)

	req, err := http.NewRequest("POST", rs.IAMURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("failed to generate access_token from URL: '%s'", rs.IAMURL)
	}

	return accessToken, nil
}

func (rs *RestService) constructIBMCloudServiceAndAuthRoot() (string, string, error) {
	if rs.Address == "" || rs.Tenant == "" || rs.Database == "" {
		return "", "", fmt.Errorf("'address', 'tenant' and 'database' must be provided to connect to TM1 > v12 in IBM Cloud")
	}
	if !rs.SSL {
		return "", "", fmt.Errorf("'ssl' must be True to connect to TM1 > v12 in IBM Cloud")
	}

	baseURL := fmt.Sprintf("https://%s/api/%s/v0/tm1/%s", rs.Address, rs.Tenant, rs.Database)

	authURL := fmt.Sprintf("%s/Configuration/ProductVersion/$value", baseURL)

	return baseURL, authURL, nil
}

func (rs *RestService) constructS2SServiceAndAuthRoot() (string, string, error) {
	if rs.Instance == "" || rs.Database == "" {
		return "", "", fmt.Errorf("'Instance' and 'Database' arguments are required for v12 authentication with 'address'")
	}

	address := rs.Address
	if address == "" {
		address = "localhost"
	}
	portPart := ""
	if rs.Port != 0 {
		portPart = fmt.Sprintf(":%s", strconv.Itoa(rs.Port))
	}
	sslPart := ""
	if rs.SSL {
		sslPart = "s"
	}

	baseURL := fmt.Sprintf("http%s://%s%s/%s/api/v1/Databases('%s')", sslPart, address, portPart, rs.Instance, rs.Database)
	authURL := fmt.Sprintf("http%s://%s%s/%s/auth/v1/session", sslPart, address, portPart, rs.Instance)

	return baseURL, authURL, nil
}

// Construct the base URL and auth URL for v11
func (rs *RestService) constructV11ServiceAndAuthRoot() (string, string, error) {
	ssl := ""
	if rs.SSL {
		ssl = "s"
	}

	address := "localhost"
	if len(rs.Address) > 0 {
		address = rs.Address
	}

	baseURL := fmt.Sprintf("http%s://%s:%d/api/v1", ssl, address, rs.Port)
	authURL := baseURL + "/Configuration/ProductVersion/$value"

	return baseURL, authURL, nil
}

// Construct the base URL and auth URL for the connection
func (rs *RestService) constructServiceAndAuthRoot() (string, string, error) {
	switch rs.AuthMode {
	case IBM_CLOUD_API_KEY:
		return rs.constructIBMCloudServiceAndAuthRoot()
	case SERVICE_TO_SERVICE:
		return rs.constructS2SServiceAndAuthRoot()
	case BASIC, WIA, CAM, CAM_SSO:
		if rs.BaseURL == "" {
			return rs.constructV11ServiceAndAuthRoot()
		}
		return rs.constructAllVersionServiceAndAuthRootFromBaseURL()
	default:
		return "", "", fmt.Errorf("unsupported authentication mode")
	}
}

// Retrieve the async response
func (rs *RestService) retrieveAsyncResponse(asyncID string) (*http.Response, error) {
	url := "/_async('" + asyncID + "')"
	asyncMode := false
	return rs.GET(url, nil, 0, &asyncMode)
}

// Cancel an async operation
func (rs *RestService) cancelAsyncOperation(asyncID string) (*http.Response, error) {
	url := rs.BaseURL + "/_async('" + asyncID + "')"
	asyncMode := false
	return rs.DELETE(url, nil, 0, &asyncMode)
}

// isDataAdmin checks if the active user is a data admin.
func (rs *RestService) IsDataAdmin() bool {
	if strings.ToLower(rs.User) == "admin" {
		return true
	}

	if rs.isDataAdmin == NULL {
		url := "/ActiveUser/Groups"
		rsp, err := rs.GET(url, nil, 0, nil)
		if err != nil {
			return false
		}

		var groupsResp Groups
		defer rsp.Body.Close()
		if err := json.NewDecoder(rsp.Body).Decode(&groupsResp); err != nil {
			return false
		}

		// Check if user is in Admin or OperationsAdmin groups
		isAdmin := false
		for _, group := range groupsResp.Value {
			normalizedGroupName := strings.ToLower(strings.ReplaceAll(group.Name, " ", ""))
			if normalizedGroupName == "admin" || normalizedGroupName == "dataadmin" {
				isAdmin = true
				break
			}
		}
		if isAdmin {
			rs.isOpsAdmin = TRUE
		} else {
			rs.isOpsAdmin = FALSE
		}
		return isAdmin
	} else {
		if rs.isOpsAdmin == TRUE {
			return true
		} else {
			return false
		}
	}
}

// IsOpsAdmin checks if the active user is an operations admin.
func (rs *RestService) IsOpsAdmin() bool {
	if strings.ToLower(rs.User) == "admin" {
		return true
	}

	if rs.isOpsAdmin == NULL {
		url := "/ActiveUser/Groups"
		rsp, err := rs.GET(url, nil, 0, nil)
		if err != nil {
			return false
		}

		var groupsResp Groups
		defer rsp.Body.Close()
		if err := json.NewDecoder(rsp.Body).Decode(&groupsResp); err != nil {
			return false
		}

		// Check if user is in Admin or OperationsAdmin groups
		isAdmin := false
		for _, group := range groupsResp.Value {
			normalizedGroupName := strings.ToLower(strings.ReplaceAll(group.Name, " ", ""))
			if normalizedGroupName == "admin" || normalizedGroupName == "operationsadmin" {
				isAdmin = true
				break
			}
		}
		if isAdmin {
			rs.isOpsAdmin = TRUE
		} else {
			rs.isOpsAdmin = FALSE
		}
		return isAdmin
	} else {
		if rs.isOpsAdmin == TRUE {
			return true
		} else {
			return false
		}
	}
}

// Extracts the TM1SessionId from the Set-Cookie header
func extractTM1SessionIDFromSetCookieHeader(authResponseHeaders http.Header) string {
	setCookieHeader := authResponseHeaders.Get("Set-Cookie")
	if setCookieHeader != "" {
		cookies := strings.Split(setCookieHeader, ";")
		for _, cookie := range cookies {
			parts := strings.SplitN(strings.TrimSpace(cookie), "=", 2)
			if len(parts) == 2 && parts[0] == "TM1SessionId" {
				return parts[1]
			}
		}
	}
	return ""
}

// Extracts the TM1SessionId from the Set-Cookie header
func extractSetCookieHeader(authResponseHeaders http.Header) (string, string) {
	setCookieHeader := authResponseHeaders.Get("Set-Cookie")
	if setCookieHeader != "" {
		cookies := strings.Split(setCookieHeader, ";")
		for _, cookie := range cookies {
			parts := strings.SplitN(strings.TrimSpace(cookie), "=", 2)
			if len(parts) == 2 {
				return parts[0], parts[1]
			}
		}
	}
	return "", ""
}

// Build the authorization token
func buildAuthorizationToken(user, password, namespace, gateway, camPassport string, verify bool) (string, error) {
	if camPassport != "" {
		return "CAMPassport " + camPassport, nil
	} else if namespace != "" {
		return buildAuthorizationTokenCam(user, password, namespace, gateway, verify)
	} else {
		return buildAuthorizationTokenBasic(user, password), nil
	}
}

// Build the authorization token for CAM authentication
func buildAuthorizationTokenCam(user, password, namespace, gateway string, verify bool) (string, error) {
	if gateway != "" {
		// not done yet
		return "", fmt.Errorf("CAM SSO is not supported yet")
	}

	// Build the CAM token for Basic authentication
	authStr := fmt.Sprintf("%s:%s:%s", user, password, namespace)
	authToken := "CAMNamespace " + base64.StdEncoding.EncodeToString([]byte(authStr))
	return authToken, nil
}

// Build the authorization token for basic authentication
func buildAuthorizationTokenBasic(user, password string) string {
	authStr := user + ":" + password
	authToken := "Basic " + base64.StdEncoding.EncodeToString([]byte(authStr))
	return authToken
}

// Determine whether SSL is enabled based on the base URL
func determineSSLBasedOnBaseURL(baseURL string) (bool, error) {
	if strings.HasPrefix(baseURL, "https://") {
		return true, nil
	} else if strings.HasPrefix(baseURL, "http://") {
		return false, nil
	} else {
		return false, fmt.Errorf("Invalid base_url: '%s'", baseURL)
	}
}

// Determine the TCP socket option based on the provided parameters
func determineTCPOption(tcpKeepalive bool, asyncRequestsMode bool) bool {
	if asyncRequestsMode != true {
		return tcpKeepalive
	}
	return false
}

// Set the connection pool size based on the provided parameters
func setConnectionPoolSize(connectionPoolSize int) int {
	if connectionPoolSize > 0 {
		return connectionPoolSize
	}
	return DefaultConnectionPoolSize
}

// Determine whether to verify option based on the provided parameters
func determineVerify(verify bool, authMode AuthenticationMode) bool {
	if authMode == IBM_CLOUD_API_KEY || authMode == SERVICE_TO_SERVICE {
		return true
	}
	return verify
}

// Determine the authentication mode based on the provided parameters
func determineAuthMode(authURL, instance, database, apiKey, iamURL, paURL, tenant, namespace, gateway string, integratedLogin bool) AuthenticationMode {
	if authURL == "" && instance == "" && database == "" && apiKey == "" && iamURL == "" && paURL == "" && tenant == "" {
		// v11
		if namespace == "" && gateway == "" {
			return BASIC
		}

		if gateway != "" {
			return CAM_SSO
		}

		if integratedLogin {
			return WIA
		}

		return CAM
	}

	// v12
	if apiKey != "" && tenant != "" {
		return IBM_CLOUD_API_KEY
	}

	return SERVICE_TO_SERVICE
}

func verifyResponse(response *http.Response) error {
	if response.StatusCode >= 400 {
		defer response.Body.Close()
		result := ErrorResponse{}
		err := json.NewDecoder(response.Body).Decode(&result)
		if err != nil {
			return fmt.Errorf("response status code (%d) indicates failure", response.StatusCode)
		}
		return fmt.Errorf("response status code (%d) indicates failure: %s", response.StatusCode, result.Error.Message)
	}
	return nil
}

// addCompactJSONHeader modifies the 'Accept' header of an http.Request to include 'tm1.compact=v0'.
// It returns the original 'Accept' header value.
func addCompactJSONHeader(req *http.Request) string {
	// Get the original 'Accept' header value.
	originalHeader := req.Header.Get("Accept")

	// Split the original header value by ';' to manipulate its parts.
	parts := strings.Split(originalHeader, ";")

	// Insert 'tm1.compact=v0' after 'application/json', assuming it is always the first part.
	// You may need to adjust this logic based on your actual header structure.
	if len(parts) > 0 && strings.TrimSpace(parts[0]) == "application/json" {
		parts = append(parts[:1], append([]string{"tm1.compact=v0"}, parts[1:]...)...)
	} else {
		// If 'application/json' is not the first part, or the header is not set,
		// simply prepend the 'tm1.compact=v0' directive.
		parts = append([]string{"tm1.compact=v0"}, parts...)
	}

	// Join the parts back together and update the 'Accept' header.
	modifiedHeader := strings.Join(parts, ";")
	req.Header.Set("Accept", modifiedHeader)

	// Return the original header value.
	return originalHeader
}

func waitTimeGenerator(timeout int, ch chan<- time.Duration) {
	// Ensure the channel is closed when finished.
	defer close(ch)

	// Initial wait times.
	ch <- 100 * time.Millisecond
	ch <- 300 * time.Millisecond
	ch <- 600 * time.Millisecond

	if timeout > 0 {
		for i := 1; i < timeout; i++ {
			ch <- 1 * time.Second
		}
	} else {
		for {
			ch <- 1 * time.Second
		}
	}
}

// buildResponseFromBinaryData creates a http.Response object from binary data.
func buildResponseFromBinaryData(data []byte) *http.Response {
	// Create a buffer from the binary data to use as the response body.
	body := io.NopCloser(io.Reader(bytes.NewBuffer(data)))

	// Construct a minimal http.Response object.
	response := &http.Response{
		StatusCode:    http.StatusOK,
		Body:          body,
		ContentLength: int64(len(data)),
		Header:        make(http.Header),
	}

	return response
}
