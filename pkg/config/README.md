# Configuration Package for go-sqlcl

This package provides a flexible configuration system for the go-sqlcl client using the functional options pattern. It allows users to customize SQLcl behavior including timeouts, paths, and output formats.

## Basic Usage

```go
import (
    "github.com/zodimo/go-sqlcl/pkg/config"
    "github.com/zodimo/go-sqlcl/pkg/sqlcl"
    "time"
)

// Create config with custom options
clientConfig := config.New(
    config.WithSQLclPath("/path/to/sql"),
    config.WithTimeout(30 * time.Second),
    config.WithOutputFormat("json"),
)

// Create client with custom config
client, err := sqlcl.NewClient(clientConfig)
```

## Available Configuration Options

### Client Configuration

| Option | Description | Default |
|--------|-------------|---------|
| WithSQLclPath | Path to SQLcl executable | "sql" |
| WithTimeout | Default timeout for operations | 30s |
| WithConnectTimeout | Timeout for database connections | 10s |
| WithQueryTimeout | Timeout for query execution | 60s |
| WithLogLevel | Log level (debug, info, warn, error) | "info" |
| WithColorOutput | Enable/disable color in output | false |
| WithStripNewlines | Strip newlines from output | true |
| WithOutputFormat | Output format (csv, json, table) | "table" |
| WithMaxBufferSize | Maximum buffer size for output | 1MB |
| WithExtraEnv | Extra environment variables | nil |
| WithWorkingDir | Working directory for SQLcl process | "" |

### Connection Configuration

```go
// Create connection options with custom settings
connOpts := config.NewConnection(
    config.WithUsername("system"),
    config.WithPassword("password"),
    config.WithConnectString("localhost:1521/orclpdb"),
    config.WithRole("SYSDBA"),
)

// Connect using these options
client.ConnectWithOptions(ctx, *connOpts)
```

| Option | Description |
|--------|-------------|
| WithUsername | Database username |
| WithPassword | Database password |
| WithConnectString | Connection string (host:port/service) |
| WithWallet | Path to Oracle wallet |
| WithTNSAdmin | Path to tnsnames.ora directory |
| WithWalletPassword | Wallet password |
| WithRole | Role to connect as (SYSDBA, SYSOPER, etc.) |
| WithProxy | Proxy user to connect through |
| WithConnectionType | Connection type |

## Using Default Configuration

For simple use cases, you can use the default configuration:

```go
// Get default configuration
config := config.DefaultConfig()

// Create client with default config
client, err := sqlcl.NewClient(config)
```

Default configuration provides reasonable timeouts and settings suitable for most common use cases. 