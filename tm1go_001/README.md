> Notice: This is an older version of TM1go and is no longer used or developed.

<img src="logo.jpg" alt="alt text" width="400"/>

# TM1go: Go Library for IBM Planning Analytics TM1

TM1go is a Go library designed to interface with IBM Planning Analytics TM1, inspired by and closely mirroring the structure and functionality of the widely recognized tm1py Python library. This library serves as a bridge for Go developers, allowing them to interact with TM1 instances with the familiar paradigms and function names found in tm1py.

At its core, TM1go aims to provide a comprehensive, intuitive, and type-safe way to automate and manipulate TM1 objects and data directly from Go applications. By translating the flexible and powerful features of tm1py into Go, TM1go opens up a new realm of possibilities for TM1 automation, reporting, and integration tasks.

## Key Features

- Familiar Structure: Mimics the tm1py library’s structure, making it easy for users of tm1py to transition to using Go for their TM1-related operations.
- Comprehensive Functionality: Covers a wide range of TM1 REST API features, including cube data manipulation, process execution, and server configuration, among others.
- Type-Safe Interactions: Leverages Go's type system to ensure safer and more predictable interactions with the TM1 REST API.
- Ease of Use: Simplifies the complexity of interacting with TM1's REST API, providing a straightforward and productive developer experience.
- Whether you're building custom applications, automating TM1 processes, or integrating TM1 with other services, TM1go offers a robust and developer-friendly pathway to achieving your objectives. Join us in extending the capabilities of TM1 within the Go ecosystem.

## Installation 

    go get github.com/andreyea/tm1go

## Usage

    import "github.com/andreyea/tm1go"

#### Config for TM1 v11 local

    var config = tm1go.TM1ServiceConfig{
		BaseURL:  "https://localhost:8010",
		User:     "admin",
		Password: "apple",
	}


#### Config for TM1 v11 Cloud

	var config = tm1go.TM1ServiceConfig{
		BaseURL:           "<base-url-for-tm1-instance>",
		User:              "<user-name>",
		Namespace:         "LDAP",
		Password:          "<password>",
		SSL:               true,
		Verify:            true,
		AsyncRequestsMode: false,
	}

#### Config for TM1 v12

    var config = tm1go.TM1ServiceConfig{
        Address:  "<ibm-workspace-url>",
        IAMURL:   "<auth-url>",
        APIKey:   "<api-key>",
        Tenant:   "<tenant-id>",
        Database: "<tm1-database-name>",
        SSL:      true,
    }

#### Config for TM1 v12 SAAS

	var config = tm1go.TM1ServiceConfig{
		Address:  "<ibm-workspace-url>",
		APIKey:   "<api-key>",
		Tenant:   "<tenant-id>",
		Database: "<tm1-database-name>",
		SSL:      true,
	}

#### Create TM1 service. 

    var tm1Service = tm1go.NewTM1Service(config)

#### Get all cubes from the service

    cubes, err := tm1Service.CubeService.GetAll()
    if err != nil {
        fmt.Println(err)
    }else{
        fmt.Println(cubes)
    }
    
# Issues and Contribution

Feel free to create an issue if you find bugs.

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.