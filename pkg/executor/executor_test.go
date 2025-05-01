package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMockCommander(t *testing.T) {
	t.Run("DefaultCommand", func(t *testing.T) {
		commander := NewMockCommander()
		cmd := commander.Command("echo", "hello")

		if cmd == nil {
			t.Fatal("Expected a Command, got nil")
		}

		// Default implementation should return nil error
		err := cmd.Run()
		if err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}
	})

	t.Run("CustomCommandFunc", func(t *testing.T) {
		errMsg := "command error"
		commander := &MockCommander{
			CommandFunc: func(name string, args ...string) Command {
				mockCmd := &MockCommand{
					name: name,
					args: args,
				}
				mockCmd.SetRunFunc(func() error {
					return errors.New(errMsg)
				})
				return mockCmd
			},
		}

		cmd := commander.Command("echo", "hello")
		err := cmd.Run()

		if err == nil || err.Error() != errMsg {
			t.Errorf("Expected error with message %q, got %v", errMsg, err)
		}
	})
}

func TestMockCommand(t *testing.T) {
	t.Run("SetDir", func(t *testing.T) {
		cmd := &MockCommand{}
		returned := cmd.SetDir("/tmp")

		if returned != cmd {
			t.Error("SetDir did not return the command itself")
		}

		if cmd.dir != "/tmp" {
			t.Errorf("Expected dir to be %q, got %q", "/tmp", cmd.dir)
		}
	})

	t.Run("SetEnv", func(t *testing.T) {
		cmd := &MockCommand{}
		env := []string{"FOO=bar"}
		returned := cmd.SetEnv(env)

		if returned != cmd {
			t.Error("SetEnv did not return the command itself")
		}

		if len(cmd.env) != 1 || cmd.env[0] != "FOO=bar" {
			t.Errorf("Expected env to be %v, got %v", env, cmd.env)
		}
	})

	t.Run("SetStdout", func(t *testing.T) {
		cmd := &MockCommand{}
		var buf bytes.Buffer
		returned := cmd.SetStdout(&buf)

		if returned != cmd {
			t.Error("SetStdout did not return the command itself")
		}

		if cmd.stdout != &buf {
			t.Error("Stdout was not set correctly")
		}
	})

	t.Run("RunContext", func(t *testing.T) {
		t.Run("DefaultBehavior", func(t *testing.T) {
			cmd := &MockCommand{}
			ctx := context.Background()

			err := cmd.RunContext(ctx)
			if err != nil {
				t.Errorf("Expected nil error, got %v", err)
			}
		})

		t.Run("CustomBehavior", func(t *testing.T) {
			errMsg := "context error"
			cmd := &MockCommand{}
			ctx := context.Background()

			cmd.SetRunContextFunc(func(ctx context.Context) error {
				return errors.New(errMsg)
			})

			err := cmd.RunContext(ctx)
			if err == nil || err.Error() != errMsg {
				t.Errorf("Expected error with message %q, got %v", errMsg, err)
			}
		})

		t.Run("Cancellation", func(t *testing.T) {
			cmd := &MockCommand{}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()

			cmd.SetRunContextFunc(func(ctx context.Context) error {
				<-ctx.Done()
				return ctx.Err()
			})

			err := cmd.RunContext(ctx)
			if err == nil {
				t.Error("Expected context deadline exceeded error, got nil")
			}
		})
	})

	t.Run("Output", func(t *testing.T) {
		t.Run("DefaultBehavior", func(t *testing.T) {
			cmd := &MockCommand{}

			output, err := cmd.Output()
			if err != nil {
				t.Errorf("Expected nil error, got %v", err)
			}

			if len(output) != 0 {
				t.Errorf("Expected empty output, got %v", output)
			}
		})

		t.Run("CustomBehavior", func(t *testing.T) {
			expected := []byte("hello world")
			cmd := &MockCommand{}

			cmd.SetOutputFunc(func() ([]byte, error) {
				return expected, nil
			})

			output, err := cmd.Output()
			if err != nil {
				t.Errorf("Expected nil error, got %v", err)
			}

			if !bytes.Equal(output, expected) {
				t.Errorf("Expected output %q, got %q", expected, output)
			}
		})
	})
}

func TestRealCommander(t *testing.T) {
	commander := &RealCommander{}

	// Test Command method returns a proper Command
	cmd := commander.Command("echo", "hello")
	assert.NotNil(t, cmd)

	// Verify it's a RealCommand
	_, ok := cmd.(*RealCommand)
	assert.True(t, ok)
}

func TestRealCommand_Run(t *testing.T) {
	// Test with echo command which should succeed
	commander := &RealCommander{}
	cmd := commander.Command("echo", "hello")
	err := cmd.Run()
	assert.NoError(t, err)

	// Test with an invalid command which should fail
	badCmd := commander.Command("commandthatdoesnotexist")
	err = badCmd.Run()
	assert.Error(t, err)
}

func TestRealCommand_Output(t *testing.T) {
	commander := &RealCommander{}
	cmd := commander.Command("echo", "hello")

	output, err := cmd.Output()
	assert.NoError(t, err)
	assert.Contains(t, string(output), "hello")
}

