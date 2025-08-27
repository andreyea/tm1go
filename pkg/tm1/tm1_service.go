package tm1

import (
	"encoding/json"
	"fmt"
	"os"
)

// TM1Service provides access to all TM1 functionality
type TM1Service struct {
	// Core REST client
	rest Client

	// Service instances (lazy-loaded)
	annotations     AnnotationService
	cells           CellService
	chores          ChoreService
	cubes           CubeService
	dimensions      DimensionService
	elements        ElementService
	files           FileService
	hierarchies     HierarchyService
	processes       ProcessService
	security        SecurityService
	subsets         SubsetService
	applications    ApplicationService
	views           ViewService
	sandboxes       SandboxService
	jobs            JobService
	users           UserService
	threads         ThreadService
	sessions        SessionService
	configuration   ConfigurationService
	auditLogs       AuditLogService
	transactionLogs TransactionLogService
	messageLogs     MessageLogService

	// Higher level modules
	powerBI    PowerBIService
	loggers    LoggerService
	server     ServerService
	monitoring MonitoringService
	git        GitService
}

// NewTM1Service creates a new TM1Service instance
func NewTM1Service(config *Config) (*TM1Service, error) {
	restService, err := NewRestService(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST service: %w", err)
	}

	// Connect to TM1
	if err := restService.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to TM1: %w", err)
	}

	return &TM1Service{
		rest: restService,
	}, nil
}

// NewTM1ServiceFromConfig creates a new TM1Service from a config file
func NewTM1ServiceFromConfig(configPath string) (*TM1Service, error) {
	config, err := LoadConfigFromFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return NewTM1Service(config)
}

// LoadConfigFromFile loads configuration from a JSON file
func LoadConfigFromFile(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	return &config, nil
}

// SaveConfigToFile saves configuration to a JSON file
func SaveConfigToFile(config *Config, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}

// Close disconnects from TM1 server and cleans up resources
func (tm1 *TM1Service) Close() error {
	if tm1.rest != nil {
		return tm1.rest.Disconnect()
	}
	return nil
}

// Connection returns the underlying REST client
func (tm1 *TM1Service) Connection() Client {
	return tm1.rest
}

// Version returns the TM1 server version
func (tm1 *TM1Service) Version() string {
	return tm1.rest.Version()
}

// Config returns the connection configuration
func (tm1 *TM1Service) Config() *Config {
	return tm1.rest.Config()
}

// IsConnected returns the connection status
func (tm1 *TM1Service) IsConnected() bool {
	return tm1.rest.IsConnected()
}

// Reconnect attempts to reconnect to the TM1 server
func (tm1 *TM1Service) Reconnect() error {
	return tm1.rest.Connect()
}

// Service accessor methods (implement lazy loading pattern)

// Annotations returns the annotation service
func (tm1 *TM1Service) Annotations() AnnotationService {
	if tm1.annotations == nil {
		tm1.annotations = NewAnnotationService(tm1.rest)
	}
	return tm1.annotations
}

// Cells returns the cell service
func (tm1 *TM1Service) Cells() CellService {
	if tm1.cells == nil {
		tm1.cells = NewCellService(tm1.rest)
	}
	return tm1.cells
}

// Chores returns the chore service
func (tm1 *TM1Service) Chores() ChoreService {
	if tm1.chores == nil {
		tm1.chores = NewChoreService(tm1.rest)
	}
	return tm1.chores
}

// Cubes returns the cube service
func (tm1 *TM1Service) Cubes() CubeService {
	if tm1.cubes == nil {
		tm1.cubes = NewCubeService(tm1.rest)
	}
	return tm1.cubes
}

// Dimensions returns the dimension service
func (tm1 *TM1Service) Dimensions() DimensionService {
	if tm1.dimensions == nil {
		tm1.dimensions = NewDimensionService(tm1.rest)
	}
	return tm1.dimensions
}

// Elements returns the element service
func (tm1 *TM1Service) Elements() ElementService {
	if tm1.elements == nil {
		tm1.elements = NewElementService(tm1.rest)
	}
	return tm1.elements
}

// Files returns the file service
func (tm1 *TM1Service) Files() FileService {
	if tm1.files == nil {
		tm1.files = NewFileService(tm1.rest)
	}
	return tm1.files
}

// Hierarchies returns the hierarchy service
func (tm1 *TM1Service) Hierarchies() HierarchyService {
	if tm1.hierarchies == nil {
		tm1.hierarchies = NewHierarchyService(tm1.rest)
	}
	return tm1.hierarchies
}

