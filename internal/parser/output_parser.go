// Package parser provides functionality for parsing SQLcl output
package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	// ErrInvalidOutput is returned when the output cannot be parsed
	ErrInvalidOutput = errors.New("invalid sqlcl output format")

	// ErrNoRows is returned when a query returns no rows
	ErrNoRows = errors.New("no rows returned")
)

// Column represents a column in a SQL query result
type Column struct {
	Name string
	Type string
}

// Row represents a row in a SQL query result
type Row struct {
	Values []string
}

// QueryResult represents the result of a SQL query
type QueryResult struct {
	Columns []Column
	Rows    []Row
	Summary string // Summary line like "n rows selected"
}

// ErrorInfo represents an extracted error from SQLcl output
type ErrorInfo struct {
	Code    string // Error code like ORA-00001
	Message string // Error message
	Line    int    // Line number where the error occurred, if available
	Column  int    // Column number where the error occurred, if available
}

// CommandStatus represents the status of a command execution
type CommandStatus struct {
	IsSuccessful bool
	Message      string
	RowsAffected int // For DML commands
}

// OutputParser parses SQLcl output
type OutputParser struct {
	// Regular expressions for parsing different parts of the output
	tableHeaderRegex     *regexp.Regexp
	tableRowRegex        *regexp.Regexp
	errorRegex           *regexp.Regexp
	rowsAffectedRegex    *regexp.Regexp
	rowsSelectedRegex    *regexp.Regexp
	commandSuccessRegex  *regexp.Regexp
	commandFailureRegex  *regexp.Regexp
	promptRegex          *regexp.Regexp
	columnLineRegex      *regexp.Regexp
	columnSeparatorRegex *regexp.Regexp
	jsonOutputRegex      *regexp.Regexp
}

// NewOutputParser creates a new OutputParser
func NewOutputParser() *OutputParser {
	return &OutputParser{
		tableHeaderRegex:     regexp.MustCompile(`-{2,}\s*\+\s*-{2,}`),
		tableRowRegex:        regexp.MustCompile(`\|\s*`),
		errorRegex:           regexp.MustCompile(`(ORA-\d{5}|SP2-\d{4})\s*:\s*(.*)`),
		rowsAffectedRegex:    regexp.MustCompile(`(\d+)\s+rows?\s+(created|updated|deleted|merged|inserted)`),
		rowsSelectedRegex:    regexp.MustCompile(`(\d+)\s+rows?\s+selected`),
		commandSuccessRegex:  regexp.MustCompile(`(Table|View|Index|Sequence|Trigger|Procedure|Function|Package)\s+created`),
		commandFailureRegex:  regexp.MustCompile(`(ERROR|Warning|ORA-\d{5}|SP2-\d{4})`),
		promptRegex:          regexp.MustCompile(`SQL>\s*$`),
		columnLineRegex:      regexp.MustCompile(`-{2,}`),
		columnSeparatorRegex: regexp.MustCompile(`\s{2,}|\t+`),
		jsonOutputRegex:      regexp.MustCompile(`^\s*\[\s*\{.*\}\s*\]\s*$`),
	}
}

// ParseQueryResult parses the output of a SQL query into a structured QueryResult
func (p *OutputParser) ParseQueryResult(output string) (*QueryResult, error) {
	// Remove the prompt and empty lines
	cleanedOutput := p.cleanOutput(output)

	// Check if the output is in JSON format
	if p.isJSONOutput(cleanedOutput) {
		return p.parseJSONOutput(cleanedOutput)
	}

	// Check if there are any results
	if p.rowsSelectedRegex.MatchString(cleanedOutput) {
		// Extract summary line (e.g., "10 rows selected")
		summaryMatch := p.rowsSelectedRegex.FindStringSubmatch(cleanedOutput)
		summary := ""
		if len(summaryMatch) >= 1 {
			summary = summaryMatch[0]
		}

		// Remove summary line from output for further processing
		cleanedOutput = strings.ReplaceAll(cleanedOutput, summary, "")

		// Check if output has table format
		if p.tableHeaderRegex.MatchString(cleanedOutput) {
			return p.parseTableOutput(cleanedOutput, summary)
		} else {
			// Handle non-tabular output (space or tab-separated values)
			return p.parseColumnOutput(cleanedOutput, summary)
		}
	} else if strings.Contains(cleanedOutput, "no rows selected") {
		// Return empty result with columns if available
		result := &QueryResult{
			Columns: []Column{},
			Rows:    []Row{},
			Summary: "no rows selected",
		}

		// Try to extract column headers if they exist
		lines := strings.Split(cleanedOutput, "\n")
		for i, line := range lines {
			if p.columnLineRegex.MatchString(line) && i > 0 {
				// Previous line should contain column headers
				headerLine := lines[i-1]
				headers := p.parseColumnHeaders(headerLine)
				for _, header := range headers {
					result.Columns = append(result.Columns, Column{Name: header, Type: ""})
				}
				break
			}
		}

		return result, nil
	}

	return nil, ErrNoRows
}

