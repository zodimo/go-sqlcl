# GO-SQLCL Documentation

This directory contains comprehensive documentation for the go-sqlcl package.

## Contents

- [API Documentation](API.md): Detailed documentation of the package API
- [Usage Guide](USAGE.md): Detailed usage instructions and patterns
- [Examples](EXAMPLES.md): Additional examples beyond the basic example

## Overview

go-sqlcl is a Go package that provides a scriptable API for Oracle's SQLcl command-line tool. The package aims to provide a scriptable API with 100% feature coverage of the SQLcl application while using the SQLcl REPL under the hood.

## Requirements

- Go 1.24 or higher
- Oracle SQLcl (available on your system path as "sql")

## Quick Start

Install the package:

```bash
go get github.com/zodimo/go-sqlcl
```

Basic example:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/zodimo/go-sqlcl/pkg/sqlcl"
)

func main() {
    // Create a new client
    client, err := sqlcl.NewClient(nil) // Use default configuration
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()
    
    // Connect to a database
    err = client.Connect(context.Background(), "username/password@localhost:1521/orcl")
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    
    // Execute a SQL query
    result, err := client.ExecuteSQL(context.Background(), "SELECT * FROM employees")
    if err != nil {
        log.Fatalf("Query failed: %v", err)
    }
    
    // Process the results
    for _, row := range result.Rows {
        fmt.Println(row)
    }
}
```

Refer to the other documentation files for more detailed information. 