// Package errors provides custom error types and error handling for the go-sqlcl package
package errors

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ErrorType represents the type of error
type ErrorType string

// Error types
const (
	TypeConnectionError ErrorType = "ConnectionError" // Error connecting to the database
	TypeSQLError        ErrorType = "SQLError"        // Error executing SQL
	TypeProcessError    ErrorType = "ProcessError"    // Error with the SQLCL process
	TypeTimeoutError    ErrorType = "TimeoutError"    // Timeout error
	TypeParserError     ErrorType = "ParserError"     // Error parsing SQLCL output
	TypeConfigError     ErrorType = "ConfigError"     // Error with configuration
	TypeUnknownError    ErrorType = "UnknownError"    // Unknown error
)

// OracleErrorCode represents an Oracle error code
type OracleErrorCode string

// Common Oracle error codes
const (
	ORA00001 OracleErrorCode = "ORA-00001" // Unique constraint violated
	ORA00942 OracleErrorCode = "ORA-00942" // Table or view does not exist
	ORA01017 OracleErrorCode = "ORA-01017" // Invalid username/password
	ORA01031 OracleErrorCode = "ORA-01031" // Insufficient privileges
	ORA12514 OracleErrorCode = "ORA-12514" // Service name not found
	ORA12541 OracleErrorCode = "ORA-12541" // No listener
	ORA12545 OracleErrorCode = "ORA-12545" // Connect failed because target host or object does not exist
	ORA03114 OracleErrorCode = "ORA-03114" // Not connected to Oracle
)

// Error represents a go-sqlcl error
type Error struct {
	Type       ErrorType       // Type of error
	Code       OracleErrorCode // Oracle error code, if applicable
	Message    string          // Error message
	SQL        string          // SQL statement that caused the error, if applicable
	LineNumber int             // Line number where the error occurred, if applicable
	Position   int             // Position in the line where the error occurred, if applicable
	Cause      error           // Original error that caused this error
}

// Error implements the error interface
func (e *Error) Error() string {
	msg := string(e.Type)

	if e.Code != "" {
		msg += ": " + string(e.Code)
	}

	msg += ": " + e.Message

	if e.SQL != "" {
		// Truncate SQL if it's too long
		sql := e.SQL
		if len(sql) > 100 {
			sql = sql[:97] + "..."
		}
		msg += " [SQL: " + sql + "]"
	}

	if e.LineNumber > 0 {
		msg += fmt.Sprintf(" at line %d", e.LineNumber)
		if e.Position > 0 {
			msg += fmt.Sprintf(", position %d", e.Position)
		}
	}

	if e.Cause != nil {
		msg += " - caused by: " + e.Cause.Error()
	}

	return msg
}

// Unwrap implements the unwrap interface for errors
func (e *Error) Unwrap() error {
	return e.Cause
}

// Is checks if this error is of the given Oracle error code
func (e *Error) Is(code OracleErrorCode) bool {
	return e.Code == code
}

// IsErrorType checks if this error is of the given error type
func (e *Error) IsErrorType(errorType ErrorType) bool {
	return e.Type == errorType
}

// NewConnectionError creates a new connection error
func NewConnectionError(message string, cause error) *Error {
	return &Error{
		Type:    TypeConnectionError,
		Message: message,
		Cause:   cause,
	}
}

// NewSQLError creates a new SQL error
func NewSQLError(message string, sql string, cause error) *Error {
	return &Error{
		Type:    TypeSQLError,
		Message: message,
		SQL:     sql,
		Cause:   cause,
	}
}

