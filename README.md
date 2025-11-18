# TM1go - IBM Planning Analytics (TM1) REST API Client for Go

A comprehensive Go client library for IBM Planning Analytics (TM1) REST API v11+, inspired by TM1py.

## Features

### ✅ Core Functionality
- **Full REST API Support**: GET, POST, PATCH, PUT, DELETE operations
- **Multiple Authentication Methods**:
  - HTTP Basic Authentication
  - CAM (Cognos Authentication Manager)
  - Session Cookie/SessionID reuse
  - Bearer Token
  - Access Token
  - CAM Passport
- **Session Management**:
  - Session cookie handling with http.Client cookie jar
  - SessionID retrieval for session reuse
  - KeepAlive mode to preserve sessions
  - Automatic session cleanup
- **Comprehensive Configuration**: 75+ config parameters matching TM1py
- **Logging Support**:
  - Config-based logging (on/off)
  - Custom logger interface
  - File-based logging
- **Advanced Features**:
  - Connection pooling configuration
  - SSL verification control
  - Proxy support (HTTP, HTTPS, SOCKS5)
  - Custom TLS certificates
  - Request timeouts
  - Async operation management
  - Compact JSON responses

### 🔶 Current Limitations
- Service sub-modules (CubeService, DimensionService, etc.) not yet implemented
- Data model classes (Cube, Dimension, Element, etc.) not yet implemented
- See [FEATURE_COMPARISON.md](./FEATURE_COMPARISON.md) for detailed comparison with TM1py

## Installation

```bash
go get github.com/andreyea/tm1go/pkg/tm1
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/andreyea/tm1go/pkg/tm1"
)

func main() {
    // Create configuration
    cfg := tm1.Config{
        Address:  "localhost",
        Port:     8882,
        User:     "admin",
        Password: "apple",
        SSL:      true,
        Logging:  true,
    }

    // Create TM1 service
    svc, err := tm1.NewTM1Service(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer svc.Close()

    ctx := context.Background()

    // Get TM1 version
    version, err := svc.Version(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("TM1 Version: %s\n", version)

    // Check if connected
    if svc.IsConnected(ctx) {
        fmt.Println("✓ Connected to TM1")
    }

    // Get current user
    user, _ := svc.WhoAmI(ctx)
    fmt.Printf("User: %+v\n", user)

    // Check admin privileges
    isAdmin, _ := svc.IsAdmin(ctx)
    fmt.Printf("Is Admin: %t\n", isAdmin)
}
```

## Configuration

### Basic Configuration

```go
cfg := tm1.Config{
    Address:  "localhost",
    Port:     8882,
    User:     "admin",
    Password: "password",
    SSL:      true,
}
```

### Advanced Configuration

```go
cfg := tm1.Config{
    // Connection
    Address:               "localhost",
    Port:                  8882,
    SSL:                   true,
    
    // Authentication
    User:                  "admin",
    Password:              "password",
    DecodeBase64:          false,
    Namespace:             "",           // CAM namespace
    Gateway:               "",           // CAM gateway URL
    IntegratedLogin:       false,
    IntegratedLoginDomain: "",
    IntegratedLoginService: "",
    IntegratedLoginHost:   "",
    IntegratedLoginDelegate: false,
    ImpersonateUser:       "",
    SessionID:             "",           // Reuse existing session
    CAMPassport:           "",
    AccessToken:           "",
    
    // Request Behavior
    AsyncRequestsMode:     false,
    Timeout:               60,           // seconds
    CancelAtTimeout:       false,
    
    // Connection Pooling
    ConnectionPoolSize:    10,
    ReconnectOnSessionTimeout: false,
    
    // Session Management
    KeepAlive:            false,        // Keep session alive after Close()
    
    // Proxies
    HTTPProxy:            "",
    HTTPSProxy:           "",
    ProxyType:            "",           // "http", "https", "socks5"
    
    // Certificates
    Verify:               true,
    CertFile:             "",           // Path to cert file
    CertData:             "",           // Cert data as string
    
    // Logging
    Logging:              true,
}
```

## Authentication

### Basic Authentication

```go
cfg := tm1.Config{
    Address:  "localhost",
    Port:     8882,
    User:     "admin",
    Password: "password",
    SSL:      true,
}
```

### CAM Authentication

