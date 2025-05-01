# Error Handling for go-sqlcl

This package provides custom error types and error handling functions for the go-sqlcl package.

## Error Types

The package defines several error types to help categorize different kinds of errors:

- `TypeConnectionError`: Error connecting to the database
- `TypeSQLError`: Error executing SQL
- `TypeProcessError`: Error with the SQLCL process
- `TypeTimeoutError`: Timeout error
- `TypeParserError`: Error parsing SQLCL output
- `TypeConfigError`: Error with configuration
- `TypeUnknownError`: Unknown error

## Oracle Error Codes

The package defines common Oracle error codes as constants:

- `ORA00001`: Unique constraint violated
- `ORA00942`: Table or view does not exist
- `ORA01017`: Invalid username/password
- `ORA01031`: Insufficient privileges
- `ORA12514`: Service name not found
- `ORA12541`: No listener
- `ORA12545`: Connect failed because target host or object does not exist
- `ORA03114`: Not connected to Oracle

## Error Creation

You can create custom errors using the provided constructor functions:

```go
// Create a connection error
err := errors.NewConnectionError("Failed to connect to database", cause)

// Create an SQL error
err := errors.NewSQLError("SQL execution failed", sqlStatement, cause)

// Create a process error
err := errors.NewProcessError("Failed to start SQLCL process", cause)

// Create a timeout error
err := errors.NewTimeoutError("Query execution timed out", sqlStatement, cause)

// Create a parser error
err := errors.NewParserError("Failed to parse SQLCL output", cause)

// Create a configuration error
err := errors.NewConfigError("Invalid configuration", cause)

// Create an unknown error
err := errors.NewUnknownError("An unknown error occurred", cause)
```

## Error Methods

The `Error` type implements the standard error interface and provides additional methods:

```go
// Get the error message
msg := err.Error()

// Check if the error is of a specific Oracle error code
if err.Is(errors.ORA00942) {
    // Handle table not found error
}

// Check if the error is of a specific type
if err.IsErrorType(errors.TypeConnectionError) {
    // Handle connection error
}

// Add SQL statement information to an error
err = err.WithSQL(sqlStatement)

// Add a cause to an error
err = err.WithCause(cause)
```

## Error Checking Functions

The package provides utility functions to check error types:

```go
// Check if the error is a connection error
if errors.IsConnectionError(err) {
    // Handle connection error
}

// Check if the error is an SQL error
if errors.IsSQLError(err) {
    // Handle SQL error
}

// Check if the error is a timeout error
if errors.IsTimeoutError(err) {
    // Handle timeout error
}

// Check if the error is an Oracle error with a specific code
if errors.IsOracleError(err, errors.ORA00942) {
    // Handle table not found error
}
```

## Parsing SQLCL Output

The package provides a function to parse SQLCL error output:

```go
// Parse SQLCL error output
err := errors.ParseSQLCLError(output)
```

This function extracts error information from the SQLCL output, including Oracle error codes, line numbers, and positions. 