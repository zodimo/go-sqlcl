package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zodimo/go-sqlcl/pkg/sqlcl"
	"github.com/zodimo/go-sqlcl/pkg/types"
)

// TestFixedClientImplementation tests the fixed client implementation
func TestFixedClientImplementation(t *testing.T) {
	// Skip test if running in CI or if explicitly disabled
	if testing.Short() || isContainerTestDisabled() {
		t.Skip("Skipping container-based test in short mode or with SKIP_CONTAINER_TESTS set")
	}

	// Get the SQLcl path from the default config
	sqlclPath := DefaultContainerTestConfig().SQLclPath

	// Test simple connection
	t.Run("SimpleConnection", func(t *testing.T) {
		// Get connection details
		host, port, err := parseConnectStr(globalConnOpts.ConnectStr)
		require.NoError(t, err, "Failed to parse connect string")

		// Create a connect string
		connectStr := fmt.Sprintf("%s/%s@%s:%s/%s",
			globalConnOpts.Username,
			globalConnOpts.Password,
			host, port, "XEPDB1")

		// Create a client config
		config := &types.ClientConfig{
			SQLclPath:      sqlclPath,
			Timeout:        30 * time.Second,
			ConnectTimeout: 15 * time.Second,
			QueryTimeout:   60 * time.Second,
		}

		// Create a fixed client
		client, err := sqlcl.NewFixedClient(config)
		require.NoError(t, err, "Failed to create fixed client")
		defer client.Close()

		// Connect to the database
		err = client.Connect(globalCtx, connectStr)
		require.NoError(t, err, "Failed to connect to the database")

		// Execute a simple query
		result, err := client.ExecuteSQL(globalCtx, "SELECT 'Fixed client test' FROM dual")
		require.NoError(t, err, "Failed to execute simple query")
		require.True(t, result.Success, "Query should be successful")
		require.Contains(t, result.Message, "Fixed client test", "Query result should contain the test string")
	})

	// Test with transaction
	t.Run("TransactionExecution", func(t *testing.T) {
		// Get connection details
		host, port, err := parseConnectStr(globalConnOpts.ConnectStr)
		require.NoError(t, err, "Failed to parse connect string")

		// Create a connect string
		connectStr := fmt.Sprintf("%s/%s@%s:%s/%s",
			globalConnOpts.Username,
			globalConnOpts.Password,
			host, port, "XEPDB1")

		// Create a client config
		config := &types.ClientConfig{
			SQLclPath:      sqlclPath,
			Timeout:        30 * time.Second,
			ConnectTimeout: 15 * time.Second,
			QueryTimeout:   60 * time.Second,
		}

		// Create a fixed client
		client, err := sqlcl.NewFixedClient(config)
		require.NoError(t, err, "Failed to create fixed client")
		defer client.Close()

		// Connect to the database
		err = client.Connect(globalCtx, connectStr)
		require.NoError(t, err, "Failed to connect to the database")

		// Use a simple transaction with dual to avoid permissions issues
		transactionSQL := `
		BEGIN
			DBMS_OUTPUT.PUT_LINE('Beginning transaction test');
			DBMS_OUTPUT.PUT_LINE('Current timestamp: ' || TO_CHAR(SYSTIMESTAMP, 'YYYY-MM-DD HH24:MI:SS.FF'));
			FOR i IN 1..3 LOOP
				DBMS_OUTPUT.PUT_LINE('Loop iteration ' || i);
			END LOOP;
			DBMS_OUTPUT.PUT_LINE('Transaction test completed');
		END;
		/
		`
		result, err := client.ExecuteSQL(globalCtx, transactionSQL)
		require.NoError(t, err, "Transaction execution should succeed")
		require.True(t, result.Success, "Transaction should be successful")
		t.Logf("Transaction result: %s", result.Message)
	})

	// Test error handling
	t.Run("ErrorHandling", func(t *testing.T) {
		// Get connection details
		host, port, err := parseConnectStr(globalConnOpts.ConnectStr)
		require.NoError(t, err, "Failed to parse connect string")

		// Create a connect string
		connectStr := fmt.Sprintf("%s/%s@%s:%s/%s",
			globalConnOpts.Username,
			globalConnOpts.Password,
			host, port, "XEPDB1")

		// Create a client config
		config := &types.ClientConfig{
			SQLclPath:      sqlclPath,
			Timeout:        30 * time.Second,
			ConnectTimeout: 15 * time.Second,
			QueryTimeout:   60 * time.Second,
		}

		// Create a fixed client
		client, err := sqlcl.NewFixedClient(config)
		require.NoError(t, err, "Failed to create fixed client")
		defer client.Close()

		// Connect to the database
		err = client.Connect(globalCtx, connectStr)
		require.NoError(t, err, "Failed to connect to the database")

		// Execute an invalid query
		result, err := client.ExecuteSQL(globalCtx, "SELECT * FROM non_existent_table")
		require.NoError(t, err, "Error should be captured in result, not returned")
		require.False(t, result.Success, "Query should fail")
		require.Contains(t, result.Message, "ORA-", "Error message should contain Oracle error code")
		t.Logf("Error message: %s", result.Message)
	})

	// Test with SQL script file
	t.Run("FileExecution", func(t *testing.T) {
		// Get connection details
		host, port, err := parseConnectStr(globalConnOpts.ConnectStr)
		require.NoError(t, err, "Failed to parse connect string")

		// Create a connect string
		connectStr := fmt.Sprintf("%s/%s@%s:%s/%s",
			globalConnOpts.Username,
			globalConnOpts.Password,
			host, port, "XEPDB1")

		// Create a client config
		config := &types.ClientConfig{
			SQLclPath:      sqlclPath,
			Timeout:        30 * time.Second,
			ConnectTimeout: 15 * time.Second,
			QueryTimeout:   60 * time.Second,
		}

		// Create a fixed client
		client, err := sqlcl.NewFixedClient(config)
		require.NoError(t, err, "Failed to create fixed client")
		defer client.Close()

		// Connect to the database
		err = client.Connect(globalCtx, connectStr)
		require.NoError(t, err, "Failed to connect to the database")

		// Create a temporary SQL script file
		tempDir, err := ioutil.TempDir("", "sqlcl-test")
		require.NoError(t, err, "Failed to create temp directory")
		defer os.RemoveAll(tempDir)

		scriptContent := `
		-- Simple test script
		SELECT 'Script execution test' FROM dual;
		SELECT 'Multiple statements' FROM dual;
		SELECT 'in one script' FROM dual;
		
		-- Create a test table
		DECLARE
		  v_count NUMBER;
		BEGIN
		  SELECT COUNT(*) INTO v_count FROM user_tables WHERE table_name = 'TEMP_TEST_TABLE';
		  IF v_count > 0 THEN
		    EXECUTE IMMEDIATE 'DROP TABLE temp_test_table';
		  END IF;
		END;
		/

		CREATE GLOBAL TEMPORARY TABLE temp_test_table (
		  id NUMBER,
		  name VARCHAR2(100)
		);

		-- Insert some data
		INSERT INTO temp_test_table VALUES (1, 'Test 1');
		INSERT INTO temp_test_table VALUES (2, 'Test 2');
		
		-- Query the data
		SELECT * FROM temp_test_table ORDER BY id;
		`

		scriptPath := filepath.Join(tempDir, "test_script.sql")
		err = ioutil.WriteFile(scriptPath, []byte(scriptContent), 0644)
		require.NoError(t, err, "Failed to write script file")

		// Execute the script file
		result, err := client.ExecuteFile(globalCtx, scriptPath)
		require.NoError(t, err, "Failed to execute script file")
		require.True(t, result.Success, "Script execution should be successful")

		// Verify script execution results
		t.Logf("Script execution result: %s", result.Message)
		require.Contains(t, result.Message, "Script execution test", "Script result should contain test string")
		require.Contains(t, result.Message, "Multiple statements", "Script result should contain multiple statements")
		require.Contains(t, result.Message, "Test 1", "Script result should contain inserted data")
		require.Contains(t, result.Message, "Test 2", "Script result should contain inserted data")

		// Clean up (drops when session ends anyway since it's a global temporary table)
		_, err = client.ExecuteSQL(globalCtx, "TRUNCATE TABLE temp_test_table")
		if err != nil {
			t.Logf("Warning: Failed to truncate temporary table: %v", err)
		}
	})
}

// Helper function to check if container tests are disabled
func isContainerTestDisabled() bool {
	return len(globalConnOpts.ConnectStr) == 0 || globalContainer == nil
}
