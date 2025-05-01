// Package integration provides integration tests for the go-sqlcl package.
// These tests verify that the package works correctly with an actual SQLcl installation.
package integration

import (
	"os"
	"time"

	"github.com/zodimo/go-sqlcl/pkg/types"
)

// TestConfig stores the configuration for integration tests.
type TestConfig struct {
	// SQLclPath is the path to the SQLcl executable
	SQLclPath string

	// Connection details
	Username    string
	Password    string
	ConnectStr  string
	TNSAdmin    string
	Wallet      string
	Role        string
	UseDefaults bool

	// Timeouts
	Timeout        time.Duration
	ConnectTimeout time.Duration
	QueryTimeout   time.Duration
}

// LoadConfig loads the test configuration from environment variables.
// It falls back to default values if environment variables are not set.
func LoadConfig() *TestConfig {
	config := &TestConfig{
		SQLclPath:      getEnvWithDefault("TEST_SQLCL_PATH", "sql"),
		Username:       getEnvWithDefault("TEST_DB_USER", ""),
		Password:       getEnvWithDefault("TEST_DB_PASSWORD", ""),
		ConnectStr:     getEnvWithDefault("TEST_DB_CONNECT", ""),
		TNSAdmin:       getEnvWithDefault("TEST_TNS_ADMIN", ""),
		Wallet:         getEnvWithDefault("TEST_WALLET", ""),
		Role:           getEnvWithDefault("TEST_DB_ROLE", ""),
		Timeout:        30 * time.Second,
		ConnectTimeout: 15 * time.Second,
		QueryTimeout:   60 * time.Second,
		UseDefaults:    os.Getenv("TEST_USE_DEFAULTS") == "true",
	}

	return config
}

// GetClientConfig returns a ClientConfig for the go-sqlcl package.
func (tc *TestConfig) GetClientConfig() *types.ClientConfig {
	return &types.ClientConfig{
		SQLclPath:      tc.SQLclPath,
		Timeout:        tc.Timeout,
		ConnectTimeout: tc.ConnectTimeout,
		QueryTimeout:   tc.QueryTimeout,
		ColorOutput:    false,
		Format:         "table",
		LogLevel:       "info",
	}
}

// GetConnectionOptions returns ConnectionOptions for the go-sqlcl package.
func (tc *TestConfig) GetConnectionOptions() types.ConnectionOptions {
	return types.ConnectionOptions{
		Username:   tc.Username,
		Password:   tc.Password,
		ConnectStr: tc.ConnectStr,
		TNSAdmin:   tc.TNSAdmin,
		Wallet:     tc.Wallet,
		Role:       tc.Role,
	}
}

// IsConnectionConfigured returns true if the connection details are configured.
func (tc *TestConfig) IsConnectionConfigured() bool {
	return tc.Username != "" && tc.Password != "" && tc.ConnectStr != ""
}

// getEnvWithDefault returns the value of the environment variable or the default value.
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
