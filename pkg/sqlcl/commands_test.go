package sqlcl

import (
	"testing"
)

// TestParseDescribeOutput tests the parseDescribeOutput function
func TestParseDescribeOutput(t *testing.T) {
	testCases := []struct {
		name          string
		output        string
		objectName    string
		expectError   bool
		objectType    string
		columnCount   int
		firstColName  string
		firstColType  string
		firstNullable bool
	}{
		{
			name: "Parse table output",
			output: `
Name                             Null?    Type
-------------------------------- -------- ----------------------------
ID                               NOT NULL NUMBER(10)
NAME                                      VARCHAR2(100)
DESCRIPTION                               CLOB
CREATED_AT                               DATE
UPDATED_AT                               TIMESTAMP(6)
SQL>
`,
			objectName:    "TEST_TABLE",
			expectError:   false,
			objectType:    "TABLE",
			columnCount:   5,
			firstColName:  "ID",
			firstColType:  "NUMBER(10)",
			firstNullable: false,
		},
		{
			name: "Parse view output",
			output: `
View TEST_VIEW is a view on the following tables:
- TABLE1
- TABLE2

Name           Null?    Type
-------------- -------- ----------------------------
ID             NOT NULL NUMBER(10)
NAME                    VARCHAR2(100)
SQL>
`,
			objectName:    "TEST_VIEW",
			expectError:   false,
			objectType:    "VIEW",
			columnCount:   2,
			firstColName:  "ID",
			firstColType:  "NUMBER(10)",
			firstNullable: false,
		},
		{
			name:        "Invalid output",
			output:      "Invalid output",
			objectName:  "TEST_INVALID",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseDescribeOutput(tc.output, tc.objectName)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.ObjectName != tc.objectName {
				t.Errorf("Expected object name %s, got %s", tc.objectName, result.ObjectName)
			}

			if result.ObjectType != tc.objectType {
				t.Errorf("Expected object type %s, got %s", tc.objectType, result.ObjectType)
			}

			if len(result.Columns) != tc.columnCount {
				t.Errorf("Expected %d columns, got %d", tc.columnCount, len(result.Columns))
			}

			if len(result.Columns) > 0 {
				if result.Columns[0].Name != tc.firstColName {
					t.Errorf("Expected first column name %s, got %s", tc.firstColName, result.Columns[0].Name)
				}

				if result.Columns[0].Type != tc.firstColType {
					t.Errorf("Expected first column type %s, got %s", tc.firstColType, result.Columns[0].Type)
				}

				if result.Columns[0].Nullable != tc.firstNullable {
					t.Errorf("Expected first column nullable %v, got %v", tc.firstNullable, result.Columns[0].Nullable)
				}
			}
		})
	}
}

// TestParseHistoryOutput tests the parseHistoryOutput function
func TestParseHistoryOutput(t *testing.T) {
	testCases := []struct {
		name          string
		output        string
		expectedCount int
		firstCommand  string
	}{
		{
			name: "Parse history with multiple entries",
			output: `
HISTORY
  1  SELECT * FROM users;
  2  CREATE TABLE test (id NUMBER);
  3  DESCRIBE users
SQL>
`,
			expectedCount: 3,
			firstCommand:  "SELECT * FROM users;",
		},
		{
			name: "Parse empty history",
			output: `
HISTORY
No history available
SQL>
`,
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			history := parseHistoryOutput(tc.output)

			if len(history) != tc.expectedCount {
				t.Errorf("Expected %d history entries, got %d", tc.expectedCount, len(history))
			}

			if tc.expectedCount > 0 && len(history) > 0 {
				if history[0] != tc.firstCommand {
					t.Errorf("Expected first command %s, got %s", tc.firstCommand, history[0])
				}
			}
		})
	}
}