func TestRealCommand_CombinedOutput(t *testing.T) {
	commander := &RealCommander{}
	cmd := commander.Command("echo", "hello")

	output, err := cmd.CombinedOutput()
	assert.NoError(t, err)
	assert.Contains(t, string(output), "hello")
}

func TestRealCommand_SetDir(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := ioutil.TempDir("", "executor-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	commander := &RealCommander{}
	cmd := commander.Command("pwd")

	// Set working directory to the temp dir
	cmd.SetDir(tmpDir)

	output, err := cmd.Output()
	assert.NoError(t, err)

	// Compare the canonical paths to handle symlinks, etc.
	expected, _ := os.Getwd()
	expected = strings.TrimSpace(tmpDir)
	actual := strings.TrimSpace(string(output))
	assert.Equal(t, expected, actual)
}

func TestRealCommand_String(t *testing.T) {
	commander := &RealCommander{}

	// Test normal command
	cmd := commander.Command("ls", "-la")
	realCmd, ok := cmd.(*RealCommand)
	assert.True(t, ok)
	str := realCmd.String()
	assert.Contains(t, str, "ls")
	assert.Contains(t, str, "-la")

	// Test command with password
	cmd = commander.Command("liquibase", "--password=secret", "update")
	realCmd, ok = cmd.(*RealCommand)
	assert.True(t, ok)
	str = realCmd.String()
	assert.Contains(t, str, "liquibase")
	assert.Contains(t, str, "--password=********")
	assert.NotContains(t, str, "secret")
}

func TestRealCommand_SetStdout(t *testing.T) {
	commander := &RealCommander{}
	cmd := commander.Command("echo", "hello")

	var buf bytes.Buffer
	cmd.SetStdout(&buf)

	err := cmd.Run()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "hello")
}

func TestMockCommander_NewHelpers(t *testing.T) {
	// Test with a simple mock
	mock := &MockCommander{
		CommandFunc: func(name string, args ...string) Command {
			return &MockCommand{
				CalledWithName: name,
				CalledWithArgs: args,
			}
		},
	}

	cmd := mock.Command("test", "arg1", "arg2")
	assert.NotNil(t, cmd)

	// Verify the LastCommand is set and args are captured
	assert.NotNil(t, mock.LastCommand)
	mockCmd := cmd.(*MockCommand)
	assert.Equal(t, "test", mockCmd.CalledWithName)
	assert.Equal(t, []string{"arg1", "arg2"}, mockCmd.CalledWithArgs)
}

func TestMockCommand_Run(t *testing.T) {
	// Test with success
	mockSuccess := &MockCommand{
		runFunc: func() error {
			return nil
		},
	}
	assert.NoError(t, mockSuccess.Run())
	assert.True(t, mockSuccess.RunCalled)

	// Test with error
	expectedErr := fmt.Errorf("command failed")
	mockError := &MockCommand{
		runFunc: func() error {
			return expectedErr
		},
	}
	err := mockError.Run()
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.True(t, mockError.RunCalled)
}

func TestNewMockCommanderWithOutput(t *testing.T) {
	// Test success case
	expectedOutput := "test output"
	mock := NewMockCommanderWithOutput(expectedOutput, nil)

	cmd := mock.Command("test")
	output, err := cmd.Output()
	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, string(output))

	// Test combined output
	combined, err := cmd.CombinedOutput()
	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, string(combined))

	// Test with run (should also succeed with no error)
	err = cmd.Run()
	assert.NoError(t, err)

	// Test with error
	expectedErr := fmt.Errorf("output error")
	mockWithErr := NewMockCommanderWithOutput("error output", expectedErr)

	cmdWithErr := mockWithErr.Command("test")
	_, err = cmdWithErr.Output()
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestNewMockCommanderWithError(t *testing.T) {
	expectedErr := fmt.Errorf("command failed")
	mock := NewMockCommanderWithError(expectedErr)

	cmd := mock.Command("test")

	// Test Run
	err := cmd.Run()
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

	// Test Output
	_, err = cmd.Output()
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

	// Test CombinedOutput
	_, err = cmd.CombinedOutput()
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestMockCommand_SetFuncs(t *testing.T) {
	mockCmd := &MockCommand{}

	// Test SetRunFunc
	expectedRunErr := fmt.Errorf("run error")
	mockCmd.SetRunFunc(func() error {
		return expectedRunErr
	})
	err := mockCmd.Run()
	assert.Equal(t, expectedRunErr, err)

	// Test SetOutputFunc
	expectedOutput := []byte("test output")
	mockCmd.SetOutputFunc(func() ([]byte, error) {
		return expectedOutput, nil
	})
	output, err := mockCmd.Output()
	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)

	// Test SetCombinedOutputFunc
	expectedCombined := []byte("combined output")
	mockCmd.SetCombinedOutputFunc(func() ([]byte, error) {
		return expectedCombined, nil
	})
	combined, err := mockCmd.CombinedOutput()
	assert.NoError(t, err)
	assert.Equal(t, expectedCombined, combined)
}

