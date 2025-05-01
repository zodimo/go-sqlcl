# GO SQLCL 

A Go package that provides a scriptable API for Oracle's SQLcl command-line tool.

## Overview

- Language: Go 1.24
- Package name: github.com/zodimo/go-sqlcl

The sqlcl application from Oracle is a command-line version of Oracle SQL Developer. It primarily runs a REPL (Read-Eval-Print Loop) in the terminal, which makes it challenging to use from a scripting perspective.

This package aims to provide a scriptable API with 100% feature coverage of the SQLcl application while using the SQLcl REPL under the hood.

## Requirements

- Go 1.24 or higher
- Oracle SQLcl (available on your system path as "sql")

## Installation

```bash
go get github.com/zodimo/go-sqlcl
```

## Basic Usage

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
    client, err := sqlcl.NewClient()
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

## Documentation

For detailed documentation, see the [docs](./docs) directory.

## Examples

For more examples, see the [examples](./examples) directory.

## SQLcl Documentation

For more information about Oracle SQLcl, see the official documentation:
- [Oracle SQLcl Documentation](https://docs.oracle.com/en/database/oracle/sql-developer-command-line/25.1/sqcug/index.html)


