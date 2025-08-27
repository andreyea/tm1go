// Package tm1 provides a comprehensive Go client library for IBM Planning Analytics (TM1) REST API.
//
// TM1go offers a clean, idiomatic Go interface to all TM1 functionality, following Go best practices
// and patterns. It provides type-safe operations, comprehensive error handling, and connection management
// for TM1 servers.
//
// Basic Usage:
//
//	config := tm1.DefaultConfig()
//	config.Address = "localhost"
//	config.Port = 12354
//	config.User = "admin"
//	config.Password = "password"
//	config.SSL = false
//
//	tm1Service, err := tm1.NewTM1Service(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer tm1Service.Close()
//
//	// Use the service
//	elements := tm1Service.Elements()
//	elementNames, err := elements.GetElementNames("Product", "Product")
//	if err != nil {
//		log.Fatal(err)
//	}
//
// The library is organized into services that correspond to different areas of TM1 functionality:
//
//   - Elements: Dimension element operations
//   - Dimensions: Dimension management
//   - Cubes: Cube operations
//   - Processes: TurboIntegrator process management
//   - Security: User and group management
//   - Cells: Cell value operations
//   - Views: View management
//   - And many more...
//
// Each service provides methods for common operations like Create, Read, Update, Delete (CRUD)
// as well as specialized operations specific to that service area.
//
// Error Handling:
//
// The library provides specific error types for different kinds of TM1 errors:
//
//   - TM1RestException: REST API errors with HTTP status codes
//   - TM1TimeoutException: Timeout errors
//   - TM1VersionDeprecationException: Version compatibility errors
//
// Authentication:
//
// Multiple authentication modes are supported:
//
//   - Basic Authentication (username/password)
//   - Windows Integrated Authentication (WIA)
//   - CAM Authentication
//   - IBM Cloud API Key
//   - Service-to-Service authentication
//
// The authentication mode is automatically determined based on the configuration provided.
//
// Configuration:
//
// Configuration can be provided directly as a Config struct or loaded from a JSON file:
//
//	tm1Service, err := tm1.NewTM1ServiceFromConfig("config.json")
//
// Thread Safety:
//
// All operations are thread-safe and can be used concurrently from multiple goroutines.
// The underlying HTTP client uses connection pooling for efficient resource utilization.
//
// For more examples and detailed documentation, see the examples directory and individual
// service documentation.
package tm1

const (
	// Version is the current version of the TM1go library
	Version = "1.0.0"

	// UserAgent is the default user agent string used in HTTP requests
	UserAgent = "TM1go/" + Version
)

// Library-wide constants
const (
	// DefaultTimeout is the default request timeout
	DefaultTimeout = 30 // seconds

	// DefaultConnectionPoolSize is the default HTTP connection pool size
	DefaultConnectionPoolSize = 10

	// DefaultSessionContext is the default session context name
	DefaultSessionContext = "TM1go"
)
