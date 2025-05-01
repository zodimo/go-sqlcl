// Package config provides configuration options for the go-sqlcl package
package config

import (
	"time"

	"github.com/zodimo/go-sqlcl/pkg/types"
)

// Option represents a functional option for configuring a ClientConfig
type Option func(*types.ClientConfig)

// DefaultSQLclPath is the default path to the SQLcl executable
const DefaultSQLclPath = "sql"

// DefaultTimeout is the default timeout for operations
const DefaultTimeout = 30 * time.Second

// DefaultConnectTimeout is the default timeout for database connections
const DefaultConnectTimeout = 10 * time.Second

// DefaultQueryTimeout is the default timeout for query execution
const DefaultQueryTimeout = 60 * time.Second

// DefaultMaxBufferSize is the default maximum buffer size for output
const DefaultMaxBufferSize = 1024 * 1024 // 1MB

// WithSQLclPath sets the path to the SQLcl executable
func WithSQLclPath(path string) Option {
	return func(c *types.ClientConfig) {
		c.SQLclPath = path
	}
}

// WithTimeout sets the default timeout for operations
func WithTimeout(timeout time.Duration) Option {
	return func(c *types.ClientConfig) {
		c.Timeout = timeout
	}
}

// WithConnectTimeout sets the timeout for database connections
func WithConnectTimeout(timeout time.Duration) Option {
	return func(c *types.ClientConfig) {
		c.ConnectTimeout = timeout
	}
}

// WithQueryTimeout sets the timeout for query execution
func WithQueryTimeout(timeout time.Duration) Option {
	return func(c *types.ClientConfig) {
		c.QueryTimeout = timeout
	}
}

// WithLogLevel sets the log level for the client
func WithLogLevel(level string) Option {
	return func(c *types.ClientConfig) {
		c.LogLevel = level
	}
}

// WithColorOutput sets whether to enable color in output
func WithColorOutput(enabled bool) Option {
	return func(c *types.ClientConfig) {
		c.ColorOutput = enabled
	}
}

// WithStripNewlines sets whether to strip newlines from output
func WithStripNewlines(enabled bool) Option {
	return func(c *types.ClientConfig) {
		c.StripNewlines = enabled
	}
}

// WithOutputFormat sets the default output format
func WithOutputFormat(format string) Option {
	return func(c *types.ClientConfig) {
		c.Format = format
	}
}

// WithMaxBufferSize sets the maximum buffer size for output
func WithMaxBufferSize(size int) Option {
	return func(c *types.ClientConfig) {
		c.MaxBufferSize = size
	}
}

// WithExtraEnv sets extra environment variables for the SQLcl process
func WithExtraEnv(env []string) Option {
	return func(c *types.ClientConfig) {
		c.ExtraEnv = env
	}
}

// WithWorkingDir sets the working directory for the SQLcl process
func WithWorkingDir(dir string) Option {
	return func(c *types.ClientConfig) {
		c.WorkingDir = dir
	}
}

// New creates a new ClientConfig with the given options
func New(options ...Option) *types.ClientConfig {
	// Start with default configuration
	config := &types.ClientConfig{
		SQLclPath:      DefaultSQLclPath,
		Timeout:        DefaultTimeout,
		LogLevel:       "info",
		ColorOutput:    false,
		StripNewlines:  true,
		Format:         "table",
		MaxBufferSize:  DefaultMaxBufferSize,
		ConnectTimeout: DefaultConnectTimeout,
		QueryTimeout:   DefaultQueryTimeout,
	}

	// Apply options
	for _, option := range options {
		option(config)
	}

	return config
}

// DefaultConfig returns a default configuration
func DefaultConfig() *types.ClientConfig {
	return New()
}

// ConnectionOption represents a functional option for configuring ConnectionOptions
type ConnectionOption func(*types.ConnectionOptions)

// WithUsername sets the username for the connection
func WithUsername(username string) ConnectionOption {
	return func(o *types.ConnectionOptions) {
		o.Username = username
	}
}

// WithPassword sets the password for the connection
func WithPassword(password string) ConnectionOption {
	return func(o *types.ConnectionOptions) {
		o.Password = password
	}
}

// WithConnectString sets the connection string
func WithConnectString(connectStr string) ConnectionOption {
	return func(o *types.ConnectionOptions) {
		o.ConnectStr = connectStr
	}
}

// WithWallet sets the path to the Oracle wallet
func WithWallet(wallet string) ConnectionOption {
	return func(o *types.ConnectionOptions) {
		o.Wallet = wallet
	}
}

// WithTNSAdmin sets the path to the tnsnames.ora directory
func WithTNSAdmin(tnsAdmin string) ConnectionOption {
	return func(o *types.ConnectionOptions) {
		o.TNSAdmin = tnsAdmin
	}
}

// WithWalletPassword sets the wallet password
func WithWalletPassword(walletPwd string) ConnectionOption {
	return func(o *types.ConnectionOptions) {
		o.WalletPwd = walletPwd
	}
}

// WithRole sets the role to connect as
func WithRole(role string) ConnectionOption {
	return func(o *types.ConnectionOptions) {
		o.Role = role
	}
}

// WithProxy sets the proxy user to connect through
func WithProxy(proxy string) ConnectionOption {
	return func(o *types.ConnectionOptions) {
		o.Proxy = proxy
	}
}

// WithConnectionType sets the connection type
func WithConnectionType(connType string) ConnectionOption {
	return func(o *types.ConnectionOptions) {
		o.ConnectType = connType
	}
}

// NewConnection creates a new ConnectionOptions with the given options
func NewConnection(options ...ConnectionOption) *types.ConnectionOptions {
	// Start with empty connection options
	connOpts := &types.ConnectionOptions{}

	// Apply options
	for _, option := range options {
		option(connOpts)
	}

	return connOpts
}
