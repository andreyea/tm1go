// Package tm1 provides a Go client library for IBM Planning Analytics (TM1) REST API.
package tm1

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// AuthenticationMode represents different authentication methods for TM1
type AuthenticationMode int

const (
	Basic AuthenticationMode = iota + 1
	WIA
	CAM
	CAMSso
	// 5 is legacy early-release of v12. Deprecate with next major release
	IBMCloudAPIKey
	ServiceToService
	PAProxy
	BasicAPIKey
	AccessToken
)

// UseV12Auth returns true if the authentication mode uses v12 authentication
func (mode AuthenticationMode) UseV12Auth() bool {
	return mode >= IBMCloudAPIKey
}

// Config represents the configuration for TM1 connection
type Config struct {
	// Core connection parameters
	Address  string `json:"address,omitempty"`
	Port     int    `json:"port,omitempty"`
	SSL      bool   `json:"ssl"`
	Instance string `json:"instance,omitempty"`
	Database string `json:"database,omitempty"`
	BaseURL  string `json:"base_url,omitempty"`
	AuthURL  string `json:"auth_url,omitempty"`

	// Authentication parameters
	User                    string `json:"user,omitempty"`
	Password                string `json:"password,omitempty"`
	DecodeB64               bool   `json:"decode_b64,omitempty"`
	Namespace               string `json:"namespace,omitempty"`
	CAMPassport             string `json:"cam_passport,omitempty"`
	SessionID               string `json:"session_id,omitempty"`
	ApplicationClientID     string `json:"application_client_id,omitempty"`
	ApplicationClientSecret string `json:"application_client_secret,omitempty"`
	APIKey                  string `json:"api_key,omitempty"`
	IAMURL                  string `json:"iam_url,omitempty"`
	PAURL                   string `json:"pa_url,omitempty"`
	CPDURL                  string `json:"cpd_url,omitempty"`
	Tenant                  string `json:"tenant,omitempty"`

	// Connection options
	SessionContext            string            `json:"session_context,omitempty"`
	Verify                    interface{}       `json:"verify,omitempty"` // can be bool or string (path to cert)
	Logging                   bool              `json:"logging,omitempty"`
	Timeout                   *time.Duration    `json:"timeout,omitempty"`
	CancelAtTimeout           bool              `json:"cancel_at_timeout,omitempty"`
	AsyncRequestsMode         bool              `json:"async_requests_mode,omitempty"`
	ConnectionPoolSize        int               `json:"connection_pool_size,omitempty"`
	IntegratedLogin           bool              `json:"integrated_login,omitempty"`
	IntegratedLoginDomain     string            `json:"integrated_login_domain,omitempty"`
	IntegratedLoginService    string            `json:"integrated_login_service,omitempty"`
	IntegratedLoginHost       string            `json:"integrated_login_host,omitempty"`
	IntegratedLoginDelegate   bool              `json:"integrated_login_delegate,omitempty"`
	Impersonate               string            `json:"impersonate,omitempty"`
	ReconnectOnSessionTimeout bool              `json:"re_connect_on_session_timeout"`
	Proxies                   map[string]string `json:"proxies,omitempty"`
	SSLContext                interface{}       `json:"ssl_context,omitempty"`
	Cert                      interface{}       `json:"cert,omitempty"` // can be string or tuple
}

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	return &Config{
		SSL:                       true,
		ConnectionPoolSize:        10,
		ReconnectOnSessionTimeout: true,
		SessionContext:            "TM1go",
		IntegratedLoginDomain:     ".",
		IntegratedLoginService:    "HTTP",
		IAMURL:                    "https://iam.cloud.ibm.com",
	}
}

// HTTPHeaders represents the default HTTP headers for TM1 requests
var HTTPHeaders = map[string]string{
	"Connection":         "keep-alive",
	"User-Agent":         "TM1go",
	"Content-Type":       "application/json; odata.streaming=true; charset=utf-8",
	"Accept":             "application/json;odata.metadata=none,text/plain",
	"TM1-SessionContext": "TM1go",
}

