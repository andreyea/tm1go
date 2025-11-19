# TM1Go Quick Reference - Cloud Connections

## Quick Start Examples

### TM1 v11 Local
```go
cfg := tm1.Config{
    Address: "localhost", Port: 12354, SSL: true,
    User: "admin", Password: "apple",
}
service, _ := tm1.NewTM1Service(cfg)
```

### IBM Cloud v12
```go
cfg := tm1.Config{
    Address: "us-east-2.planninganalytics.cloud.ibm.com",
    Tenant: "YC4B2M1AG2Y6", Database: "Planning Sample",
    APIKey: "your-api-key", SSL: true,
    AsyncRequestsMode: true, // REQUIRED
}
service, _ := tm1.NewTM1Service(cfg)
```

### Service-to-Service v12
```go
cfg := tm1.Config{
    Address: "localhost", Port: 8001, SSL: true,
    Instance: "tm1s1", Database: "Planning Sample",
    ApplicationClientID: "app", ApplicationClientSecret: "secret",
}
service, _ := tm1.NewTM1Service(cfg)
```

### PA Proxy (CPD)
```go
cfg := tm1.Config{
    Address: "pa-workspace.company.com", SSL: true,
    Database: "Planning Sample",
    CPDUrl: "https://cpd-zen.apps.company.com",
    AccessToken: "cpd-token",
}
service, _ := tm1.NewTM1Service(cfg)
```

### CAM Authentication
```go
cfg := tm1.Config{
    Address: "localhost", Port: 8011, SSL: true,
    User: "john.doe", Password: "pwd", Namespace: "LDAP",
}
service, _ := tm1.NewTM1Service(cfg)
```

### Session Reuse
```go
// First connection
cfg1 := tm1.Config{...KeepAlive: true}
service1, _ := tm1.NewTM1Service(cfg1)
sessionID := service1.SessionID()
service1.Close()

// Second connection
cfg2 := tm1.Config{SessionID: sessionID, ...}
service2, _ := tm1.NewTM1Service(cfg2)
```

## Authentication Modes

| Mode | Required Fields | Use Case |
|------|----------------|----------|
| Basic | `User`, `Password` | Traditional TM1 v11 |
| CAM | `User`, `Password`, `Namespace` | Enterprise LDAP/AD |
| CAM Passport | `CAMPassport` | Pre-authenticated CAM |
| API Key | `APIKey`, `Tenant`, `Database` | IBM Cloud v12 |
| Service-to-Service | `ApplicationClientID`, `ApplicationClientSecret`, `Instance`, `Database` | v12 automation |
| Access Token | `AccessToken` | OAuth2/JWT |
| Session ID | `SessionID` | Session reuse |

## URL Patterns

| Deployment | Pattern |
|------------|---------|
| v11 Local | `https://host:port/api/v1` |
| IBM Cloud | `https://region.cloud.ibm.com/api/tenant/v0/tm1/database` |
| Service-to-Service | `https://host:port/instance/api/v1/Databases('database')` |
| PA Proxy | `https://host/tm1/database/api/v1` |

## Common Config Fields

```go
// Connection
Address:  "hostname"        // Server address
Port:     12354             // Server port (v11, v12 local)
SSL:      true              // Use HTTPS
BaseURL:  ""                // Override with full URL

// v12 Cloud
Tenant:   "TENANT_ID"       // IBM Cloud tenant
Database: "DB Name"         // Database name
Instance: "tm1s1"           // Instance name (v12 local)

// Authentication
User:     "admin"           // Username
Password: "pwd"             // Password
APIKey:   "key"             // IBM Cloud API key
AccessToken: "token"        // Bearer token
SessionID: "session"        // Reuse session

// Service-to-Service
ApplicationClientID: "id"       // Named app client ID
ApplicationClientSecret: "sec"  // Named app secret

// CAM
Namespace: "LDAP"          // CAM namespace
CAMPassport: "passport"    // CAM passport token

// Behavior
AsyncRequestsMode: true    // Required for IBM Cloud
Timeout: 60                // Seconds
KeepAlive: true           // Don't logout on Close()
ReConnectOnSessionTimeout: true    // Auto-reconnect
ReConnectOnRemoteDisconnect: true  // Auto-reconnect

// Security
DecodeBase64: true        // Decode password from base64
SkipSSLVerification: true // Dev only - disable SSL verify
Verify: "/path/cert.cer"  // SSL certificate path
```

## Important Notes

### IBM Cloud
- **ALWAYS** set `AsyncRequestsMode: true`
- Default 60s timeout will cause issues
- Use `Timeout: 120` or higher

### Session Reuse
- Set `KeepAlive: true` to keep session after Close()
- Store `service.SessionID()` for reuse
- Remember to `Logout()` when completely done

### Error Handling
```go
service, err := tm1.NewTM1Service(cfg)
if err != nil {
    if httpErr, ok := err.(*tm1.HTTPError); ok {
        switch httpErr.StatusCode {
        case 401: // Authentication failed
        case 403: // Access denied
        case 404: // Server not found
        }
    }
}
```

### Best Practices
1. Use environment variables for credentials
2. Enable auto-reconnection in production
3. Set appropriate timeouts for operations
4. Always close connections: `defer service.Close()`
5. Use session reuse for connection pooling

## Testing

```bash
# Run all tests
go test ./pkg/tm1 -v

# Run cloud tests only
go test ./pkg/tm1 -v -run "Cloud|Auth"

# Check coverage
go test ./pkg/tm1 -cover

# Build examples
go build examples/test_cloud_connections.go
```

## Documentation

- Full Guide: [CLOUD_CONNECTIVITY.md](CLOUD_CONNECTIVITY.md)
- Implementation: [CLOUD_IMPLEMENTATION_SUMMARY.md](CLOUD_IMPLEMENTATION_SUMMARY.md)
- Examples: [examples/test_cloud_connections.go](examples/test_cloud_connections.go)
