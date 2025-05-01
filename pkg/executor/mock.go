// Package executor provides an abstraction over the os/exec package
// to make command execution testable and mockable.
package executor

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

// MockCommander implements the Commander interface for testing purposes.
// It provides a way to replace real command execution with controlled
// behavior in tests, enabling deterministic and predictable testing.
type MockCommander struct {
	// CommandFunc allows customizing the Command method behavior for testing.
	// If set, it will be called instead of the default implementation.
	CommandFunc func(name string, args ...string) Command

	// LastCommand tracks the most recently created command for verification in tests.
	// This allows test code to inspect what command would have been executed.
	LastCommand Command

	// LookPathFunc allows customizing the LookPath method behavior for testing.
	// If set, it will be called instead of the default implementation.
	LookPathFunc func(file string) (string, error)

	// LookPathCalled tracks whether LookPath has been called for verification in tests.
	LookPathCalled bool

	// LastLookPathFile tracks the file argument passed to the most recent LookPath call.
	LastLookPathFile string
}

// Command calls the CommandFunc if it is set, otherwise returns a default MockCommand.
// It records the created command in LastCommand for later verification.
func (m *MockCommander) Command(name string, args ...string) Command {
	var cmd Command
	if m.CommandFunc != nil {
		cmd = m.CommandFunc(name, args...)
	} else {
		cmd = &MockCommand{
			name: name,
			args: args,
		}
	}
	m.LastCommand = cmd
	return cmd
}

// LookPath searches for an executable named file in the directories named by the PATH environment variable.
// In the mock implementation, it either calls the custom LookPathFunc if set, or returns a default mock path.
func (m *MockCommander) LookPath(file string) (string, error) {
	m.LookPathCalled = true
	m.LastLookPathFile = file
	if m.LookPathFunc != nil {
		return m.LookPathFunc(file)
	}
	// Default implementation (successful lookup)
	return "/mock/path/" + file, nil
}

// NewMockCommander creates a new MockCommander with default behavior.
// The returned MockCommander will create MockCommand instances with no
// custom behavior overrides.
func NewMockCommander() *MockCommander {
	return &MockCommander{}
}

// MockCommand implements the Command interface for testing purposes.
// It allows test code to control the behavior of command execution
// without actually executing any external processes.
type MockCommand struct {
	name       string                          // The name of the command
	args       []string                        // The arguments passed to the command
	dir        string                          // The working directory for the command
	env        []string                        // The environment variables for the command
	stdin      io.Reader                       // The standard input for the command
	stdout     io.Writer                       // The standard output for the command
	stderr     io.Writer                       // The standard error for the command
	runFunc    func() error                    // Custom function to be called by Run
	startFunc  func() error                    // Custom function to be called by Start
	waitFunc   func() error                    // Custom function to be called by Wait
	outFunc    func() ([]byte, error)          // Custom function to be called by Output
	combFunc   func() ([]byte, error)          // Custom function to be called by CombinedOutput
	runCtxFunc func(ctx context.Context) error // Custom function to be called by RunContext

	// Track method calls for verification in tests
	RunCalled            bool // Whether Run has been called
	StartCalled          bool // Whether Start has been called
	WaitCalled           bool // Whether Wait has been called
	OutputCalled         bool // Whether Output has been called
	CombinedOutputCalled bool // Whether CombinedOutput has been called

	// Additional fields for tracking call information
	CalledWithName string   // Records the name used when creating the command
	CalledWithArgs []string // Records the args used when creating the command
}

// SetDir sets the working directory for the mock command and returns the command.
// This allows method chaining for convenient command configuration.
func (m *MockCommand) SetDir(dir string) Command {
	m.dir = dir
	return m
}

// SetEnv sets the environment variables for the mock command and returns the command.
// This allows method chaining for convenient command configuration.
func (m *MockCommand) SetEnv(env []string) Command {
	m.env = env
	return m
}

// SetStdin sets the standard input for the mock command and returns the command.
// This allows method chaining for convenient command configuration.
func (m *MockCommand) SetStdin(r io.Reader) Command {
	m.stdin = r
	return m
}