// RequestOptions represents options for HTTP requests
type RequestOptions struct {
	Headers         map[string]string
	AsyncMode       *bool
	ReturnAsyncID   bool
	Timeout         *time.Duration
	CancelAtTimeout bool
	Encoding        string
	Context         context.Context
}

// Response represents a generic HTTP response
type Response struct {
	StatusCode int
	Body       []byte
	Headers    map[string][]string
}

// AsyncResponse represents an async operation response
type AsyncResponse struct {
	ID       string
	Status   string
	Location string
}

// Client interface defines the contract for TM1 client operations
type Client interface {
	// HTTP operations
	GET(url string, opts *RequestOptions) (*Response, error)
	POST(url string, data []byte, opts *RequestOptions) (*Response, error)
	PATCH(url string, data []byte, opts *RequestOptions) (*Response, error)
	PUT(url string, data []byte, opts *RequestOptions) (*Response, error)
	DELETE(url string, opts *RequestOptions) (*Response, error)

	// Connection management
	Connect() error
	Disconnect() error
	IsConnected() bool

	// Configuration
	Config() *Config
	Version() string
}

// Element represents a TM1 element
type Element struct {
	Name       string                 `json:"Name"`
	Type       ElementType            `json:"Type"`
	Level      int                    `json:"Level"`
	Index      int                    `json:"Index"`
	Attributes map[string]interface{} `json:"Attributes,omitempty"`
}

// ElementType represents the type of an element with flexible JSON parsing
type ElementType int

// UnmarshalJSON implements custom JSON unmarshaling for ElementType
func (et *ElementType) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as integer first
	var intVal int
	if err := json.Unmarshal(data, &intVal); err == nil {
		*et = ElementType(intVal)
		return nil
	}

	// Try to unmarshal as string
	var strVal string
	if err := json.Unmarshal(data, &strVal); err == nil {
		switch strings.ToLower(strVal) {
		case "numeric", "1":
			*et = ElementTypeNumeric
		case "string", "2":
			*et = ElementTypeString
		case "consolidated", "3":
			*et = ElementTypeConsolidated
		default:
			return fmt.Errorf("unknown element type: %s", strVal)
		}
		return nil
	}

	return fmt.Errorf("cannot unmarshal element type from %s", string(data))
}

// ElementType constants
const (
	ElementTypeNumeric      ElementType = 1
	ElementTypeString       ElementType = 2
	ElementTypeConsolidated ElementType = 3
)

// ElementAttribute represents an element attribute
type ElementAttribute struct {
	Name string `json:"Name"`
	Type string `json:"Type"`
}

// Edge represents a hierarchy edge (parent-child relationship)
type Edge struct {
	ParentName    string  `json:"ParentName"`
	ComponentName string  `json:"ComponentName"`
	Weight        float64 `json:"Weight"`
}

// ValueArray represents a TM1 OData value array response
type ValueArray[T any] struct {
	Value []T `json:"value"`
}

// MDXExecuteParams represents parameters for MDX execution
type MDXExecuteParams struct {
	MDX               string
	TopRecords        *int
	MemberProperties  []string
	ParentProperties  []string
	ElementProperties []string
}

// CellsetAxis represents a cellset axis from MDX execution
type CellsetAxis struct {
	Tuples []struct {
		Members []struct {
			Name       string                 `json:"Name"`
			UniqueName string                 `json:"UniqueName"`
			Attributes map[string]interface{} `json:"Attributes,omitempty"`
			Parent     *struct {
				Name       string                 `json:"Name"`
				Attributes map[string]interface{} `json:"Attributes,omitempty"`
			} `json:"Parent,omitempty"`
			Element *struct {
				Name       string                 `json:"Name"`
				Type       int                    `json:"Type"`
				Attributes map[string]interface{} `json:"Attributes,omitempty"`
			} `json:"Element,omitempty"`
		} `json:"Members"`
	} `json:"Tuples"`
}

// Level represents a hierarchy level
type Level struct {
	Name string `json:"Name"`
}
