package parser

import (
	"reflect"
	"testing"
)

func TestOutputParser_ParseQueryResult_TableFormat(t *testing.T) {
	parser := NewOutputParser()

	// Sample output in table format
	output := `
EMPLOYEE_ID  FIRST_NAME  LAST_NAME   SALARY     
-----------  ----------  ----------  ---------- 
100          Steven      King        24000      
101          Neena       Kochhar     17000      
102          Lex         De Haan     17000      
3 rows selected

SQL> `

	result, err := parser.ParseQueryResult(output)
	if err != nil {
		t.Fatalf("Failed to parse query result: %v", err)
	}

	// Check result
	expectedColumns := []Column{
		{Name: "EMPLOYEE_ID", Type: ""},
		{Name: "FIRST_NAME", Type: ""},
		{Name: "LAST_NAME", Type: ""},
		{Name: "SALARY", Type: ""},
	}

	if !reflect.DeepEqual(result.Columns, expectedColumns) {
		t.Errorf("Expected columns %v, got %v", expectedColumns, result.Columns)
	}

	expectedRows := []Row{
		{Values: []string{"100", "Steven", "King", "24000"}},
		{Values: []string{"101", "Neena", "Kochhar", "17000"}},
		{Values: []string{"102", "Lex", "De Haan", "17000"}},
	}

	if len(result.Rows) != len(expectedRows) {
		t.Fatalf("Expected %d rows, got %d", len(expectedRows), len(result.Rows))
	}

	for i, row := range result.Rows {
		if !reflect.DeepEqual(row.Values, expectedRows[i].Values) {
			t.Errorf("Row %d: expected %v, got %v", i, expectedRows[i].Values, row.Values)
		}
	}

	if result.Summary != "3 rows selected" {
		t.Errorf("Expected summary '3 rows selected', got '%s'", result.Summary)
	}
}

func TestOutputParser_ParseQueryResult_NoRows(t *testing.T) {
	parser := NewOutputParser()

	// Sample output with no rows
	output := `
EMPLOYEE_ID  FIRST_NAME  LAST_NAME   SALARY     
-----------  ----------  ----------  ---------- 
no rows selected

SQL> `

	result, err := parser.ParseQueryResult(output)
	if err != nil {
		t.Fatalf("Failed to parse query result: %v", err)
	}

	// Check result
	expectedColumns := []Column{
		{Name: "EMPLOYEE_ID", Type: ""},
		{Name: "FIRST_NAME", Type: ""},
		{Name: "LAST_NAME", Type: ""},
		{Name: "SALARY", Type: ""},
	}

	if !reflect.DeepEqual(result.Columns, expectedColumns) {
		t.Errorf("Expected columns %v, got %v", expectedColumns, result.Columns)
	}

	if len(result.Rows) != 0 {
		t.Errorf("Expected 0 rows, got %d", len(result.Rows))
	}

	if result.Summary != "no rows selected" {
		t.Errorf("Expected summary 'no rows selected', got '%s'", result.Summary)
	}
}

func TestOutputParser_ParseQueryResult_JSONFormat(t *testing.T) {
	parser := NewOutputParser()

	// Sample output in JSON format
	output := `
[
  {
    "EMPLOYEE_ID": 100,
    "FIRST_NAME": "Steven",
    "LAST_NAME": "King",
    "SALARY": 24000
  },
  {
    "EMPLOYEE_ID": 101,
    "FIRST_NAME": "Neena",
    "LAST_NAME": "Kochhar",
    "SALARY": 17000
  },
  {
    "EMPLOYEE_ID": 102,
    "FIRST_NAME": "Lex",
    "LAST_NAME": "De Haan",
    "SALARY": 17000
  }
]
SQL> `

	result, err := parser.ParseQueryResult(output)
	if err != nil {
		t.Fatalf("Failed to parse JSON query result: %v", err)
	}

	// Check columns (order might be different due to map iteration)
	expectedColumnNames := map[string]bool{
		"EMPLOYEE_ID": true,
		"FIRST_NAME":  true,
		"LAST_NAME":   true,
		"SALARY":      true,
	}

	if len(result.Columns) != len(expectedColumnNames) {
		t.Errorf("Expected %d columns, got %d", len(expectedColumnNames), len(result.Columns))
	}

	for _, col := range result.Columns {
		if !expectedColumnNames[col.Name] {
			t.Errorf("Unexpected column name: %s", col.Name)
		}
	}

	// Check number of rows
	if len(result.Rows) != 3 {
		t.Fatalf("Expected 3 rows, got %d", len(result.Rows))
	}

	// Check summary
	if result.Summary != "3 rows selected" {
		t.Errorf("Expected summary '3 rows selected', got '%s'", result.Summary)
	}
}

func TestOutputParser_ParseError(t *testing.T) {
	parser := NewOutputParser()

	// Sample error output
	output := `ORA-00942: table or view does not exist
Error at line 1
SQL> `

	errorInfo, err := parser.ParseError(output)
	if err != nil {
		t.Fatalf("Failed to parse error: %v", err)
	}

	// Check error info
	if errorInfo.Code != "ORA-00942" {
		t.Errorf("Expected error code 'ORA-00942', got '%s'", errorInfo.Code)
	}

	if errorInfo.Message != "table or view does not exist" {
		t.Errorf("Expected error message 'table or view does not exist', got '%s'", errorInfo.Message)
	}

	if errorInfo.Line != 1 {
		t.Errorf("Expected error at line 1, got %d", errorInfo.Line)
	}
}

func TestOutputParser_ParseCommandStatus_DDL(t *testing.T) {
	parser := NewOutputParser()

	// Sample output for a successful DDL command
	output := `Table created.

SQL> `

	status, err := parser.ParseCommandStatus(output)
	if err != nil {
		t.Fatalf("Failed to parse command status: %v", err)
	}

	// Check status
	if !status.IsSuccessful {
		t.Errorf("Expected command to be successful")
	}

	if status.RowsAffected != 0 {
		t.Errorf("Expected 0 rows affected, got %d", status.RowsAffected)
	}
}

func TestOutputParser_ParseCommandStatus_DML(t *testing.T) {
	parser := NewOutputParser()

	// Sample output for a successful DML command
	output := `2 rows updated.

SQL> `

	status, err := parser.ParseCommandStatus(output)
	if err != nil {
		t.Fatalf("Failed to parse command status: %v", err)
	}

	// Check status
	if !status.IsSuccessful {
		t.Errorf("Expected command to be successful")
	}

	if status.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", status.RowsAffected)
	}
}

func TestOutputParser_ParseCommandStatus_Error(t *testing.T) {
	parser := NewOutputParser()

	// Sample output for a failed command
	output := `ORA-00001: unique constraint (HR.EMP_PK) violated

SQL> `

	status, err := parser.ParseCommandStatus(output)
	if err != nil {
		t.Fatalf("Failed to parse command status: %v", err)
	}

	// Check status
	if status.IsSuccessful {
		t.Errorf("Expected command to fail")
	}
}