// isJSONOutput checks if the output is in JSON format
func (p *OutputParser) isJSONOutput(output string) bool {
	// Check if the output starts with [ and ends with ]
	trimmedOutput := strings.TrimSpace(output)
	return strings.HasPrefix(trimmedOutput, "[") &&
		strings.HasSuffix(trimmedOutput, "]") &&
		strings.Contains(trimmedOutput, "{") &&
		strings.Contains(trimmedOutput, "}")
}

// parseJSONOutput parses output in JSON format
func (p *OutputParser) parseJSONOutput(output string) (*QueryResult, error) {
	// Extract the JSON part from the output
	jsonStart := strings.Index(output, "[")
	jsonEnd := strings.LastIndex(output, "]") + 1

	if jsonStart < 0 || jsonEnd <= jsonStart {
		return nil, ErrInvalidOutput
	}

	jsonStr := output[jsonStart:jsonEnd]

	// Parse the JSON into a slice of maps
	var data []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	result := &QueryResult{
		Summary: fmt.Sprintf("%d rows selected", len(data)),
	}

	// Extract column names from the first row
	if len(data) > 0 {
		for key := range data[0] {
			result.Columns = append(result.Columns, Column{Name: key, Type: ""})
		}
	}

	// Extract row values
	for _, item := range data {
		var rowValues []string
		for _, col := range result.Columns {
			value := ""
			if val, ok := item[col.Name]; ok {
				// Convert the value to string
				switch v := val.(type) {
				case string:
					value = v
				case float64:
					value = fmt.Sprintf("%g", v)
				case bool:
					value = fmt.Sprintf("%t", v)
				case nil:
					value = "NULL"
				default:
					value = fmt.Sprintf("%v", v)
				}
			}
			rowValues = append(rowValues, value)
		}
		result.Rows = append(result.Rows, Row{Values: rowValues})
	}

	return result, nil
}

// ParseError parses the output to extract error information
func (p *OutputParser) ParseError(output string) (*ErrorInfo, error) {
	errorMatch := p.errorRegex.FindStringSubmatch(output)
	if len(errorMatch) < 3 {
		return nil, ErrInvalidOutput
	}

	errorInfo := &ErrorInfo{
		Code:    errorMatch[1],
		Message: strings.TrimSpace(errorMatch[2]),
		Line:    -1,
		Column:  -1,
	}

	// Look for line and column information
	lineColRegex := regexp.MustCompile(`line\s+(\d+)(?:,\s*column\s+(\d+))?`)
	lineColMatch := lineColRegex.FindStringSubmatch(output)
	if len(lineColMatch) >= 2 {
		lineNum := 0
		fmt.Sscanf(lineColMatch[1], "%d", &lineNum)
		errorInfo.Line = lineNum

		if len(lineColMatch) >= 3 && lineColMatch[2] != "" {
			colNum := 0
			fmt.Sscanf(lineColMatch[2], "%d", &colNum)
			errorInfo.Column = colNum
		}
	}

	return errorInfo, nil
}

// ParseCommandStatus parses the output to determine the status of a command execution
func (p *OutputParser) ParseCommandStatus(output string) (*CommandStatus, error) {
	// Check for successful DDL commands
	if p.commandSuccessRegex.MatchString(output) {
		return &CommandStatus{
			IsSuccessful: true,
			Message:      p.commandSuccessRegex.FindString(output),
			RowsAffected: 0,
		}, nil
	}

	// Check for DML commands that affect rows
	rowsAffectedMatch := p.rowsAffectedRegex.FindStringSubmatch(output)
	if len(rowsAffectedMatch) >= 3 {
		rowsAffected := 0
		fmt.Sscanf(rowsAffectedMatch[1], "%d", &rowsAffected)
		return &CommandStatus{
			IsSuccessful: true,
			Message:      rowsAffectedMatch[0],
			RowsAffected: rowsAffected,
		}, nil
	}

	// Check for command failures
	if p.commandFailureRegex.MatchString(output) {
		return &CommandStatus{
			IsSuccessful: false,
			Message:      p.commandFailureRegex.FindString(output),
			RowsAffected: 0,
		}, nil
	}

	// Default case, assume command was successful if no errors found
	return &CommandStatus{
		IsSuccessful: true,
		Message:      "Command executed successfully",
		RowsAffected: 0,
	}, nil
}

