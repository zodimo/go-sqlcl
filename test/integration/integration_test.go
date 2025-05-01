// Package integration provides integration tests for the go-sqlcl package.
package integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zodimo/go-sqlcl/pkg/sqlcl"
	"github.com/zodimo/go-sqlcl/pkg/types"
)

var (
	testConfig *TestConfig
	client     *sqlcl.Client
)

// TestMain is the main entry point for integration tests.
func TestMain(m *testing.M) {
	// Load test configuration
	testConfig = LoadConfig()

	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting integration tests")

	// Check if we have a SQLcl installation
	if !validateSQLcl() {
		log.Printf("WARNING: SQLcl not found at '%s'. Integration tests will be skipped.", testConfig.SQLclPath)
		log.Printf("Set TEST_SQLCL_PATH environment variable to specify the SQLcl path.")
		os.Exit(0)
	}

	// Run tests
	exitCode := m.Run()

	// Clean up
	if client != nil {
		client.Close()
	}

	os.Exit(exitCode)
}

// validateSQLcl checks if the SQLcl executable is available.
func validateSQLcl() bool {
	_, err := os.Stat(testConfig.SQLclPath)
	if err == nil {
		return true
	}

	// Try with $PATH resolution if it's not an absolute path
	if !strings.Contains(testConfig.SQLclPath, "/") {
		cmd := exec.Command("which", testConfig.SQLclPath)
		output, err := cmd.CombinedOutput()
		return err == nil && len(output) > 0
	}

	return false
}

// setupClient creates a new SQLcl client for testing.
func setupClient(t *testing.T) *sqlcl.Client {
	clientConfig := testConfig.GetClientConfig()
	client, err := sqlcl.NewClient(clientConfig)
	require.NoError(t, err, "Failed to create client")
	return client
}

// setupConnection establishes a database connection if credentials are available.
func setupConnection(t *testing.T, client *sqlcl.Client) bool {
	if !testConfig.IsConnectionConfigured() && !testConfig.UseDefaults {
		t.Skip("Database connection not configured. Skipping test.")
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), testConfig.ConnectTimeout)
	defer cancel()

	connectionOpts := testConfig.GetConnectionOptions()
	err := client.ConnectWithOptions(ctx, connectionOpts)

	if err != nil {
		t.Skipf("Failed to connect to database: %v", err)
		return false
	}

	return true
}

// TestClientCreation tests that a client can be created.
func TestClientCreation(t *testing.T) {
	client := setupClient(t)
	defer client.Close()

	assert.NotNil(t, client, "Client should not be nil")
}

// TestDatabaseConnection tests connecting to a database.
func TestDatabaseConnection(t *testing.T) {
	client := setupClient(t)
	defer client.Close()

	connected := setupConnection(t, client)
	assert.True(t, connected, "Should be able to connect to the database")
}

// TestBasicQuery tests executing a basic SQL query.
func TestBasicQuery(t *testing.T) {
	client := setupClient(t)
	defer client.Close()

	if !setupConnection(t, client) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), testConfig.QueryTimeout)
	defer cancel()

	// Execute a simple SQL query
	sqlQuery := "SELECT user, sysdate FROM dual"
	result, err := client.ExecuteSQL(ctx, sqlQuery)
	require.NoError(t, err, "Failed to execute query")
	require.NotNil(t, result, "Result should not be nil")

	// Verify result structure
	assert.GreaterOrEqual(t, len(result.Columns), 2, "Result should have at least 2 columns")
	assert.GreaterOrEqual(t, len(result.Rows), 1, "Result should have at least 1 row")

	// Verify column names
	assert.Contains(t, strings.ToUpper(result.Columns[0].Name), "USER", "First column should be USER")
	assert.Contains(t, strings.ToUpper(result.Columns[1].Name), "SYSDATE", "Second column should be SYSDATE")

	// Check row values
	assert.NotEmpty(t, result.Rows[0].Values[0], "Username should not be empty")
	assert.NotEmpty(t, result.Rows[0].Values[1], "Date should not be empty")
}

// TestSQLCLSpecificCommands tests executing SQLcl-specific commands.
func TestSQLCLSpecificCommands(t *testing.T) {
	client := setupClient(t)
	defer client.Close()

	if !setupConnection(t, client) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), testConfig.Timeout)
	defer cancel()

	// Test showing user
	cmd := types.Command{
		SQL:           "SHOW USER",
		ReturnResults: true,
	}

	result, err := client.ExecuteCommand(ctx, cmd)
	require.NoError(t, err, "Failed to execute SHOW USER command")
	require.NotNil(t, result, "Result should not be nil")
	assert.Contains(t, result.Summary, "USER", "Result should contain USER")

	// Test a different SQLcl-specific command
	cmd = types.Command{
		SQL:           "SHOW PARAMETER NLS_DATE_FORMAT",
		ReturnResults: true,
	}

	result, err = client.ExecuteCommand(ctx, cmd)
	require.NoError(t, err, "Failed to execute SHOW PARAMETER command")
	require.NotNil(t, result, "Result should not be nil")
	assert.Contains(t, result.Summary, "NLS_DATE_FORMAT", "Result should contain NLS_DATE_FORMAT")
}