// SetStdout sets the standard output for the mock command and returns the command.
// This allows method chaining for convenient command configuration.
func (m *MockCommand) SetStdout(w io.Writer) Command {
	m.stdout = w
	return m
}

// SetStderr sets the standard error for the mock command and returns the command.
// This allows method chaining for convenient command configuration.
func (m *MockCommand) SetStderr(w io.Writer) Command {
	m.stderr = w
	return m
}

// Run executes the mock command. If runFunc is set, it calls that function.
// Otherwise, it returns nil. It also tracks that the method was called.
func (m *MockCommand) Run() error {
	m.RunCalled = true
	if m.runFunc != nil {
		return m.runFunc()
	}
	return nil
}

// Start starts the mock command. If startFunc is set, it calls that function.
// Otherwise, it returns nil. It also tracks that the method was called.
func (m *MockCommand) Start() error {
	m.StartCalled = true
	if m.startFunc != nil {
		return m.startFunc()
	}
	return nil
}

// Wait waits for the mock command to complete. If waitFunc is set, it calls that function.
// Otherwise, it returns nil. It also tracks that the method was called.
func (m *MockCommand) Wait() error {
	m.WaitCalled = true
	if m.waitFunc != nil {
		return m.waitFunc()
	}
	return nil
}

// Output returns the mock command's output. If outFunc is set, it calls that function.
// Otherwise, it returns an empty byte slice and nil error. It also tracks that the method was called.
func (m *MockCommand) Output() ([]byte, error) {
	m.OutputCalled = true
	if m.outFunc != nil {
		return m.outFunc()
	}
	return []byte{}, nil
}

// CombinedOutput returns the mock command's combined output. If combFunc is set, it calls that function.
// Otherwise, it returns an empty byte slice and nil error. It also tracks that the method was called.
func (m *MockCommand) CombinedOutput() ([]byte, error) {
	m.CombinedOutputCalled = true
	if m.combFunc != nil {
		return m.combFunc()
	}
	return []byte{}, nil
}

// StdoutPipe returns a pipe that will be connected to the command's standard output.
// In the mock implementation, it returns a ReadCloser that reads from an empty string.
func (m *MockCommand) StdoutPipe() (io.ReadCloser, error) {
	return ioutil.NopCloser(strings.NewReader("")), nil
}

// StderrPipe returns a pipe that will be connected to the command's standard error.
// In the mock implementation, it returns a ReadCloser that reads from an empty string.
func (m *MockCommand) StderrPipe() (io.ReadCloser, error) {
	return ioutil.NopCloser(strings.NewReader("")), nil
}

// StdinPipe returns a pipe that will be connected to the command's standard input.
// In the mock implementation, it returns a WriteCloser that discards all writes.
func (m *MockCommand) StdinPipe() (io.WriteCloser, error) {
	return &nopWriteCloser{w: io.Discard}, nil
}

// RunContext executes the mock command within the given context.
// If runCtxFunc is set, it calls that function. Otherwise, it calls Run.
func (m *MockCommand) RunContext(ctx context.Context) error {
	if m.runCtxFunc != nil {
		return m.runCtxFunc(ctx)
	}
	return m.Run()
}

// String returns a string representation of the command.
// This is useful for logging and debugging in tests.
func (m *MockCommand) String() string {
	return fmt.Sprintf("%s %s", m.name, strings.Join(m.args, " "))
}

// nopWriteCloser wraps an io.Writer and adds a no-op Close method.
// It is used to implement StdinPipe in the mock implementation.
type nopWriteCloser struct {
	w io.Writer
}

// Write delegates to the wrapped writer.
func (n *nopWriteCloser) Write(p []byte) (int, error) {
	return n.w.Write(p)
}

// Close implements io.Closer with a no-op close operation.
func (n *nopWriteCloser) Close() error {
	return nil
}

// GetCommandName returns the name of the command for verification in tests.
// This allows test code to verify what command would have been executed.
func (m *MockCommand) GetCommandName() string {
	return m.name
}

// GetCommandArgs returns the arguments of the command for verification in tests.
// This allows test code to verify what arguments would have been passed.
func (m *MockCommand) GetCommandArgs() []string {
	return m.args
}

