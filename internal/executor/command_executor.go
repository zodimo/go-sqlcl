// Package executor provides functionality for executing commands on the SQLcl process
package executor

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/zodimo/go-sqlcl/internal/process"
)

var (
	// ErrCommandTimeout is returned when a command execution times out
	ErrCommandTimeout = errors.New("command execution timed out")

	// ErrCommandFailed is returned when a command execution fails
	ErrCommandFailed = errors.New("command execution failed")

	// ErrProcessNotAvailable is returned when the SQLcl process is not available
	ErrProcessNotAvailable = errors.New("sqlcl process is not available")
)

// CommandType represents the type of command being executed
type CommandType int

const (
	// QueryCommand represents a SQL query command
	QueryCommand CommandType = iota

	// DDLCommand represents a Data Definition Language command
	DDLCommand

	// SQLCLCommand represents a SQLcl-specific command
	SQLCLCommand
)

// CommandResult represents the result of executing a command
type CommandResult struct {
	Output       string
	Error        string
	IsSuccessful bool
	CommandType  CommandType
	Duration     time.Duration
}

// Option is a functional option for configuring CommandExecutor
type Option func(*CommandExecutor)

// WithTimeout sets the default timeout for command execution
func WithTimeout(timeout time.Duration) Option {
	return func(e *CommandExecutor) {
		e.timeout = timeout
	}
}

// WithPromptPattern sets the pattern used to detect command completion
func WithPromptPattern(pattern string) Option {
	return func(e *CommandExecutor) {
		e.promptPattern = regexp.MustCompile(pattern)
	}
}

// CommandExecutor executes commands on a SQLcl process
type CommandExecutor struct {
	process       *process.SQLCLProcess
	timeout       time.Duration
	promptPattern *regexp.Regexp
	mutex         sync.Mutex
}

// DefaultTimeout is the default timeout for command execution
const DefaultTimeout = 60 * time.Second

// DefaultPromptPattern is the default pattern used to detect command completion
const DefaultPromptPattern = `SQL>\s*$`

// NewCommandExecutor creates a new CommandExecutor with the given SQLCLProcess
func NewCommandExecutor(sqlclProcess *process.SQLCLProcess, options ...Option) *CommandExecutor {
	executor := &CommandExecutor{
		process:       sqlclProcess,
		timeout:       DefaultTimeout,
		promptPattern: regexp.MustCompile(DefaultPromptPattern),
	}

	for _, option := range options {
		option(executor)
	}

	return executor
}

// Execute executes a command on the SQLcl process and returns the result
func (e *CommandExecutor) Execute(ctx context.Context, command string) (*CommandResult, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.process == nil {
		return nil, ErrProcessNotAvailable
	}

	if !e.process.IsRunning() {
		return nil, process.ErrProcessNotRunning
	}

	// Create a context with timeout if the provided context doesn't have a deadline
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, e.timeout)
		defer cancel()
	}

	startTime := time.Now()

	// Send the command to the SQLcl process
	if err := e.process.SendCommand(ctx, command); err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	// Read the output from the SQLcl process
	output, err := e.process.ReadOutput(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read output: %w", err)
	}

	// Read any error output from the SQLcl process
	errorOutput, err := e.process.ReadError(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read error output: %w", err)
	}

	duration := time.Since(startTime)

	// Determine the command type
	commandType := e.determineCommandType(command)

	// Check if the command completed successfully
	isSuccessful := e.isCommandSuccessful(output, errorOutput)

	return &CommandResult{
		Output:       output,
		Error:        errorOutput,
		IsSuccessful: isSuccessful,
		CommandType:  commandType,
		Duration:     duration,
	}, nil
}

// ExecuteQuery executes a SQL query command
func (e *CommandExecutor) ExecuteQuery(ctx context.Context, query string) (*CommandResult, error) {
	return e.Execute(ctx, query)
}

// ExecuteDDL executes a DDL command
func (e *CommandExecutor) ExecuteDDL(ctx context.Context, ddl string) (*CommandResult, error) {
	return e.Execute(ctx, ddl)
}

// ExecuteSQLCLCommand executes a SQLcl-specific command
func (e *CommandExecutor) ExecuteSQLCLCommand(ctx context.Context, command string) (*CommandResult, error) {
	return e.Execute(ctx, command)
}

// determineCommandType determines the type of command being executed
func (e *CommandExecutor) determineCommandType(command string) CommandType {
	command = strings.TrimSpace(strings.ToUpper(command))

	// Check if the command is a SQLcl-specific command
	sqlclCommands := []string{"SET", "SHOW", "DESCRIBE", "DESC", "HELP", "EXIT", "QUIT", "CONNECT", "CONN"}
	for _, cmd := range sqlclCommands {
		if strings.HasPrefix(command, cmd) {
			return SQLCLCommand
		}
	}

	// Check if the command is a DDL command
	ddlCommands := []string{"CREATE", "ALTER", "DROP", "TRUNCATE", "COMMENT", "GRANT", "REVOKE"}
	for _, cmd := range ddlCommands {
		if strings.HasPrefix(command, cmd) {
			return DDLCommand
		}
	}

	// Default to query command
	return QueryCommand
}

// isCommandSuccessful determines if a command executed successfully
func (e *CommandExecutor) isCommandSuccessful(output, errorOutput string) bool {
	// Check for common error indicators in output
	errorPatterns := []string{
		"ORA-[0-9]+",
		"ERROR",
		"SP2-[0-9]+",
	}

	for _, pattern := range errorPatterns {
		if regexp.MustCompile(pattern).MatchString(output) {
			return false
		}
	}

	// Check if the error output contains any error indicators
	if errorOutput != "" {
		return false
	}

	return true
}
