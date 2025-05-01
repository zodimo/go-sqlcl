# Usage Guide

This document provides detailed instructions on how to use the go-sqlcl package.

## Table of Contents

- [Installation](#installation)
- [Basic Usage](#basic-usage)
- [Connection](#connection)
- [SQL Execution](#sql-execution)
- [Error Handling](#error-handling)
- [Working with Results](#working-with-results)
- [SQLcl-specific Commands](#sqlcl-specific-commands)
- [Advanced Configuration](#advanced-configuration)
- [Best Practices](#best-practices)

## Installation

### Prerequisites

Before using the go-sqlcl package, ensure you have:

1. **Go 1.24 or higher** installed on your system
2. **Oracle SQLcl** installed and available on your system PATH as "sql"

You can check your Go version with:
```bash
go version
```

You can check if SQLcl is available on your PATH with:
```bash
sql -v
```

### Installing the Package

Install the package using Go modules:

```bash
go get github.com/zodimo/go-sqlcl
```

## Basic Usage

The following example demonstrates a basic usage of the package:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/zodimo/go-sqlcl/pkg/sqlcl"
)

func main() {
    // Create a new client with default configuration
    client, err := sqlcl.NewClient(nil)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close() // Always close the client to free resources
    
    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Connect to a database
    err = client.Connect(ctx, "username/password@localhost:1521/orcl")
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    
    // Execute a SQL query
    result, err := client.ExecuteSQL(ctx, "SELECT * FROM employees WHERE rownum <= 10")
    if err != nil {
        log.Fatalf("Query failed: %v", err)
    }
    
    // Print column names
    for _, col := range result.Columns {
        fmt.Printf("%s\t", col.Name)
    }
    fmt.Println()
    
    // Print rows
    for _, row := range result.Rows {
        for _, val := range row.Values {
            fmt.Printf("%v\t", val)
        }
        fmt.Println()
    }
}
```

## Connection

### Simple Connection

For basic usage, you can connect using a simple connection string:

```go
// Format: username/password@host:port/service
err := client.Connect(ctx, "system/password@localhost:1521/XEPDB1")
```

### Connection with Options

For more advanced connections, use the `ConnectWithOptions` method:

```go
connectionOpts := types.ConnectionOptions{
    Username:   "system",
    Password:   "password",
    ConnectStr: "localhost:1521/XEPDB1",
    // Optional settings:
    TNSAdmin:   "/path/to/tnsadmin",
    Wallet:     "/path/to/wallet",
    Role:       "SYSDBA",
}
err := client.ConnectWithOptions(ctx, connectionOpts)
```

### Connecting with Environment Variables

Using environment variables for sensitive connection information:

```go
connectionOpts := types.ConnectionOptions{
    Username:   os.Getenv("DB_USER"),
    Password:   os.Getenv("DB_PASSWORD"),
    ConnectStr: os.Getenv("DB_CONNECT"),
}

// Validate the connection details
if connectionOpts.Username == "" || connectionOpts.Password == "" || connectionOpts.ConnectStr == "" {
    log.Fatal("Missing required environment variables")
}

err = client.ConnectWithOptions(ctx, connectionOpts)
```

## SQL Execution

### Executing SQL Statements

```go
// Simple query
result, err := client.ExecuteSQL(ctx, "SELECT * FROM employees")

// Query with timeout
queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
result, err := client.ExecuteSQL(queryCtx, "SELECT * FROM large_table")

// DML statements
result, err := client.ExecuteSQL(ctx, "INSERT INTO employees VALUES (1, 'John Doe', 5000)")
fmt.Println(result.Summary) // Will print something like "1 row inserted"
```

### Executing SQL Files

```go
// Execute a SQL script file
result, err := client.ExecuteFile(ctx, "/path/to/script.sql")
```

### Executing Multiple Statements

```go
// Multiple statements with semicolons
sql := `
CREATE TABLE temp_employees AS SELECT * FROM employees;
UPDATE temp_employees SET salary = salary * 1.1;
SELECT COUNT(*) FROM temp_employees;
`
result, err := client.ExecuteSQL(ctx, sql)
```

## Error Handling

### Handling Connection Errors

```go
err = client.Connect(ctx, "username/password@localhost:1521/orcl")
if err != nil {
    if oraErr, ok := err.(*types.Error); ok {
        switch oraErr.Code {
        case types.ORA01017:
            log.Fatalf("Invalid username or password")
        case types.ORA12514:
            log.Fatalf("Service name not found")
        case types.ORA12541:
            log.Fatalf("No listener: check if the database is running")
        default:
            log.Fatalf("Oracle error: %v", err)
        }
    } else {
        log.Fatalf("Connection error: %v", err)
    }
}
```

### Handling Query Errors

```go
result, err := client.ExecuteSQL(ctx, "SELECT * FROM non_existent_table")
if err != nil {
    if oraErr, ok := err.(*types.Error); ok {
        if oraErr.Code == types.ORA00942 {
            log.Fatalf("Table doesn't exist")
        } else {
            log.Fatalf("Oracle error: %v", err)
        }
    } else {
        log.Fatalf("Query error: %v", err)
    }
}
```

## Working with Results

### Processing Query Results

```go
result, err := client.ExecuteSQL(ctx, "SELECT employee_id, first_name, last_name, salary FROM employees")
if err != nil {
    log.Fatalf("Query failed: %v", err)
}

// Print column names and types
for _, col := range result.Columns {
    fmt.Printf("%s (%s)\t", col.Name, col.Type)
}
fmt.Println()

// Print rows
for _, row := range result.Rows {
    for i, val := range row.Values {
        fmt.Printf("%v\t", val)
    }
    fmt.Println()
}

// Print summary information
fmt.Println(result.Summary) // e.g., "14 rows selected"
```

### Handling NULL Values

NULL values are represented as `nil` in the result rows:

```go
result, err := client.ExecuteSQL(ctx, "SELECT employee_id, manager_id FROM employees")
if err != nil {
    log.Fatalf("Query failed: %v", err)
}

for _, row := range result.Rows {
    employeeID := row.Values[0]
    managerID := row.Values[1]
    
    if managerID == nil {
        fmt.Printf("Employee %v has no manager\n", employeeID)
    } else {
        fmt.Printf("Employee %v has manager %v\n", employeeID, managerID)
    }
}
```

## SQLcl-specific Commands

### Setting Options

```go
// Set page size
err := client.SetPageSize(ctx, 50)

// Set line size
err := client.SetLineSize(ctx, 120)

// Enable/disable feedback
err := client.SetFeedback(ctx, true)

// Enable/disable timing
err := client.SetTiming(ctx, true)

// Enable/disable headings
err := client.SetHeading(ctx, true)

// Set SQL terminator
err := client.SetSQLTerminator(ctx, "/")

// Set output format
err := client.SetOutputFormat(ctx, "json")
```

### Executing Commands

```go
// Execute a command using the Command struct
cmd := types.Command{
    SQL:           "SHOW USER",
    ReturnResults: true,
}
result, err := client.ExecuteCommand(ctx, cmd)
fmt.Println(result.Output)

// Using command-specific methods
output, err := client.ShowCommand(ctx, "PAGESIZE")
fmt.Println(output)

// Getting help
helpText, err := client.GetHelp(ctx, "SELECT")
fmt.Println(helpText)

// Describe an object
tableDesc, err := client.DescribeObject(ctx, "EMPLOYEES")
fmt.Printf("Table: %s\n", tableDesc.ObjectName)
for _, col := range tableDesc.Columns {
    nullableStr := "NOT NULL"
    if col.Nullable {
        nullableStr = "NULL"
    }
    fmt.Printf("  %s %s %s\n", col.Name, col.Type, nullableStr)
}
```

## Advanced Configuration

### Custom Configuration

```go
// Create a configuration with custom settings
config := &types.ClientConfig{
    SQLclPath:      "/custom/path/to/sql",  // Custom path to SQLcl
    Timeout:        60 * time.Second,       // Extended default timeout
    ConnectTimeout: 30 * time.Second,       // Extended connection timeout
    QueryTimeout:   120 * time.Second,      // Extended query timeout
    ColorOutput:    false,                  // Disable color output
    Format:         "json",                 // Default to JSON output
    LogLevel:       "debug",                // Use debug log level
    MaxBufferSize:  2 * 1024 * 1024,        // 2MB buffer size
    ExtraEnv:       []string{"NLS_LANG=AMERICAN_AMERICA.AL32UTF8"},
    WorkingDir:     "/path/to/working/dir", // Custom working directory
}

// Create client with custom configuration
client, err := sqlcl.NewClient(config)
```

## Best Practices

### Resource Management

Always close the client when done:

```go
client, err := sqlcl.NewClient(nil)
if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}
defer client.Close() // Important to release resources
```

### Timeout Management

Use contexts with timeouts to avoid hanging operations:

```go
// Context with timeout for connection
connectCtx, cancelConnect := context.WithTimeout(context.Background(), 15*time.Second)
defer cancelConnect()
err := client.Connect(connectCtx, "username/password@localhost:1521/orcl")

// Context with timeout for query
queryCtx, cancelQuery := context.WithTimeout(context.Background(), 30*time.Second)
defer cancelQuery()
result, err := client.ExecuteSQL(queryCtx, "SELECT * FROM large_table")
```

### Environment Variables

Store sensitive information in environment variables rather than hardcoding:

```go
connectionOpts := types.ConnectionOptions{
    Username:   os.Getenv("DB_USER"),
    Password:   os.Getenv("DB_PASSWORD"),
    ConnectStr: os.Getenv("DB_CONNECT"),
}
```

### Error Handling

Handle errors appropriately for each context:

```go
if err != nil {
    if oraErr, ok := err.(*types.Error); ok {
        // Handle Oracle-specific errors
        handleOracleError(oraErr)
    } else if errors.Is(err, context.DeadlineExceeded) {
        // Handle timeout errors
        handleTimeout(err)
    } else {
        // Handle other errors
        handleGenericError(err)
    }
}
```

### Logging

Configure logging appropriately:

```go
// Configure logging
log.SetFlags(log.LstdFlags | log.Lshortfile)

// Create client with debug logging
config := &types.ClientConfig{
    LogLevel: "debug",
}
client, err := sqlcl.NewClient(config)
```

### Security Considerations

- Don't hardcode credentials in your code
- Use connection pooling in production environments
- Consider using wallets for secure connections
- Implement proper query parameter validation 