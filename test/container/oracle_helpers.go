package container

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zodimo/go-sqlcl/pkg/sqlcl"
	"github.com/zodimo/go-sqlcl/pkg/types"
)

// CreateTestClient creates a new client connected to the test container
func CreateTestClient(t *testing.T) types.Client {
	t.Helper()

	// Create a client config
	config := &types.ClientConfig{
		SQLclPath:      DefaultContainerTestConfig().SQLclPath,
		Timeout:        30 * time.Second,
		ConnectTimeout: 15 * time.Second,
		QueryTimeout:   60 * time.Second,
		ColorOutput:    false,
	}

	// Use the fixed client implementation for more reliable tests
	client, err := sqlcl.NewFixedClient(config)
	require.NoError(t, err, "Failed to create SQLcl client")

	// Connect to the database
	// Build connection string using the format we know works from our direct tests
	host, port, err := parseConnectStr(globalConnOpts.ConnectStr)
	require.NoError(t, err, "Failed to parse connect string")

	// For Oracle 18c, use a connect string with service name explicitly specified
	connString := fmt.Sprintf("%s/%s@%s:%s/%s",
		globalConnOpts.Username,
		globalConnOpts.Password,
		host, port, "XEPDB1")

	t.Logf("Connecting with: %s", connString)

	// Try to connect with retries
	var connectErr error
	maxRetries := 3
	retryDelay := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		connectErr = client.Connect(globalCtx, connString)
		if connectErr == nil {
			break
		}
		t.Logf("Connection attempt %d failed: %v. Retrying in %v...", i+1, connectErr, retryDelay)
		time.Sleep(retryDelay)
	}
	require.NoError(t, connectErr, "Failed to connect to the database after %d attempts", maxRetries)

	return client
}

// ExecuteScript executes a SQL script against the Oracle container
func ExecuteScript(t *testing.T, client *sqlcl.Client, scriptContent string) error {
	t.Helper()

	// Execute the script
	_, err := client.ExecuteSQL(globalCtx, scriptContent)
	return err
}

// ExecuteScriptFile executes a SQL script file against the Oracle container
func ExecuteScriptFile(t *testing.T, client *sqlcl.Client, scriptPath string) error {
	t.Helper()

	// Read the script file
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read script file: %w", err)
	}

	// Execute the script
	return ExecuteScript(t, client, string(content))
}

// SetupTestSchema sets up a test schema in the Oracle container
func SetupTestSchema(t *testing.T, client *sqlcl.Client, schemaName string) error {
	t.Helper()

	// Create the schema
	createSchemaSQL := fmt.Sprintf(`
		-- Create user
		CREATE USER %s IDENTIFIED BY %[1]s;
		-- Grant permissions
		GRANT CONNECT, RESOURCE, CREATE VIEW, CREATE PROCEDURE, CREATE SYNONYM TO %[1]s;
		-- Grant tablespace
		ALTER USER %[1]s QUOTA UNLIMITED ON USERS;
	`, strings.ToUpper(schemaName))

	return ExecuteScript(t, client, createSchemaSQL)
}

// DropTestSchema drops a test schema from the Oracle container
func DropTestSchema(t *testing.T, client *sqlcl.Client, schemaName string) error {
	t.Helper()

	// Drop the schema
	dropSchemaSQL := fmt.Sprintf(`
		DROP USER %s CASCADE;
	`, strings.ToUpper(schemaName))

	return ExecuteScript(t, client, dropSchemaSQL)
}

// ExecuteSQLDirect executes SQL directly using SQL*Plus instead of SQLcl
func ExecuteSQLDirect(t *testing.T, sql string) error {
	t.Helper()

	// Get connection details
	host, port, err := parseConnectStr(globalConnOpts.ConnectStr)
	if err != nil {
		return fmt.Errorf("failed to parse connect string: %w", err)
	}

	// Create SQL*Plus command
	sqlPlusCmd := fmt.Sprintf("echo '%s' | sqlplus -s %s/%s@%s:%s/%s",
		sql,
		globalConnOpts.Username,
		globalConnOpts.Password,
		host, port, "XEPDB1")

	// Run the command
	cmd := exec.Command("bash", "-c", sqlPlusCmd)
	output, err := cmd.CombinedOutput()
	t.Logf("SQL*Plus output: %s", string(output))

	if err != nil {
		return fmt.Errorf("error executing SQL: %w, output: %s", err, string(output))
	}

	return nil
}
