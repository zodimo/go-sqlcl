package sqlcl_test

import (
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestDummy is a simple test to ensure testcontainers-go is properly imported
func TestDummy(t *testing.T) {
	// This test doesn't do anything but import the packages
	_ = testcontainers.ContainerRequest{}
	_ = wait.Strategy(nil)
}
