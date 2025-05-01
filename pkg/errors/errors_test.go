package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestErrorInterface(t *testing.T) {
	// Test that our Error type implements the error interface
	var _ error = &Error{}
}

func TestErrorCreation(t *testing.T) {
	tests := []struct {
		name          string
		createError   func() *Error
		expectedType  ErrorType
		expectedCode  OracleErrorCode
		checkSQLSet   bool
		checkCauseSet bool
	}{
		{
			name: "ConnectionError",
			createError: func() *Error {
				return NewConnectionError("Failed to connect", nil)
			},
			expectedType: TypeConnectionError,
		},
		{
			name: "SQLError",
			createError: func() *Error {
				return NewSQLError("SQL syntax error", "SELECT * FROM", nil)
			},
			expectedType: TypeSQLError,
			checkSQLSet:  true,
		},
		{
			name: "ProcessError",
			createError: func() *Error {
				return NewProcessError("Failed to start process", nil)
			},
			expectedType: TypeProcessError,
		},
		{
			name: "TimeoutError",
			createError: func() *Error {
				return NewTimeoutError("Operation timed out", "SELECT * FROM long_table", nil)
			},
			expectedType: TypeTimeoutError,
			checkSQLSet:  true,
		},
		{
			name: "ParserError",
			createError: func() *Error {
				return NewParserError("Failed to parse output", nil)
			},
			expectedType: TypeParserError,
		},
		{
			name: "ConfigError",
			createError: func() *Error {
				return NewConfigError("Invalid config", nil)
			},
			expectedType: TypeConfigError,
		},
		{
			name: "UnknownError",
			createError: func() *Error {
				return NewUnknownError("Unknown error", nil)
			},
			expectedType: TypeUnknownError,
		},
		{
			name: "ErrorWithCause",
			createError: func() *Error {
				cause := fmt.Errorf("original error")
				return NewSQLError("SQL error with cause", "SELECT * FROM", cause)
			},
			expectedType:  TypeSQLError,
			checkSQLSet:   true,
			checkCauseSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.createError()

			// Check error type
			if err.Type != tt.expectedType {
				t.Errorf("Expected error type %s, got %s", tt.expectedType, err.Type)
			}

			// Check if code is set when expected
			if tt.expectedCode != "" && err.Code != tt.expectedCode {
				t.Errorf("Expected error code %s, got %s", tt.expectedCode, err.Code)
			}

			// Check if SQL is set when expected
			if tt.checkSQLSet && err.SQL == "" {
				t.Errorf("Expected SQL to be set, but it was empty")
			}

			// Check if cause is set when expected
			if tt.checkCauseSet && err.Cause == nil {
				t.Errorf("Expected cause to be set, but it was nil")
			}

			// Check that Error() returns a non-empty string
			if err.Error() == "" {
				t.Errorf("Error() returned an empty string")
			}
		})
	}
}

func TestWithSQL(t *testing.T) {
	err := NewConnectionError("Failed to connect", nil)
	sql := "SELECT * FROM users"
	err = err.WithSQL(sql)

	if err.SQL != sql {
		t.Errorf("Expected SQL %s, got %s", sql, err.SQL)
	}
}

func TestWithCause(t *testing.T) {
	cause := fmt.Errorf("original error")
	err := NewConnectionError("Failed to connect", nil)
	err = err.WithCause(cause)

	if err.Cause != cause {
		t.Errorf("Expected cause to be set correctly")
	}
}

