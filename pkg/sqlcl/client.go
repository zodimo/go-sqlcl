// Package sqlcl provides the main client API for interacting with SQLCL
package sqlcl

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zodimo/go-sqlcl/internal/executor"
	"github.com/zodimo/go-sqlcl/internal/parser"
	"github.com/zodimo/go-sqlcl/internal/process"
	"github.com/zodimo/go-sqlcl/pkg/types"
)

// Client implements the types.Client interface and provides methods for interacting with SQLCL
type Client struct {
	config      *types.ClientConfig
	process     *process.SQLCLProcess
	executor    *executor.CommandExecutor
	parser      *parser.OutputParser
	mutex       sync.Mutex
	isConnected bool
}

// NewClient creates a new SQLCL client with the given configuration
func NewClient(config *types.ClientConfig) (*Client, error) {
	if config == nil {
		config = types.DefaultConfig()
	}

	// Create SQLCLProcess with configuration
	sqlclProcess := process.NewSQLCLProcess(
		process.WithPath(config.SQLclPath),
		process.WithTimeout(config.Timeout),
	)

	// Create CommandExecutor
	cmdExecutor := executor.NewCommandExecutor(
		sqlclProcess,
		executor.WithTimeout(config.QueryTimeout),
	)

	// Create OutputParser
	outputParser := parser.NewOutputParser()

	return &Client{
		config:      config,
		process:     sqlclProcess,
		executor:    cmdExecutor,
		parser:      outputParser,
		isConnected: false,
	}, nil
}

// Connect establishes a connection to the database using a simple connection string
func (c *Client) Connect(ctx context.Context, connectString string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Start the SQLCL process if not already running
	if !c.process.IsRunning() {
		if err := c.process.Start(ctx); err != nil {
			return fmt.Errorf("failed to start SQLCL process: %w", err)
		}
	}

	// Create a connection command
	command := "CONNECT " + connectString

	// Set a timeout for the connection if the context doesn't have a deadline
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, c.config.ConnectTimeout)
		defer cancel()
	}

	// Execute the connect command
	result, err := c.executor.Execute(ctx, command)
	if err != nil {
		return fmt.Errorf("failed to execute connect command: %w", err)
	}

	// Check if the connection was successful
	if !result.IsSuccessful {
		// Try to parse the error
		errorInfo, err := c.parser.ParseError(result.Output)
		if err != nil {
			return fmt.Errorf("connection failed: %s", result.Error)
		}

		return &types.Error{
			Code:    types.ErrorCode(errorInfo.Code),
			Message: errorInfo.Message,
			SQL:     command,
		}
	}

	c.isConnected = true
	return nil
}

// ConnectWithOptions establishes a connection to the database with detailed options
func (c *Client) ConnectWithOptions(ctx context.Context, options types.ConnectionOptions) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Start the SQLCL process if not already running
	if !c.process.IsRunning() {
		if err := c.process.Start(ctx); err != nil {
			return fmt.Errorf("failed to start SQLCL process: %w", err)
		}
	}

	// Build connection string based on options
	var connectCommand string

	// Check for TNS connection vs. basic connection
	if options.ConnectStr != "" {
		// Basic format: username/password@host:port/service
		connectCommand = fmt.Sprintf("CONNECT %s/%s@%s",
			options.Username,
			options.Password,
			options.ConnectStr)
	} else {
		// Just username and password
		connectCommand = fmt.Sprintf("CONNECT %s/%s",
			options.Username,
			options.Password)
	}

	// Add role if specified
	if options.Role != "" {
		connectCommand += " AS " + options.Role
	}

	// Execute connection command with timeout
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, c.config.ConnectTimeout)
		defer cancel()
	}

	// Execute the connect command
	result, err := c.executor.Execute(ctx, connectCommand)
	if err != nil {
		return fmt.Errorf("failed to execute connect command: %w", err)
	}

	// Check if the connection was successful
	if !result.IsSuccessful {
		// Try to parse the error
		errorInfo, err := c.parser.ParseError(result.Output)
		if err != nil {
			return fmt.Errorf("connection failed: %s", result.Error)
		}

		return &types.Error{
			Code:    types.ErrorCode(errorInfo.Code),
			Message: errorInfo.Message,
			SQL:     connectCommand,
		}
	}

	// If wallet or TNS admin were specified, set them after connection
	if options.Wallet != "" {
		_, err := c.executor.ExecuteSQLCLCommand(ctx, fmt.Sprintf("SET WALLET %s", options.Wallet))
		if err != nil {
			return fmt.Errorf("failed to set wallet: %w", err)
		}
	}

	if options.TNSAdmin != "" {
		_, err := c.executor.ExecuteSQLCLCommand(ctx, fmt.Sprintf("SET TNS_ADMIN %s", options.TNSAdmin))
		if err != nil {
			return fmt.Errorf("failed to set TNS_ADMIN: %w", err)
		}
	}

	c.isConnected = true
	return nil
}

