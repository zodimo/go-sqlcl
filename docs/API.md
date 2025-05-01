# API Documentation

This document provides detailed documentation for the go-sqlcl package API.

## Table of Contents

- [Client](#client)
  - [Configuration](#configuration)
  - [Connection](#connection)
  - [SQL Execution](#sql-execution)
  - [Commands](#commands)
- [Types](#types)
  - [QueryResult](#queryresult)
  - [ConnectionOptions](#connectionoptions)
  - [Error Handling](#error-handling)

## Client

The `Client` struct is the main entry point for interacting with SQLcl.

### Creating a Client

```go
// With default configuration
client, err := sqlcl.NewClient(nil)

// With custom configuration
config := &types.ClientConfig{
    SQLclPath:      "sql",            // Path to SQLcl executable
    Timeout:        30 * time.Second, // Default operation timeout
    ConnectTimeout: 15 * time.Second, // Connection timeout
    QueryTimeout:   60 * time.Second, // Query execution timeout
    ColorOutput:    false,            // Disable color output
    Format:         "table",          // Default output format
    LogLevel:       "info",           // Log level
}
client, err := sqlcl.NewClient(config)
```

### Configuration

The `ClientConfig` struct provides configuration options for the SQLcl client:

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| SQLclPath | string | Path to the SQLcl executable | "sql" (on PATH) |
| Timeout | time.Duration | Default timeout for operations | 30 seconds |
| LogLevel | string | Log level (debug, info, warn, error) | "info" |
| ColorOutput | bool | Whether to enable color output | false |
| StripNewlines | bool | Whether to strip newlines from output | true |
| Format | string | Output format (csv, json, table, etc.) | "table" |
| MaxBufferSize | int | Maximum buffer size for output | 1MB |
| ExtraEnv | []string | Extra environment variables | nil |
| WorkingDir | string | Working directory for the SQLcl process | "" |
| ConnectTimeout | time.Duration | Timeout for connection | 10 seconds |
| QueryTimeout | time.Duration | Timeout for queries | 60 seconds |

### Connection

The client provides two methods for connecting to a database:

#### Simple Connection

```go
// Connect with a simple connection string
err := client.Connect(ctx, "username/password@localhost:1521/orcl")
```

#### Connection with Options

```go
// Connect with detailed options
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

### SQL Execution

#### Execute SQL Statements

```go
// Execute a SQL query
result, err := client.ExecuteSQL(ctx, "SELECT * FROM employees")

// Process the results
for i, col := range result.Columns {
    fmt.Printf("Column %d: %s (%s)\n", i, col.Name, col.Type)
}

for _, row := range result.Rows {
    for i, val := range row.Values {
        fmt.Printf("%v ", val)
    }
    fmt.Println()
}
```

#### Execute SQL Files

```go
// Execute SQL from a file
result, err := client.ExecuteFile(ctx, "/path/to/script.sql")
```

### Commands

The client provides methods for executing SQLcl-specific commands:

#### General Command Execution

```go
// Execute a SQLcl command
cmd := types.Command{
    SQL:           "SHOW USER",
    ReturnResults: true,
}
result, err := client.ExecuteCommand(ctx, cmd)
```

#### Common Commands

```go
// Set options
client.SetPageSize(ctx, 50)
client.SetLineSize(ctx, 100)
client.SetFeedback(ctx, true)
client.SetTiming(ctx, true)
client.SetHeading(ctx, true)
client.SetOutputFormat(ctx, "json")

// Show options
output, err := client.ShowCommand(ctx, "PAGESIZE")

// Get help
helpText, err := client.GetHelp(ctx, "SELECT")

// Get command history
history, err := client.GetHistory(ctx)

// Describe database objects
tableDesc, err := client.DescribeObject(ctx, "EMPLOYEES")
for _, col := range tableDesc.Columns {
    fmt.Printf("%s %s %s\n", col.Name, col.Type, 
               col.Nullable ? "NULL" : "NOT NULL")
}
```

## Types

### QueryResult

The `QueryResult` struct represents the result of a SQL query:

```go
type QueryResult struct {
    Columns []Column  // Columns in the result
    Rows    []Row     // Rows in the result
    Summary string    // Summary line like "n rows selected"
    Success bool      // Whether the query was successful
    Message string    // Message from the execution (if any)
}

type Column struct {
    Name string      // Name of the column
    Type ColumnType  // Data type of the column
}

type Row struct {
    Values []interface{}  // Values in the row
}
```

### ConnectionOptions

The `ConnectionOptions` struct provides options for connecting to a database:

```go
type ConnectionOptions struct {
    Username    string  // Database username
    Password    string  // Database password
    ConnectStr  string  // Connection string (host:port/service)
    Wallet      string  // Path to Oracle wallet
    TNSAdmin    string  // Path to tnsnames.ora directory
    WalletPwd   string  // Wallet password
    Role        string  // Role to connect as (SYSDBA, SYSOPER, etc.)
    Proxy       string  // Proxy user to connect through
    ConnectType string  // Connection type (Basic, TNS, etc.)
}
```

### Error Handling

The package provides a custom `Error` type for handling Oracle errors:

```go
// Check for specific Oracle errors
if err != nil {
    if oraErr, ok := err.(*types.Error); ok {
        switch oraErr.Code {
        case types.ORA01017:
            log.Fatalf("Authentication failed: %v", err)
        case types.ORA12514:
            log.Fatalf("Database service not found: %v", err)
        default:
            log.Fatalf("Database error: %v", err)
        }
    } else {
        log.Fatalf("Failed to execute query: %v", err)
    }
}
```

Common Oracle error codes are predefined as constants in the `types` package:

| Constant | Error Code | Description |
|----------|------------|-------------|
| ORA00001 | ORA-00001 | Unique constraint violated |
| ORA00942 | ORA-00942 | Table or view does not exist |
| ORA01017 | ORA-01017 | Invalid username/password |
| ORA01031 | ORA-01031 | Insufficient privileges |
| ORA12514 | ORA-12514 | Service name not found |
| ORA12541 | ORA-12541 | No listener |
| ORA12545 | ORA-12545 | Connect failed, target host or object does not exist | 