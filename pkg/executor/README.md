# Executor Package

This package provides an abstraction layer for executing external commands, making it easier to test and mock command execution in Go applications. It wraps the functionality of the `os/exec` package with interfaces that can be easily mocked for testing.

## Overview

The executor package defines two main interfaces:

- `Commander`: Creates executable commands
- `Command`: Represents an executable command with methods to configure and run it

It also provides concrete implementations:

- `RealCommander`/`RealCommand`: Production implementation that wraps `os/exec`
- `MockCommander`/`MockCommand`: Test implementation that simulates command execution

## Usage

### Basic Usage

```go
import "github.com/zodimo/go-sqlcl/pkg/executor"

// Create a commander
commander := &executor.RealCommander{}

// Create a command
cmd := commander.Command("echo", "hello")

// Run the command
err := cmd.Run()
if err != nil {
    // Handle error
}

// Or get the output
output, err := cmd.Output()
if err != nil {
    // Handle error
}
fmt.Println(string(output))
```

### Setting Command Properties

```go
// Set working directory
cmd.SetDir("/tmp")

// Set environment variables
cmd.SetEnv([]string{"VAR=value"})

// Redirect stdout/stderr
cmd.SetStdout(os.Stdout)
cmd.SetStderr(os.Stderr)

// Provide stdin
cmd.SetStdin(strings.NewReader("input data"))
```

### Using Pipes

```go
// Get stdout pipe
stdout, err := cmd.StdoutPipe()
if err != nil {
    // Handle error
}

// Start the command
if err := cmd.Start(); err != nil {
    // Handle error
}

// Read from stdout pipe
data, err := io.ReadAll(stdout)
if err != nil {
    // Handle error
}

// Wait for command to finish
if err := cmd.Wait(); err != nil {
    // Handle error
}
```

### Using Context

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Run with context to enable timeout
err := cmd.RunContext(ctx)
if err != nil {
    // Check if error was due to context timeout
    if errors.Is(err, context.DeadlineExceeded) {
        // Handle timeout
    }
    // Handle other errors
}
```

## Making Code Testable

To make your code testable, accept a Commander parameter in functions that execute commands:

```go
func ExecCommand(args []string, commander executor.Commander) error {
    // Use default commander if none provided
    if commander == nil {
        commander = &executor.RealCommander{}
    }
    
    cmd := commander.Command("sql", args...)
    return cmd.Run()
}
```

## Testing with Mocks

### Basic Mocking

```go
func TestExecCommand(t *testing.T) {
    // Create a mock commander
    mockCommander := executor.NewMockCommander()
    
    // Call the function with the mock
    err := ExecCommand([]string{"update"}, mockCommander)
    assert.NoError(t, err)
    
    // Verify command was called with expected arguments
    mockCmd := mockCommander.LastCommand.(*executor.MockCommand)
    assert.Equal(t, "sql", mockCmd.GetCommandName())
    assert.Contains(t, mockCmd.GetCommandArgs(), "update")
}
```

### Simulating Specific Outputs

```go
func TestExecCommandOutput(t *testing.T) {
    // Create a mock commander that returns specific output
    expectedOutput := "command output"
    mockCommander := executor.NewMockCommanderWithOutput(expectedOutput, nil)
    
    // Get the output
    cmd := mockCommander.Command("sql", "status")
    output, err := cmd.Output()
    
    // Verify results
    assert.NoError(t, err)
    assert.Equal(t, expectedOutput, string(output))
}
```

### Simulating Errors

```go
func TestExecCommandError(t *testing.T) {
    // Create a mock commander that fails
    expectedErr := fmt.Errorf("command failed")
    mockCommander := executor.NewMockCommanderWithError(expectedErr)
    
    // Call the function with the mock
    err := ExecCommand([]string{"update"}, mockCommander)
    
    // Verify the error was returned
    assert.Error(t, err)
    assert.Equal(t, expectedErr, err)
}
```

### Custom Command Behavior

```go
func TestExecCommandCustom(t *testing.T) {
    mockCommander := executor.NewMockCommander()
    mockCmd := &executor.MockCommand{}
    
    // Configure the mock command with custom behavior
    mockCmd.SetRunFunc(func() error {
        // Custom logic here
        return nil
    })
    
    // Configure commander to return our custom command
    mockCommander.CommandFunc = func(name string, args ...string) executor.Command {
        mockCmd.CalledWithName = name
        mockCmd.CalledWithArgs = args
        return mockCmd
    }
    
    // Use the mock
    ExecCommand([]string{"update"}, mockCommander)
    
    // Verify
    assert.True(t, mockCmd.RunCalled)
    assert.Equal(t, "sql", mockCmd.CalledWithName)
}
```

## Best Practices

1. **Dependency Injection**: Always accept a Commander parameter in functions that execute external commands
2. **Default Behavior**: Provide a default RealCommander when none is provided
3. **Sensitive Information**: Use the String() method for logging commands as it masks sensitive information like passwords
4. **Test Coverage**: Create tests that verify both success and error cases
5. **Mocking Strategy**: Use the helper functions (NewMockCommander, NewMockCommanderWithOutput, etc.) to create mocks with the desired behavior
6. **Service Pattern**: For complex command sequences, consider creating a specialized service that encapsulates the command logic
7. **Context Support**: Use RunContext to support timeouts and cancellation of long-running commands
8. **Error Handling**: Be explicit about handling command errors and differentiate between different error types when possible
9. **LookPath Pattern**: Similar to Command execution, always accept a Commander parameter in functions that use LookPath to make them testable

## Using LookPath

The `Commander` interface also provides a `LookPath` method that abstracts `os/exec.LookPath`:

```go
// Find the path to an executable
commander := &executor.RealCommander{}
path, err := commander.LookPath("sql")
if err != nil {
    // Handle error (executable not found)
}
fmt.Println("sql path:", path)
```

### Finding and Executing Commands

A common pattern is to first find the path to an executable and then run it:

```go
// Find the path to an executable and then execute it
commander := &executor.RealCommander{}

// Find the executable path
sqlbasePath, err := commander.LookPath("sql")
if err != nil {
    return fmt.Errorf("sql not found in PATH: %w", err)
}

// Execute the command using the found path
cmd := commander.Command(sqlbasePath, "update")
if err := cmd.Run(); err != nil {
    return fmt.Errorf("failed to run sql update: %w", err)
}
```

### Mocking LookPath in Tests

The package provides a helper function to create mock commanders with custom LookPath behavior:

```go
func TestMyFunction(t *testing.T) {
    // Create a mock commander with custom LookPath behavior
    mockCommander := executor.NewMockCommanderWithLookPath("/mock/path/sql", nil)
    
    // Call the function under test with the mock
    result := MyFunction(mockCommander)
    
    // Verify LookPath was called with the expected argument
    assert.True(t, mockCommander.LookPathCalled)
    assert.Equal(t, "sql", mockCommander.LastLookPathFile)
}

// Testing the error case
func TestMyFunctionLookPathError(t *testing.T) {
    // Create a mock commander that returns an error for LookPath
    expectedErr := fmt.Errorf("executable not found")
    mockCommander := executor.NewMockCommanderWithLookPath("", expectedErr)
    
    // Call the function under test with the mock
    err := MyFunction(mockCommander)
    
    // Verify the error was handled properly
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not found in PATH")
}
```