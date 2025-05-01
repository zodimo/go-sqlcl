package container

import (
	"context"
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
		SQLclPath:       "sql",
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
