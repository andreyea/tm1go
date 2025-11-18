# TM1go vs TM1py Feature Comparison

This document compares the features implemented in TM1go (Go implementation) with TM1py (Python reference implementation).

## Core Components

### RestService

| Feature | TM1py | TM1go | Status |
|---------|-------|-------|--------|
| HTTP Methods (GET, POST, PATCH, PUT, DELETE) | ✓ | ✓ | ✅ Complete |
| Authentication (Basic, CAM, SessionID, Bearer, AccessToken) | ✓ | ✓ | ✅ Complete |
| Session Management (Cookie Jar) | ✓ | ✓ | ✅ Complete |
| Logging | ✓ | ✓ | ✅ Complete |
| SSL Verification Control | ✓ | ✓ | ✅ Complete |
| Proxy Support | ✓ | ✓ | ✅ Complete |
| Connection Pooling | ✓ | ✓ | ✅ Complete |
| Request Timeouts | ✓ | ✓ | ✅ Complete |
| Custom Headers | ✓ | ✓ | ✅ Complete |
| Version Caching | ✓ | ⚠️ | 🔶 Partial (not cached in Go) |
| Compact JSON Header | ✓ | ✓ | ✅ Complete |
| Async Operations | ✓ | ✓ | ✅ Complete |
| Logout/Session Close | ✓ | ✓ | ✅ Complete |
| KeepAlive Support | ✓ | ✓ | ✅ Complete |
| Session ID Retrieval | ✓ | ✓ | ✅ Complete |

### TM1Service Core Methods

| Feature | TM1py | TM1go | Status |
|---------|-------|-------|--------|
| Version() | ✓ | ✓ | ✅ Complete |
| Metadata() | ✓ | ✓ | ✅ Complete |
| Ping() | ✓ | ✓ | ✅ Complete |
| Close() | ✓ | ✓ | ✅ Complete |
| Logout() | ✓ | ✓ | ✅ Complete |
| SessionID() | ✓ | ✓ | ✅ Complete |
| WhoAmI() | ✓ | ✓ | ✅ Complete |
| IsConnected() | ✓ | ✓ | ✅ Complete |
| Reconnect() | ✓ | ✓ | ✅ Complete |
| IsAdmin() | ✓ | ✓ | ✅ Complete |
| IsDataAdmin() | ✓ | ✓ | ✅ Complete |
| IsSecurityAdmin() | ✓ | ✓ | ✅ Complete |
| IsOpsAdmin() | ✓ | ✓ | ✅ Complete |
| SandboxingDisabled() | ✓ | ✓ | ✅ Complete |

### Configuration Options

| Category | TM1py | TM1go | Status |
|----------|-------|-------|--------|
| Connection (Address, Port, SSL) | ✓ | ✓ | ✅ Complete |
| Authentication (User, Password, CAM, etc.) | ✓ | ✓ | ✅ Complete |
| Integrated Login | ✓ | ✓ | ✅ Complete |
| Request Behavior (Async, Timeouts) | ✓ | ✓ | ✅ Complete |
| Connection Pooling | ✓ | ✓ | ✅ Complete |
| Session Management (KeepAlive, SessionID) | ✓ | ✓ | ✅ Complete |
| Proxies | ✓ | ✓ | ✅ Complete |
| Certificates | ✓ | ✓ | ✅ Complete |
| Logging | ✓ | ✓ | ✅ Complete |
| Base64 Password Decoding | ✓ | ✓ | ✅ Complete |

## Service Sub-Modules

These are specialized service classes in TM1py that provide domain-specific functionality. **Not yet implemented in TM1go.**

| Service | TM1py | TM1go | Status | Priority |
|---------|-------|-------|--------|----------|
| ApplicationService | ✓ | ❌ | 🔴 Not Started | High |
| AnnotationService | ✓ | ❌ | 🔴 Not Started | Medium |
| CellService | ✓ | ❌ | 🔴 Not Started | High |
| ChoreService | ✓ | ❌ | 🔴 Not Started | High |
| CubeService | ✓ | ❌ | 🔴 Not Started | High |
| DimensionService | ✓ | ❌ | 🔴 Not Started | High |
| ElementService | ✓ | ❌ | 🔴 Not Started | High |
| HierarchyService | ✓ | ❌ | 🔴 Not Started | High |
| MonitoringService | ✓ | ❌ | 🔴 Not Started | Medium |
| PowerBiService | ✓ | ❌ | 🔴 Not Started | Low |
| ProcessService | ✓ | ❌ | 🔴 Not Started | High |
| SecurityService | ✓ | ❌ | 🔴 Not Started | Medium |
| ServerService | ✓ | ❌ | 🔴 Not Started | Medium |
| SubsetService | ✓ | ❌ | 🔴 Not Started | Medium |
| ViewService | ✓ | ❌ | 🔴 Not Started | High |
| FileService | ✓ | ❌ | 🔴 Not Started | Low |
| GitService | ✓ | ❌ | 🔴 Not Started | Low |
| SandboxService | ✓ | ❌ | 🔴 Not Started | Medium |

## Data Models

TM1py has model classes for various TM1 objects. **Not yet implemented in TM1go.**