```go
cfg := tm1.Config{
    Address:   "localhost",
    Port:      8882,
    Namespace: "LDAP",
    User:      "user@domain.com",
    Password:  "password",
    SSL:       true,
}
```

### Session Reuse

```go
// First connection
cfg1 := tm1.Config{
    Address:   "localhost",
    Port:      8882,
    User:      "admin",
    Password:  "password",
    SSL:       true,
    KeepAlive: true,  // Keep session alive
}

svc1, _ := tm1.NewTM1Service(cfg1)
sessionID := svc1.SessionID()
svc1.Close()  // Closes client but keeps TM1 session alive

// Reuse session
cfg2 := tm1.Config{
    Address:   "localhost",
    Port:      8882,
    SessionID: sessionID,
    SSL:       true,
}

svc2, _ := tm1.NewTM1Service(cfg2)
// Same session, no re-authentication needed
```

### Bearer Token

```go
cfg := tm1.Config{
    Address:     "localhost",
    Port:        8882,
    AccessToken: "your-bearer-token",
    SSL:         true,
}
```

## Logging

### Config-based Logging

```go
cfg := tm1.Config{
    Address: "localhost",
    Port:    8882,
    User:    "admin",
    SSL:     true,
    Logging: true,  // Enable default logging
}
```

### Custom Logger

```go
type CustomLogger struct{}

func (l *CustomLogger) Printf(format string, args ...any) {
    log.Printf("[CUSTOM] "+format, args...)
}

svc, _ := tm1.NewTM1Service(cfg, tm1.WithLogger(&CustomLogger{}))
```

### File Logger

```go
type FileLogger struct {
    file *os.File
}

func (l *FileLogger) Printf(format string, args ...any) {
    fmt.Fprintf(l.file, format+"\n", args...)
}

file, _ := os.Create("tm1_requests.log")
defer file.Close()

svc, _ := tm1.NewTM1Service(cfg, tm1.WithLogger(&FileLogger{file: file}))
```

## Advanced Features

### Session Management

```go
// Get current session ID
sessionID := svc.SessionID()
fmt.Printf("Session ID: %s\n", sessionID)

// Check if connected
if svc.IsConnected(ctx) {
    fmt.Println("Connected")
}

// Reconnect with new config
newCfg := tm1.Config{...}
err := svc.Reconnect(newCfg)
```

### User Information

```go
// Get current user
user, _ := svc.WhoAmI(ctx)

// Check privileges
isAdmin, _ := svc.IsAdmin(ctx)
isDataAdmin, _ := svc.IsDataAdmin(ctx)
isSecurityAdmin, _ := svc.IsSecurityAdmin(ctx)
isOpsAdmin, _ := svc.IsOpsAdmin(ctx)
sandboxingDisabled, _ := svc.SandboxingDisabled(ctx)
```

### Async Operations

```go
// Retrieve async operation result
resp, err := svc.RetrieveAsyncResponse(ctx, asyncID)

// Cancel async operation
err := svc.CancelAsyncOperation(ctx, asyncID)

// Cancel running operation
err := svc.CancelRunningOperation(ctx, threadID)
```

### Compact JSON

```go
rest := svc.Rest()
originalHeader := rest.AddCompactJSONHeader()
// Requests now use compact JSON format
// Restore: rest.headers.Set("Accept", originalHeader)
```

### Direct REST Access

```go
rest := svc.Rest()

// GET request
resp, err := rest.Get(ctx, "/Cubes")

// POST with body
body := strings.NewReader(`{"Name":"NewCube"}`)
resp, err := rest.Post(ctx, "/Cubes", body)

// With custom headers
resp, err := rest.Get(ctx, "/Cubes", 
    tm1.WithHeader("Custom-Header", "value"))

// With query parameters
resp, err := rest.Get(ctx, "/Cubes",
    tm1.WithQueryValue("$select", "Name"))
```

## TM1Service Methods

### Core Methods

| Method | Description |
|--------|-------------|
| `Version(ctx)` | Get TM1 server version |
| `Metadata(ctx)` | Get REST API metadata |
| `Ping(ctx)` | Check server connectivity |
| `Close()` | Close connection and logout |
| `Logout(ctx)` | Explicitly logout |
| `SessionID()` | Get current session ID |

### User Methods

