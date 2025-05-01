package executor

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zodimo/go-sqlcl/internal/process"
)

// MockSQLCLProcess implements a mock of the SQLCLProcess for testing
type MockSQLCLProcess struct {
	running       bool
	commandSent   string
	outputToSend  string
	errorToSend   string
	sendError     error
	readError     error
	readErrError  error
	executeDelay  time.Duration
	executeCount  int
	promptPattern *regexp.Regexp
}

// Ensure MockSQLCLProcess implements SQLCLProcessInterface
var _ SQLCLProcessInterface = (*MockSQLCLProcess)(nil)

func NewMockSQLCLProcess() *MockSQLCLProcess {
	return &MockSQLCLProcess{
		running:       true,
		promptPattern: regexp.MustCompile(DefaultPromptPattern),
	}
}

func (m *MockSQLCLProcess) IsRunning() bool {
	return m.running
}

func (m *MockSQLCLProcess) SendCommand(ctx context.Context, command string) error {
	m.commandSent = command
	m.executeCount++

	if m.sendError != nil {
		return m.sendError
	}

	// Simulate execution delay
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(m.executeDelay):
		// Continue execution
	}

	return nil
}

func (m *MockSQLCLProcess) ReadOutput(ctx context.Context) (string, error) {
	if m.readError != nil {
		return "", m.readError
	}

	// Simulate execution delay
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(m.executeDelay):
		// Continue execution
	}

	return m.outputToSend + "\nSQL> ", nil
}

func (m *MockSQLCLProcess) ReadError(ctx context.Context) (string, error) {
	if m.readErrError != nil {
		return "", m.readErrError
	}

	return m.errorToSend, nil
}

// WithRunning sets the running state
func (m *MockSQLCLProcess) WithRunning(running bool) *MockSQLCLProcess {
	m.running = running
	return m
}

// WithOutputToSend sets the output to send when ReadOutput is called
func (m *MockSQLCLProcess) WithOutputToSend(output string) *MockSQLCLProcess {
	m.outputToSend = output
	return m
}

// WithErrorToSend sets the error to send when ReadError is called
func (m *MockSQLCLProcess) WithErrorToSend(err string) *MockSQLCLProcess {
	m.errorToSend = err
	return m
}

// WithSendError sets the error to return when SendCommand is called
func (m *MockSQLCLProcess) WithSendError(err error) *MockSQLCLProcess {
	m.sendError = err
	return m
}

// WithReadError sets the error to return when ReadOutput is called
func (m *MockSQLCLProcess) WithReadError(err error) *MockSQLCLProcess {
	m.readError = err
	return m
}

// WithReadErrError sets the error to return when ReadError is called
func (m *MockSQLCLProcess) WithReadErrError(err error) *MockSQLCLProcess {
	m.readErrError = err
	return m
}

// WithExecuteDelay sets the delay to simulate execution time
func (m *MockSQLCLProcess) WithExecuteDelay(delay time.Duration) *MockSQLCLProcess {
	m.executeDelay = delay
	return m
}

// TestNewCommandExecutor tests the creation of a new CommandExecutor
func TestNewCommandExecutor(t *testing.T) {
	mockProcess := NewMockSQLCLProcess()

	// Test with default options
	executor := NewCommandExecutor(mockProcess)
	assert.NotNil(t, executor)
	assert.Equal(t, DefaultTimeout, executor.timeout)
	assert.Equal(t, DefaultPromptPattern, executor.promptPattern.String())

	// Test with custom options
	customTimeout := 30 * time.Second
	customPattern := "Custom>"
	executor = NewCommandExecutor(
		mockProcess,
		WithTimeout(customTimeout),
		WithPromptPattern(customPattern),
	)
	assert.NotNil(t, executor)
	assert.Equal(t, customTimeout, executor.timeout)
	assert.Equal(t, customPattern, executor.promptPattern.String())
}

// TestExecuteSuccessful tests successful command execution
func TestExecuteSuccessful(t *testing.T) {
	mockProcess := NewMockSQLCLProcess().
		WithOutputToSend("Query executed successfully\n1 rows selected").
		WithErrorToSend("")

	executor := NewCommandExecutor(mockProcess)

	result, err := executor.Execute(context.Background(), "SELECT * FROM dual")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Query executed successfully\n1 rows selected\nSQL> ", result.Output)
	assert.Equal(t, "", result.Error)
	assert.True(t, result.IsSuccessful)
	assert.Equal(t, QueryCommand, result.CommandType)
	assert.Greater(t, result.Duration, time.Duration(0))
	assert.Equal(t, "SELECT * FROM dual", mockProcess.commandSent)
}

