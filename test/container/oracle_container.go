// Package container provides container-based testing infrastructure for Oracle XE
package container

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/zodimo/go-sqlcl/pkg/types"
)

const (
	// Default Oracle XE container image
	defaultOracleImage = "gvenzl/oracle-xe:latest"

	// Default Oracle ports
	defaultOraclePort     = "1521"
	defaultOracleHostPort = "1521"

	// Default Oracle credentials
	defaultUser     = "system"
	defaultPassword = "oracle"
	defaultDatabase = "XEPDB1"

	// Default healthcheck settings
	defaultStartupTimeout = 120 * time.Second
	defaultRetrySleep     = 2 * time.Second
	defaultRetryCount     = 30
)

// OracleContainerConfig stores the configuration for the Oracle XE container
type OracleContainerConfig struct {
	Image            string
	Tag              string
	Port             string
	HostPort         string
	User             string
	Password         string
	Database         string
	StartupTimeout   time.Duration
	RetryCount       int
	RetrySleep       time.Duration
	AdditionalEnvs   map[string]string
	AdditionalLabels map[string]string
}

// DefaultOracleContainerConfig returns the default configuration for the Oracle XE container
func DefaultOracleContainerConfig() *OracleContainerConfig {
	return &OracleContainerConfig{
		Image:          "gvenzl/oracle-xe",
		Tag:            "latest",
		Port:           defaultOraclePort,
		HostPort:       defaultOracleHostPort,
		User:           defaultUser,
		Password:       defaultPassword,
		Database:       defaultDatabase,
		StartupTimeout: defaultStartupTimeout,
		RetryCount:     defaultRetryCount,
		RetrySleep:     defaultRetrySleep,
		AdditionalEnvs: map[string]string{
			"ORACLE_PASSWORD": defaultPassword,
		},
		AdditionalLabels: map[string]string{},
	}
}

// StartOracleContainer starts an Oracle XE container and returns a container instance
// along with connection options compatible with the existing client
func StartOracleContainer(ctx context.Context, config *OracleContainerConfig) (testcontainers.Container, types.ConnectionOptions, error) {
	if config == nil {
		config = DefaultOracleContainerConfig()
	}

	// Prepare environment variables
	env := map[string]string{
		"ORACLE_PASSWORD": config.Password,
	}

	// Add any additional environment variables
	for k, v := range config.AdditionalEnvs {
		env[k] = v
	}

	// Convert port string to nat.Port
	natPort := nat.Port(config.Port + "/tcp")

	// Create container request
	req := testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("%s:%s", config.Image, config.Tag),
		ExposedPorts: []string{fmt.Sprintf("%s:%s", config.Port, config.HostPort)},
		Env:          env,
		WaitingFor: wait.ForAll(
			wait.ForLog("DATABASE IS READY TO USE!").WithStartupTimeout(config.StartupTimeout),
			wait.ForListeningPort(natPort),
		),
		Labels: config.AdditionalLabels,
	}

	// Start container
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		return nil, types.ConnectionOptions{}, fmt.Errorf("failed to start container: %w", err)
	}

	// Get connection details
	connOptions, err := GetConnectionOptionsFromContainer(ctx, container, config)
	if err != nil {
		// If getting connection options fails, try to terminate the container
		_ = container.Terminate(ctx)
		return nil, types.ConnectionOptions{}, err
	}

	return container, connOptions, nil
}

// GetConnectionOptionsFromContainer extracts connection details from a running Oracle container
func GetConnectionOptionsFromContainer(ctx context.Context, container testcontainers.Container, config *OracleContainerConfig) (types.ConnectionOptions, error) {
	host, err := container.Host(ctx)
	if err != nil {
		return types.ConnectionOptions{}, fmt.Errorf("failed to get container host: %w", err)
	}

	// Convert port string to nat.Port
	natPort := nat.Port(config.Port + "/tcp")

	mappedPort, err := container.MappedPort(ctx, natPort)
	if err != nil {
		return types.ConnectionOptions{}, fmt.Errorf("failed to get mapped port: %w", err)
	}

	portNum, err := strconv.Atoi(mappedPort.Port())
	if err != nil {
		return types.ConnectionOptions{}, fmt.Errorf("invalid port number: %w", err)
	}

	// Create the Oracle connection string
	connectStr := fmt.Sprintf("%s:%d/%s", host, portNum, config.Database)

	// Return connection options
	return types.ConnectionOptions{
		Username:   config.User,
		Password:   config.Password,
		ConnectStr: connectStr,
	}, nil
}
