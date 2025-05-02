// Package process provides functionality for managing the SQLcl process
package process

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

var (
	// ErrScriptExecutionFailed is returned when the script execution fails
	ErrScriptExecutionFailed = errors.New("sqlcl script execution failed")
)

// ScriptBasedSQLCLProcess implements a more reliable SQLcl process using script files
type ScriptBasedSQLCLProcess struct {
	path        string        // Path to the SQLcl executable
	tempDir     string        // Directory to store temporary script files
	connectStr  string        // Connection string to use for SQLcl
	mutex       sync.Mutex    // Mutex to protect concurrent access
	timeout     time.Duration // Default timeout for operations
	cleanup     func()        // Function to clean up temporary resources
	initialized bool          // Whether the process has been initialized
}

// NewScriptBasedSQLCLProcess creates a new script-based SQLcl process
func NewScriptBasedSQLCLProcess(path string, connectStr string, timeout time.Duration) (*ScriptBasedSQLCLProcess, error) {
	// Create a temporary directory for script files
	tempDir, err := ioutil.TempDir("", "sqlcl-scripts")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// Create a cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return &ScriptBasedSQLCLProcess{
		path:        path,
		tempDir:     tempDir,
		connectStr:  connectStr,
		timeout:     timeout,
		cleanup:     cleanup,
		initialized: true,
	}, nil
}

// ExecuteSQL executes a SQL statement using a temporary script file
func (p *ScriptBasedSQLCLProcess) ExecuteSQL(ctx context.Context, sql string) (string, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.initialized {
		return "", fmt.Errorf("process not initialized")
	}

	// Create a unique script file name
	scriptFileName := fmt.Sprintf("script_%d.sql", time.Now().UnixNano())
	scriptPath := filepath.Join(p.tempDir, scriptFileName)

	// Create the script file with SQL and exit command
	err := ioutil.WriteFile(scriptPath, []byte(sql+"\nexit;\n"), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create script file: %w", err)
	}
	defer os.Remove(scriptPath)

	// Build the command to execute the script
	// Format: sql -S username/password@host:port/service @scriptPath
	cmd := exec.CommandContext(ctx, p.path, "-S", p.connectStr, "@"+scriptPath)

	// Execute the command with timeout
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, p.timeout)
		defer cancel()
	}

	// Execute and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Return the output even if there's an error, as it may contain valuable error information
		return string(output), fmt.Errorf("%w: %v", ErrScriptExecutionFailed, err)
	}

	return string(output), nil
}

// Close cleans up resources
func (p *ScriptBasedSQLCLProcess) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.cleanup != nil {
		p.cleanup()
	}

	return nil
}

// UpdateConnectString updates the connection string
func (p *ScriptBasedSQLCLProcess) UpdateConnectString(connectStr string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.connectStr = connectStr
}

// IsInitialized returns whether the process has been initialized
func (p *ScriptBasedSQLCLProcess) IsInitialized() bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.initialized
}