// ExecuteSQL executes a SQL statement and returns the result
func (c *Client) ExecuteSQL(ctx context.Context, sql string) (*types.QueryResult, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.isConnected {
		return nil, fmt.Errorf("not connected to database")
	}

	// Set timeout if not already set in context
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, c.config.QueryTimeout)
		defer cancel()
	}

	// Execute the SQL using the command executor
	result, err := c.executor.Execute(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to execute SQL: %w", err)
	}

	// Parse the result based on the command type
	if result.CommandType == executor.QueryCommand {
		// For queries, parse the result into a structured QueryResult
		queryResult, err := c.parser.ParseQueryResult(result.Output)
		if err != nil && err != parser.ErrNoRows {
			return nil, fmt.Errorf("failed to parse query result: %w", err)
		}

		// Convert internal QueryResult to public QueryResult
		publicResult := &types.QueryResult{
			Success: result.IsSuccessful,
			Message: result.Error,
		}

		if queryResult != nil {
			// Map columns
			publicResult.Columns = make([]types.Column, len(queryResult.Columns))
			for i, col := range queryResult.Columns {
				publicResult.Columns[i] = types.Column{
					Name: col.Name,
					Type: types.TypeUnknown, // Default to unknown type
				}
			}

			// Map rows
			publicResult.Rows = make([]types.Row, len(queryResult.Rows))
			for i, row := range queryResult.Rows {
				// Convert string values to interfaces
				values := make([]interface{}, len(row.Values))
				for j, val := range row.Values {
					values[j] = val
				}
				publicResult.Rows[i] = types.Row{Values: values}
			}

			publicResult.Summary = queryResult.Summary
		} else {
			// No rows or error parsing
			publicResult.Summary = "No rows returned"
			publicResult.Columns = []types.Column{}
			publicResult.Rows = []types.Row{}
		}

		return publicResult, nil
	} else {
		// For non-query commands, parse the command status
		cmdStatus, err := c.parser.ParseCommandStatus(result.Output)
		if err != nil {
			return nil, fmt.Errorf("failed to parse command status: %w", err)
		}

		// Create a simple QueryResult for non-query commands
		return &types.QueryResult{
			Success: cmdStatus.IsSuccessful,
			Message: cmdStatus.Message,
			Summary: fmt.Sprintf("%d rows affected", cmdStatus.RowsAffected),
		}, nil
	}
}

