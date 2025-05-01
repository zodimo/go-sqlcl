# Testing for go-sqlcl

This directory contains tests for the go-sqlcl package.

## Test Directories

- **integration/**: Integration tests that test the package against a real SQLcl installation.

## Running Tests

### Run All Tests

```sh
go test ./...
```

### Run Integration Tests Only

```sh
go test -v ./test/integration
```

### Run With Coverage

```sh
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test Prerequisites

Different tests may have different prerequisites. See the README in each test directory for specific requirements. 