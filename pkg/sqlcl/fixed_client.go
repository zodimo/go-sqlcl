// Package sqlcl provides the main client API for interacting with SQLCL
package sqlcl

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/zodimo/go-sqlcl/internal/parser"
	"github.com/zodimo/go-sqlcl/internal/process"
	"github.com/zodimo/go-sqlcl/pkg/types"
)

// FixedClient implements a more reliable client for interacting with SQLCL
// using a script-based approach instead of keeping a long-running process
type FixedClient struct {
	config      *types.ClientConfig
	process     *process.ScriptBasedSQLCLProcess
	parser      *parser.OutputParser
	mutex       sync.Mutex
	isConnected bool
}

// NewFixedClient creates a new SQLCL client with the given configuration
func NewFixedClient(config *types.ClientConfig) (*FixedClient, error) {
	if config == nil {
		config = types.DefaultConfig()
	}

	// Create OutputParser
	outputParser := parser.NewOutputParser()

	return &FixedClient{
		config:      config,
		parser:      outputParser,
		isConnected: false,
	}, nil
}

// Connect establishes a connection to the database using a simple connection string
func (c *FixedClient) Connect(ctx context.Context, connectString string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Create script-based process with connection string
	var err error
	c.process, err = process.NewScriptBasedSQLCLProcess(
		c.config.SQLclPath,
		connectString,
		c.config.Timeout,
	)
	if err != nil {
		return fmt.Errorf("failed to create script-based process: %w", err)
	}

	// Test the connection with a simple query
	output, err := c.process.ExecuteSQL(ctx, "SELECT 1 FROM dual")
	if err != nil {
		// Check if the error is related to authentication
		if strings.Contains(output, "ORA-01017") {
			return &types.Error{
				Code:    types.ORA01017,
				Message: "Invalid username/password",
				SQL:     "CONNECT " + connectString,
			}
		}
		return fmt.Errorf("connection test failed: %w", err)
	}

	c.isConnected = true
	return nil
}

// ConnectWithOptions establishes a connection to the database with detailed options
func (c *FixedClient) ConnectWithOptions(ctx context.Context, options types.ConnectionOptions) error {
	// Build connection string based on options
	var connectString string

	// Check for TNS connection vs. basic connection
	if options.ConnectStr != "" {
		// Basic format: username/password@host:port/service
		connectString = fmt.Sprintf("%s/%s@%s",
			options.Username,
			options.Password,
			options.ConnectStr)
	} else {
		// Just username and password
		connectString = fmt.Sprintf("%s/%s",
			options.Username,
			options.Password)
	}

	// Connect with the built connection string
	return c.Connect(ctx, connectString)
}

// ExecuteSQL executes a SQL statement and returns the result
func (c *FixedClient) ExecuteSQL(ctx context.Context, sql string) (*types.QueryResult, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.isConnected || !c.process.IsInitialized() {
		return nil, fmt.Errorf("not connected to database")
	}

	// Execute the SQL using the script-based process
	output, err := c.process.ExecuteSQL(ctx, sql)

	// Check for Oracle errors in the output even if the process execution succeeded
	hasOracleError := strings.Contains(output, "ORA-") ||
		strings.Contains(output, "ERROR at line") ||
		strings.Contains(output, "SP2-") ||
		strings.Contains(output, "PLS-")

	// Create a basic result with the status and output
	result := &types.QueryResult{
		Success: err == nil && !hasOracleError,
		Message: output,
	}

	// If there was an error from the process or an Oracle error in the output
	if err != nil || hasOracleError {
		if err != nil {
			result.Message = fmt.Sprintf("Error: %v\nOutput: %s", err, output)
		}
		result.Success = false
		result.Summary = "SQL execution failed"
		return result, nil
	}

	// Try to parse the result based on the command type
	// Determine if it's a query (SELECT) or DML/DDL
	isQuery := strings.HasPrefix(strings.ToUpper(strings.TrimSpace(sql)), "SELECT")

	if isQuery {
		// For queries, parse the result into a structured QueryResult
		queryResult, err := c.parser.ParseQueryResult(output)
		if err != nil && err != parser.ErrNoRows {
			result.Success = false
			result.Message = fmt.Sprintf("Failed to parse query result: %v", err)
			return result, nil
		}

		if queryResult != nil {
			// Map columns
			result.Columns = make([]types.Column, len(queryResult.Columns))
			for i, col := range queryResult.Columns {
				result.Columns[i] = types.Column{
					Name: col.Name,
					Type: types.TypeUnknown, // Default to unknown type
				}
			}

			// Map rows
			result.Rows = make([]types.Row, len(queryResult.Rows))
			for i, row := range queryResult.Rows {
				// Convert string values to interfaces
				values := make([]interface{}, len(row.Values))
				for j, val := range row.Values {
					values[j] = val
				}
				result.Rows[i] = types.Row{Values: values}
			}

			result.Summary = queryResult.Summary
		} else {
			// No rows or error parsing
			result.Summary = "No rows returned"
			result.Columns = []types.Column{}
			result.Rows = []types.Row{}
		}
	} else {
		// For DML/DDL, just return success and the raw output
		result.Summary = "Command executed successfully"
	}

	return result, nil
}

