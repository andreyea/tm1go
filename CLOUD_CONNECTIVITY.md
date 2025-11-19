# TM1Go Cloud Connectivity Guide

This guide covers all authentication methods and cloud deployment patterns supported by TM1Go.

## Table of Contents

- [Overview](#overview)
- [Supported Deployment Types](#supported-deployment-types)
- [Authentication Modes](#authentication-modes)
- [Configuration Examples](#configuration-examples)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

TM1Go supports connecting to:
- **TM1 v11 Local** - Traditional on-premise TM1 servers
- **TM1 v11 Cloud** - Legacy cloud deployments
- **IBM Planning Analytics v12 Cloud** - IBM Cloud with API Key authentication
- **IBM Planning Analytics v12 Service-to-Service** - Named application authentication
- **Planning Analytics Workspace** - Cloud Pak for Data (CPD/ZEN) proxy

## Supported Deployment Types

### 1. TM1 v11 Local (Traditional)

Traditional on-premise TM1 servers using HTTP/HTTPS.

**URL Pattern**: `https://hostname:port/api/v1`

**Configuration Fields**:
- `Address` - Server hostname or IP
- `Port` - HTTPPortNumber from tm1s.cfg
- `SSL` - Enable HTTPS
- `User` - Username
- `Password` - Password

### 2. IBM Cloud v12 (API Key)

IBM Cloud Planning Analytics with IAM authentication.

**URL Pattern**: `https://region.planninganalytics.cloud.ibm.com/api/tenant/v0/tm1/database`

**Configuration Fields**:
- `Address` - IBM Cloud regional URL (e.g., `us-east-2.planninganalytics.cloud.ibm.com`)
- `Tenant` - Tenant ID (e.g., `YC4B2M1AG2Y6`)
- `Database` - Database name
- `APIKey` - IBM Cloud API Key
- `IAMUrl` - IAM token endpoint (optional, defaults to `https://iam.cloud.ibm.com`)
- `AsyncRequestsMode` - **Required: `true`** to avoid 60s gateway timeout

### 3. Service-to-Service v12 (Named Application)

Planning Analytics Engine v12 with named application authentication.

**URL Pattern**: `https://hostname:port/instance/api/v1/Databases('database')`

**Configuration Fields**:
- `Address` - Server hostname
- `Port` - Server port
- `Instance` - TM1 instance name (e.g., `tm1s1`)
- `Database` - Database name
- `ApplicationClientID` - Named application client ID
- `ApplicationClientSecret` - Named application secret
- `SSL` - Enable HTTPS

### 4. Planning Analytics Proxy (CPD/ZEN)

Planning Analytics Workspace through Cloud Pak for Data proxy.

**URL Pattern**: `https://hostname/tm1/database/api/v1`

**Configuration Fields**:
- `Address` - PA Workspace URL
- `Database` - Database name
- `CPDUrl` - Cloud Pak for Data ZEN URL
- `User` - CPD username
- `Password` - CPD password
- `AccessToken` - Pre-obtained CPD access token (alternative to username/password)
- `SSL` - Enable HTTPS

## Authentication Modes

### 1. Basic Authentication

Standard username/password authentication.

```go
cfg := tm1.Config{
    Address:  "localhost",
    Port:     12354,
    User:     "admin",
    Password: "apple",
    SSL:      true,
}
```

**When to use**: Traditional TM1 v11 servers with local or LDAP users.

### 2. CAM Authentication

Cognos Access Manager authentication with namespace support.

```go
cfg := tm1.Config{
    Address:   "tm1-server.company.com",
    Port:      8011,
    User:      "john.doe",
    Password:  "password123",
    Namespace: "LDAP",
    Gateway:   "cam-server.company.com:9300",
    SSL:       true,
}
```

**When to use**: Enterprise deployments with Cognos Analytics integration.

### 3. CAM Passport

Pre-authenticated CAM passport token.

```go
cfg := tm1.Config{
    Address:     "tm1-server.company.com",
    Port:        8011,
    CAMPassport: "your-cam-passport-value",
    SSL:         true,
}
```

**When to use**: Integrating with applications that already have CAM authentication.

### 4. IBM Cloud API Key

IBM Cloud IAM authentication using API keys.

```go
cfg := tm1.Config{
    Address:           "us-east-2.planninganalytics.cloud.ibm.com",
    Tenant:            "YC4B2M1AG2Y6",
    Database:          "Planning Sample",
    APIKey:            "your-ibm-cloud-api-key",
    IAMUrl:            "https://iam.cloud.ibm.com", // Optional
    SSL:               true,
    AsyncRequestsMode: true, // Important!
}
```

**When to use**: IBM Cloud Planning Analytics deployments.

**Important**: Always set `AsyncRequestsMode: true` for IBM Cloud to avoid 60-second gateway timeouts on long-running operations.

### 5. Service-to-Service (Named Application)

Programmatic authentication using named application credentials.

```go
cfg := tm1.Config{
    Address:                 "localhost",
    Port:                    8001,
    Instance:                "tm1s1",
    Database:                "Planning Sample",
    ApplicationClientID:     "my-application",
    ApplicationClientSecret: "my-application-secret",
    SSL:                     true,
}
```

**When to use**: v12 deployments with service accounts for automation.

### 6. Access Token (Bearer Token)

Pre-obtained OAuth2 or JWT token authentication.

```go
cfg := tm1.Config{
    Address:     "tm1-server.company.com",
    Port:        8011,
    AccessToken: "your-bearer-token",
    SSL:         true,
}
```

**When to use**: Integrating with OAuth2 providers (Azure AD, Okta, etc.).

### 7. Session ID Reuse

Reuse an existing TM1 session without re-authenticating.

```go
// First connection
cfg1 := tm1.Config{
    Address:   "localhost",
    Port:      12354,
    User:      "admin",
    Password:  "apple",
    SSL:       true,
    KeepAlive: true, // Don't logout on Close()
}
service1, _ := tm1.NewTM1Service(cfg1)
sessionID := service1.SessionID()
service1.Close() // Closes connection but keeps session alive

// Second connection - reuse session
cfg2 := tm1.Config{
    Address:   "localhost",
    Port:      12354,
    SSL:       true,
    SessionID: sessionID, // No credentials needed
}
service2, _ := tm1.NewTM1Service(cfg2)
```

**When to use**: 
- Connection pooling in web applications
- Sharing sessions across multiple service instances
- Avoiding repeated authentication overhead

### 8. Base64 Encoded Password

Store passwords in base64 encoding for basic obfuscation.

```go
cfg := tm1.Config{
    Address:      "localhost",
    Port:         12354,
    User:         "admin",
    Password:     "YXBwbGU=", // "apple" base64 encoded
    DecodeBase64: true,
    SSL:          true,
}
```

**When to use**: Configuration files where passwords should not be plain text.

**Note**: This is NOT encryption - use proper secret management in production.

### 9. Windows Integrated Authentication (WIA)

**Status**: Not yet implemented in Go (requires SSPI library).

Windows authentication using current user credentials.

```go
cfg := tm1.Config{
    Address:                "localhost",
    Port:                   12354,
    IntegratedLogin:        true,
    IntegratedLoginDomain:  "DOMAIN",
    IntegratedLoginService: "HTTP",
    SSL:                    true,
}
```

**When to use**: Windows environments with Active Directory.

## Configuration Examples

### Complete IBM Cloud Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/andreyea/tm1go/pkg/tm1"
)

func main() {
    cfg := tm1.Config{
        // Connection
        Address:  "us-east-2.planninganalytics.cloud.ibm.com",
        Tenant:   "YC4B2M1AG2Y6",
        Database: "Planning Sample",
        SSL:      true,
        
        // Authentication
        APIKey: "your-ibm-cloud-api-key-here",
        IAMUrl: "https://iam.cloud.ibm.com",
        
        // Behavior
        AsyncRequestsMode: true,  // Required for IBM Cloud
        CancelAtTimeout:   false, // Keep operations running on timeout
        Timeout:           120,   // 120 seconds
        
        // Connection management
        ReConnectOnSessionTimeout:   true,
        ReConnectOnRemoteDisconnect: true,
        KeepAlive:                   false, // Logout on Close()
    }

    service, err := tm1.NewTM1Service(cfg)
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer service.Close()

    ctx := context.Background()
    
    // Verify connection
    version, err := service.Version(ctx)
    if err != nil {
        log.Fatalf("Failed to get version: %v", err)
    }
    fmt.Printf("Connected to TM1 v%s\n", version)

    // Use services
    processes, err := service.Processes.GetAllNames(ctx, false)
    if err != nil {
        log.Fatalf("Failed to get processes: %v", err)
    }
    fmt.Printf("Found %d processes\n", len(processes))
}
```

### Complete Service-to-Service Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/andreyea/tm1go/pkg/tm1"
)

func main() {
    cfg := tm1.Config{
        // Connection
        Address:  "pa-server.company.com",
        Port:     8001,
        Instance: "tm1s1",
        Database: "Planning Sample",
        SSL:      true,
        
        // Service-to-Service Authentication
        ApplicationClientID:     "automation-app",
        ApplicationClientSecret: "app-secret-key",
        
        // Optional: Connection pool settings
        ConnectionPoolSize: 10,
        PoolConnections:    1,
        Timeout:            60,
        
        // Reconnection
        ReConnectOnSessionTimeout:   true,
        ReConnectOnRemoteDisconnect: true,
    }

    service, err := tm1.NewTM1Service(cfg)
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer service.Close()

    ctx := context.Background()
    
    // Execute process
    result, err := service.Processes.Execute(ctx, "UpdateData", nil)
    if err != nil {
        log.Fatalf("Failed to execute process: %v", err)
    }
    fmt.Printf("Process completed with status: %s\n", result.Status)
}
```

## Best Practices

### 1. Always Set AsyncRequestsMode for IBM Cloud

IBM Cloud has a 60-second gateway timeout. Enable async mode to prevent timeout errors:

```go
cfg := tm1.Config{
    AsyncRequestsMode: true,
    CancelAtTimeout:   false,
    Timeout:           120,
}
```

### 2. Use Session Reuse for Connection Pooling

In web applications, reuse sessions to avoid authentication overhead:

```go
// Store session ID in cache/database
sessionID := service.SessionID()

// Reuse in subsequent requests
cfg := tm1.Config{
    SessionID: sessionID,
    // ... other settings
}
```

### 3. Enable Automatic Reconnection

Let TM1Go handle transient connection issues:

```go
cfg := tm1.Config{
    ReConnectOnSessionTimeout:   true,
    ReConnectOnRemoteDisconnect: true,
}
```

### 4. Configure Appropriate Timeouts

Set timeouts based on your operations:

```go
cfg := tm1.Config{
    Timeout: 300, // 5 minutes for long-running operations
}
```

### 5. Use KeepAlive Strategically

Keep sessions alive when sharing across instances:

```go
cfg := tm1.Config{
    KeepAlive: true, // Don't logout on Close()
}
```

**Important**: Always call `Logout()` explicitly when done:

```go
defer service.Logout(context.Background())
```

### 6. Secure Credential Management

Never hardcode credentials. Use environment variables or secret management:

```go
cfg := tm1.Config{
    User:     os.Getenv("TM1_USER"),
    Password: os.Getenv("TM1_PASSWORD"),
}
```

### 7. Handle Errors Gracefully

Always check for connection and authentication errors:

```go
service, err := tm1.NewTM1Service(cfg)
if err != nil {
    if httpErr, ok := err.(*tm1.HTTPError); ok {
        switch httpErr.StatusCode {
        case 401:
            log.Fatal("Authentication failed - check credentials")
        case 403:
            log.Fatal("Access denied - insufficient privileges")
        case 404:
            log.Fatal("Server not found - check address and port")
        default:
            log.Fatalf("Connection failed: %v", err)
        }
    }
    log.Fatalf("Failed to connect: %v", err)
}
```

## Troubleshooting

### IBM Cloud: 60-Second Timeout Errors

**Problem**: Operations fail after 60 seconds with timeout errors.

**Solution**: Enable async requests mode:

```go
cfg.AsyncRequestsMode = true
cfg.Timeout = 120 // or higher
```

### Authentication Failures

**Problem**: 401 Unauthorized errors.

**Solutions**:
1. Verify credentials are correct
2. For IBM Cloud, verify API key is valid and has correct permissions
3. For CAM, verify namespace and gateway settings
4. For service-to-service, verify named application exists and secret is correct

### Connection Refused Errors

**Problem**: Cannot connect to server.

**Solutions**:
1. Verify `Address` and `Port` are correct
2. Verify `SSL` setting matches server configuration
3. Check firewall rules
4. Verify TM1 server is running
5. For IBM Cloud, verify region-specific URL is correct

### Session Timeout Issues

**Problem**: Session expires during long-running operations.

**Solutions**:
1. Enable automatic reconnection:
   ```go
   cfg.ReConnectOnSessionTimeout = true
   ```
2. Increase session timeout on TM1 server
3. Use `KeepAlive: true` and refresh connections periodically

### SSL Certificate Errors

**Problem**: SSL verification failures.

**Solutions**:
1. For development, disable SSL verification:
   ```go
   cfg.SkipSSLVerification = true
   ```
2. For production, provide certificate:
   ```go
   cfg.Verify = "/path/to/certificate.cer"
   ```
3. Ensure server certificate is valid and not expired

### Service-to-Service Authentication Errors

**Problem**: Named application authentication fails.

**Solutions**:
1. Verify named application exists in TM1:
   - Connect to TM1 as admin
   - Check Applications folder in TM1 Architect
2. Verify `ApplicationClientID` and `ApplicationClientSecret` are correct
3. Verify the application has necessary permissions
4. Check `Instance` and `Database` names are correct (case-sensitive)

### IBM Cloud IAM Token Generation Failures

**Problem**: Cannot generate IAM access token.

**Solutions**:
1. Verify API key is valid and not revoked
2. Check `IAMUrl` is correct (default: `https://iam.cloud.ibm.com`)
3. Verify network can reach IBM IAM endpoint
4. Check API key has necessary IAM permissions

## API Reference

### Config Fields Reference

| Field | Type | Description | Required For |
|-------|------|-------------|--------------|
| `Address` | string | Server hostname/IP | All deployments |
| `Port` | int | Server port | v11, v12 on-premise |
| `SSL` | bool | Use HTTPS | All deployments |
| `User` | string | Username | Basic, CAM |
| `Password` | string | Password | Basic, CAM |
| `Namespace` | string | CAM namespace | CAM |
| `CAMPassport` | string | CAM passport token | CAM Passport |
| `SessionID` | string | Existing session ID | Session reuse |
| `AccessToken` | string | OAuth2/JWT token | Token auth |
| `APIKey` | string | IBM Cloud API key | IBM Cloud |
| `IAMUrl` | string | IBM IAM URL | IBM Cloud |
| `Tenant` | string | IBM Cloud tenant ID | IBM Cloud |
| `Database` | string | Database name | v12, PA Proxy |
| `Instance` | string | Instance name | v12 S2S |
| `ApplicationClientID` | string | Named app client ID | v12 S2S |
| `ApplicationClientSecret` | string | Named app secret | v12 S2S |
| `CPDUrl` | string | CPD/ZEN URL | PA Proxy |
| `BaseURL` | string | Full base URL | Custom URLs |
| `AsyncRequestsMode` | bool | Use async for IBM Cloud | IBM Cloud |
| `CancelAtTimeout` | bool | Cancel on timeout | Optional |
| `Timeout` | time.Duration | Request timeout | Optional |
| `ReConnectOnSessionTimeout` | bool | Auto-reconnect on timeout | Optional |
| `ReConnectOnRemoteDisconnect` | bool | Auto-reconnect on disconnect | Optional |
| `KeepAlive` | bool | Don't logout on Close() | Optional |
| `DecodeBase64` | bool | Decode password from base64 | Optional |

## Version Compatibility

| Feature | TM1 v11 | PA v12 Local | PA v12 IBM Cloud | PA v12 S2S | PA Workspace |
|---------|---------|--------------|------------------|------------|--------------|
| Basic Auth | ✅ | ✅ | ❌ | ❌ | ✅ |
| CAM Auth | ✅ | ✅ | ❌ | ❌ | ❌ |
| API Key | ❌ | ❌ | ✅ | ❌ | ❌ |
| Service-to-Service | ❌ | ✅ | ❌ | ✅ | ❌ |
| Access Token | ✅ | ✅ | ✅ | ✅ | ✅ |
| Session Reuse | ✅ | ✅ | ✅ | ✅ | ✅ |
| Async Mode | N/A | N/A | **Required** | Optional | Optional |

## Additional Resources

- [TM1 REST API Documentation](https://www.ibm.com/docs/en/planning-analytics)
- [IBM Cloud Planning Analytics](https://cloud.ibm.com/catalog/services/planning-analytics)
- [Planning Analytics Engine v12 Documentation](https://www.ibm.com/docs/en/planning-analytics/2.0.0)
- [Example Code](examples/test_cloud_connections.go)

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/andreyea/tm1go).