func TestNewMockCommanderWithOutputs(t *testing.T) {
	stdout := "standard output"
	combined := "combined output"

	mock := NewMockCommanderWithOutputs(stdout, combined, nil)
	cmd := mock.Command("test")

	// Test stdout
	output, err := cmd.Output()
	assert.NoError(t, err)
	assert.Equal(t, stdout, string(output))

	// Test combined
	comb, err := cmd.CombinedOutput()
	assert.NoError(t, err)
	assert.Equal(t, combined, string(comb))
}

func TestMockCommand_GetMethods(t *testing.T) {
	name := "test-command"
	args := []string{"arg1", "arg2"}
	dir := "/test/dir"
	env := []string{"VAR=value"}

	mockCmd := &MockCommand{
		name: name,
		args: args,
		dir:  dir,
		env:  env,
	}

	assert.Equal(t, name, mockCmd.GetCommandName())
	assert.Equal(t, args, mockCmd.GetCommandArgs())
	assert.Equal(t, dir, mockCmd.GetDir())
	assert.Equal(t, env, mockCmd.GetEnv())
}

func TestRealCommand_RunContext(t *testing.T) {
	// Skip if in short mode
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Create a temporary directory for testing
	tmpDir, err := ioutil.TempDir("", "executor-context-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup test environment variables
	env := []string{"TEST_VAR=test_value"}

	// Test 1: Verify all command settings are properly propagated
	t.Run("PropagateSettings", func(t *testing.T) {
		// Setup buffers for stdout and stderr
		var stdout, stderr bytes.Buffer
		stdinContent := "test input"
		stdin := strings.NewReader(stdinContent)

		// Create a command that echoes environment variable
		commander := &RealCommander{}
		cmd := commander.Command("sh", "-c", "echo $TEST_VAR && pwd && cat")

		// Set all properties
		cmd.SetDir(tmpDir)
		cmd.SetEnv(env)
		cmd.SetStdout(&stdout)
		cmd.SetStderr(&stderr)
		cmd.SetStdin(stdin)

		// Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Run the command with context
		err := cmd.RunContext(ctx)

		// Verify results
		assert.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, "test_value", "Environment variable not propagated")
		assert.Contains(t, output, tmpDir, "Working directory not propagated")
		assert.Contains(t, output, "test input", "Stdin not propagated")
	})

	// Test 2: Test context cancellation
	t.Run("Cancellation", func(t *testing.T) {
		commander := &RealCommander{}
		// Command that sleeps for 10 seconds
		cmd := commander.Command("sleep", "10")

		// Set directory and env to verify these are also properly handled in cancellation case
		cmd.SetDir(tmpDir)
		cmd.SetEnv(env)

		// Create a context that we'll cancel immediately
		ctx, cancel := context.WithCancel(context.Background())

		// Start a goroutine to cancel the context shortly after the command starts
		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		// Run the command with context that will be cancelled
		err := cmd.RunContext(ctx)

		// Verify the command was cancelled
		assert.Error(t, err)
		// The error message will typically contain "killed" when a process is terminated
		assert.Contains(t, err.Error(), "killed", "Context cancellation should terminate the process")
	})

	// Test 3: Test with timeout context
	t.Run("Timeout", func(t *testing.T) {
		commander := &RealCommander{}
		// Command that sleeps for 5 seconds
		cmd := commander.Command("sleep", "5")

		// Create a context with a short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		// Run the command with context that will timeout
		err := cmd.RunContext(ctx)

		// Verify the command was cancelled due to timeout
		assert.Error(t, err)
		// The error message will typically contain "killed" when a process is terminated due to context timeout
		assert.Contains(t, err.Error(), "killed", "Context timeout should terminate the process")
	})
}

func TestRealCommanderLookPath(t *testing.T) {
	// Skip if in short mode
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	commander := &RealCommander{}
	path, err := commander.LookPath("echo")

	assert.NoError(t, err)
	assert.NotEmpty(t, path)
}

func TestMockCommanderLookPath(t *testing.T) {
	expectedPath := "/mock/path/test"
	commander := NewMockCommanderWithLookPath(expectedPath, nil)

	path, err := commander.LookPath("test")

	assert.NoError(t, err)
	assert.Equal(t, expectedPath, path)
	assert.True(t, commander.LookPathCalled)
	assert.Equal(t, "test", commander.LastLookPathFile)
}

func TestMockCommanderLookPathError(t *testing.T) {
	expectedErr := fmt.Errorf("not found")
	commander := NewMockCommanderWithLookPath("", expectedErr)

	path, err := commander.LookPath("test")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Empty(t, path)
}

func TestNewMockCommanderWithLookPath(t *testing.T) {
	expectedPath := "/custom/path"
	expectedErr := fmt.Errorf("custom error")

	commander := NewMockCommanderWithLookPath(expectedPath, expectedErr)

	// Verify the commander has the expected LookPathFunc
	path, err := commander.LookPath("anything")

	assert.Equal(t, expectedPath, path)
	assert.Equal(t, expectedErr, err)
}
