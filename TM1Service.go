package tm1go

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
	VerifyCertPath            string            // The path to a certificate file for verification
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
	Gateway                   string
	Headers                   map[string]string
	Session                   interface{}
	//httpClient                *http.Client
}

// NewTM1Service creates a new TM1Service instance
func NewTM1Service(config TM1ServiceConfig) *TM1Service {

	var tm1Service = &TM1Service{
		restClient: NewRestClient(config),
	}
	tm1Service.restClient.connect()

	// Attach all services to the TM1Service
	tm1Service.ObjectService = NewObjectService(tm1Service.restClient)
	tm1Service.ConfigurationService = NewConfigurationService(tm1Service.restClient)
	tm1Service.CubeService = NewCubeService(tm1Service.restClient, tm1Service.ObjectService, tm1Service.DimensionService)
	tm1Service.ElementService = NewElementService(tm1Service.restClient, tm1Service.ObjectService)
	tm1Service.BatchService = NewBatchService(tm1Service.restClient)
	tm1Service.HierarchyService = NewHierarchyService(tm1Service.restClient, tm1Service.ObjectService)
	tm1Service.DimensionService = NewDimensionService(tm1Service.restClient, tm1Service.ObjectService, tm1Service.HierarchyService)
	tm1Service.SubsetService = NewSubsetService(tm1Service.restClient)
	tm1Service.ProcessService = NewProcessService(tm1Service.restClient, tm1Service.ObjectService)
	tm1Service.CellService = NewCellService(tm1Service.restClient, tm1Service.CubeService)
	tm1Service.FileService = NewFileService(tm1Service.restClient, tm1Service.ObjectService)
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
