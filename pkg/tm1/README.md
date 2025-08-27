# TM1go - Go Client Library for IBM Planning Analytics (TM1)

TM1go is a comprehensive Go client library for interacting with IBM Planning Analytics (TM1) REST API. It provides a clean, idiomatic Go interface to all TM1 functionality, following Go best practices and patterns.

## Features

- **Complete TM1 REST API Coverage**: Access to all TM1 services including dimensions, cubes, processes, security, and more
- **Type Safety**: Strongly typed interfaces and data structures
- **Connection Management**: Automatic connection handling with reconnection support
- **Multiple Authentication Modes**: Support for Basic, WIA, CAM, IBM Cloud API Key, and more
- **Error Handling**: Comprehensive error types with detailed information
- **Concurrent Safe**: Thread-safe operations for multi-goroutine usage
- **Extensible**: Clean interfaces allow for easy testing and mocking

## Installation

```bash
go get github.com/andreyea/tm1go
```

## Quick Start

### Basic Connection

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/andreyea/tm1go/pkg/tm1"
)

func main() {
    // Create configuration
    config := tm1.DefaultConfig()
    config.Address = "localhost"
    config.Port = 12354
    config.User = "admin"
    config.Password = "apple"
    config.SSL = false
    
    // Create TM1 service
    tm1Service, err := tm1.NewTM1Service(config)
    if err != nil {
        log.Fatal("Failed to create TM1 service:", err)
    }
    defer tm1Service.Close()
    
    // Get server version
    fmt.Printf("Connected to TM1 version: %s\n", tm1Service.Version())
    
    // Use services
    elements := tm1Service.Elements()
    elementNames, err := elements.GetElementNames("Product", "Product")
    if err != nil {
        log.Fatal("Failed to get elements:", err)
    }
    
    fmt.Printf("Found %d elements in Product dimension\n", len(elementNames))
}
```

### Configuration from File

```go
// Load configuration from JSON file
tm1Service, err := tm1.NewTM1ServiceFromConfig("config.json")
if err != nil {
    log.Fatal("Failed to load config:", err)
}
defer tm1Service.Close()
```

Example `config.json`:
```json
{
  "address": "localhost",
  "port": 12354,
  "user": "admin",
  "password": "apple",
  "ssl": false,
  "session_context": "MyGoApp",
  "timeout": "30s",
  "connection_pool_size": 10
}
```

## Architecture

### Core Components

1. **TM1Service**: Main service that provides access to all TM1 functionality
2. **RestService**: Low-level HTTP client handling authentication and requests
3. **Service Interfaces**: Clean interfaces for each TM1 service area
4. **Type Definitions**: Strongly typed data structures for TM1 objects

### Service Structure

```
TM1Service
├── Elements()      - Element operations
├── Dimensions()    - Dimension operations  
├── Cubes()         - Cube operations
├── Processes()     - Process operations
├── Security()      - Security operations
├── Cells()         - Cell operations
├── Views()         - View operations
└── ...             - And many more
```

## Configuration Options

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `Address` | `string` | TM1 server address | |
| `Port` | `int` | HTTP port number | |
| `SSL` | `bool` | Use SSL/TLS | `true` |
| `User` | `string` | Username | |
| `Password` | `string` | Password | |
| `SessionContext` | `string` | Application name | `"TM1go"` |
| `Timeout` | `*time.Duration` | Request timeout | |
| `ConnectionPoolSize` | `int` | HTTP connection pool size | `10` |
| `ReconnectOnSessionTimeout` | `bool` | Auto-reconnect on timeout | `true` |

## Authentication Modes

TM1go supports multiple authentication modes:

- **Basic Authentication**: Username/password
- **Windows Integrated Authentication (WIA)**: Windows authentication
- **CAM Authentication**: Cognos Access Manager
- **IBM Cloud API Key**: For IBM Cloud deployments
- **Service-to-Service**: Application-to-application authentication

## Error Handling

TM1go provides comprehensive error types:

```go
// Check for specific TM1 errors
if err != nil {
    switch e := err.(type) {
    case *tm1.TM1RestException:
        fmt.Printf("REST API error: %s (Status: %d)\n", e.Message, e.StatusCode)
    case *tm1.TM1TimeoutException:
        fmt.Printf("Timeout after %.2f seconds\n", e.Timeout)
    default:
        fmt.Printf("Other error: %v\n", err)
    }
}
```

## Examples

### Working with Elements

```go
elements := tm1Service.Elements()

// Get an element
element, err := elements.Get("Product", "Product", "iPhone")
if err != nil {
    log.Fatal(err)
}

// Create a new element
newElement := &tm1.Element{
    Name: "NewProduct",
    Type: tm1.ElementTypeString,
}
err = elements.Create("Product", "Product", newElement)

// Get all leaf elements
leafElements, err := elements.GetLeafElements("Product", "Product")

// Execute MDX
mdxParams := tm1.MDXExecuteParams{
    MDX: "{[Product].[Product].Members}",
    MemberProperties: []string{"Name", "Attributes/Price"},
}
result, err := elements.ExecuteSetMDX(mdxParams)
```

### Working with Cubes

```go
cubes := tm1Service.Cubes()

// Get all cube names
cubeNames, err := cubes.GetNames()

// Check if cube exists
exists, err := cubes.Exists("Sales")

// Get cube structure
cube, err := cubes.Get("Sales")
```

## Best Practices

1. **Connection Management**: Always call `Close()` when done:
   ```go
   defer tm1Service.Close()
   ```

2. **Error Handling**: Check errors and handle specific TM1 error types

3. **Configuration**: Use configuration files for production deployments

4. **Concurrency**: TM1go is thread-safe, but be mindful of TM1 server limits

5. **Resource Management**: Use appropriate connection pool sizes for your workload

## Testing

TM1go is designed to be easily testable:

```go
// Mock the client for testing
type MockClient struct{}

func (m *MockClient) GET(url string, opts *tm1.RequestOptions) (*tm1.Response, error) {
    // Return mock response
    return &tm1.Response{
        StatusCode: 200,
        Body: []byte(`{"value": []}`),
    }, nil
}

// Use mock in tests
mockClient := &MockClient{}
elementService := tm1.NewElementServiceImpl(mockClient)
```

## Contributing

Contributions are welcome! Please follow Go best practices and include tests for new functionality.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Related Projects

- [TM1py](https://github.com/cubewise-code/tm1py) - Python client library that inspired this Go implementation
- [TM1js](https://github.com/cubewise-code/tm1js) - JavaScript client library
