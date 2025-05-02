package container

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zodimo/go-sqlcl/pkg/sqlcl"
	"github.com/zodimo/go-sqlcl/pkg/types"
)

// TestUseCaseMigrations tests executing database migrations using go-sqlcl with the Oracle container
func TestUseCaseMigrations(t *testing.T) {
	// Skip test if running in CI or if explicitly disabled
	if os.Getenv("SKIP_CONTAINER_TESTS") != "" {
		t.Skip("Skipping container test as SKIP_CONTAINER_TESTS is set")
	}

	// Create a fixed client for testing
	client := CreateFixedClient(t)
	defer client.Close()

	// Test case 1: Create a simple table and verify it exists
	t.Run("SimpleTableCreation", func(t *testing.T) {
		// Create a test table using the system user
		createTableSQL := `
		CREATE TABLE test_simple (
			id NUMBER PRIMARY KEY,
			name VARCHAR2(100)
		);
		`

		result, err := client.ExecuteSQL(globalCtx, createTableSQL)
		require.NoError(t, err, "Failed to create test table")
		t.Logf("Create table result: %s", result.Message)

		// Verify the table exists
		verifySQL := `SELECT table_name FROM user_tables WHERE table_name = 'TEST_SIMPLE';`
		result, err = client.ExecuteSQL(globalCtx, verifySQL)
		require.NoError(t, err, "Failed to verify table creation")
		t.Logf("Verify table result: %s", result.Message)
		require.Contains(t, result.Message, "TEST_SIMPLE", "TEST_SIMPLE table should exist")

		// Clean up
		dropSQL := `DROP TABLE test_simple;`
		_, err = client.ExecuteSQL(globalCtx, dropSQL)
		require.NoError(t, err, "Failed to drop test table")
	})

	// Test case 2: Create a user and execute script under that user
	t.Run("UserMigrationScript", func(t *testing.T) {
		// Create a test user
		username := "TESTUSER1"
		createUserSQL := fmt.Sprintf(`
		-- Create user
		CREATE USER %s IDENTIFIED BY %[1]s;
		-- Grant permissions
		GRANT CONNECT, RESOURCE, CREATE VIEW, CREATE PROCEDURE, CREATE SYNONYM TO %[1]s;
		-- Grant tablespace
		ALTER USER %[1]s QUOTA UNLIMITED ON USERS;
		`, username)

		result, err := client.ExecuteSQL(globalCtx, createUserSQL)
		require.NoError(t, err, "Failed to create test user")
		t.Logf("Create user result: %s", result.Message)

		// Connect as the test user
		testUserConnectStr := fmt.Sprintf("%s/%s@%s", username, username, globalConnOpts.ConnectStr)

		// Create a new client connected as the test user
		clientConfig := &types.ClientConfig{
			SQLclPath:      DefaultContainerTestConfig().SQLclPath,
			Timeout:        30 * time.Second,
			ConnectTimeout: 15 * time.Second,
			QueryTimeout:   60 * time.Second,
			ColorOutput:    false,
		}

		testUserClient, err := CreateCustomClient(t, clientConfig, testUserConnectStr)
		require.NoError(t, err, "Failed to create client for test user")
		defer testUserClient.Close()

		// Create a test table
		createTableSQL := `
		CREATE TABLE user_test_table (
			id NUMBER PRIMARY KEY,
			name VARCHAR2(100),
			description VARCHAR2(200)
		);
		`

		result, err = testUserClient.ExecuteSQL(globalCtx, createTableSQL)
		require.NoError(t, err, "Failed to create test table")
		t.Logf("Create test table result: %s", result.Message)

		// Verify the table exists
		verifySQL := `SELECT table_name FROM user_tables WHERE table_name = 'USER_TEST_TABLE';`
		result, err = testUserClient.ExecuteSQL(globalCtx, verifySQL)
		require.NoError(t, err, "Failed to verify table creation")
		t.Logf("Verify table result: %s", result.Message)
		require.Contains(t, result.Message, "USER_TEST_TABLE", "USER_TEST_TABLE table should exist")

		// Clean up
		_, err = client.ExecuteSQL(globalCtx, fmt.Sprintf("DROP USER %s CASCADE;", username))
		require.NoError(t, err, "Failed to drop test user")
	})

	// Test case 3: Full migration script execution
	t.Run("FullMigrationScript", func(t *testing.T) {
		// Create a test user
		username := "TESTUSER2"
		createUserSQL := fmt.Sprintf(`
		-- Create user
		CREATE USER %s IDENTIFIED BY %[1]s;
		-- Grant permissions
		GRANT CONNECT, RESOURCE, CREATE VIEW, CREATE PROCEDURE, CREATE SYNONYM TO %[1]s;
		-- Grant tablespace
		ALTER USER %[1]s QUOTA UNLIMITED ON USERS;
		`, username)

		result, err := client.ExecuteSQL(globalCtx, createUserSQL)
		require.NoError(t, err, "Failed to create test user")
		t.Logf("Create user result: %s", result.Message)

		// Connect as the test user
		testUserConnectStr := fmt.Sprintf("%s/%s@%s", username, username, globalConnOpts.ConnectStr)

		// Create a new client connected as the test user
		clientConfig := &types.ClientConfig{
			SQLclPath:      DefaultContainerTestConfig().SQLclPath,
			Timeout:        30 * time.Second,
			ConnectTimeout: 15 * time.Second,
			QueryTimeout:   60 * time.Second,
			ColorOutput:    false,
		}

		testUserClient, err := CreateCustomClient(t, clientConfig, testUserConnectStr)
		require.NoError(t, err, "Failed to create client for test user")
		defer testUserClient.Close()

		// Get migration scripts from testdata directory
		createTablesScriptPath := filepath.Join("testdata", "migrations", "create_tables.sql")
		createTablesScript, err := os.ReadFile(createTablesScriptPath)
		if err != nil {
			t.Logf("Warning: Could not read migration script file: %v", err)
			t.Skip("Skipping test as migration script file is not available")
			return
		}

		insertDataScriptPath := filepath.Join("testdata", "migrations", "insert_data.sql")
		insertDataScript, err := os.ReadFile(insertDataScriptPath)
		if err != nil {
			t.Logf("Warning: Could not read data insertion script file: %v", err)
			t.Skip("Skipping test as data insertion script file is not available")
			return
		}

		// Execute the table creation script
		result, err = testUserClient.ExecuteSQL(globalCtx, string(createTablesScript))
		require.NoError(t, err, "Failed to execute create tables script")
		t.Logf("Create tables script result: %s", result.Message)

		// Directly check if each table exists individually
		for _, tableName := range []string{"DEPARTMENTS", "EMPLOYEES", "JOB_HISTORY"} {
			verifyTableSQL := fmt.Sprintf(`SELECT table_name FROM user_tables WHERE table_name = '%s';`, tableName)
			result, err = testUserClient.ExecuteSQL(globalCtx, verifyTableSQL)
			require.NoError(t, err, fmt.Sprintf("Failed to verify if %s table exists", tableName))
			t.Logf("%s table check result: %s", tableName, result.Message)
			require.Contains(t, result.Message, tableName, fmt.Sprintf("%s table should exist", tableName))
		}

		// Execute the data insertion script
		result, err = testUserClient.ExecuteSQL(globalCtx, string(insertDataScript))
		require.NoError(t, err, "Failed to execute insert data script")
		t.Logf("Insert data script result: %s", result.Message)

		// Verify data was inserted
		verifyEmployeesSQL := `SELECT COUNT(*) FROM employees;`
		result, err = testUserClient.ExecuteSQL(globalCtx, verifyEmployeesSQL)
		require.NoError(t, err, "Failed to verify employees count")
		t.Logf("Verify employees count result: %s", result.Message)

		// Note: The insert_data.sql adds 10 employees, so we check for "10"
		require.Contains(t, result.Message, "10", "Should have inserted 10 employees")

		// Test a complex query joining multiple tables
		complexQuerySQL := `
		SELECT e.first_name, e.last_name, d.department_name
		FROM employees e
		JOIN departments d ON e.department_id = d.department_id
		WHERE e.salary > 10000
		ORDER BY e.salary DESC;
		`

		result, err = testUserClient.ExecuteSQL(globalCtx, complexQuerySQL)
		require.NoError(t, err, "Failed to execute complex join query")
		t.Logf("Complex query result: %s", result.Message)

		// The highest paid employees should be King, Kochhar, and De Haan in the Administration department
		require.Contains(t, result.Message, "King", "Result should contain employee King")
		require.Contains(t, result.Message, "Administration", "Result should contain Administration department")

		// Clean up
		_, err = client.ExecuteSQL(globalCtx, fmt.Sprintf("DROP USER %s CASCADE;", username))
		require.NoError(t, err, "Failed to drop test user")
	})
}

// CreateCustomClient creates a client with custom configuration
func CreateCustomClient(t *testing.T, config *types.ClientConfig, connectString string) (types.Client, error) {
	t.Helper()

	// Create a fixed client
	fixedClient, err := sqlcl.NewFixedClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create fixed client: %w", err)
	}

	// Connect with the specified connect string
	err = fixedClient.Connect(globalCtx, connectString)
	if err != nil {
		fixedClient.Close()
		return nil, fmt.Errorf("failed to connect with %s: %w", connectString, err)
	}

	return fixedClient, nil
}
