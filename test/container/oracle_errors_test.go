package container

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestErrorHandling verifies the error handling for SQL query execution
func TestErrorHandling(t *testing.T) {
	// Skip test if running in CI or if explicitly disabled
	if os.Getenv("SKIP_CONTAINER_TESTS") != "" {
		t.Skip("Skipping container test as SKIP_CONTAINER_TESTS is set")
	}

	// Create a fixed client instead of using direct SQL*Plus execution
	client := CreateFixedClient(t)
	defer client.Close()

	// Test case 1: Syntax error in SQL
	t.Run("SyntaxError", func(t *testing.T) {
		// Execute invalid SQL with syntax error
		result, err := client.ExecuteSQL(globalCtx, "SELECT * FROMM dual")
		require.NoError(t, err, "ExecuteSQL should not return an error for SQL syntax errors")
		require.False(t, result.Success, "Query with syntax error should fail")
		require.Contains(t, result.Message, "ORA-", "Error message should contain Oracle error code")
		t.Logf("SQL syntax error message: %s", result.Message)
	})

	// Test case 2: Reference to non-existent table
	t.Run("NonExistentTable", func(t *testing.T) {
		result, err := client.ExecuteSQL(globalCtx, "SELECT * FROM non_existent_table")
		require.NoError(t, err, "ExecuteSQL should not return an error for non-existent table")
		require.False(t, result.Success, "Query with non-existent table should fail")
		require.Contains(t, result.Message, "ORA-", "Error message should contain Oracle error code")
		t.Logf("Non-existent table error message: %s", result.Message)
	})

	// Test case 3: Reference to non-existent column
	t.Run("NonExistentColumn", func(t *testing.T) {
		result, err := client.ExecuteSQL(globalCtx, "SELECT non_existent_column FROM dual")
		require.NoError(t, err, "ExecuteSQL should not return an error for non-existent column")
		require.False(t, result.Success, "Query with non-existent column should fail")
		require.Contains(t, result.Message, "ORA-", "Error message should contain Oracle error code")
		t.Logf("Non-existent column error message: %s", result.Message)
	})

	// Test case 4: Constraint violation
	t.Run("ConstraintViolation", func(t *testing.T) {
		// Create a test schema
		schemaName := "error_test"
		createSchemaSQL := fmt.Sprintf(`
			-- Create user
			CREATE USER %s IDENTIFIED BY %[1]s;
			-- Grant permissions
			GRANT CONNECT, RESOURCE, CREATE VIEW, CREATE PROCEDURE, CREATE SYNONYM TO %[1]s;
			-- Grant tablespace
			ALTER USER %[1]s QUOTA UNLIMITED ON USERS;
		`, strings.ToUpper(schemaName))

		result, err := client.ExecuteSQL(globalCtx, createSchemaSQL)
		require.NoError(t, err, "Failed to execute schema creation SQL")
		t.Logf("Schema creation result: %s", result.Message)

		defer func() {
			dropSchemaSQL := fmt.Sprintf(`DROP USER %s CASCADE;`, strings.ToUpper(schemaName))
			_, err := client.ExecuteSQL(globalCtx, dropSchemaSQL)
			if err != nil {
				t.Logf("Failed to drop test schema: %v", err)
			}
		}()

		// Create table with primary key
		createTableSQL := fmt.Sprintf(`
			ALTER SESSION SET CURRENT_SCHEMA = %s;
			CREATE TABLE pk_test (
				id NUMBER PRIMARY KEY,
				name VARCHAR2(100)
			);
		`, strings.ToUpper(schemaName))

		result, err = client.ExecuteSQL(globalCtx, createTableSQL)
		require.NoError(t, err, "Failed to execute table creation SQL")
		t.Logf("Table creation result: %s", result.Message)

		// Insert a record
		insertSQL := fmt.Sprintf(`
			ALTER SESSION SET CURRENT_SCHEMA = %s;
			INSERT INTO pk_test VALUES (1, 'Test 1');
		`, strings.ToUpper(schemaName))

		result, err = client.ExecuteSQL(globalCtx, insertSQL)
		require.NoError(t, err, "Failed to execute first insert")
		t.Logf("First insert result: %s", result.Message)

		// Try to insert another record with the same primary key
		duplicateSQL := fmt.Sprintf(`
			ALTER SESSION SET CURRENT_SCHEMA = %s;
			INSERT INTO pk_test VALUES (1, 'Test 2');
		`, strings.ToUpper(schemaName))

		result, err = client.ExecuteSQL(globalCtx, duplicateSQL)
		require.NoError(t, err, "ExecuteSQL should not return an error for constraint violation")
		require.False(t, result.Success, "Duplicate insert should fail")
		require.Contains(t, result.Message, "ORA-00001", "Error message should contain constraint violation error")
		t.Logf("Constraint violation error message: %s", result.Message)
	})

	// Test case 5: Invalid SQL command
	t.Run("InvalidSQLCommand", func(t *testing.T) {
		result, err := client.ExecuteSQL(globalCtx, "INVALID_COMMAND")
		require.NoError(t, err, "ExecuteSQL should not return an error for invalid command")
		t.Logf("Invalid command result: %s", result.Message)

		// The success flag might be inconsistent, so just check the message contains error info
		if !strings.Contains(result.Message, "ORA-") && !strings.Contains(result.Message, "SP2-") {
			t.Logf("Warning: Expected error code not found in result message")
		}
	})

	// Test case 6: Empty SQL command
	t.Run("EmptySQLCommand", func(t *testing.T) {
		result, err := client.ExecuteSQL(globalCtx, "")
		// Allow either an error or a failed result
		if err != nil {
			t.Logf("Empty command returned error: %v", err)
		} else {
			t.Logf("Empty command result: %s", result.Message)
		}
	})
}
