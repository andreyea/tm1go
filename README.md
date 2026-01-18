<img src="logo.jpg" alt="TM1go" width="400"/>

# TM1go

TM1go is a Go client for IBM Planning Analytics / TM1 REST API. It provides a high-level service facade (`TM1Service`) with focused sub-services for common TM1 domains (dimensions, hierarchies, elements, subsets, processes, and cells), plus a lower-level REST client for custom calls.

This library mirrors many concepts from TM1py while embracing Go’s type system and patterns.

## Installation

```bash
go get github.com/andreyea/tm1go
```

## Quick start

```go
package main

import (
	"context"
	"fmt"

	"github.com/andreyea/tm1go/pkg/tm1"
)

func main() {
	cfg := tm1.Config{
		Address:  "localhost",
		Port:     8882,
		SSL:      true,
		User:     "admin",
		Password: "apple",
		Logging:  true,
	}

	client, err := tm1.NewTM1Service(cfg)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ctx := context.Background()

	dims, err := client.Dimensions.GetAllNames(ctx, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Dimensions: %v\n", dims)
}
```

## Connecting to different TM1 deployments

> Tip: You can also set `BaseURL` directly if you already know the full REST root URL.

### TM1 v11 (on‑prem)

```go
cfg := tm1.Config{
	Address:  "localhost",
	Port:     8010,
	SSL:      true,
	User:     "admin",
	Password: "apple",
}
```

### TM1 v11 with CAM / LDAP namespace

```go
cfg := tm1.Config{
	Address:   "tm1.company.com",
	Port:      8010,
	SSL:       true,
	User:      "jdoe",
	Password:  "secret",
	Namespace: "LDAP",
}
```

### Planning Analytics v12 (IBM Cloud / SaaS)

```go
cfg := tm1.Config{
	Address:  "us-east-1.planninganalytics.saas.ibm.com",
	Tenant:   "YOUR_TENANT",
	Database: "YOUR_DATABASE",
	APIKey:   "YOUR_IBM_CLOUD_API_KEY",
	SSL:      true,
}
```

## Common tasks

### Execute an MDX query

```go
mdx := "SELECT {[Year].[2024]} ON ROWS, {[Measure].[Revenue]} ON COLUMNS FROM [Sales]"
cells, err := client.Cells.ExecuteMDX(ctx, mdx, nil, "")
if err != nil {
	panic(err)
}
for coords, cell := range cells {
	fmt.Printf("%s => %v\n", coords, cell["Value"])
}
```

### Write multiple values

```go
coords := [][]string{
	{"q1", "e1", "Value"},
	{"q1", "e2", "Value"},
	{"q1", "e3", "Value"},
}
values := []interface{}{1, 2, 3}
dimensions := []string{"d1", "d2", "Measure"}

if err := client.Cells.WriteValuesByCoords(ctx, "3D", coords, values, dimensions, ""); err != nil {
	panic(err)
}
```

### Execute a process

```go
params := map[string]interface{}{
    "pLogOutput":           "0",
    "pStrictErrorHandling": "0",
    "pWaitSec":             "1",
}

ok, status, msg, err := client.Processes.ExecuteWithReturn(ctx, "}bedrock.server.wait", params, nil, false)
if err != nil {
    panic(err)
}
fmt.Printf("Success=%v Status=%s Message=%s\n", ok, status, msg)
```

### Work with dimensions and hierarchies

```go
dim, err := client.Dimensions.Get(ctx, "Account")
if err != nil {
	panic(err)
}
fmt.Println("Hierarchies:", dim.HierarchyNames())

elements, err := client.Elements.GetElementNames(ctx, "Account", "Account")
if err != nil {
	panic(err)
}
fmt.Printf("Elements: %d\n", len(elements))
```

### Subsets (public or private)

```go
names, err := client.Subsets.GetAllNames(ctx, "Account", "Account", false)
if err != nil {
	panic(err)
}
fmt.Printf("Public subsets: %v\n", names)
```

## Service overview

`TM1Service` exposes:

- `Dimensions` – create/update/delete dimensions and hierarchies
- `Hierarchies` – manage hierarchies, elements, and attributes
- `Elements` – element CRUD, attributes, and edge operations
- `Subsets` – dynamic/static subset operations
- `Processes` – execute, debug, and manage TurboIntegrator processes
- `Cells` – cell read/write, MDX execution, views, and cellsets

You can also access the underlying REST client with `client.Rest()` for custom endpoints.

## Notes

- `Close()` logs out and closes idle connections (unless `KeepAlive` is `true`).
- `AsyncRequestsMode` enables TM1 async requests and automatically polls completion.
- Windows Integrated Authentication is not implemented yet (returns an error if enabled).

## Examples

Coming soon

