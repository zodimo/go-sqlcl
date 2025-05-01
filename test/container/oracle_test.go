package container

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/zodimo/go-sqlcl/pkg/sqlcl"
	"github.com/zodimo/go-sqlcl/pkg/types"
)

var (
	// Global variables to store container and client instances
	globalContainer testcontainers.Container
	globalClient    *sqlcl.Client
	globalConnOpts  types.ConnectionOptions
	globalCtx       context.Context
	globalCancel    context.CancelFunc
)

// TestMain sets up and tears down the test environment for all tests in this package
func TestMain(m *testing.M) {
	// Skip container tests if running in CI or if explicitly disabled
	if os.Getenv("SKIP_CONTAINER_TESTS") != "" {
		log.Println("Skipping container tests")
		os.Exit(0)
	}

	// Create a global context with timeout
	globalCtx, globalCancel = context.WithTimeout(context.Background(), 15*time.Minute)
	defer globalCancel()

	// Setup the container environment
	var err error
	exitCode := 1

	// Setup the container
	log.Println("Starting Oracle container...")
	globalContainer, globalConnOpts, err = StartOracleContainer(globalCtx, nil)
	if err != nil {
		log.Printf("Failed to start Oracle container: %v\n", err)
		os.Exit(exitCode)
	}

	// Create a SQLcl client
	globalClient, err = createSQLcLClient(DefaultContainerTestConfig())
	if err != nil {
		log.Printf("Failed to create SQLcl client: %v\n", err)
		terminateContainer()
		os.Exit(exitCode)
	}

	// Run the tests
	log.Println("Running tests...")
	exitCode = m.Run()

	// Clean up
	log.Println("Cleaning up test environment...")
	if globalClient != nil {
		if err := globalClient.Close(); err != nil {
			log.Printf("Failed to close client: %v\n", err)
		}
	}

	terminateContainer()
	os.Exit(exitCode)
}

// terminateContainer terminates the global container
func terminateContainer() {
	if globalContainer != nil {
		if err := globalContainer.Terminate(globalCtx); err != nil {
			log.Printf("Failed to terminate container: %v\n", err)
		}
	}
}

// CreateTestClient creates a new client connected to the test container
func CreateTestClient(t *testing.T) *sqlcl.Client {
	t.Helper()

	// Create a copy of the client
	client, err := createSQLcLClient(DefaultContainerTestConfig())
	require.NoError(t, err, "Failed to create SQLcl client")

	// Connect to the database
	err = client.Connect(globalCtx, fmt.Sprintf("%s/%s@%s",
		globalConnOpts.Username,
		globalConnOpts.Password,
		globalConnOpts.ConnectStr))
	require.NoError(t, err, "Failed to connect to the database")

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

// GetContainerConnectionOptions returns the connection options for the container
func GetContainerConnectionOptions() types.ConnectionOptions {
	return globalConnOpts
}

// GetGlobalContext returns the global context
func GetGlobalContext() context.Context {
	return globalCtx
}
