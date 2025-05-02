# Testing for go-sqlcl

This directory contains tests for the go-sqlcl package.

## Test Directories

- **integration/**: Integration tests that test the package against a real SQLcl installation.
- **container/**: Container-based tests that use Docker containers to run tests against an Oracle database.

## Test Approaches

### Standard Integration Tests

The integration tests require a local Oracle SQLcl installation and optionally a local Oracle database. These tests validate that the package works correctly with real database operations but require more manual setup.

### Container-Based Tests

The container-based tests use [testcontainers-go](https://github.com/testcontainers/testcontainers-go) to automatically create and manage Docker containers running Oracle XE. This approach provides the following benefits:

- **Isolated testing environment**: Each test gets a fresh database instance, eliminating test interference
- **No manual setup**: No need to install or configure Oracle locally
- **Reproducible tests**: Tests run in the same environment every time
- **CI/CD friendly**: Easy to incorporate into automated testing workflows

#### Prerequisites for Container-Based Testing

- Docker installed and running on your system
- Go 1.24 or higher
- Oracle SQLcl installed (on your system path as "sql")
- Internet connection (first run will download the Oracle XE Docker image)

#### Running Container-Based Tests

```sh
# Run all container tests
go test -v ./test/container

# Run a specific container test
go test -v ./test/container -run TestBasicSQL

# Skip container tests (useful for quick runs)
SKIP_CONTAINER_TESTS=1 go test ./...
```

#### Container Test Types

The container tests cover various use cases:

1. **Basic SQL Operations**: Testing simple SQL commands and queries
2. **Transaction Handling**: Testing commit and rollback operations
3. **Error Handling**: Testing how the package handles various error conditions
4. **Concurrent Operations**: Testing concurrent connections and operations
5. **Database Migrations**: Testing migration scripts and schema changes

## Running Tests

### Run All Tests

```sh
go test ./...
```

### Run Integration Tests Only

```sh
go test -v ./test/integration
```

### Run Container Tests Only

```sh
go test -v ./test/container
```

### Skip Container Tests

```sh
SKIP_CONTAINER_TESTS=1 go test ./...
```

### Run With Coverage

```sh
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test Prerequisites

Different tests may have different prerequisites:

- **Unit tests**: No special requirements
- **Integration tests**: SQLcl installed, optionally an Oracle database
- **Container tests**: Docker installed, SQLcl installed, internet connection 