// GetDir returns the working directory of the command for verification in tests.
// This allows test code to verify what working directory would have been set.
func (m *MockCommand) GetDir() string {
	return m.dir
}

// GetEnv returns the environment variables of the command for verification in tests.
// This allows test code to verify what environment variables would have been set.
func (m *MockCommand) GetEnv() []string {
	return m.env
}

// SetRunFunc sets a custom function to be called when Run is invoked.
// It returns the MockCommand to allow method chaining.
func (m *MockCommand) SetRunFunc(f func() error) *MockCommand {
	m.runFunc = f
	return m
}

// SetStartFunc sets a custom function to be called when Start is invoked.
// It returns the MockCommand to allow method chaining.
func (m *MockCommand) SetStartFunc(f func() error) *MockCommand {
	m.startFunc = f
	return m
}

// SetWaitFunc sets a custom function to be called when Wait is invoked.
// It returns the MockCommand to allow method chaining.
func (m *MockCommand) SetWaitFunc(f func() error) *MockCommand {
	m.waitFunc = f
	return m
}

// SetOutputFunc sets a custom function to be called when Output is invoked.
// It returns the MockCommand to allow method chaining.
func (m *MockCommand) SetOutputFunc(f func() ([]byte, error)) *MockCommand {
	m.outFunc = f
	return m
}

// SetCombinedOutputFunc sets a custom function to be called when CombinedOutput is invoked.
// It returns the MockCommand to allow method chaining.
func (m *MockCommand) SetCombinedOutputFunc(f func() ([]byte, error)) *MockCommand {
	m.combFunc = f
	return m
}

// SetRunContextFunc sets a custom function to be called when RunContext is invoked.
// It returns the MockCommand to allow method chaining.
func (m *MockCommand) SetRunContextFunc(f func(ctx context.Context) error) *MockCommand {
	m.runCtxFunc = f
	return m
}

// NewMockCommanderWithOutput creates a MockCommander that returns the specified output.
// This helper function simplifies test setup by configuring common command behaviors.
func NewMockCommanderWithOutput(output string, err error) *MockCommander {
	return &MockCommander{
		CommandFunc: func(name string, args ...string) Command {
			return &MockCommand{
				CalledWithName: name,
				CalledWithArgs: args,
				outFunc: func() ([]byte, error) {
					return []byte(output), err
				},
				combFunc: func() ([]byte, error) {
					return []byte(output), err
				},
				runFunc: func() error {
					return err
				},
			}
		},
	}
}

// NewMockCommanderWithOutputs creates a MockCommander that returns different outputs for stdout and combined.
// This helper function allows specifying different outputs for standard and combined output,
// which is useful for testing code that handles these outputs differently.
func NewMockCommanderWithOutputs(stdout, combined string, err error) *MockCommander {
	return &MockCommander{
		CommandFunc: func(name string, args ...string) Command {
			return &MockCommand{
				CalledWithName: name,
				CalledWithArgs: args,
				outFunc: func() ([]byte, error) {
					return []byte(stdout), err
				},
				combFunc: func() ([]byte, error) {
					return []byte(combined), err
				},
				runFunc: func() error {
					return err
				},
			}
		},
	}
}

// NewMockCommanderWithError creates a MockCommander that returns the specified error.
// This helper function simplifies test setup for error handling code paths.
func NewMockCommanderWithError(err error) *MockCommander {
	return &MockCommander{
		CommandFunc: func(name string, args ...string) Command {
			return &MockCommand{
				CalledWithName: name,
				CalledWithArgs: args,
				outFunc: func() ([]byte, error) {
					return nil, err
				},
				combFunc: func() ([]byte, error) {
					return nil, err
				},
				runFunc: func() error {
					return err
				},
			}
		},
	}
}

// NewMockCommanderWithLookPath creates a new MockCommander with custom LookPath behavior.
// The returned MockCommander will return the specified path and error when LookPath is called.
func NewMockCommanderWithLookPath(path string, err error) *MockCommander {
	m := NewMockCommander()
	m.LookPathFunc = func(file string) (string, error) {
		return path, err
	}
	return m
}