// NewProcessError creates a new process error
func NewProcessError(message string, cause error) *Error {
	return &Error{
		Type:    TypeProcessError,
		Message: message,
		Cause:   cause,
	}
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(message string, sql string, cause error) *Error {
	return &Error{
		Type:    TypeTimeoutError,
		Message: message,
		SQL:     sql,
		Cause:   cause,
	}
}

// NewParserError creates a new parser error
func NewParserError(message string, cause error) *Error {
	return &Error{
		Type:    TypeParserError,
		Message: message,
		Cause:   cause,
	}
}

// NewConfigError creates a new configuration error
func NewConfigError(message string, cause error) *Error {
	return &Error{
		Type:    TypeConfigError,
		Message: message,
		Cause:   cause,
	}
}

// NewUnknownError creates a new unknown error
func NewUnknownError(message string, cause error) *Error {
	return &Error{
		Type:    TypeUnknownError,
		Message: message,
		Cause:   cause,
	}
}

// ParseSQLCLError attempts to parse SQLCL error output and extract error information
func ParseSQLCLError(output string) *Error {
	// Extract Oracle error code (ORA-XXXXX)
	oraPattern := regexp.MustCompile(`(ORA-\d{5})`)
	match := oraPattern.FindString(output)

	var code OracleErrorCode
	if match != "" {
		code = OracleErrorCode(match)
	}

	// Extract line and position information
	linePattern := regexp.MustCompile(`(?i)Line:\s*(\d+)`)
	posPattern := regexp.MustCompile(`(?i)Position:\s*(\d+)`)

	lineMatch := linePattern.FindStringSubmatch(output)
	posMatch := posPattern.FindStringSubmatch(output)

	var lineNum, pos int
	if len(lineMatch) > 1 {
		lineNum, _ = strconv.Atoi(lineMatch[1])
	}

	if len(posMatch) > 1 {
		pos, _ = strconv.Atoi(posMatch[1])
	}

	// Determine error type based on code or content
	errorType := TypeUnknownError

	if strings.Contains(output, "not connected") || code == ORA03114 {
		errorType = TypeConnectionError
	} else if strings.Contains(output, "timeout") || strings.Contains(output, "timed out") {
		errorType = TypeTimeoutError
	} else if code != "" {
		errorType = TypeSQLError
	} else if strings.Contains(output, "process") || strings.Contains(output, "executable") {
		errorType = TypeProcessError
	}

	// Clean message by removing extra whitespace and newlines
	message := output
	message = strings.TrimSpace(message)
	message = regexp.MustCompile(`\s+`).ReplaceAllString(message, " ")

	return &Error{
		Type:       errorType,
		Code:       code,
		Message:    message,
		LineNumber: lineNum,
		Position:   pos,
	}
}

// WithSQL adds SQL statement information to an error
func (e *Error) WithSQL(sql string) *Error {
	e.SQL = sql
	return e
}

// WithCause adds a cause to an error
func (e *Error) WithCause(cause error) *Error {
	e.Cause = cause
	return e
}

// IsConnectionError checks if the error is a connection error
func IsConnectionError(err error) bool {
	var sqlclErr *Error
	if ok := errorAs(err, &sqlclErr); ok {
		return sqlclErr.Type == TypeConnectionError
	}
	return false
}

// IsSQLError checks if the error is a SQL error
func IsSQLError(err error) bool {
	var sqlclErr *Error
	if ok := errorAs(err, &sqlclErr); ok {
		return sqlclErr.Type == TypeSQLError
	}
	return false
}

// IsTimeoutError checks if the error is a timeout error
func IsTimeoutError(err error) bool {
	var sqlclErr *Error
	if ok := errorAs(err, &sqlclErr); ok {
		return sqlclErr.Type == TypeTimeoutError
	}
	return false
}

// IsOracleError checks if the error is from Oracle with the specified code
func IsOracleError(err error, code OracleErrorCode) bool {
	var sqlclErr *Error
	if ok := errorAs(err, &sqlclErr); ok {
		return sqlclErr.Code == code
	}
	return false
}

// errorAs is a helper function for type assertion
func errorAs(err error, target interface{}) bool {
	switch target.(type) {
	case **Error:
		sqlclErr, ok := err.(*Error)
		if !ok {
			return false
		}
		*target.(**Error) = sqlclErr
		return true
	default:
		return false
	}
}