func TestParseSQLCLError(t *testing.T) {
	tests := []struct {
		name           string
		output         string
		expectedType   ErrorType
		expectedCode   OracleErrorCode
		expectedLine   int
		expectedPos    int
		shouldContain  string
		shouldNotMatch []string
	}{
		{
			name:          "ORA-01017",
			output:        "ERROR: ORA-01017: invalid username/password; logon denied",
			expectedType:  TypeSQLError,
			expectedCode:  ORA01017,
			shouldContain: "invalid username/password",
		},
		{
			name:          "ORA-00942",
			output:        "ERROR:\nORA-00942: table or view does not exist\nLine: 1, Position: 15",
			expectedType:  TypeSQLError,
			expectedCode:  ORA00942,
			expectedLine:  1,
			expectedPos:   15,
			shouldContain: "table or view does not exist",
		},
		{
			name:          "Not connected",
			output:        "ERROR: not connected to database",
			expectedType:  TypeConnectionError,
			shouldContain: "not connected to database",
		},
		{
			name:          "Timeout error",
			output:        "ERROR: operation timed out after 30 seconds",
			expectedType:  TypeTimeoutError,
			shouldContain: "operation timed out",
		},
		{
			name:          "Process error",
			output:        "ERROR: failed to start SQLCL process: executable not found",
			expectedType:  TypeProcessError,
			shouldContain: "executable not found",
		},
		{
			name:          "Unknown error",
			output:        "ERROR: something went wrong",
			expectedType:  TypeUnknownError,
			shouldContain: "something went wrong",
		},
		{
			name:           "Multiline error",
			output:         "ERROR:\nSomething went wrong\nwith multiple\nlines of output",
			expectedType:   TypeUnknownError,
			shouldContain:  "Something went wrong with multiple lines of output",
			shouldNotMatch: []string{"\n"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ParseSQLCLError(tt.output)

			// Check error type
			if err.Type != tt.expectedType {
				t.Errorf("Expected error type %s, got %s", tt.expectedType, err.Type)
			}

			// Check error code if expected
			if tt.expectedCode != "" && err.Code != tt.expectedCode {
				t.Errorf("Expected error code %s, got %s", tt.expectedCode, err.Code)
			}

			// Check line number if expected
			if tt.expectedLine > 0 && err.LineNumber != tt.expectedLine {
				t.Errorf("Expected line number %d, got %d", tt.expectedLine, err.LineNumber)
			}

			// Check position if expected
			if tt.expectedPos > 0 && err.Position != tt.expectedPos {
				t.Errorf("Expected position %d, got %d", tt.expectedPos, err.Position)
			}

			// Check that message contains expected substring
			if tt.shouldContain != "" && !contains(err.Message, tt.shouldContain) {
				t.Errorf("Expected message to contain '%s', got '%s'", tt.shouldContain, err.Message)
			}

			// Check that message does not contain any of the shouldNotMatch strings
			for _, s := range tt.shouldNotMatch {
				if contains(err.Message, s) {
					t.Errorf("Expected message not to contain '%s', but it did: '%s'", s, err.Message)
				}
			}
		})
	}
}

func TestError_Is(t *testing.T) {
	err := &Error{
		Type:    TypeSQLError,
		Code:    ORA01017,
		Message: "invalid username/password",
	}

	if !err.Is(ORA01017) {
		t.Errorf("Expected error to be ORA-01017")
	}

	if err.Is(ORA00942) {
		t.Errorf("Expected error not to be ORA-00942")
	}
}

func TestIsErrorFunctions(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		checkFunction func(error) bool
		expected      bool
	}{
		{
			name: "IsConnectionError with connection error",
			err: &Error{
				Type: TypeConnectionError,
			},
			checkFunction: IsConnectionError,
			expected:      true,
		},
		{
			name: "IsConnectionError with SQL error",
			err: &Error{
				Type: TypeSQLError,
			},
			checkFunction: IsConnectionError,
			expected:      false,
		},
		{
			name: "IsSQLError with SQL error",
			err: &Error{
				Type: TypeSQLError,
			},
			checkFunction: IsSQLError,
			expected:      true,
		},
		{
			name: "IsTimeoutError with timeout error",
			err: &Error{
				Type: TypeTimeoutError,
			},
			checkFunction: IsTimeoutError,
			expected:      true,
		},
		{
			name:          "IsConnectionError with non-sqlcl error",
			err:           errors.New("regular error"),
			checkFunction: IsConnectionError,
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.checkFunction(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsOracleError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		code     OracleErrorCode
		expected bool
	}{
		{
			name: "IsOracleError with matching code",
			err: &Error{
				Code: ORA01017,
			},
			code:     ORA01017,
			expected: true,
		},
		{
			name: "IsOracleError with non-matching code",
			err: &Error{
				Code: ORA00942,
			},
			code:     ORA01017,
			expected: false,
		},
		{
			name:     "IsOracleError with non-sqlcl error",
			err:      errors.New("regular error"),
			code:     ORA01017,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsOracleError(tt.err, tt.code)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
