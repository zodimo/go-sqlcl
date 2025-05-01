# Development Guidelines for go-sqlcl

## Project Overview

- **Purpose**: Provide a scriptable API for Oracle's SQLCL application in Go
- **Target language**: Go 1.24
- **Package name**: github.com/zodimo/go-sqlcl
- **Goal**: 100% feature coverage of Oracle SQLCL while using the SQLCL REPL under the hood

## Project Architecture

### Directory Structure

- `/cmd`: CLI applications and entry points
- `/internal`: Private package code
- `/pkg`: Public API packages
- `/test`: Integration tests and test resources
- `/docs`: Documentation
- `/examples`: Example code demonstrating library usage

### Core Components

- **SQLCL Process Manager**: Handles starting, stopping, and monitoring the SQLCL process
- **Command Executor**: Sends commands to SQLCL and processes responses
- **Output Parser**: Parses SQLCL output into structured data
- **API Layer**: Provides a clean Go API for each SQLCL command
- **Error Handler**: Manages and recovers from SQLCL errors

## Coding Standards

### Naming Conventions

- Use `CamelCase` for exported names and `camelCase` for non-exported names
- Prefix interfaces with `I` (e.g., `ICommandExecutor`)
- Use descriptive, clear names that convey purpose
- Avoid abbreviations except for standard ones (e.g., `SQL`, `HTTP`)

### Code Organization

- One API function per SQLCL command
- Group related commands in separate files
- Use interfaces to enable mocking for testing
- Limit file size to 500 lines maximum

### Error Handling

- Return errors rather than using panic
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Create custom error types for specific error scenarios
- Do not discard errors; always handle or return them

### Documentation

- Document all exported functions, types, and constants
- Include examples in godoc format
- Maintain a separate document for each major component
- Write documentation assuming the reader has no knowledge of SQLCL

## Feature Implementation Standards

### Command Wrapping

- Create one Go function per SQLCL command
- Use function parameters to represent command options
- Return structured data instead of raw output strings
- Maintain consistent parameter naming across related commands

### SQLCL Process Management

- Support configuration of SQLCL path and Java options
- Handle process termination gracefully
- Implement timeout mechanisms for all commands
- Support resource cleanup on errors

### Output Processing

- Parse all output into structured data types
- Preserve raw output for debugging purposes
- Handle partial and malformed output gracefully
- Support streaming for large result sets

### Interactive Mode

- Provide both synchronous and asynchronous APIs
- Handle interactive prompts automatically when possible
- Supply callback functions for user input when required
- Maintain state between commands in a session

## Testing Requirements

### Unit Tests

- Achieve at least 80% code coverage with unit tests
- Mock the SQLCL process for unit testing
- Test error handling and edge cases
- Use table-driven tests for commands with multiple options

### Integration Tests

- Test against a real SQLCL installation
- Create fixtures for test data
- Test the full command set against a test database
- Include performance tests for critical operations

### Continuous Integration

- Run tests on each PR
- Enforce code coverage requirements
- Test against multiple Go versions
- Test against multiple SQLCL versions

## Third-Party Dependencies

### Approved Libraries

- Standard library preferred for most functionality
- Use `github.com/stretchr/testify` for testing
- Use `github.com/spf13/cobra` for CLI applications if needed
- Minimize external dependencies

### Dependency Management

- Use Go modules
- Pin dependency versions
- Review all dependencies for license compatibility
- Document all dependencies and their purposes

## Workflow Standards

### Development Workflow

- Use feature branches
- Submit changes via pull requests
- Follow conventional commits (`feat:`, `fix:`, `docs:`, etc.)
- Update documentation with code changes

### Release Workflow

- Semantic versioning
- Create release notes for each version
- Build and publish binaries for multiple platforms
- Tag releases in Git

## File Interaction Standards

### Configuration

- Use environment variables for runtime configuration
- Support configuration files in YAML or JSON
- Validate all configuration parameters
- Provide sensible defaults for all settings

### Concurrent Access

- Make all API functions safe for concurrent use
- Use synchronization primitives where necessary
- Document any concurrency limitations
- Test concurrent access scenarios

## Decision Standards

### When to Use Goroutines

- Use goroutines for long-running operations
- Handle SQLCL output processing in separate goroutines
- Provide context for cancellation
- Avoid goroutine leaks by using WaitGroups

### Command Timeouts

- Set reasonable default timeouts for all operations
- Allow user configuration of timeouts
- Provide clear error messages on timeout
- Cancel ongoing operations when timeouts occur

### Large Result Sets

- Support streaming for large results
- Implement pagination where applicable
- Use efficient memory management for large datasets
- Provide progress indicators for long-running queries

## Prohibitions

- Do NOT use global state
- Do NOT execute arbitrary shell commands
- Do NOT implement functionality already provided by SQLCL
- Do NOT store sensitive information in logs or error messages
- Do NOT use deprecated Go features
- Do NOT override existing SQLCL configuration without explicit user action
- Do NOT assume SQLCL version compatibility; always check version

## Examples

### Command Wrapper Implementation

```go
// ExecuteSQL executes a SQL query and returns the results
func (c *Client) ExecuteSQL(ctx context.Context, sql string) (*SQLResult, error) {
    if sql == "" {
        return nil, errors.New("sql query cannot be empty")
    }
    
    result, err := c.executor.Execute(ctx, sql)
    if err != nil {
        return nil, fmt.Errorf("failed to execute SQL: %w", err)
    }
    
    return c.parser.ParseSQLResult(result)
}
```

### Error Handling Example

```go
// Connect establishes a connection to the database
func (c *Client) Connect(ctx context.Context, connectString string) error {
    if connectString == "" {
        return errors.New("connect string cannot be empty")
    }
    
    cmd := fmt.Sprintf("CONNECT %s", connectString)
    output, err := c.executor.Execute(ctx, cmd)
    if err != nil {
        return fmt.Errorf("SQLCL execution error: %w", err)
    }
    
    if strings.Contains(output, "ORA-") {
        return NewOracleError(output)
    }
    
    c.connected = true
    return nil
}
```