package container

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"
	"time"

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
	defaultConfig := DefaultOracleContainerConfig()
	log.Printf("Using Docker image: %s:%s", defaultConfig.Image, defaultConfig.Tag)
	log.Printf("Using database user: %s", defaultConfig.User)
	log.Printf("Using environment variables: %v", defaultConfig.AdditionalEnvs)

	globalContainer, globalConnOpts, err = StartOracleContainer(globalCtx, nil)
	if err != nil {
		log.Printf("Failed to start Oracle container: %v\n", err)
		os.Exit(exitCode)
	}

	// Try direct SQL execution through the container to test connectivity
	checkConnectivity(globalContainer)

	// Run the tests
	log.Println("Running tests...")
	exitCode = m.Run()

	// Clean up
	log.Println("Cleaning up test environment...")
	terminateContainer()
	os.Exit(exitCode)
}

// checkConnectivity tests connectivity directly to the container using SQL command
func checkConnectivity(container testcontainers.Container) {
	// Test connectivity directly using the container's SQL command
	log.Println("Testing direct SQL connectivity to container...")

	// Use SQL*Plus directly in the container to verify connectivity
	cmd := "echo 'select 1 from dual;' | sqlplus -s system/oracle@localhost:1521/XEPDB1"
	exitCode, reader, err := container.Exec(context.Background(), []string{"bash", "-c", cmd})
	if err != nil {
		log.Printf("Failed to execute SQL command: %v\n", err)
		return
	}

	output, err := io.ReadAll(reader)
	if err != nil {
		log.Printf("Failed to read command output: %v\n", err)
		return
	}

	log.Printf("SQL execution result (exit code %d): %s", exitCode, string(output))
	if exitCode != 0 {
		log.Printf("SQL command failed with exit code %d", exitCode)
	} else {
		log.Println("Direct SQL connectivity test succeeded!")
	}

	// Try running a simple query using the system account to show database information
	cmd = "echo 'select * from v$instance;' | sqlplus -s sys/oracle@localhost:1521/XEPDB1 as sysdba"
	exitCode, reader, err = container.Exec(context.Background(), []string{"bash", "-c", cmd})
	if err != nil {
		log.Printf("Failed to execute instance query: %v\n", err)
		return
	}

	output, err = io.ReadAll(reader)
	if err != nil {
		log.Printf("Failed to read instance query output: %v\n", err)
		return
	}

	log.Printf("Database instance info: %s", string(output))
}

// terminateContainer terminates the global container
func terminateContainer() {
	if globalContainer != nil {
		if err := globalContainer.Terminate(globalCtx); err != nil {
			log.Printf("Failed to terminate container: %v\n", err)
		}
	}
}

// GetContainerConnectionOptions returns the connection options for the container
func GetContainerConnectionOptions() types.ConnectionOptions {
	return globalConnOpts
}

// GetGlobalContext returns the global context
func GetGlobalContext() context.Context {
	return globalCtx
}

// Helper function to parse host and port from connect string
func parseConnectStr(connectStr string) (string, string, error) {
	parts := strings.Split(connectStr, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid connect string format: %s", connectStr)
	}

	host := parts[0]

	serviceParts := strings.Split(parts[1], "/")
	if len(serviceParts) != 2 {
		return "", "", fmt.Errorf("invalid connect string format: %s", connectStr)
	}

	port := serviceParts[0]

	return host, port, nil
}
