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
	assert.Equal(t, "gvenzl/oracle-xe", config.Image, "Default image should be gvenzl/oracle-xe")
	assert.Equal(t, "18-slim", config.Tag, "Default image tag should be 18-slim")
	assert.Equal(t, "system", config.User, "Default username should be system")
	assert.Equal(t, "oracle", config.Password, "Default password should be oracle")
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
