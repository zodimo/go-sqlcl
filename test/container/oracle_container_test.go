package container

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultOracleContainerConfig verifies that the default configuration is correctly set
func TestDefaultOracleContainerConfig(t *testing.T) {
	config := DefaultOracleContainerConfig()

	assert.Equal(t, "gvenzl/oracle-xe", config.Image)
	assert.Equal(t, "latest", config.Tag)
	assert.Equal(t, "1521", config.Port)
	assert.Equal(t, "system", config.User)
	assert.Equal(t, "oracle", config.Password)
	assert.Equal(t, "XEPDB1", config.Database)
	assert.Equal(t, 120*time.Second, config.StartupTimeout)
}

// TestOracleContainer tests the Oracle container setup
// This test is marked as skipped by default because it requires Docker
// and takes a long time to run
func TestOracleContainer(t *testing.T) {
	t.Skip("Skipping container test (requires Docker and takes time)")

	// Prepare test context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create a container with default configuration
	container, connOptions, err := StartOracleContainer(ctx, nil)

	// Clean up after the test
	if container != nil {
		defer container.Terminate(ctx)
	}

	// Check the results
	require.NoError(t, err)
	require.NotNil(t, container)

	// Verify connection options
	assert.Equal(t, "system", connOptions.Username)
	assert.Equal(t, "oracle", connOptions.Password)
	assert.Contains(t, connOptions.ConnectStr, "/XEPDB1")
}
