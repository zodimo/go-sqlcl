// Package types defines the public API data types for the go-sqlcl package
package types

import (
	"context"
	"time"
)

// ColumnType represents the data type of a column in a SQL query result
type ColumnType string

// Common column types
const (
	TypeString   ColumnType = "STRING"
	TypeNumber   ColumnType = "NUMBER"
	TypeDate     ColumnType = "DATE"
	TypeDateTime ColumnType = "DATETIME"
	TypeBoolean  ColumnType = "BOOLEAN"
	TypeBLOB     ColumnType = "BLOB"
	TypeCLOB     ColumnType = "CLOB"
	TypeXML      ColumnType = "XML"
	TypeJSON     ColumnType = "JSON"
	TypeUnknown  ColumnType = "UNKNOWN"
)

// Column represents a column in a SQL query result
type Column struct {
	Name string     // Name of the column
	Type ColumnType // Data type of the column
}

// Row represents a row in a SQL query result
type Row struct {
	Values []interface{} // Values in the row, can be strings, numbers, etc.
}

// QueryResult represents the result of a SQL query
type QueryResult struct {
	Columns []Column // Columns in the result
	Rows    []Row    // Rows in the result
	Summary string   // Summary line like "n rows selected"
	Success bool     // Whether the query was successful
	Message string   // Message from the execution (if any)
}

// ConnectionOptions represents options for connecting to a database
type ConnectionOptions struct {
	Username    string // Database username
	Password    string // Database password
	ConnectStr  string // Connection string (host:port/service)
	Wallet      string // Path to Oracle wallet
	TNSAdmin    string // Path to tnsnames.ora directory
	WalletPwd   string // Wallet password
	Role        string // Role to connect as (SYSDBA, SYSOPER, etc.)
	Proxy       string // Proxy user to connect through
	ConnectType string // Connection type (Basic, TNS, etc.)
}

// ClientConfig represents the configuration for the SQLcl client
type ClientConfig struct {
	SQLclPath      string        // Path to the SQLcl executable
	Timeout        time.Duration // Default timeout for operations
	LogLevel       string        // Log level (debug, info, warn, error)
	ColorOutput    bool          // Whether to enable color output
	StripNewlines  bool          // Whether to strip newlines from output
	Format         string        // Output format (csv, json, table, etc.)
	MaxBufferSize  int           // Maximum buffer size for output
	ExtraEnv       []string      // Extra environment variables
	WorkingDir     string        // Working directory for the SQLcl process
	ConnectTimeout time.Duration // Timeout for connection
	QueryTimeout   time.Duration // Timeout for queries
}

// DefaultConfig returns the default configuration for the SQLcl client
func DefaultConfig() *ClientConfig {
	return &ClientConfig{
		SQLclPath:      "sql", // Default to using 'sql' on PATH
		Timeout:        30 * time.Second,
		LogLevel:       "info",
		ColorOutput:    false,
		StripNewlines:  true,
		Format:         "table",
		MaxBufferSize:  1024 * 1024, // 1MB
		ConnectTimeout: 10 * time.Second,
		QueryTimeout:   60 * time.Second,
	}
}

// Command represents a SQLcl command to be executed
type Command struct {
	SQL           string            // SQL statement to execute
	Parameters    map[string]string // Named parameters for the SQL
	Timeout       time.Duration     // Timeout for this specific command
	Format        string            // Output format for this command
	IgnoreErrors  bool              // Whether to continue on errors
	ReturnResults bool              // Whether to return results
}

// ErrorCode represents an Oracle error code
type ErrorCode string

// Common Oracle error codes
const (
	ORA00001 ErrorCode = "ORA-00001" // Unique constraint violated
	ORA00942 ErrorCode = "ORA-00942" // Table or view does not exist
	ORA01017 ErrorCode = "ORA-01017" // Invalid username/password
	ORA01031 ErrorCode = "ORA-01031" // Insufficient privileges
	ORA12514 ErrorCode = "ORA-12514" // Service name not found
	ORA12541 ErrorCode = "ORA-12541" // No listener
	ORA12545 ErrorCode = "ORA-12545" // Connect failed because target host or object does not exist
)

// Error represents an error from the SQLcl client
type Error struct {
	Code    ErrorCode // Error code like ORA-00001
	Message string    // Error message
	SQL     string    // SQL statement that caused the error
	Line    int       // Line number where the error occurred, if available
	Column  int       // Column number where the error occurred, if available
}

// Error implements the error interface
func (e *Error) Error() string {
	return string(e.Code) + ": " + e.Message
}

// Is checks if this error is of the given error code
func (e *Error) Is(code ErrorCode) bool {
	return e.Code == code
}

// Client defines the interface for interacting with the SQLcl client
type Client interface {
	// Connect establishes a connection to the database
	Connect(ctx context.Context, connectString string) error

	// ConnectWithOptions establishes a connection to the database with options
	ConnectWithOptions(ctx context.Context, options ConnectionOptions) error

	// ExecuteSQL executes a SQL statement and returns the result
	ExecuteSQL(ctx context.Context, sql string) (*QueryResult, error)

	// ExecuteCommand executes a command and returns the result
	ExecuteCommand(ctx context.Context, command Command) (*QueryResult, error)

	// Close closes the connection to the database
	Close() error
}
