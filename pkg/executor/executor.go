// Package executor provides an abstraction over the os/exec package
// to make command execution testable and mockable. It simplifies unit testing
// of code that depends on executing external commands by providing interfaces
// that can be easily mocked.
package executor

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Commander defines the interface for creating executable commands.
// It wraps the functionality of os/exec.Command and allows for
// dependency injection and mocking in tests.
type Commander interface {
	// Command creates a new Command with the given name and arguments.
	// The name is the path or name of the command to run, and args are the
	// command line arguments to pass to the command.
	Command(name string, args ...string) Command

	// LookPath searches for an executable named file in the directories named by the PATH environment variable.
	// If file contains a slash, it is tried directly and the PATH is not consulted.
	LookPath(file string) (string, error)
}

// Command defines the interface for an executable command.
// It wraps the functionality of os/exec.Cmd, providing methods to
// configure, start, and interact with external processes.
type Command interface {
	// SetDir sets the working directory for the command.
	// If dir is empty, the command is executed in the calling process's current directory.
	SetDir(dir string) Command

	// SetEnv sets the environment variables for the command.
	// If env is nil, the command inherits the calling process's environment.
	// If env is not nil, it must be in the form key=value.
	SetEnv(env []string) Command

	// SetStdin sets the standard input for the command.
	// If r is nil, the command reads from an empty pipe.
	SetStdin(r io.Reader) Command

	// SetStdout sets the standard output for the command.
	// If w is nil, the command writes to a discarded output.
	SetStdout(w io.Writer) Command

	// SetStderr sets the standard error for the command.
	// If w is nil, the command writes to a discarded output.
	SetStderr(w io.Writer) Command

	// Run starts the command and waits for it to complete.
	// It returns an error if the command cannot be started or
	// exits with a non-zero status.
	Run() error

	// Start starts the command but does not wait for it to complete.
	// The Wait method must be called to release associated resources.
	Start() error

	// Wait waits for the command to complete after calling Start.
	// It must be called after a successful call to Start.
	Wait() error

	// Output runs the command and returns its standard output.
	// Any returned error will include the output in the error message.
	Output() ([]byte, error)

	// CombinedOutput runs the command and returns its combined standard output and standard error.
	// This is useful for capturing all output from a command in a single operation.
	CombinedOutput() ([]byte, error)

	// StdoutPipe returns a pipe that will be connected to the command's standard output.
	// The pipe will be closed automatically after the command completes.
	StdoutPipe() (io.ReadCloser, error)

	// StderrPipe returns a pipe that will be connected to the command's standard error.
	// The pipe will be closed automatically after the command completes.
	StderrPipe() (io.ReadCloser, error)

	// StdinPipe returns a pipe that will be connected to the command's standard input.
	// The pipe must be closed explicitly after writing is complete.
	StdinPipe() (io.WriteCloser, error)

	// RunContext is like Run but includes a context for cancellation.
	// If the context is canceled before the command completes, the command is killed.
	RunContext(ctx context.Context) error

	// String returns a string representation of the command.
	// This is useful for logging and debugging purposes.
	// Sensitive information like passwords will be masked.
	String() string
}

// RealCommander creates real commands that use os/exec.
// It is the default implementation of the Commander interface.
type RealCommander struct{}

// Command creates a new RealCommand that wraps os/exec.Command.
// It implements the Commander interface by creating a new RealCommand
// instance for the given command name and arguments.
func (c RealCommander) Command(name string, args ...string) Command {
	return &RealCommand{cmd: exec.Command(name, args...)}
}

// LookPath searches for an executable named file in the directories named by the PATH environment variable.
// It delegates to the underlying exec.LookPath function.
func (c RealCommander) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// RealCommand wraps exec.Cmd to implement the Command interface.
// It provides a concrete implementation that delegates to the standard
// library's os/exec package.
type RealCommand struct {
	cmd *exec.Cmd
}

// Run starts the command and waits for it to complete.
// It delegates to the underlying exec.Cmd.Run method.
func (c *RealCommand) Run() error {
	return c.cmd.Run()
}