// ExecuteFile executes a SQL script file
func (c *FixedClient) ExecuteFile(ctx context.Context, filePath string) (*types.QueryResult, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.isConnected || !c.process.IsInitialized() {
		return nil, fmt.Errorf("not connected to database")
	}

	// Use the @ command to execute the script file
	fileCommand := fmt.Sprintf("@%s", filePath)
	output, err := c.process.ExecuteSQL(ctx, fileCommand)

	// Create a basic result with the status and output
	result := &types.QueryResult{
		Success: err == nil,
		Message: output,
		Summary: "Script executed successfully",
	}

	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Error: %v\nOutput: %s", err, output)
		result.Summary = "Script execution failed"
	}

	return result, nil
}

// Close releases resources
func (c *FixedClient) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.process != nil {
		err := c.process.Close()
		c.isConnected = false
		return err
	}

	return nil
}

// IsConnected returns whether a connection has been established
func (c *FixedClient) IsConnected() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.isConnected && c.process != nil && c.process.IsInitialized()
}

// ExecuteCommand executes a command and returns the result
func (c *FixedClient) ExecuteCommand(ctx context.Context, command types.Command) (*types.QueryResult, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.isConnected || !c.process.IsInitialized() {
		return nil, fmt.Errorf("not connected to database")
	}

	// Apply command-specific timeout if specified
	var cancel context.CancelFunc
	if command.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, command.Timeout)
		defer cancel()
	} else if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, c.config.QueryTimeout)
		defer cancel()
	}

	// If there are parameters, substitute them in the SQL
	sql := command.SQL
	for name, value := range command.Parameters {
		// Replace parameters in the form :name with their values
		sql = strings.ReplaceAll(sql, ":"+name, value)
	}

	// Execute the SQL
	output, err := c.process.ExecuteSQL(ctx, sql)

	// Check for Oracle errors in the output even if the process execution succeeded
	hasOracleError := strings.Contains(output, "ORA-") ||
		strings.Contains(output, "ERROR at line") ||
		strings.Contains(output, "SP2-") ||
		strings.Contains(output, "PLS-")

	// Create a result
	result := &types.QueryResult{
		Success: err == nil && !hasOracleError,
		Message: output,
	}

	// If there was an error or command should ignore errors
	if err != nil || hasOracleError {
		if err != nil {
			result.Message = fmt.Sprintf("Error: %v\nOutput: %s", err, output)
		}

		if command.IgnoreErrors {
			// If we're supposed to ignore errors, mark as success
			result.Success = true
			result.Summary = "Command executed with errors (ignored)"
		} else {
			result.Success = false
			result.Summary = "Command execution failed"
		}

		return result, nil
	}

	// If it's a query, try to parse the result
	isQuery := strings.HasPrefix(strings.ToUpper(strings.TrimSpace(sql)), "SELECT")

	if isQuery && command.ReturnResults {
		// Parse the query result
		queryResult, err := c.parser.ParseQueryResult(output)
		if err != nil && err != parser.ErrNoRows {
			result.Success = false
			result.Message = fmt.Sprintf("Failed to parse query result: %v", err)
			return result, nil
		}

		if queryResult != nil {
			// Map columns
			result.Columns = make([]types.Column, len(queryResult.Columns))
			for i, col := range queryResult.Columns {
				result.Columns[i] = types.Column{
					Name: col.Name,
					Type: types.TypeUnknown, // Default to unknown type
				}
			}

			// Map rows
			result.Rows = make([]types.Row, len(queryResult.Rows))
			for i, row := range queryResult.Rows {
				// Convert string values to interfaces
				values := make([]interface{}, len(row.Values))
				for j, val := range row.Values {
					values[j] = val
				}
				result.Rows[i] = types.Row{Values: values}
			}

			result.Summary = queryResult.Summary
		} else {
			// No rows or error parsing
			result.Summary = "No rows returned"
			result.Columns = []types.Column{}
			result.Rows = []types.Row{}
		}
	} else {
		// For DML/DDL, just return success and the raw output
		result.Summary = "Command executed successfully"
	}

	return result, nil
}
