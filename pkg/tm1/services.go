package tm1

// Service interfaces define the contracts for various TM1 services

// AnnotationService provides operations for TM1 annotations
type AnnotationService interface {
	// Add annotation service methods here
}

// CellService provides operations for TM1 cells
type CellService interface {
	// Add cell service methods here
}

// ChoreService provides operations for TM1 chores
type ChoreService interface {
	// Add chore service methods here
}

// CubeService provides operations for TM1 cubes
type CubeService interface {
	// Add cube service methods here
}

// DimensionService provides operations for TM1 dimensions
type DimensionService interface {
	// Add dimension service methods here
}

// ElementService provides operations for TM1 elements
type ElementService interface {
	// Basic CRUD operations
	Get(dimensionName, hierarchyName, elementName string) (*Element, error)
	Create(dimensionName, hierarchyName string, element *Element) error
	Update(dimensionName, hierarchyName string, element *Element) error
	Delete(dimensionName, hierarchyName, elementName string) error
	Exists(dimensionName, hierarchyName, elementName string) (bool, error)

	// Element retrieval operations
	GetElements(dimensionName, hierarchyName string) ([]Element, error)
	GetElementNames(dimensionName, hierarchyName string) ([]string, error)
	GetLeafElements(dimensionName, hierarchyName string) ([]Element, error)
	GetConsolidatedElements(dimensionName, hierarchyName string) ([]Element, error)
	GetNumericElements(dimensionName, hierarchyName string) ([]Element, error)
	GetStringElements(dimensionName, hierarchyName string) ([]Element, error)

	// Count operations
	GetNumberOfElements(dimensionName, hierarchyName string) (int, error)

	// Hierarchy operations
	GetEdges(dimensionName, hierarchyName string) ([]Edge, error)
	AddEdges(dimensionName, hierarchyName string, edges []Edge) error
	RemoveEdge(dimensionName, hierarchyName, parentName, componentName string) error

	// Attribute operations
	GetElementAttributes(dimensionName, hierarchyName string) ([]ElementAttribute, error)

	// MDX operations
	ExecuteSetMDX(params MDXExecuteParams) (*CellsetAxis, error)
}

// FileService provides operations for TM1 files
type FileService interface {
	// Add file service methods here
}

// HierarchyService provides operations for TM1 hierarchies
type HierarchyService interface {
	// Add hierarchy service methods here
}

// ProcessService provides operations for TM1 processes
type ProcessService interface {
	// Add process service methods here
}

// SecurityService provides operations for TM1 security
type SecurityService interface {
	// Add security service methods here
}

// SubsetService provides operations for TM1 subsets
type SubsetService interface {
	// Add subset service methods here
}

// ApplicationService provides operations for TM1 applications
type ApplicationService interface {
	// Add application service methods here
}

// ViewService provides operations for TM1 views
type ViewService interface {
	// Add view service methods here
}

// SandboxService provides operations for TM1 sandboxes
type SandboxService interface {
	// Add sandbox service methods here
}

// JobService provides operations for TM1 jobs
type JobService interface {
	// Add job service methods here
}

// UserService provides operations for TM1 users
type UserService interface {
	// Add user service methods here
}

// ThreadService provides operations for TM1 threads
type ThreadService interface {
	// Add thread service methods here
}

// SessionService provides operations for TM1 sessions
type SessionService interface {
	// Add session service methods here
}

// ConfigurationService provides operations for TM1 configuration
type ConfigurationService interface {
	// Add configuration service methods here
}

// AuditLogService provides operations for TM1 audit logs
type AuditLogService interface {
	// Add audit log service methods here
}

// TransactionLogService provides operations for TM1 transaction logs
type TransactionLogService interface {
	// Add transaction log service methods here
}

// MessageLogService provides operations for TM1 message logs
type MessageLogService interface {
	// Add message log service methods here
}

// PowerBIService provides operations for Power BI integration
type PowerBIService interface {
	// Add Power BI service methods here
}

// LoggerService provides operations for TM1 loggers
type LoggerService interface {
	// Add logger service methods here
}

// ServerService provides operations for TM1 server
type ServerService interface {
	// Add server service methods here
}

// MonitoringService provides operations for TM1 monitoring
type MonitoringService interface {
	// Add monitoring service methods here
}

// GitService provides operations for TM1 git integration
type GitService interface {
	// Add git service methods here
}

// Constructor functions for services (to be implemented)

func NewAnnotationService(client Client) AnnotationService {
	// TODO: Implement
	return nil
}

func NewCellService(client Client) CellService {
	// TODO: Implement
	return nil
}

func NewChoreService(client Client) ChoreService {
	// TODO: Implement
	return nil
}

func NewCubeService(client Client) CubeService {
	// TODO: Implement
	return nil
}

func NewDimensionService(client Client) DimensionService {
	// TODO: Implement
	return nil
}

// NewElementService is implemented in element_service.go

func NewFileService(client Client) FileService {
	// TODO: Implement
	return nil
}

func NewHierarchyService(client Client) HierarchyService {
	// TODO: Implement
	return nil
}

func NewProcessService(client Client) ProcessService {
	// TODO: Implement
	return nil
}

func NewSecurityService(client Client) SecurityService {
	// TODO: Implement
	return nil
}

func NewSubsetService(client Client) SubsetService {
	// TODO: Implement
	return nil
}

func NewApplicationService(client Client) ApplicationService {
	// TODO: Implement
	return nil
}

func NewViewService(client Client) ViewService {
	// TODO: Implement
	return nil
}

func NewSandboxService(client Client) SandboxService {
	// TODO: Implement
	return nil
}

func NewJobService(client Client) JobService {
	// TODO: Implement
	return nil
}

func NewUserService(client Client) UserService {
	// TODO: Implement
	return nil
}

func NewThreadService(client Client) ThreadService {
	// TODO: Implement
	return nil
}

func NewSessionService(client Client) SessionService {
	// TODO: Implement
	return nil
}

func NewConfigurationService(client Client) ConfigurationService {
	// TODO: Implement
	return nil
}

func NewAuditLogService(client Client) AuditLogService {
	// TODO: Implement
	return nil
}

func NewTransactionLogService(client Client) TransactionLogService {
	// TODO: Implement
	return nil
}

func NewMessageLogService(client Client) MessageLogService {
	// TODO: Implement
	return nil
}

func NewPowerBIService(client Client) PowerBIService {
	// TODO: Implement
	return nil
}

func NewLoggerService(client Client) LoggerService {
	// TODO: Implement
	return nil
}

func NewServerService(client Client) ServerService {
	// TODO: Implement
	return nil
}

func NewMonitoringService(client Client) MonitoringService {
	// TODO: Implement
	return nil
}

func NewGitService(client Client) GitService {
	// TODO: Implement
	return nil
}
