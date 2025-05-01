// Package process provides functionality for managing the SQLcl process
package process

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

var (
	// ErrProcessNotRunning is returned when an operation is attempted on a non-running process
	ErrProcessNotRunning = errors.New("sqlcl process is not running")

	// ErrProcessAlreadyRunning is returned when attempting to start an already running process
	ErrProcessAlreadyRunning = errors.New("sqlcl process is already running")

	// ErrProcessTimeout is returned when a process operation times out
	ErrProcessTimeout = errors.New("sqlcl process operation timed out")
)

// DefaultSQLCLPath is the default path to the SQLcl executable
const DefaultSQLCLPath = "sql"

// DefaultTimeout is the default timeout for operations
const DefaultTimeout = 30 * time.Second

// Option is a functional option for configuring SQLCLProcess
type Option func(*SQLCLProcess)

// WithPath sets the path to the SQLcl binary
func WithPath(path string) Option {
	return func(p *SQLCLProcess) {
		p.path = path
	}
}

// WithTimeout sets the default timeout for operations
func WithTimeout(timeout time.Duration) Option {
	return func(p *SQLCLProcess) {
		p.timeout = timeout
	}
}

// WithArgs sets additional command-line arguments for SQLcl
func WithArgs(args []string) Option {
	return func(p *SQLCLProcess) {
		p.args = args
	}
}

// SQLCLProcess manages a SQLcl process
type SQLCLProcess struct {
	path      string
	args      []string
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	isRunning bool
	mutex     sync.Mutex
	timeout   time.Duration
}

// NewSQLCLProcess creates a new SQLCLProcess instance
func NewSQLCLProcess(options ...Option) *SQLCLProcess {
	process := &SQLCLProcess{
		path:    DefaultSQLCLPath,
		timeout: DefaultTimeout,
	}

	for _, option := range options {
		option(process)
	}

	return process
}

// Start starts the SQLcl process
func (p *SQLCLProcess) Start(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.isRunning {
		return ErrProcessAlreadyRunning
	}

	// Create a new command with the SQLcl path and any additional arguments
	args := append([]string{"-S"}, p.args...)
	p.cmd = exec.CommandContext(ctx, p.path, args...)

	// Set up pipes for stdin, stdout, and stderr
	var err error

	p.stdin, err = p.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	p.stdout, err = p.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	p.stderr, err = p.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start sqlcl process: %w", err)
	}

	p.isRunning = true

	return nil
}

// Stop stops the SQLcl process
func (p *SQLCLProcess) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.isRunning {
		return ErrProcessNotRunning
	}

	// Send exit command to SQLcl
	if _, err := p.stdin.Write([]byte("exit\n")); err != nil {
		// If we can't send the exit command, force kill the process
		if err := p.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill sqlcl process: %w", err)
		}
	}

	// Wait for the process to exit with a timeout
	done := make(chan error, 1)
	go func() {
		done <- p.cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("error waiting for sqlcl process to exit: %w", err)
		}
	case <-time.After(p.timeout):
		// If the process doesn't exit gracefully, force kill it
		if err := p.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill sqlcl process after timeout: %w", err)
		}
		return ErrProcessTimeout
	}

	// Close pipes
	p.stdin.Close()

	p.isRunning = false

	return nil
}

// IsRunning returns whether the process is running
func (p *SQLCLProcess) IsRunning() bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.isRunning
}

// SendCommand sends a command to the SQLcl process
func (p *SQLCLProcess) SendCommand(ctx context.Context, command string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.isRunning {
		return ErrProcessNotRunning
	}

	// Add a newline to the command if it doesn't end with one
	if len(command) == 0 || command[len(command)-1] != '\n' {
		command = command + "\n"
	}

	// Create a context with timeout if the provided context doesn't have a deadline
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, p.timeout)
		defer cancel()
	}

	// Send the command
	done := make(chan error, 1)
	go func() {
		_, err := p.stdin.Write([]byte(command))
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to send command to sqlcl: %w", err)
		}
	case <-ctx.Done():
		return fmt.Errorf("send command timeout: %w", ctx.Err())
	}

	return nil
}

// ReadOutput reads the stdout of the SQLcl process
func (p *SQLCLProcess) ReadOutput(ctx context.Context) (string, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.isRunning {
		return "", ErrProcessNotRunning
	}

	// Create a context with timeout if the provided context doesn't have a deadline
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, p.timeout)
		defer cancel()
	}

	// Create a scanner to read from stdout
	scanner := bufio.NewScanner(p.stdout)
	var output string

	// Read from stdout until we get the prompt or context is done
	done := make(chan bool, 1)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			output += line + "\n"
			// Check if this line indicates command completion
			if line == "SQL>" {
				break
			}
		}
		done <- true
	}()

	select {
	case <-done:
		if err := scanner.Err(); err != nil {
			return output, fmt.Errorf("error reading from sqlcl stdout: %w", err)
		}
	case <-ctx.Done():
		return output, fmt.Errorf("read output timeout: %w", ctx.Err())
	}

	return output, nil
}

// ReadError reads the stderr of the SQLcl process
func (p *SQLCLProcess) ReadError(ctx context.Context) (string, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.isRunning {
		return "", ErrProcessNotRunning
	}

	// Create a context with timeout if the provided context doesn't have a deadline
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, p.timeout)
		defer cancel()
	}

	// Create a scanner to read from stderr
	scanner := bufio.NewScanner(p.stderr)
	var output string

	// Read from stderr until the context is done
	done := make(chan bool, 1)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			output += line + "\n"
		}
		done <- true
	}()

	select {
	case <-done:
		if err := scanner.Err(); err != nil {
			return output, fmt.Errorf("error reading from sqlcl stderr: %w", err)
		}
	case <-ctx.Done():
		return output, fmt.Errorf("read error timeout: %w", ctx.Err())
	}

	return output, nil
}
