# Examples

This document provides additional examples for using the go-sqlcl package beyond the basic examples in the README.

## Table of Contents

- [Connection Examples](#connection-examples)
- [Query Examples](#query-examples)
- [Transaction Examples](#transaction-examples)
- [Command Examples](#command-examples)
- [Complex Scenarios](#complex-scenarios)

## Connection Examples

### Connecting with System User

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/zodimo/go-sqlcl/pkg/sqlcl"
    "github.com/zodimo/go-sqlcl/pkg/types"
)

func main() {
    client, err := sqlcl.NewClient(nil)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()
    
    // Connect as SYSDBA
    opts := types.ConnectionOptions{
        Username:   "sys",
        Password:   "password",
        ConnectStr: "localhost:1521/XEPDB1",
        Role:       "SYSDBA",
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()
    
    err = client.ConnectWithOptions(ctx, opts)
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    
    fmt.Println("Connected as SYSDBA")
}
```

### Connecting with Oracle Wallet

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/zodimo/go-sqlcl/pkg/sqlcl"
    "github.com/zodimo/go-sqlcl/pkg/types"
)

func main() {
    client, err := sqlcl.NewClient(nil)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()
    
    // Connect with wallet
    opts := types.ConnectionOptions{
        Username:   "username",
        Password:   "password",
        ConnectStr: "localhost:1521/orcl",
        Wallet:     "/path/to/wallet",
        TNSAdmin:   "/path/to/tnsadmin",
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()
    
    err = client.ConnectWithOptions(ctx, opts)
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    
    fmt.Println("Connected using wallet")
}
```

## Query Examples

### Query with Parameter Substitution

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/zodimo/go-sqlcl/pkg/sqlcl"
    "github.com/zodimo/go-sqlcl/pkg/types"
)

func main() {
    client, err := sqlcl.NewClient(nil)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Connect to database
    err = client.Connect(ctx, "username/password@localhost:1521/orcl")
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    
    // Execute a query with bind variables
    cmd := types.Command{
        SQL: "SELECT * FROM employees WHERE department_id = :dept_id AND salary > :min_salary",
        Parameters: map[string]string{
            "dept_id":    "10",
            "min_salary": "5000",
        },
        ReturnResults: true,
    }
    
    result, err := client.ExecuteCommand(ctx, cmd)
    if err != nil {
        log.Fatalf("Query failed: %v", err)
    }
    
    // Print results
    for _, row := range result.Rows {
        fmt.Println(row.Values)
    }
}
```

### Handling LOBs

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
    client, err := sqlcl.NewClient(nil)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Connect to database
    err = client.Connect(ctx, "username/password@localhost:1521/orcl")
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    
    // Set long format to display CLOBs
    err = client.SetCommand(ctx, "long", "999999")
    if err != nil {
        log.Fatalf("Failed to set LONG format: %v", err)
    }
    
    // Query with CLOB data
    result, err := client.ExecuteSQL(ctx, "SELECT id, name, description FROM documents")
    if err != nil {
        log.Fatalf("Query failed: %v", err)
    }
    
    // Process results
    for i, row := range result.Rows {
        id := row.Values[0]
        name := row.Values[1]
        description := row.Values[2]
        
        fmt.Printf("Document %d:\n", i+1)
        fmt.Printf("  ID: %v\n", id)
        fmt.Printf("  Name: %v\n", name)
        fmt.Printf("  Description: %v\n", description)
        fmt.Println("---")
    }
}
```

## Transaction Examples

### Managing Transactions

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
    client, err := sqlcl.NewClient(nil)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()
    
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()
    
    // Connect to database
    err = client.Connect(ctx, "username/password@localhost:1521/orcl")
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    
    // Start transaction
    _, err = client.ExecuteSQL(ctx, "BEGIN")
    if err != nil {
        log.Fatalf("Failed to start transaction: %v", err)
    }
    
    // Perform operations within transaction
    _, err = client.ExecuteSQL(ctx, "INSERT INTO employees (id, name, salary) VALUES (101, 'John Doe', 5000)")
    if err != nil {
        // Rollback on error
        _, rollbackErr := client.ExecuteSQL(ctx, "ROLLBACK")
        if rollbackErr != nil {
            log.Printf("Rollback failed: %v", rollbackErr)
        }
        log.Fatalf("Insert failed: %v", err)
    }
    
    _, err = client.ExecuteSQL(ctx, "UPDATE departments SET manager_id = 101 WHERE id = 10")
    if err != nil {
        // Rollback on error
        _, rollbackErr := client.ExecuteSQL(ctx, "ROLLBACK")
        if rollbackErr != nil {
            log.Printf("Rollback failed: %v", rollbackErr)
        }
        log.Fatalf("Update failed: %v", err)
    }
    
    // Commit transaction
    _, err = client.ExecuteSQL(ctx, "COMMIT")
    if err != nil {
        log.Fatalf("Commit failed: %v", err)
    }
    
    fmt.Println("Transaction completed successfully")
}
```

## Command Examples

### Advanced SQL Command Features

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/zodimo/go-sqlcl/pkg/sqlcl"
    "github.com/zodimo/go-sqlcl/pkg/types"
)

func main() {
    client, err := sqlcl.NewClient(nil)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Connect to database
    err = client.Connect(ctx, "username/password@localhost:1521/orcl")
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    
    // Set output formatting
    err = client.SetOutputFormat(ctx, "json")
    if err != nil {
        log.Printf("Failed to set output format: %v", err)
    }
    
    // Enable SQL timing
    err = client.SetTiming(ctx, true)
    if err != nil {
        log.Printf("Failed to enable timing: %v", err)
    }
    
    // Execute command with custom timeout
    cmd := types.Command{
        SQL:           "SELECT * FROM employees",
        Timeout:       5 * time.Second,
        Format:        "json",
        IgnoreErrors:  false,
        ReturnResults: true,
    }
    
    result, err := client.ExecuteCommand(ctx, cmd)
    if err != nil {
        log.Fatalf("Command execution failed: %v", err)
    }
    
    fmt.Println(result.Summary)
}
```

### Using SQLcl-specific Commands

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
    client, err := sqlcl.NewClient(nil)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Connect to database
    err = client.Connect(ctx, "username/password@localhost:1521/orcl")
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    
    // Describe a table
    tableDesc, err := client.DescribeObject(ctx, "EMPLOYEES")
    if err != nil {
        log.Fatalf("Failed to describe table: %v", err)
    }
    
    fmt.Printf("Table: %s (%s)\n", tableDesc.ObjectName, tableDesc.ObjectType)
    for _, col := range tableDesc.Columns {
        nullableStr := "NOT NULL"
        if col.Nullable {
            nullableStr = "NULL"
        }
        
        defaultStr := ""
        if col.Default != "" {
            defaultStr = fmt.Sprintf("DEFAULT %s", col.Default)
        }
        
        fmt.Printf("  %s %s %s %s\n", col.Name, col.Type, nullableStr, defaultStr)
    }
    
    // Get command history
    history, err := client.GetHistory(ctx)
    if err != nil {
        log.Fatalf("Failed to get command history: %v", err)
    }
    
    fmt.Println("\nCommand History:")
    for i, cmd := range history {
        fmt.Printf("%d: %s\n", i+1, cmd)
    }
    
    // Show current settings
    pageSize, err := client.ShowCommand(ctx, "PAGESIZE")
    if err != nil {
        log.Printf("Failed to get pagesize: %v", err)
    } else {
        fmt.Printf("Current pagesize: %s\n", pageSize)
    }
    
    lineSize, err := client.ShowCommand(ctx, "LINESIZE")
    if err != nil {
        log.Printf("Failed to get linesize: %v", err)
    } else {
        fmt.Printf("Current linesize: %s\n", lineSize)
    }
}
```

## Complex Scenarios

### Data Export Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "strings"
    "time"
    
    "github.com/zodimo/go-sqlcl/pkg/sqlcl"
)

func main() {
    client, err := sqlcl.NewClient(nil)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()
    
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()
    
    // Connect to database
    err = client.Connect(ctx, "username/password@localhost:1521/orcl")
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    
    // Set CSV format
    err = client.SetOutputFormat(ctx, "csv")
    if err != nil {
        log.Printf("Failed to set output format: %v", err)
    }
    
    // Query data for export
    result, err := client.ExecuteSQL(ctx, "SELECT * FROM employees")
    if err != nil {
        log.Fatalf("Query failed: %v", err)
    }
    
    // Create export file
    file, err := os.Create("employees_export.csv")
    if err != nil {
        log.Fatalf("Failed to create file: %v", err)
    }
    defer file.Close()
    
    // Write header
    headers := make([]string, len(result.Columns))
    for i, col := range result.Columns {
        headers[i] = col.Name
    }
    file.WriteString(strings.Join(headers, ",") + "\n")
    
    // Write data
    for _, row := range result.Rows {
        values := make([]string, len(row.Values))
        for i, val := range row.Values {
            if val == nil {
                values[i] = ""
            } else {
                values[i] = fmt.Sprintf("%v", val)
            }
        }
        file.WriteString(strings.Join(values, ",") + "\n")
    }
    
    fmt.Printf("Exported %d rows to employees_export.csv\n", len(result.Rows))
}
```

### Schema Comparison Tool

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/zodimo/go-sqlcl/pkg/sqlcl"
    "github.com/zodimo/go-sqlcl/pkg/types"
)

func main() {
    // Create two clients for different environments
    sourceClient, err := sqlcl.NewClient(nil)
    if err != nil {
        log.Fatalf("Failed to create source client: %v", err)
    }
    defer sourceClient.Close()
    
    targetClient, err := sqlcl.NewClient(nil)
    if err != nil {
        log.Fatalf("Failed to create target client: %v", err)
    }
    defer targetClient.Close()
    
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()
    
    // Connect to source database
    err = sourceClient.Connect(ctx, "username/password@source-db:1521/orcl")
    if err != nil {
        log.Fatalf("Failed to connect to source: %v", err)
    }
    
    // Connect to target database
    err = targetClient.Connect(ctx, "username/password@target-db:1521/orcl")
    if err != nil {
        log.Fatalf("Failed to connect to target: %v", err)
    }
    
    // Get tables from source
    sourceResult, err := sourceClient.ExecuteSQL(ctx, 
        "SELECT table_name FROM user_tables ORDER BY table_name")
    if err != nil {
        log.Fatalf("Failed to get source tables: %v", err)
    }
    
    // Get tables from target
    targetResult, err := targetClient.ExecuteSQL(ctx, 
        "SELECT table_name FROM user_tables ORDER BY table_name")
    if err != nil {
        log.Fatalf("Failed to get target tables: %v", err)
    }
    
    // Extract table names
    sourceTables := make(map[string]bool)
    for _, row := range sourceResult.Rows {
        tableName := fmt.Sprintf("%v", row.Values[0])
        sourceTables[tableName] = true
    }
    
    targetTables := make(map[string]bool)
    for _, row := range targetResult.Rows {
        tableName := fmt.Sprintf("%v", row.Values[0])
        targetTables[tableName] = true
    }
    
    // Compare tables
    fmt.Println("Schema Comparison Results:")
    fmt.Println("=========================")
    
    fmt.Println("\nTables in source but not in target:")
    for table := range sourceTables {
        if !targetTables[table] {
            fmt.Println("  " + table)
        }
    }
    
    fmt.Println("\nTables in target but not in source:")
    for table := range targetTables {
        if !sourceTables[table] {
            fmt.Println("  " + table)
        }
    }
    
    // Compare common tables structure (simplified example)
    fmt.Println("\nColumn differences in common tables:")
    for table := range sourceTables {
        if targetTables[table] {
            compareTableStructure(ctx, sourceClient, targetClient, table)
        }
    }
}

func compareTableStructure(ctx context.Context, sourceClient, targetClient *sqlcl.Client, tableName string) {
    // Describe table in source
    sourceDesc, err := sourceClient.DescribeObject(ctx, tableName)
    if err != nil {
        log.Printf("Failed to describe source table %s: %v", tableName, err)
        return
    }
    
    // Describe table in target
    targetDesc, err := targetClient.DescribeObject(ctx, tableName)
    if err != nil {
        log.Printf("Failed to describe target table %s: %v", tableName, err)
        return
    }
    
    // Create maps for comparing columns
    sourceColumns := make(map[string]struct{
        Type     string
        Nullable bool
    })
    
    for _, col := range sourceDesc.Columns {
        sourceColumns[col.Name] = struct{
            Type     string
            Nullable bool
        }{
            Type:     col.Type,
            Nullable: col.Nullable,
        }
    }
    
    // Compare columns
    for _, col := range targetDesc.Columns {
        sourceCol, exists := sourceColumns[col.Name]
        if !exists {
            fmt.Printf("  Table %s: Column %s exists in target but not in source\n", 
                tableName, col.Name)
            continue
        }
        
        if sourceCol.Type != col.Type {
            fmt.Printf("  Table %s: Column %s has different type (source: %s, target: %s)\n", 
                tableName, col.Name, sourceCol.Type, col.Type)
        }
        
        if sourceCol.Nullable != col.Nullable {
            nullableSource := "NOT NULL"
            if sourceCol.Nullable {
                nullableSource = "NULL"
            }
            
            nullableTarget := "NOT NULL"
            if col.Nullable {
                nullableTarget = "NULL"
            }
            
            fmt.Printf("  Table %s: Column %s has different nullability (source: %s, target: %s)\n", 
                tableName, col.Name, nullableSource, nullableTarget)
        }
        
        delete(sourceColumns, col.Name)
    }
    
    // Check for columns in source but not target
    for colName := range sourceColumns {
        fmt.Printf("  Table %s: Column %s exists in source but not in target\n", 
            tableName, colName)
    }
}
```

### Batch Processing

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
    client, err := sqlcl.NewClient(nil)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()
    
    ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
    defer cancel()
    
    // Connect to database
    err = client.Connect(ctx, "username/password@localhost:1521/orcl")
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    
    // Get data for batch processing
    result, err := client.ExecuteSQL(ctx, "SELECT employee_id FROM employees")
    if err != nil {
        log.Fatalf("Query failed: %v", err)
    }
    
    // Get employee IDs
    employeeIDs := make([]interface{}, 0, len(result.Rows))
    for _, row := range result.Rows {
        employeeIDs = append(employeeIDs, row.Values[0])
    }
    
    fmt.Printf("Processing %d employees in batches\n", len(employeeIDs))
    
    // Process in batches of 25
    batchSize := 25
    totalBatches := (len(employeeIDs) + batchSize - 1) / batchSize
    
    for i := 0; i < totalBatches; i++ {
        start := i * batchSize
        end := (i + 1) * batchSize
        if end > len(employeeIDs) {
            end = len(employeeIDs)
        }
        
        batch := employeeIDs[start:end]
        fmt.Printf("Processing batch %d/%d (%d employees)\n", i+1, totalBatches, len(batch))
        
        // Start transaction for batch
        _, err := client.ExecuteSQL(ctx, "BEGIN")
        if err != nil {
            log.Printf("Failed to start transaction: %v", err)
            continue
        }
        
        success := true
        // Process each employee in the batch
        for _, id := range batch {
            // Example: Update salary for each employee
            updateSQL := fmt.Sprintf(
                "UPDATE employees SET salary = salary * 1.05 WHERE employee_id = %v", id)
            
            _, err := client.ExecuteSQL(ctx, updateSQL)
            if err != nil {
                log.Printf("Failed to update employee %v: %v", id, err)
                success = false
                break
            }
        }
        
        // Commit or rollback the batch
        if success {
            _, err := client.ExecuteSQL(ctx, "COMMIT")
            if err != nil {
                log.Printf("Failed to commit batch %d: %v", i+1, err)
            } else {
                fmt.Printf("Batch %d/%d completed successfully\n", i+1, totalBatches)
            }
        } else {
            _, err := client.ExecuteSQL(ctx, "ROLLBACK")
            if err != nil {
                log.Printf("Failed to rollback batch %d: %v", i+1, err)
            } else {
                fmt.Printf("Batch %d/%d rolled back due to errors\n", i+1, totalBatches)
            }
        }
    }
    
    fmt.Println("Batch processing completed")
} 