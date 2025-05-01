# Integration Tests for go-sqlcl

This directory contains integration tests for the go-sqlcl package. These tests verify that the package works correctly with a real SQLcl installation.

## Prerequisites

To run these tests, you need:

1. A SQLcl installation (either in your PATH or specified via environment variables)
2. An accessible Oracle database (optional, but recommended for full test coverage)

## Environment Variables

The tests use the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `TEST_SQLCL_PATH` | Path to the SQLcl executable | `sql` (assumes in PATH) |
| `TEST_DB_USER` | Database username | |
| `TEST_DB_PASSWORD` | Database password | |
| `TEST_DB_CONNECT` | Connection string (hostname:port/service) | |
| `TEST_TNS_ADMIN` | Path to the TNS admin directory (optional) | |
| `TEST_WALLET` | Path to the wallet directory (optional) | |
| `TEST_DB_ROLE` | Database role (optional) | |
| `TEST_USE_DEFAULTS` | Whether to use default connection settings (true/false) | `false` |

## Running the Tests

### With SQLcl Only (No Database)

If you only have SQLcl installed but no database, you can run:

```sh
go test -v ./test/integration
```

The tests will detect that no database connection is available and skip the tests that require a database.

### With Database Connection

To run all tests including those that connect to a database:

```sh
export TEST_DB_USER=your_username
export TEST_DB_PASSWORD=your_password
export TEST_DB_CONNECT=localhost:1521/XEPDB1
go test -v ./test/integration
```

### Specifying SQLcl Path

If SQLcl is not in your PATH, specify its location:

```sh
export TEST_SQLCL_PATH=/path/to/sqlcl/bin/sql
go test -v ./test/integration
```

## Test Coverage

The integration tests cover:

1. Client creation and configuration
2. Database connection
3. Basic SQL query execution
4. SQLcl-specific command execution
5. Transaction support
6. Error handling
7. Command wrappers (SET, SHOW, DESCRIBE)
8. Sequential and concurrent query execution

## Troubleshooting

### SQLcl Not Found

If the tests fail with a message like "SQLcl not found", make sure SQLcl is installed and either in your PATH or specified using the `TEST_SQLCL_PATH` environment variable.

### Database Connection Failed

If the tests skip database tests with a message about connection failure, check your database connection details and ensure the database is accessible from your machine.

### Test Timeout

If tests timeout, you might need to adjust the timeout values in the test configuration. These timeouts are set to reasonable defaults, but might need adjustment for slow environments. 