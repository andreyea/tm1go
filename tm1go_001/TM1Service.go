package tm1go

import (
	"encoding/json"
	"os"
)

// TM1Service is a service for interacting with TM1
type TM1Service struct {
	restClient           *RestService
	ConfigurationService *ConfigurationService
	CubeService          *CubeService
	ObjectService        *ObjectService
	DimensionService     *DimensionService
	ElementService       *ElementService
	BatchService         *BatchService
	HierarchyService     *HierarchyService
	SubsetService        *SubsetService
	ProcessService       *ProcessService
	CellService          *CellService
	FileService          *FileService
	SandboxService       *SandboxService
	ViewService          *ViewService
}

// TM1ServiceConfig is a configuration for TM1Service
type TM1ServiceConfig struct {
	Address                   string            `json:"Address,omitempty"`                   // The address of the target server or service
	Port                      int               `json:"Port,omitempty"`                      // The HTTP port number for communication
	SSL                       bool              `json:"SS,omitempty"`                        // Indicates whether SSL/TLS encryption is enabled
	Instance                  string            `json:"Instance,omitempty"`                  // The name of the specific instance or system
	Database                  string            `json:"Database,omitempty"`                  // The name of the database or data source
	BaseURL                   string            `json:"BaseURL,omitempty"`                   // The base URL for API endpoints
	AuthURL                   string            `json:"AuthURL,omitempty"`                   // The URL for authentication services
	User                      string            `json:"User,omitempty"`                      // The username or user identifier
	Password                  string            `json:"Password,omitempty"`                  // The user's password
	DecodeB64                 bool              `json:"DecodeB64,omitempty"`                 // Specifies if the password is base64 encoded
	Namespace                 string            `json:"Namespace,omitempty"`                 // An optional namespace for access control
	CAMPassport               string            `json:"CAMPassport,omitempty"`               // The CAM (Cognos Access Manager) passport
	SessionID                 map[string]string `json:"SessionID,omitempty"`                 // The unique session identifier
	ApplicationClientID       string            `json:"ApplicationClientID,omitempty"`       // The client ID for application integration
	ApplicationClientSecret   string            `json:"ApplicationClientSecret,omitempty"`   // The client secret for application integration
	APIKey                    string            `json:"APIKey,omitempty"`                    // The API Key for authentication
	IAMURL                    string            `json:"IAMURL,omitempty"`                    // The IBM Cloud IAM (Identity and Access Management) URL
	PAURL                     string            `json:"PAURL,omitempty"`                     // The URL for the Planning Analytics Engine
	Tenant                    string            `json:"Tenant,omitempty"`                    // The tenant identifier
	SessionContext            string            `json:"SessionContext,omitempty"`            // The name of the application context
	Verify                    bool              `json:"Verify,omitempty"`                    //
	VerifyCertPath            string            `json:"VerifyCertPath,omitempty"`            // The path to a certificate file for verification
	Logging                   bool              `json:"Logging,omitempty"`                   // Specifies whether verbose HTTP logging is enabled
	Timeout                   float64           `json:"Timeout,omitempty"`                   // The maximum time to wait for the first byte in seconds
	CancelAtTimeout           bool              `json:"CancelAtTimeout,omitempty"`           // Indicates whether operations should be aborted on timeout
	AsyncRequestsMode         bool              `json:"AsyncRequestsMode,omitempty"`         // Enables a mode to avoid 60s timeouts on IBM Cloud
	TCPKeepAlive              bool              `json:"TCPKeepAlive,omitempty"`              // Maintains the TCP connection continuously
	ConnectionPoolSize        int               `json:"ConnectionPoolSize,omitempty"`        // Size of the connection pool in a multi-threaded environment
	IntegratedLogin           bool              `json:"IntegratedLogin,omitempty"`           // Enables IntegratedSecurityMode3
	IntegratedLoginDomain     string            `json:"IntegratedLoginDomain,omitempty"`     // The NT Domain name for integrated login
	IntegratedLoginService    string            `json:"IntegratedLoginService,omitempty"`    // The Kerberos Service type for remote Service Principal Name
	IntegratedLoginHost       string            `json:"IntegratedLoginHost,omitempty"`       // The host name for Service Principal Name
	IntegratedLoginDelegate   bool              `json:"IntegratedLoginDelegate,omitempty"`   // Indicates whether user credentials are delegated to the server
	Impersonate               string            `json:"Impersonate,omitempty"`               // The name of the user to impersonate
	ReconnectOnSessionTimeout bool              `json:"ReconnectOnSessionTimeout,omitempty"` // Attempts to reconnect once if the session times out
	Proxies                   map[string]string `json:"Proxies,omitempty"`                   // A dictionary of proxy settings
	Gateway                   string            `json:"Gateway,omitempty"`
	Headers                   map[string]string `json:"Headers,omitempty"`
	Session                   interface{}       `json:"Session,omitempty"`
}

// Save saves the TM1ServiceConfig to a file
func (c *TM1ServiceConfig) Save(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(c)
}

// Load loads the TM1ServiceConfig from a file
func (c *TM1ServiceConfig) Load(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(c)
}

// NewTM1Service creates a new TM1Service instance
func NewTM1Service(config TM1ServiceConfig) *TM1Service {

	var tm1Service = &TM1Service{
		restClient: NewRestClient(config),
	}
	tm1Service.restClient.connect()

	// Attach all services to the TM1Service
	tm1Service.ObjectService = NewObjectService(tm1Service.restClient)
	tm1Service.ProcessService = NewProcessService(tm1Service.restClient, tm1Service.ObjectService)
	tm1Service.ConfigurationService = NewConfigurationService(tm1Service.restClient)
	tm1Service.CubeService = NewCubeService(tm1Service.restClient, tm1Service.ObjectService, tm1Service.DimensionService, tm1Service.ProcessService)
	tm1Service.ElementService = NewElementService(tm1Service.restClient, tm1Service.ObjectService)
	tm1Service.BatchService = NewBatchService(tm1Service.restClient)
	tm1Service.HierarchyService = NewHierarchyService(tm1Service.restClient, tm1Service.ObjectService)
	tm1Service.DimensionService = NewDimensionService(tm1Service.restClient, tm1Service.ObjectService, tm1Service.HierarchyService)
	tm1Service.SubsetService = NewSubsetService(tm1Service.restClient)
	tm1Service.FileService = NewFileService(tm1Service.restClient, tm1Service.ObjectService)
	tm1Service.CellService = NewCellService(tm1Service.restClient, tm1Service.CubeService, tm1Service.FileService, tm1Service.ProcessService)
	tm1Service.SandboxService = NewSandboxService(tm1Service.restClient)
	tm1Service.ViewService = NewViewService(tm1Service.restClient, tm1Service.ObjectService)

	return tm1Service
}

// Logout logs out of the TM1 service
func (s *TM1Service) Logout() error {
	return s.restClient.logout()
}

// Connect connects to the TM1 service
func (s *TM1Service) Connect() bool {
	return s.restClient.connect()
}
