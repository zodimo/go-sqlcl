# Basic Go-SQLCL Example

This example demonstrates how to use the go-sqlcl package to connect to an Oracle database and execute SQL queries.

## Features Demonstrated

- Creating a client with custom configuration
- Connecting to a database using environment variables
- Executing a simple SQL query
- Processing and displaying the query results
- Error handling
- Executing SQLCL-specific commands

## Prerequisites

1. Oracle SQLcl installed and available in your PATH
2. Go 1.18 or higher
3. Access to an Oracle database

## Running the Example

1. Set up the environment variables:

```bash
export DB_USER=your_username
export DB_PASSWORD=your_password
export DB_CONNECT=hostname:port/service_name
```

For example:

```bash
export DB_USER=hr
export DB_PASSWORD=hr
export DB_CONNECT=localhost:1521/XEPDB1
```

2. Run the example:

```bash
go run main.go
```

## Demo Mode

If you run the example without setting the required environment variables, it will enter "demo mode" which shows the code structure but doesn't attempt an actual database connection.

## Example Output

When run successfully, you should see output similar to:

```
2023/05/01 10:00:00 main.go:20: Starting go-sqlcl basic example
2023/05/01 10:00:01 main.go:62: Successfully connected to database
Query executed successfully!
Summary: 1 row selected

USER | SYSDATE
------------------------------
HR | 2023-05-01 10:00:01

--- Executing a SQLCL-specific command ---
USER is "HR"

Basic example completed successfully
```

## Error Handling

The example demonstrates how to handle different types of Oracle errors, including:
- Authentication failures (ORA-01017)
- Service name not found (ORA-12514)
- Other database errors 