// Start starts the command but does not wait for it to complete.
// It delegates to the underlying exec.Cmd.Start method.
func (c *RealCommand) Start() error {
	return c.cmd.Start()
}

// Wait waits for the command to complete after calling Start.
// It delegates to the underlying exec.Cmd.Wait method.
func (c *RealCommand) Wait() error {
	return c.cmd.Wait()
}

// Output runs the command and returns its standard output.
// It delegates to the underlying exec.Cmd.Output method.
func (c *RealCommand) Output() ([]byte, error) {
	return c.cmd.Output()
}

// CombinedOutput runs the command and returns its combined standard output and standard error.
// It delegates to the underlying exec.Cmd.CombinedOutput method.
func (c *RealCommand) CombinedOutput() ([]byte, error) {
	return c.cmd.CombinedOutput()
}

// SetDir sets the working directory for the command.
// It returns the command instance to allow method chaining.
func (c *RealCommand) SetDir(dir string) Command {
	c.cmd.Dir = dir
	return c
}

// SetEnv sets the environment variables for the command.
// It returns the command instance to allow method chaining.
func (c *RealCommand) SetEnv(env []string) Command {
	c.cmd.Env = env
	return c
}

// SetStdout sets the standard output for the command.
// It returns the command instance to allow method chaining.
func (c *RealCommand) SetStdout(stdout io.Writer) Command {
	c.cmd.Stdout = stdout
	return c
}

// SetStderr sets the standard error for the command.
// It returns the command instance to allow method chaining.
func (c *RealCommand) SetStderr(stderr io.Writer) Command {
	c.cmd.Stderr = stderr
	return c
}

// SetStdin sets the standard input for the command.
// It returns the command instance to allow method chaining.
func (c *RealCommand) SetStdin(stdin io.Reader) Command {
	c.cmd.Stdin = stdin
	return c
}

// StdoutPipe returns a pipe that will be connected to the command's standard output.
// It delegates to the underlying exec.Cmd.StdoutPipe method.
func (c *RealCommand) StdoutPipe() (io.ReadCloser, error) {
	return c.cmd.StdoutPipe()
}

// StderrPipe returns a pipe that will be connected to the command's standard error.
// It delegates to the underlying exec.Cmd.StderrPipe method.
func (c *RealCommand) StderrPipe() (io.ReadCloser, error) {
	return c.cmd.StderrPipe()
}

// StdinPipe returns a pipe that will be connected to the command's standard input.
// It delegates to the underlying exec.Cmd.StdinPipe method.
func (c *RealCommand) StdinPipe() (io.WriteCloser, error) {
	return c.cmd.StdinPipe()
}

// RunContext is like Run but includes a context for cancellation.
// It creates a new command using exec.CommandContext and runs it.
func (c *RealCommand) RunContext(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, c.cmd.Path, c.cmd.Args[1:]...)
	cmd.Dir = c.cmd.Dir
	cmd.Env = c.cmd.Env
	cmd.Stdin = c.cmd.Stdin
	cmd.Stdout = c.cmd.Stdout
	cmd.Stderr = c.cmd.Stderr
	return cmd.Run()
}

// String returns a string representation of the command.
// It masks sensitive arguments containing passwords or keys for security.
func (c *RealCommand) String() string {
	if c.cmd == nil || len(c.cmd.Args) == 0 {
		return ""
	}

	args := make([]string, len(c.cmd.Args))
	copy(args, c.cmd.Args)

	// Mask sensitive arguments like passwords
	for i, arg := range args {
		if strings.HasPrefix(arg, "--password=") ||
			strings.HasPrefix(arg, "-p=") ||
			strings.HasPrefix(arg, "--pass=") ||
			strings.HasPrefix(arg, "--key=") {
			args[i] = strings.Split(arg, "=")[0] + "=********"
		}
	}

	return fmt.Sprintf("%s %s", c.cmd.Path, strings.Join(args[1:], " "))
}
