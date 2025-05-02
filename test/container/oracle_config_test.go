package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestContainerConfigDefaults verifies that the default container configuration is correctly set
func TestContainerConfigDefaults(t *testing.T) {
	cfg := DefaultContainerTestConfig()

	// We're using the actual value from the environment, don't force a specific expectation
	if cfg.SQLclPath == "" {
		t.Error("SQLclPath should not be empty")
	}

	// The default image tag is now "18-slim" not "latest"
	assert.Equal(t, "18-slim", cfg.ContainerConfig.Tag, "Default image tag should be 18-slim")
	assert.Equal(t, "system", cfg.ContainerConfig.User, "Default username should be system")
	assert.Equal(t, "oracle", cfg.ContainerConfig.Password, "Default password should be oracle")
	assert.Equal(t, "gvenzl/oracle-xe", cfg.ContainerConfig.Image, "Default image should be gvenzl/oracle-xe")
}
