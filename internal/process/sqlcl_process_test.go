package process

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

// mockPipe implements ReadCloser and WriteCloser for testing
type mockPipe struct {
	buffer      *bytes.Buffer
	closed      bool
	mu          sync.Mutex
	writeErr    error
	readErr     error
	closeErr    error
	writeCalled bool
	readCalled  bool
	closeCalled bool
	// Add a delay feature for simulating timeouts
	readDelay time.Duration
}

func newMockPipe() *mockPipe {
	return &mockPipe{
		buffer: new(bytes.Buffer),
	}
}

func (mp *mockPipe) Read(p []byte) (n int, err error) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.readCalled = true

	// Simulate delay for timeout testing
	if mp.readDelay > 0 {
		time.Sleep(mp.readDelay)
	}

	if mp.closed {
		return 0, io.EOF
	}
	if mp.readErr != nil {
		return 0, mp.readErr
	}
	return mp.buffer.Read(p)
}

func (mp *mockPipe) Write(p []byte) (n int, err error) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.writeCalled = true
	if mp.closed {
		return 0, errors.New("write on closed pipe")
	}
	if mp.writeErr != nil {
		return 0, mp.writeErr
	}
	return mp.buffer.Write(p)
}

func (mp *mockPipe) Close() error {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.closeCalled = true
	mp.closed = true
	return mp.closeErr
}

func (mp *mockPipe) WriteString(s string) (n int, err error) {
	return mp.Write([]byte(s))
}

// TestNewSQLCLProcess tests the creation of a new SQLCLProcess
func TestNewSQLCLProcess(t *testing.T) {
	// Test default values
	process := NewSQLCLProcess()
	if process.path != DefaultSQLCLPath {
		t.Errorf("Expected default path %s, got %s", DefaultSQLCLPath, process.path)
	}
	if process.timeout != DefaultTimeout {
		t.Errorf("Expected default timeout %s, got %s", DefaultTimeout, process.timeout)
	}

	// Test with options
	customPath := "/custom/path/to/sql"
	customTimeout := 60 * time.Second
	customArgs := []string{"--silent", "--no-banner"}

	process = NewSQLCLProcess(
		WithPath(customPath),
		WithTimeout(customTimeout),
		WithArgs(customArgs),
	)

	if process.path != customPath {
		t.Errorf("Expected custom path %s, got %s", customPath, process.path)
	}
	if process.timeout != customTimeout {
		t.Errorf("Expected custom timeout %s, got %s", customTimeout, process.timeout)
	}
	if len(process.args) != len(customArgs) {
		t.Errorf("Expected %d args, got %d", len(customArgs), len(process.args))
	}
	for i, arg := range customArgs {
		if process.args[i] != arg {
			t.Errorf("Expected arg %s at position %d, got %s", arg, i, process.args[i])
		}
	}
}

// TestIsRunning tests the IsRunning method
func TestIsRunning(t *testing.T) {
	// Test when process is running
	process := &SQLCLProcess{
		isRunning: true,
	}
	if !process.IsRunning() {
		t.Error("IsRunning() returned false when process is running")
	}

	// Test when process is not running
	process = &SQLCLProcess{
		isRunning: false,
	}
	if process.IsRunning() {
		t.Error("IsRunning() returned true when process is not running")
	}
}

// TestSendCommand tests the SendCommand method
func TestSendCommand(t *testing.T) {
	tests := []struct {
		name       string
		command    string
		isRunning  bool
		writeErr   error
		wantErr    bool
		wantErrMsg string
		timeout    bool
	}{
		{
			name:      "successful command",
			command:   "select * from dual",
			isRunning: true,
			wantErr:   false,
		},
		{
			name:       "not running",
			command:    "select * from dual",
			isRunning:  false,
			wantErr:    true,
			wantErrMsg: "sqlcl process is not running",
		},
		{
			name:       "write error",
			command:    "select * from dual",
			isRunning:  true,
			writeErr:   errors.New("write error"),
			wantErr:    true,
			wantErrMsg: "failed to send command to sqlcl: write error",
		},
		{
			name:       "timeout",
			command:    "select * from dual",
			isRunning:  true,
			timeout:    true,
			wantErr:    true,
			wantErrMsg: "send command timeout: context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock stdin
			mockStdin := newMockPipe()
			mockStdin.writeErr = tt.writeErr

			// Create process under test with the mock stdin
			process := &SQLCLProcess{
				isRunning: tt.isRunning,
				stdin:     mockStdin,
				timeout:   10 * time.Millisecond, // Short timeout for tests
			}

			// Create context with timeout if needed
			ctx := context.Background()
			if tt.timeout {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 1*time.Nanosecond)
				defer cancel()
				time.Sleep(1 * time.Millisecond) // Ensure the context is canceled
			}

			// Execute the method
			err := process.SendCommand(ctx, tt.command)

			// Verify results
			if (err != nil) != tt.wantErr {
				t.Errorf("SendCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("SendCommand() error = %v, wantErrMsg %v", err, tt.wantErrMsg)
				return
			}

			// Check that the command was written to stdin correctly
			if tt.isRunning && !tt.wantErr && !tt.timeout {
				if !mockStdin.writeCalled {
					t.Error("SendCommand() did not write to stdin")
				}

				// Check that the command has a newline at the end
				written := mockStdin.buffer.String()
				if !strings.HasSuffix(written, "\n") {
					t.Errorf("SendCommand() did not append newline to command: %s", written)
				}

				// Check that the command was written correctly
				expectedCommand := tt.command
				if !strings.HasSuffix(expectedCommand, "\n") {
					expectedCommand += "\n"
				}
				if written != expectedCommand {
					t.Errorf("SendCommand() wrote %q, expected %q", written, expectedCommand)
				}
			}
		})
	}
}