// TestExecuteWithError tests command execution with error
func TestExecuteWithError(t *testing.T) {
	mockProcess := NewMockSQLCLProcess().
		WithOutputToSend("ORA-00942: table or view does not exist").
		WithErrorToSend("")

	executor := NewCommandExecutor(mockProcess)

	result, err := executor.Execute(context.Background(), "SELECT * FROM nonexistent_table")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "ORA-00942: table or view does not exist\nSQL> ", result.Output)
	assert.Equal(t, "", result.Error)
	assert.False(t, result.IsSuccessful)
	assert.Equal(t, QueryCommand, result.CommandType)
}

// TestExecuteWithStderr tests command execution with stderr output
func TestExecuteWithStderr(t *testing.T) {
	mockProcess := NewMockSQLCLProcess().
		WithOutputToSend("Some output").
		WithErrorToSend("Error in SQL command")

	executor := NewCommandExecutor(mockProcess)

	result, err := executor.Execute(context.Background(), "SELECT * FROM table")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Some output\nSQL> ", result.Output)
	assert.Equal(t, "Error in SQL command", result.Error)
	assert.False(t, result.IsSuccessful)
}

// TestExecuteWithProcessNotRunning tests command execution when process is not running
func TestExecuteWithProcessNotRunning(t *testing.T) {
	mockProcess := NewMockSQLCLProcess().
		WithRunning(false)

	executor := NewCommandExecutor(mockProcess)

	result, err := executor.Execute(context.Background(), "SELECT * FROM dual")

	assert.Error(t, err)
	assert.Equal(t, process.ErrProcessNotRunning, err)
	assert.Nil(t, result)
}

// TestExecuteWithNilProcess tests command execution with nil process
func TestExecuteWithNilProcess(t *testing.T) {
	executor := NewCommandExecutor(nil)

	result, err := executor.Execute(context.Background(), "SELECT * FROM dual")

	assert.Error(t, err)
	assert.Equal(t, ErrProcessNotAvailable, err)
	assert.Nil(t, result)
}

// TestExecuteWithSendCommandError tests command execution when sending command fails
func TestExecuteWithSendCommandError(t *testing.T) {
	sendErr := errors.New("send command error")
	mockProcess := NewMockSQLCLProcess().
		WithSendError(sendErr)

	executor := NewCommandExecutor(mockProcess)

	result, err := executor.Execute(context.Background(), "SELECT * FROM dual")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send command")
	assert.Nil(t, result)
}

// TestExecuteWithReadOutputError tests command execution when reading output fails
func TestExecuteWithReadOutputError(t *testing.T) {
	readErr := errors.New("read output error")
	mockProcess := NewMockSQLCLProcess().
		WithReadError(readErr)

	executor := NewCommandExecutor(mockProcess)

	result, err := executor.Execute(context.Background(), "SELECT * FROM dual")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read output")
	assert.Nil(t, result)
}

// TestExecuteWithReadErrorOutputError tests command execution when reading error output fails
func TestExecuteWithReadErrorOutputError(t *testing.T) {
	readErrErr := errors.New("read error output error")
	mockProcess := NewMockSQLCLProcess().
		WithReadErrError(readErrErr)

	executor := NewCommandExecutor(mockProcess)

	result, err := executor.Execute(context.Background(), "SELECT * FROM dual")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read error output")
	assert.Nil(t, result)
}

// TestExecuteWithTimeout tests command execution with timeout
func TestExecuteWithTimeout(t *testing.T) {
	mockProcess := NewMockSQLCLProcess().
		WithExecuteDelay(200 * time.Millisecond)

	executor := NewCommandExecutor(mockProcess, WithTimeout(100*time.Millisecond))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	result, err := executor.Execute(ctx, "SELECT * FROM dual")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
	assert.Nil(t, result)
}

// TestExecuteQuery tests the ExecuteQuery method
func TestExecuteQuery(t *testing.T) {
	mockProcess := NewMockSQLCLProcess().
		WithOutputToSend("Query result")

	executor := NewCommandExecutor(mockProcess)

	result, err := executor.ExecuteQuery(context.Background(), "SELECT * FROM dual")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Query result\nSQL> ", result.Output)
	assert.Equal(t, QueryCommand, result.CommandType)
}

