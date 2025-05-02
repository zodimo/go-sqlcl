package container

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestSQLclRawExecution tests the direct execution of SQLcl commands
// This is primarily a debug test and should be skipped in regular test runs
func TestSQLclRawExecution(t *testing.T) {
	// Skip this test by default as it takes too long to run
	// Only run it when explicitly requested with ENABLE_SQLCL_DEBUG_TESTS=true
	if os.Getenv("ENABLE_SQLCL_DEBUG_TESTS") != "true" {
		t.Skip("Skipping SQLcl raw execution debug test (set ENABLE_SQLCL_DEBUG_TESTS=true to run)")
	}

	// Test SQLcl direct execution
	t.Run("DirectSQLclExecution", func(t *testing.T) {
		// Get the path to the SQLcl executable
		sqlclPath := os.Getenv("SQLCL_PATH")
		if sqlclPath == "" {
			// Default to the one typically used in the container tests
			sqlclPath = "/home/jaco/Sources/sqlcl-25.1.1.113.2054/bin/sql"
		}

		// Ensure the file exists
		if _, err := os.Stat(sqlclPath); os.IsNotExist(err) {
			t.Skipf("SQLcl not found at %s: %v", sqlclPath, err)
			return
		}

		// Set up command with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Prepare a simple query
		connectStr := "system/oracle@localhost:1521/XEPDB1"
		args := []string{connectStr, "-S", "-L"}

		cmd := exec.CommandContext(ctx, sqlclPath, args...)

		// Set up pipes for stdin/stdout
		stdin, err := cmd.StdinPipe()
		require.NoError(t, err, "Failed to create stdin pipe")

		var outBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = os.Stderr

		// Start the command
		err = cmd.Start()
		require.NoError(t, err, "Failed to start SQLcl command")

		// Write the SQL commands
		_, err = io.WriteString(stdin, "SELECT 1 FROM dual;\nexit;\n")
		require.NoError(t, err, "Failed to write SQL commands")
		stdin.Close()

		// Set up a goroutine to copy stdout to our buffer with a timeout
		done := make(chan struct{})
		go func() {
			err = cmd.Wait()
			close(done)
		}()

		// Wait for command to complete with timeout
		select {
		case <-done:
			// Command completed
			t.Logf("SQLcl command output: %s", outBuf.String())
		case <-time.After(25 * time.Second):
			t.Error("SQLcl command timed out after 25 seconds")
			cmd.Process.Kill()
			t.FailNow()
		}
	})
}