// TestTransaction tests transaction support.
func TestTransaction(t *testing.T) {
	client := setupClient(t)
	defer client.Close()

	if !setupConnection(t, client) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), testConfig.Timeout)
	defer cancel()

	// Start a transaction
	_, err := client.ExecuteSQL(ctx, "BEGIN")
	require.NoError(t, err, "Failed to start transaction")

	// Create a temporary table
	tempTableName := fmt.Sprintf("TEMP_TEST_%d", time.Now().UnixNano())
	createTableSQL := fmt.Sprintf("CREATE GLOBAL TEMPORARY TABLE %s (id NUMBER, name VARCHAR2(100)) ON COMMIT DROP", tempTableName)

	_, err = client.ExecuteSQL(ctx, createTableSQL)
	require.NoError(t, err, "Failed to create temporary table")

	// Insert some data
	insertSQL := fmt.Sprintf("INSERT INTO %s VALUES (1, 'Test Data')", tempTableName)
	_, err = client.ExecuteSQL(ctx, insertSQL)
	require.NoError(t, err, "Failed to insert data")

	// Verify the data is there
	selectSQL := fmt.Sprintf("SELECT * FROM %s", tempTableName)
	result, err := client.ExecuteSQL(ctx, selectSQL)
	require.NoError(t, err, "Failed to select data")
	require.NotNil(t, result, "Result should not be nil")
	assert.Equal(t, 1, len(result.Rows), "Should have 1 row")

	// Rollback the transaction
	_, err = client.ExecuteSQL(ctx, "ROLLBACK")
	require.NoError(t, err, "Failed to rollback transaction")

	// Try to select from the table again (should fail because the table was dropped)
	_, err = client.ExecuteSQL(ctx, selectSQL)
	assert.Error(t, err, "Should fail because table was dropped after rollback")
}

// TestErrorHandling tests error handling.
func TestErrorHandling(t *testing.T) {
	client := setupClient(t)
	defer client.Close()

	if !setupConnection(t, client) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), testConfig.Timeout)
	defer cancel()

	// Execute an invalid SQL query
	_, err := client.ExecuteSQL(ctx, "SELECT * FROM nonexistent_table")
	assert.Error(t, err, "Should return an error for invalid SQL")

	// Check if it's an Oracle error
	oraErr, ok := err.(*types.Error)
	if assert.True(t, ok, "Error should be of type *types.Error") {
		// ORA-00942: table or view does not exist
		assert.Contains(t, oraErr.Message, "table or view does not exist", "Error message should indicate non-existent table")
	}
}

// TestCommandWrappers tests the SQLcl-specific command wrappers.
func TestCommandWrappers(t *testing.T) {
	client := setupClient(t)
	defer client.Close()

	if !setupConnection(t, client) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), testConfig.Timeout)
	defer cancel()

	// Test the SET wrapper
	err := client.SetCommand(ctx, "PAGESIZE", "50")
	require.NoError(t, err, "Failed to set PAGESIZE")

	// Verify the setting was applied
	result, err := client.ShowCommand(ctx, "PAGESIZE")
	require.NoError(t, err, "Failed to show PAGESIZE parameter")
	assert.Contains(t, result, "50", "PAGESIZE should be set to 50")

	// Test the DESCRIBE wrapper
	describeResult, err := client.DescribeObject(ctx, "DUAL")
	if err == nil {
		assert.Contains(t, describeResult.ObjectName, "DUAL", "Should describe the DUAL table")
		foundDummyColumn := false
		for _, col := range describeResult.Columns {
			if strings.Contains(strings.ToUpper(col.Name), "DUMMY") {
				foundDummyColumn = true
				break
			}
		}
		assert.True(t, foundDummyColumn, "DUAL table should have a DUMMY column")
	} else {
		// Some databases might restrict access to the DUAL table
		t.Log("Couldn't describe DUAL table, skipping assertion")
	}
}

// TestMultipleQueries tests executing multiple queries in sequence.
func TestMultipleQueries(t *testing.T) {
	client := setupClient(t)
	defer client.Close()

	if !setupConnection(t, client) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), testConfig.Timeout)
	defer cancel()

	// Execute a series of queries
	queries := []string{
		"SELECT 1 FROM dual",
		"SELECT user FROM dual",
		"SELECT sysdate FROM dual",
		"SELECT systimestamp FROM dual",
	}

	for i, query := range queries {
		result, err := client.ExecuteSQL(ctx, query)
		require.NoError(t, err, "Failed to execute query %d: %s", i, query)
		require.NotNil(t, result, "Result should not be nil for query %d", i)
		assert.GreaterOrEqual(t, len(result.Rows), 1, "Should have at least one row for query %d", i)
	}
}

// TestConcurrentQueries tests executing multiple queries concurrently.
func TestConcurrentQueries(t *testing.T) {
	client := setupClient(t)
	defer client.Close()

	if !setupConnection(t, client) {
		return
	}

	// Note: SQLcl doesn't support multiple concurrent queries in a single session
	// This test creates multiple clients to test concurrent execution

	queries := []string{
		"SELECT 1 FROM dual",
		"SELECT user FROM dual",
		"SELECT sysdate FROM dual",
		"SELECT systimestamp FROM dual",
	}

	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup
	wg.Add(len(queries))

	// Execute each query in a separate goroutine
	for i, query := range queries {
		go func(index int, sql string) {
			defer wg.Done()

			// Create a new client for this query
			queryClient := setupClient(t)
			defer queryClient.Close()

			// Only proceed if we can connect
			if !setupConnection(t, queryClient) {
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), testConfig.Timeout)
			defer cancel()

			// Execute the query
			result, err := queryClient.ExecuteSQL(ctx, sql)
			assert.NoError(t, err, "Failed to execute concurrent query %d", index)
			assert.NotNil(t, result, "Result should not be nil for concurrent query %d", index)
		}(i, query)
	}

	// Wait for all queries to complete
	wg.Wait()
}
