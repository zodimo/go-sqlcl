package container

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/zodimo/go-sqlcl/pkg/sqlcl"
	"github.com/zodimo/go-sqlcl/pkg/types"
)

// ContainerTestConfig holds configuration for container-based tests
type ContainerTestConfig struct {
	// SQLcl configuration
	SQLclPath string
	Timeout   time.Duration

	// Container configuration
	ContainerConfig *OracleContainerConfig

	// Test context
	Context context.Context
	Cancel  context.CancelFunc
}

// DefaultContainerTestConfig returns a default configuration for container-based tests
func DefaultContainerTestConfig() *ContainerTestConfig {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)

	return &ContainerTestConfig{
		SQLclPath:       "/home/jaco/Sources/sqlcl-25.1.1.113.2054/bin/sql",
		Timeout:         30 * time.Second,
		ContainerConfig: DefaultOracleContainerConfig(),
		Context:         ctx,
		Cancel:          cancel,
	}
}

// SetupContainerTest initializes a test environment with an Oracle container
// and returns the container, a configured client, and connection options
func SetupContainerTest(t *testing.T, config *ContainerTestConfig) (testcontainers.Container, *sqlcl.Client, types.ConnectionOptions) {
	if config == nil {
		config = DefaultContainerTestConfig()
	}

	// Start the Oracle container
	container, connOptions, err := StartOracleContainer(config.Context, config.ContainerConfig)
	require.NoError(t, err, "Failed to start Oracle container")
	require.NotNil(t, container, "Container should not be nil")

	// Create a SQLcl client
	client, err := createSQLcLClient(config)
	require.NoError(t, err, "Failed to create SQLcl client")

	return container, client, connOptions
}

// createSQLcLClient creates a SQLcl client with the given configuration
func createSQLcLClient(config *ContainerTestConfig) (*sqlcl.Client, error) {
	clientConfig := &types.ClientConfig{
		SQLclPath:      config.SQLclPath,
		Timeout:        config.Timeout,
		ConnectTimeout: 15 * time.Second,
		QueryTimeout:   60 * time.Second,
		ColorOutput:    false,
		Format:         "table",
		LogLevel:       "info",
	}

	return sqlcl.NewClient(clientConfig)
}

// TeardownContainerTest cleans up resources after a container test
func TeardownContainerTest(t *testing.T, ctx context.Context, container testcontainers.Container, client *sqlcl.Client) {
	// Close the client if it exists
	if client != nil {
		err := client.Close()
		require.NoError(t, err, "Failed to close client")
	}

	// Terminate the container if it exists
	if container != nil {
		err := container.Terminate(ctx)
		require.NoError(t, err, "Failed to terminate container")
	}
}

// CreateFixedClient creates a new client using the fixed implementation
func CreateFixedClient(t *testing.T) types.Client {
	t.Helper()

	// Get container test config
	config := DefaultContainerTestConfig()

	// Create client config
	clientConfig := &types.ClientConfig{
		SQLclPath:      config.SQLclPath,
		Timeout:        config.Timeout,
		ConnectTimeout: 15 * time.Second,
		QueryTimeout:   60 * time.Second,
		ColorOutput:    false,
		Format:         "table",
		LogLevel:       "info",
	}

	// Create a fixed client instead of the standard client
	fixedClient, err := sqlcl.NewFixedClient(clientConfig)
	require.NoError(t, err, "Failed to create SQLcl fixed client")

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
		connectErr = fixedClient.Connect(globalCtx, connString)
		if connectErr == nil {
			break
		}
		t.Logf("Connection attempt %d failed: %v. Retrying in %v...", i+1, connectErr, retryDelay)
		time.Sleep(retryDelay)
	}
	require.NoError(t, connectErr, "Failed to connect to the database after %d attempts", maxRetries)

	return fixedClient
}