// Helper methods

// cleanOutput removes SQL prompt and empty lines from the output
func (p *OutputParser) cleanOutput(output string) string {
	// Remove SQL prompt
	cleanedOutput := p.promptRegex.ReplaceAllString(output, "")

	// Remove empty lines
	lines := strings.Split(cleanedOutput, "\n")
	var nonEmptyLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}

	return strings.Join(nonEmptyLines, "\n")
}

// parseTableOutput parses output in table format (with | separators)
func (p *OutputParser) parseTableOutput(output, summary string) (*QueryResult, error) {
	lines := strings.Split(output, "\n")
	if len(lines) < 3 { // Need at least header, separator, and one data row
		return nil, ErrInvalidOutput
	}

	var result QueryResult
	result.Summary = summary

	// Find the header line and separator line
	var headerLine, separatorLine int
	for i, line := range lines {
		if p.tableHeaderRegex.MatchString(line) {
			separatorLine = i
			headerLine = i - 1
			break
		}
	}

	if headerLine < 0 || separatorLine <= headerLine {
		return nil, ErrInvalidOutput
	}

	// Parse column names
	headerParts := strings.Split(strings.TrimSpace(lines[headerLine]), "|")
	for _, part := range headerParts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result.Columns = append(result.Columns, Column{Name: trimmed, Type: ""})
		}
	}

	// Parse data rows
	for i := separatorLine + 1; i < len(lines); i++ {
		line := lines[i]
		// Skip separator and summary lines
		if p.tableHeaderRegex.MatchString(line) ||
			p.rowsSelectedRegex.MatchString(line) ||
			strings.TrimSpace(line) == "" {
			continue
		}

		// Parse row
		rowParts := strings.Split(line, "|")
		var rowValues []string
		for _, part := range rowParts {
			trimmed := strings.TrimSpace(part)
			if part != rowParts[0] && part != rowParts[len(rowParts)-1] { // Skip first and last empty parts
				rowValues = append(rowValues, trimmed)
			}
		}

		if len(rowValues) > 0 {
			result.Rows = append(result.Rows, Row{Values: rowValues})
		}
	}

	return &result, nil
}

// parseColumnOutput parses output in column format (space/tab separated)
func (p *OutputParser) parseColumnOutput(output, summary string) (*QueryResult, error) {
	lines := strings.Split(output, "\n")
	if len(lines) < 2 { // Need at least header and separator
		return nil, ErrInvalidOutput
	}

	var result QueryResult
	result.Summary = summary

	// Find the header line and separator line
	var headerLine, separatorLine int
	for i, line := range lines {
		if p.columnLineRegex.MatchString(line) {
			separatorLine = i
			headerLine = i - 1
			break
		}
	}

	if headerLine < 0 || separatorLine <= headerLine {
		return nil, ErrInvalidOutput
	}

	// Parse column names
	headers := p.parseColumnHeaders(lines[headerLine])
	for _, header := range headers {
		result.Columns = append(result.Columns, Column{Name: header, Type: ""})
	}

	// Parse data rows
	for i := separatorLine + 1; i < len(lines); i++ {
		line := lines[i]
		// Skip separator and summary lines
		if p.columnLineRegex.MatchString(line) ||
			p.rowsSelectedRegex.MatchString(line) ||
			strings.TrimSpace(line) == "" {
			continue
		}

		// Parse row values using same spacing as headers
		rowValues := p.parseRowValues(line, headers)
		if len(rowValues) > 0 {
			result.Rows = append(result.Rows, Row{Values: rowValues})
		}
	}

	return &result, nil
}

// parseColumnHeaders parses column headers from a space/tab separated header line
func (p *OutputParser) parseColumnHeaders(headerLine string) []string {
	// First try to split by multiple spaces
	parts := p.columnSeparatorRegex.Split(headerLine, -1)

	// Filter out empty parts
	var headers []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			headers = append(headers, trimmed)
		}
	}

	return headers
}

// parseRowValues parses row values using the same positions as headers
func (p *OutputParser) parseRowValues(line string, headers []string) []string {
	// This is a simplistic approach - in a real implementation we'd need to
	// track the positions of each column in the header line and use those
	// to extract values from the data lines
	parts := p.columnSeparatorRegex.Split(line, -1)

	// Filter out empty parts
	var values []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			values = append(values, trimmed)
		}
	}

	// Ensure we have the same number of values as headers
	if len(values) > len(headers) {
		values = values[:len(headers)]
	} else if len(values) < len(headers) {
		// Pad with empty values
		for i := len(values); i < len(headers); i++ {
			values = append(values, "")
		}
	}

	return values
}