// Processes returns the process service
func (tm1 *TM1Service) Processes() ProcessService {
	if tm1.processes == nil {
		tm1.processes = NewProcessService(tm1.rest)
	}
	return tm1.processes
}

// Security returns the security service
func (tm1 *TM1Service) Security() SecurityService {
	if tm1.security == nil {
		tm1.security = NewSecurityService(tm1.rest)
	}
	return tm1.security
}

// Subsets returns the subset service
func (tm1 *TM1Service) Subsets() SubsetService {
	if tm1.subsets == nil {
		tm1.subsets = NewSubsetService(tm1.rest)
	}
	return tm1.subsets
}

// Applications returns the application service
func (tm1 *TM1Service) Applications() ApplicationService {
	if tm1.applications == nil {
		tm1.applications = NewApplicationService(tm1.rest)
	}
	return tm1.applications
}

// Views returns the view service
func (tm1 *TM1Service) Views() ViewService {
	if tm1.views == nil {
		tm1.views = NewViewService(tm1.rest)
	}
	return tm1.views
}

// Sandboxes returns the sandbox service
func (tm1 *TM1Service) Sandboxes() SandboxService {
	if tm1.sandboxes == nil {
		tm1.sandboxes = NewSandboxService(tm1.rest)
	}
	return tm1.sandboxes
}

// Jobs returns the job service
func (tm1 *TM1Service) Jobs() JobService {
	if tm1.jobs == nil {
		tm1.jobs = NewJobService(tm1.rest)
	}
	return tm1.jobs
}

// Users returns the user service
func (tm1 *TM1Service) Users() UserService {
	if tm1.users == nil {
		tm1.users = NewUserService(tm1.rest)
	}
	return tm1.users
}

// Threads returns the thread service
func (tm1 *TM1Service) Threads() ThreadService {
	if tm1.threads == nil {
		tm1.threads = NewThreadService(tm1.rest)
	}
	return tm1.threads
}

// Sessions returns the session service
func (tm1 *TM1Service) Sessions() SessionService {
	if tm1.sessions == nil {
		tm1.sessions = NewSessionService(tm1.rest)
	}
	return tm1.sessions
}

// Configuration returns the configuration service
func (tm1 *TM1Service) Configuration() ConfigurationService {
	if tm1.configuration == nil {
		tm1.configuration = NewConfigurationService(tm1.rest)
	}
	return tm1.configuration
}

// AuditLogs returns the audit log service
func (tm1 *TM1Service) AuditLogs() AuditLogService {
	if tm1.auditLogs == nil {
		tm1.auditLogs = NewAuditLogService(tm1.rest)
	}
	return tm1.auditLogs
}

// TransactionLogs returns the transaction log service
func (tm1 *TM1Service) TransactionLogs() TransactionLogService {
	if tm1.transactionLogs == nil {
		tm1.transactionLogs = NewTransactionLogService(tm1.rest)
	}
	return tm1.transactionLogs
}

// MessageLogs returns the message log service
func (tm1 *TM1Service) MessageLogs() MessageLogService {
	if tm1.messageLogs == nil {
		tm1.messageLogs = NewMessageLogService(tm1.rest)
	}
	return tm1.messageLogs
}

// PowerBI returns the Power BI service
func (tm1 *TM1Service) PowerBI() PowerBIService {
	if tm1.powerBI == nil {
		tm1.powerBI = NewPowerBIService(tm1.rest)
	}
	return tm1.powerBI
}

// Loggers returns the logger service
func (tm1 *TM1Service) Loggers() LoggerService {
	if tm1.loggers == nil {
		tm1.loggers = NewLoggerService(tm1.rest)
	}
	return tm1.loggers
}

// Server returns the server service
func (tm1 *TM1Service) Server() ServerService {
	if tm1.server == nil {
		tm1.server = NewServerService(tm1.rest)
	}
	return tm1.server
}

// Monitoring returns the monitoring service
func (tm1 *TM1Service) Monitoring() MonitoringService {
	if tm1.monitoring == nil {
		tm1.monitoring = NewMonitoringService(tm1.rest)
	}
	return tm1.monitoring
}

// Git returns the git service
func (tm1 *TM1Service) Git() GitService {
	if tm1.git == nil {
		tm1.git = NewGitService(tm1.rest)
	}
	return tm1.git
}

// Whoami returns the current user
func (tm1 *TM1Service) Whoami() (string, error) {
	// This would call the security service to get current user
	// Implementation would depend on the SecurityService
	return "", fmt.Errorf("not implemented")
}

// Metadata returns API metadata
func (tm1 *TM1Service) Metadata() (map[string]interface{}, error) {
	resp, err := tm1.rest.GET("/$metadata", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(resp.Body, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return metadata, nil
}