| Method | Description |
|--------|-------------|
| `WhoAmI(ctx)` | Get current user info |
| `IsAdmin(ctx)` | Check if user is admin |
| `IsDataAdmin(ctx)` | Check if user is data admin |
| `IsSecurityAdmin(ctx)` | Check if user is security admin |
| `IsOpsAdmin(ctx)` | Check if user is ops admin |
| `SandboxingDisabled(ctx)` | Check if sandboxing is disabled |

### Connection Methods

| Method | Description |
|--------|-------------|
| `IsConnected(ctx)` | Check if connected to TM1 |
| `Reconnect(cfg, opts...)` | Reconnect with new config |

### Async Methods

| Method | Description |
|--------|-------------|
| `RetrieveAsyncResponse(ctx, asyncID)` | Get async operation result |
| `CancelAsyncOperation(ctx, asyncID)` | Cancel async operation |
| `CancelRunningOperation(ctx, threadID)` | Cancel running operation |

## Examples

See the [examples](./examples) directory for complete working examples:

- [main.go](./examples/main.go) - Basic connection and version retrieval
- [test_session.go](./examples/test_session.go) - Session reuse with SessionID
- [test_keepalive.go](./examples/test_keepalive.go) - KeepAlive feature demonstration
- [test_logging.go](./examples/test_logging.go) - Comprehensive logging examples
- [test_features.go](./examples/test_features.go) - All features demonstration
- [test_async_and_features.go](./examples/test_async_and_features.go) - Async operations and admin features
- [config_test.go](./examples/config_test.go) - Configuration options testing

## Project Structure

```
tm1go/
├── pkg/
│   └── tm1/
│       ├── config.go           # Configuration struct and helpers
│       ├── rest_service.go     # Low-level REST client
│       ├── tm1_service.go      # High-level TM1 service
│       ├── options.go          # Functional options patterns
│       └── errors.go           # Error types
├── examples/                   # Example applications
├── FEATURE_COMPARISON.md       # Detailed comparison with TM1py
└── README.md                   # This file
```

## Comparison with TM1py

TM1go currently implements the **core foundation** of TM1py:
- ✅ Complete REST communication layer
- ✅ All authentication methods
- ✅ Session management
- ✅ Comprehensive configuration (75+ parameters)
- ✅ Logging capabilities
- ✅ Basic TM1Service functionality

**Main differences:**
- 🔴 Service sub-modules (CubeService, DimensionService, etc.) not yet implemented
- 🔴 Data model classes not yet implemented

See [FEATURE_COMPARISON.md](./FEATURE_COMPARISON.md) for a detailed breakdown.

## Testing

Run the examples:

```bash
cd examples
go run main.go                     # Basic connection test
go run test_session.go             # Session reuse
go run test_keepalive.go          # KeepAlive feature
go run test_logging.go            # Logging examples
go run test_features.go           # All features
go run test_async_and_features.go # Async and admin features
```

## Requirements

- Go 1.16+
- IBM Planning Analytics (TM1) 11.8+ with REST API enabled

## License

See [LICENSE](./LICENSE) file.

## Contributing

Contributions welcome! Priority areas:
1. Service sub-modules (CubeService, DimensionService, etc.)
2. Data model structs (Cube, Dimension, Element, etc.)
3. Utility functions (MDX helpers, formatters)
4. Test coverage

## Credits

Inspired by [TM1py](https://github.com/cubewise-code/tm1py) - the excellent Python library for TM1 REST API.

## Support

For issues and questions:
- GitHub Issues: [Report an issue](https://github.com/andreyea/tm1go/issues)
- TM1 REST API Documentation: IBM Planning Analytics documentation

## Roadmap

### v0.1 (Current)
- ✅ Core REST communication
- ✅ Authentication
- ✅ Session management
- ✅ Basic TM1Service methods

### v0.2 (Planned)
- 🔜 CubeService
- 🔜 DimensionService
- 🔜 ElementService
- 🔜 HierarchyService
- 🔜 Basic data models

### v0.3 (Future)
- 🔜 ProcessService
- 🔜 ChoreService
- 🔜 ViewService
- 🔜 CellService

### v1.0 (Goal)
- 🔜 Full feature parity with TM1py
- 🔜 Comprehensive test coverage
- 🔜 Complete documentation
