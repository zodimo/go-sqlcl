package container

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestContainerConfigDefaults verifies that the default container configuration is correctly set
func TestContainerConfigDefaults(t *testing.T) {
	config := DefaultContainerTestConfig()

	// Verify config values
	assert.Equal(t, "sql", config.SQLclPath)
	assert.Equal(t, 30*time.Second, config.Timeout)

	// Verify container config
	containerConfig := config.ContainerConfig
	assert.Equal(t, "gvenzl/oracle-xe", containerConfig.Image)
	assert.Equal(t, "latest", containerConfig.Tag)
	assert.Equal(t, "1521", containerConfig.Port)
	assert.Equal(t, "system", containerConfig.User)
	assert.Equal(t, "oracle", containerConfig.Password)
	assert.Equal(t, "XEPDB1", containerConfig.Database)
}