| Model | TM1py | TM1go | Status |
|-------|-------|-------|--------|
| Cube | ✓ | ❌ | 🔴 Not Started |
| Dimension | ✓ | ❌ | 🔴 Not Started |
| Hierarchy | ✓ | ❌ | 🔴 Not Started |
| Element | ✓ | ❌ | 🔴 Not Started |
| ElementAttribute | ✓ | ❌ | 🔴 Not Started |
| Subset | ✓ | ❌ | 🔴 Not Started |
| View (NativeView, MDXView) | ✓ | ❌ | 🔴 Not Started |
| Process | ✓ | ❌ | 🔴 Not Started |
| Chore | ✓ | ❌ | 🔴 Not Started |
| ChoreStartTime | ✓ | ❌ | 🔴 Not Started |
| ChoreTask | ✓ | ❌ | 🔴 Not Started |
| Annotation | ✓ | ❌ | 🔴 Not Started |
| User | ✓ | ❌ | 🔴 Not Started |
| Group | ✓ | ❌ | 🔴 Not Started |
| Application | ✓ | ❌ | 🔴 Not Started |
| Sandbox | ✓ | ❌ | 🔴 Not Started |

## Utility Features

| Feature | TM1py | TM1go | Status |
|---------|-------|-------|--------|
| MDXUtils | ✓ | ❌ | 🔴 Not Started |
| CaseAndSpaceInsensitiveDict | ✓ | ❌ | 🔴 Not Started |
| CaseAndSpaceInsensitiveSet | ✓ | ❌ | 🔴 Not Started |
| CaseAndSpaceInsensitiveTuplesDict | ✓ | ❌ | 🔴 Not Started |
| Utilities (format helpers, etc.) | ✓ | ✓ | 🔶 Partial |

## Testing

| Aspect | TM1py | TM1go | Status |
|--------|-------|-------|--------|
| Connection Tests | ✓ | ✓ | ✅ Complete |
| Session Management Tests | ✓ | ✓ | ✅ Complete |
| Logging Tests | ✓ | ✓ | ✅ Complete |
| Service Tests | ✓ | ❌ | 🔴 Not Started |
| Model Tests | ✓ | ❌ | 🔴 Not Started |

## Legend

- ✅ Complete: Feature fully implemented and tested
- 🔶 Partial: Feature partially implemented
- 🔴 Not Started: Feature not yet implemented
- ⚠️ Different Implementation: Feature implemented differently than Python

## Summary

### What's Complete
- ✅ Core REST communication layer
- ✅ All authentication methods
- ✅ Session management (KeepAlive, SessionID, reuse)
- ✅ Comprehensive configuration options (75+ parameters)
- ✅ Logging (config-based, custom, file)
- ✅ Connection pooling, SSL, proxies
- ✅ Basic TM1Service methods (Version, Metadata, Ping, etc.)
- ✅ Admin privilege checks
- ✅ Async operation management
- ✅ Compact JSON support

### What's Missing
- 🔴 Service sub-modules (20+ specialized services)
- 🔴 Data model classes (Cube, Dimension, Element, etc.)
- 🔴 Domain-specific operations (CubeService, ProcessService, etc.)
- 🔴 Utility classes (MDXUtils, case-insensitive collections)
- 🔴 Comprehensive test coverage for services

### Architecture Notes

**TM1py Architecture:**
- `TM1Service` acts as a facade that aggregates 20+ specialized service classes
- Each service (e.g., `CubeService`, `DimensionService`) is instantiated as a property of `TM1Service`
- Services take `RestService` as a dependency and provide domain-specific methods
- Model classes represent TM1 objects with validation and serialization logic

**TM1go Current Architecture:**
- `RestService` provides low-level HTTP communication
- `TM1Service` wraps `RestService` with basic helper methods
- Configuration is comprehensive and matches TM1py
- Authentication and session management are fully functional

**Recommended Next Steps:**
1. Implement high-priority service modules (Cubes, Dimensions, Processes, Views)
2. Create data model structs for TM1 objects
3. Add service-specific operations following TM1py patterns
4. Implement utility functions (MDX helpers, formatters)
5. Add comprehensive test coverage

## Example Usage Comparison

### TM1py
```python
from TM1py import TM1Service

with TM1Service(address='localhost', port=8882, user='admin', password='', ssl=True) as tm1:
    version = tm1.version
    cubes = tm1.cubes.get_all()
    data = tm1.cells.get_value(cube='Sales', elements=('2023', 'Q1', 'Revenue'))
```

### TM1go (Current)
```go
cfg := tm1.Config{Address: "localhost", Port: 8882, User: "admin", Password: "", SSL: true}
svc, _ := tm1.NewTM1Service(cfg)
defer svc.Close()

ctx := context.Background()
version, _ := svc.Version(ctx)
// Service modules not yet implemented
```

### TM1go (Future with Services)
```go
cfg := tm1.Config{Address: "localhost", Port: 8882, User: "admin", Password: "", SSL: true}
svc, _ := tm1.NewTM1Service(cfg)
defer svc.Close()

ctx := context.Background()
version, _ := svc.Version(ctx)
cubes, _ := svc.Cubes.GetAll(ctx)
data, _ := svc.Cells.GetValue(ctx, "Sales", []string{"2023", "Q1", "Revenue"})
```

## Conclusion

TM1go has a **solid foundation** with:
- Complete REST communication layer
- Full authentication and session management
- Comprehensive configuration matching TM1py
- Basic TM1Service functionality

**The main gap** is the lack of service sub-modules and data models, which provide the higher-level, domain-specific functionality that makes TM1py so powerful. Implementing these would give TM1go feature parity with TM1py.