// ExecuteCommand executes a command with options and returns the result
func (c *Client) ExecuteCommand(ctx context.Context, command types.Command) (*types.QueryResult, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.isConnected {
		return nil, fmt.Errorf("not connected to database")
	}

	// Use command-specific timeout if specified, otherwise use default
	execTimeout := c.config.QueryTimeout
	if command.Timeout > 0 {
		execTimeout = command.Timeout
	}

	// Set timeout if not already set in context
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, execTimeout)
		defer cancel()
	}

	// Set the output format if specified
	if command.Format != "" {
		_, err := c.executor.ExecuteSQLCLCommand(ctx, fmt.Sprintf("SET FEEDBACK %s", command.Format))
		if err != nil {
			return nil, fmt.Errorf("failed to set output format: %w", err)
		}
	}

	// Replace named parameters in the SQL
	sql := command.SQL
	for name, value := range command.Parameters {
		// Simple parameter replacement
		paramName := ":" + name
		sql = fmt.Sprintf("VARIABLE %s VARCHAR2(4000)", name) + "\n" +
			fmt.Sprintf("BEGIN %s := '%s'; END;", paramName, value) + "\n" +
			sql
	}

	// Execute the command
	result, err := c.executor.Execute(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	// If ignoring errors is set and command failed, don't return an error
	if command.IgnoreErrors && !result.IsSuccessful {
		return &types.QueryResult{
			Success: false,
			Message: result.Error,
		}, nil
	}

	// If the command was successful but we don't need results, return minimal info
	if !command.ReturnResults {
		return &types.QueryResult{
			Success: result.IsSuccessful,
			Message: result.Error,
		}, nil
	}

	// Otherwise process the results as in ExecuteSQL
	if result.CommandType == executor.QueryCommand {
		queryResult, err := c.parser.ParseQueryResult(result.Output)
		if err != nil && err != parser.ErrNoRows {
			return nil, fmt.Errorf("failed to parse query result: %w", err)
		}

		publicResult := &types.QueryResult{
			Success: result.IsSuccessful,
			Message: result.Error,
		}

		if queryResult != nil {
			// Map columns
			publicResult.Columns = make([]types.Column, len(queryResult.Columns))
			for i, col := range queryResult.Columns {
				publicResult.Columns[i] = types.Column{
					Name: col.Name,
					Type: types.TypeUnknown,
				}
			}

			// Map rows
			publicResult.Rows = make([]types.Row, len(queryResult.Rows))
			for i, row := range queryResult.Rows {
				values := make([]interface{}, len(row.Values))
				for j, val := range row.Values {
					values[j] = val
				}
				publicResult.Rows[i] = types.Row{Values: values}
			}

			publicResult.Summary = queryResult.Summary
		} else {
			// No rows or error parsing
			publicResult.Summary = "No rows returned"
			publicResult.Columns = []types.Column{}
			publicResult.Rows = []types.Row{}
		}

		return publicResult, nil
	} else {
		// Non-query commands
		cmdStatus, err := c.parser.ParseCommandStatus(result.Output)
		if err != nil {
			return nil, fmt.Errorf("failed to parse command status: %w", err)
		}

		return &types.QueryResult{
			Success: cmdStatus.IsSuccessful,
			Message: cmdStatus.Message,
			Summary: fmt.Sprintf("%d rows affected", cmdStatus.RowsAffected),
		}, nil
	}
}

// Close closes the connection to the database and stops the SQLCL process
func (c *Client) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.process.IsRunning() {
		// Try to disconnect first if connected
		if c.isConnected {
			// Use a short timeout for disconnection
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, _ = c.executor.ExecuteSQLCLCommand(ctx, "DISCONNECT")
			c.isConnected = false
		}

		// Stop the SQLCL process
		if err := c.process.Stop(); err != nil {
			return fmt.Errorf("failed to stop SQLCL process: %w", err)
		}
	}

	return nil
}

// IsConnected returns whether the client is connected to a database
func (c *Client) IsConnected() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.isConnected
}

// GetConfig returns the client configuration
func (c *Client) GetConfig() *types.ClientConfig {
	return c.config
}

// SetOutputFormat sets the output format for subsequent commands
func (c *Client) SetOutputFormat(ctx context.Context, format string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.process.IsRunning() {
		return fmt.Errorf("SQLCL process is not running")
	}

	// Set the format using the SET command
	_, err := c.executor.ExecuteSQLCLCommand(ctx, fmt.Sprintf("SET FEEDBACK %s", format))
	if err != nil {
		return fmt.Errorf("failed to set output format: %w", err)
	}

	c.config.Format = format
	return nil
}

// ExecuteFile executes SQL commands from a file
func (c *Client) ExecuteFile(ctx context.Context, filePath string) (*types.QueryResult, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.isConnected {
		return nil, fmt.Errorf("not connected to database")
	}

	// Use @ command to execute a script file
	command := fmt.Sprintf("@%s", filePath)

	// Set timeout if not already set in context
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, c.config.QueryTimeout)
		defer cancel()
	}

	// Execute the command
	result, err := c.executor.Execute(ctx, command)
	if err != nil {
		return nil, fmt.Errorf("failed to execute file: %w", err)
	}

	// Return a simple result since file execution can have multiple statements
	return &types.QueryResult{
		Success: result.IsSuccessful,
		Message: result.Error,
		Summary: "File executed",
	}, nil
}