// TestExecuteDDL tests the ExecuteDDL method
func TestExecuteDDL(t *testing.T) {
	mockProcess := NewMockSQLCLProcess().
		WithOutputToSend("Table created")

	executor := NewCommandExecutor(mockProcess)

	result, err := executor.ExecuteDDL(context.Background(), "CREATE TABLE test (id NUMBER)")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Table created\nSQL> ", result.Output)
	assert.Equal(t, DDLCommand, result.CommandType)
}

// TestExecuteSQLCLCommand tests the ExecuteSQLCLCommand method
func TestExecuteSQLCLCommand(t *testing.T) {
	mockProcess := NewMockSQLCLProcess().
		WithOutputToSend("Command executed")

	executor := NewCommandExecutor(mockProcess)

	result, err := executor.ExecuteSQLCLCommand(context.Background(), "SET LINESIZE 100")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Command executed\nSQL> ", result.Output)
	assert.Equal(t, SQLCLCommand, result.CommandType)
}

// TestDetermineCommandType tests the determineCommandType method
func TestDetermineCommandType(t *testing.T) {
	mockProcess := NewMockSQLCLProcess()
	executor := NewCommandExecutor(mockProcess)

	// Test query commands
	assert.Equal(t, QueryCommand, executor.determineCommandType("SELECT * FROM dual"))
	assert.Equal(t, QueryCommand, executor.determineCommandType("INSERT INTO table VALUES (1)"))
	assert.Equal(t, QueryCommand, executor.determineCommandType("UPDATE table SET col = 1"))
	assert.Equal(t, QueryCommand, executor.determineCommandType("DELETE FROM table"))

	// Test DDL commands
	assert.Equal(t, DDLCommand, executor.determineCommandType("CREATE TABLE test (id NUMBER)"))
	assert.Equal(t, DDLCommand, executor.determineCommandType("ALTER TABLE test ADD col NUMBER"))
	assert.Equal(t, DDLCommand, executor.determineCommandType("DROP TABLE test"))
	assert.Equal(t, DDLCommand, executor.determineCommandType("TRUNCATE TABLE test"))
	assert.Equal(t, DDLCommand, executor.determineCommandType("COMMENT ON TABLE test IS 'comment'"))
	assert.Equal(t, DDLCommand, executor.determineCommandType("GRANT SELECT ON test TO user"))
	assert.Equal(t, DDLCommand, executor.determineCommandType("REVOKE SELECT ON test FROM user"))

	// Test SQLcl commands
	assert.Equal(t, SQLCLCommand, executor.determineCommandType("SET LINESIZE 100"))
	assert.Equal(t, SQLCLCommand, executor.determineCommandType("SHOW ERRORS"))
	assert.Equal(t, SQLCLCommand, executor.determineCommandType("DESCRIBE test"))
	assert.Equal(t, SQLCLCommand, executor.determineCommandType("DESC test"))
	assert.Equal(t, SQLCLCommand, executor.determineCommandType("HELP INDEX"))
	assert.Equal(t, SQLCLCommand, executor.determineCommandType("EXIT"))
	assert.Equal(t, SQLCLCommand, executor.determineCommandType("QUIT"))
	assert.Equal(t, SQLCLCommand, executor.determineCommandType("CONNECT user/pass@db"))
	assert.Equal(t, SQLCLCommand, executor.determineCommandType("CONN user/pass@db"))

	// Test case insensitivity
	assert.Equal(t, QueryCommand, executor.determineCommandType("select * from dual"))
	assert.Equal(t, DDLCommand, executor.determineCommandType("create table test (id number)"))
	assert.Equal(t, SQLCLCommand, executor.determineCommandType("set linesize 100"))
}

// TestIsCommandSuccessful tests the isCommandSuccessful method
func TestIsCommandSuccessful(t *testing.T) {
	mockProcess := NewMockSQLCLProcess()
	executor := NewCommandExecutor(mockProcess)

	// Test successful outputs
	assert.True(t, executor.isCommandSuccessful("Command executed successfully", ""))
	assert.True(t, executor.isCommandSuccessful("Table created", ""))
	assert.True(t, executor.isCommandSuccessful("1 row selected", ""))

	// Test error patterns in output
	assert.False(t, executor.isCommandSuccessful("ORA-00942: table or view does not exist", ""))
	assert.False(t, executor.isCommandSuccessful("ERROR: syntax error", ""))
	assert.False(t, executor.isCommandSuccessful("SP2-0640: Not connected", ""))

	// Test error output
	assert.False(t, executor.isCommandSuccessful("Command executed", "Error in executing command"))
}