// TestReadOutput tests the ReadOutput method
func TestReadOutput(t *testing.T) {
	tests := []struct {
		name       string
		isRunning  bool
		output     string
		readErr    error
		wantErr    bool
		wantErrMsg string
		timeout    bool
	}{
		{
			name:      "successful read",
			isRunning: true,
			output:    "output line 1\noutput line 2\nSQL>",
			wantErr:   false,
		},
		{
			name:       "not running",
			isRunning:  false,
			wantErr:    true,
			wantErrMsg: "sqlcl process is not running",
		},
		{
			name:       "read error",
			isRunning:  true,
			readErr:    errors.New("read error"),
			wantErr:    true,
			wantErrMsg: "error reading from sqlcl stdout: read error",
		},
		{
			name:       "timeout",
			isRunning:  true,
			output:     "output that never completes", // No SQL> prompt to terminate
			timeout:    true,
			wantErr:    true,
			wantErrMsg: "read output timeout: context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock stdout
			mockStdout := newMockPipe()
			mockStdout.readErr = tt.readErr

			// For timeout test, set a read delay that's longer than the context timeout
			if tt.timeout {
				mockStdout.readDelay = 50 * time.Millisecond
			}

			if tt.output != "" {
				mockStdout.WriteString(tt.output)
			}

			// Create process under test with the mock stdout
			process := &SQLCLProcess{
				isRunning: tt.isRunning,
				stdout:    mockStdout,
				timeout:   10 * time.Millisecond, // Short timeout for tests
			}

			// Create context with timeout if needed
			ctx := context.Background()
			if tt.timeout {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 5*time.Millisecond)
				defer cancel()
			}

			// Execute the method
			output, err := process.ReadOutput(ctx)

			// Verify results
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("ReadOutput() error = %v, wantErrMsg %v", err, tt.wantErrMsg)
				return
			}

			// Check that the output matches what we expected
			if tt.isRunning && !tt.wantErr && !tt.timeout {
				expectedOutput := tt.output
				if !strings.HasSuffix(expectedOutput, "\n") {
					expectedOutput += "\n"
				}
				if output != expectedOutput {
					t.Errorf("ReadOutput() output = %q, expected %q", output, expectedOutput)
				}
			}
		})
	}
}

// TestReadError tests the ReadError method
func TestReadError(t *testing.T) {
	tests := []struct {
		name       string
		isRunning  bool
		output     string
		readErr    error
		wantErr    bool
		wantErrMsg string
		timeout    bool
	}{
		{
			name:      "successful read",
			isRunning: true,
			output:    "error line 1\nerror line 2\n",
			wantErr:   false,
		},
		{
			name:       "not running",
			isRunning:  false,
			wantErr:    true,
			wantErrMsg: "sqlcl process is not running",
		},
		{
			name:       "read error",
			isRunning:  true,
			readErr:    errors.New("read error"),
			wantErr:    true,
			wantErrMsg: "error reading from sqlcl stderr: read error",
		},
		{
			name:       "timeout",
			isRunning:  true,
			timeout:    true,
			wantErr:    true,
			wantErrMsg: "read error timeout: context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock stderr
			mockStderr := newMockPipe()
			mockStderr.readErr = tt.readErr

			// For timeout test, set a read delay that's longer than the context timeout
			if tt.timeout {
				mockStderr.readDelay = 50 * time.Millisecond
			}

			if tt.output != "" {
				mockStderr.WriteString(tt.output)
			}

			// Create process under test with the mock stderr
			process := &SQLCLProcess{
				isRunning: tt.isRunning,
				stderr:    mockStderr,
				timeout:   10 * time.Millisecond, // Short timeout for tests
			}

			// Create context with timeout if needed
			ctx := context.Background()
			if tt.timeout {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 5*time.Millisecond)
				defer cancel()
			}

			// Execute the method
			output, err := process.ReadError(ctx)

			// Verify results
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadError() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("ReadError() error = %v, wantErrMsg %v", err, tt.wantErrMsg)
				return
			}

			// Check that the output matches what we expected
			if tt.isRunning && !tt.wantErr && !tt.timeout {
				if output != tt.output {
					t.Errorf("ReadError() output = %q, expected %q", output, tt.output)
				}
			}
		})
	}
}

// TestStartStopErrors tests the error conditions for Start and Stop
func TestStartStopErrors(t *testing.T) {
	t.Run("Start_AlreadyRunning", func(t *testing.T) {
		process := &SQLCLProcess{
			isRunning: true,
		}

		err := process.Start(context.Background())

		if err == nil {
			t.Error("Start() did not return an error when process is already running")
		}
		if err != ErrProcessAlreadyRunning {
			t.Errorf("Start() returned %v, expected %v", err, ErrProcessAlreadyRunning)
		}
	})

	t.Run("Stop_NotRunning", func(t *testing.T) {
		process := &SQLCLProcess{
			isRunning: false,
		}

		err := process.Stop()

		if err == nil {
			t.Error("Stop() did not return an error when process is not running")
		}
		if err != ErrProcessNotRunning {
			t.Errorf("Stop() returned %v, expected %v", err, ErrProcessNotRunning)
		}
	})
}
