// Package sqlcl provides the main client API for interacting with SQLCL
package sqlcl

import (
	"context"
	"fmt"
	"strings"

	"github.com/zodimo/go-sqlcl/pkg/types"
)

// DescribeResult represents the result of a DESCRIBE command
type DescribeResult struct {
	ObjectName string
	ObjectType string
	Columns    []DescribeColumn
	Properties map[string]string
}

// DescribeColumn represents a column in a DESCRIBE result
type DescribeColumn struct {
	Name     string
	Type     string
	Nullable bool
	Default  string
}

// SetCommand sets a SQLcl option to a specified value
func (c *Client) SetCommand(ctx context.Context, option, value string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.process.IsRunning() {
		return fmt.Errorf("sqlcl process is not running")
	}

	// Create a context with timeout if the provided context doesn't have a deadline
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, c.config.QueryTimeout)
		defer cancel()
	}

	// Execute the SET command
	command := fmt.Sprintf("SET %s %s", option, value)
	result, err := c.executor.ExecuteSQLCLCommand(ctx, command)
	if err != nil {
		return fmt.Errorf("failed to execute SET command: %w", err)
	}

	if !result.IsSuccessful {
		return fmt.Errorf("SET command failed: %s", result.Error)
	}

	return nil
}

// ShowCommand executes a SHOW command and returns the result
func (c *Client) ShowCommand(ctx context.Context, option string) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.process.IsRunning() {
		return "", fmt.Errorf("sqlcl process is not running")
	}

	// Create a context with timeout if the provided context doesn't have a deadline
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, c.config.QueryTimeout)
		defer cancel()
	}

	// Execute the SHOW command
	command := fmt.Sprintf("SHOW %s", option)
	result, err := c.executor.ExecuteSQLCLCommand(ctx, command)
	if err != nil {
		return "", fmt.Errorf("failed to execute SHOW command: %w", err)
	}

	if !result.IsSuccessful {
		return "", fmt.Errorf("SHOW command failed: %s", result.Error)
	}

	// Clean and return the output
	output := strings.TrimSpace(result.Output)
	// Remove the command echo and SQL prompt
	output = strings.TrimPrefix(output, command)
	output = strings.TrimSpace(output)
	return output, nil
}

// DescribeObject executes a DESCRIBE command on a database object
func (c *Client) DescribeObject(ctx context.Context, objectName string) (*DescribeResult, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.process.IsRunning() {
		return nil, fmt.Errorf("sqlcl process is not running")
	}

	if !c.isConnected {
		return nil, fmt.Errorf("not connected to database")
	}

	// Create a context with timeout if the provided context doesn't have a deadline
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, c.config.QueryTimeout)
		defer cancel()
	}

	// Execute the DESCRIBE command
	command := fmt.Sprintf("DESCRIBE %s", objectName)
	result, err := c.executor.ExecuteSQLCLCommand(ctx, command)
	if err != nil {
		return nil, fmt.Errorf("failed to execute DESCRIBE command: %w", err)
	}

	if !result.IsSuccessful {
		// Check for specific error patterns
		if strings.Contains(result.Output, "does not exist") {
			return nil, &types.Error{
				Code:    types.ORA00942, // Table or view does not exist
				Message: fmt.Sprintf("Object %s does not exist", objectName),
				SQL:     command,
			}
		}
		return nil, fmt.Errorf("DESCRIBE command failed: %s", result.Error)
	}

	// Parse the DESCRIBE output
	return parseDescribeOutput(result.Output, objectName)
}

// GetHelp executes a HELP command and returns the help text
func (c *Client) GetHelp(ctx context.Context, topic string) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.process.IsRunning() {
		return "", fmt.Errorf("sqlcl process is not running")
	}

	// Create a context with timeout if the provided context doesn't have a deadline
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, c.config.QueryTimeout)
		defer cancel()
	}

	// Execute the HELP command
	command := "HELP"
	if topic != "" {
		command = fmt.Sprintf("HELP %s", topic)
	}
	result, err := c.executor.ExecuteSQLCLCommand(ctx, command)
	if err != nil {
		return "", fmt.Errorf("failed to execute HELP command: %w", err)
	}

	if !result.IsSuccessful {
		return "", fmt.Errorf("HELP command failed: %s", result.Error)
	}

	// Clean and return the output
	output := strings.TrimSpace(result.Output)
	// Remove the command echo and SQL prompt
	output = strings.TrimPrefix(output, command)
	output = strings.TrimSpace(output)
	return output, nil
}

// GetHistory executes a HISTORY command and returns the command history
func (c *Client) GetHistory(ctx context.Context) ([]string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.process.IsRunning() {
		return nil, fmt.Errorf("sqlcl process is not running")
	}

	// Create a context with timeout if the provided context doesn't have a deadline
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, c.config.QueryTimeout)
		defer cancel()
	}

	// Execute the HISTORY command
	command := "HISTORY"
	result, err := c.executor.ExecuteSQLCLCommand(ctx, command)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HISTORY command: %w", err)
	}

	if !result.IsSuccessful {
		return nil, fmt.Errorf("HISTORY command failed: %s", result.Error)
	}

	// Parse the history output
	return parseHistoryOutput(result.Output), nil
}

// SetPageSize sets the number of rows displayed per page
func (c *Client) SetPageSize(ctx context.Context, size int) error {
	return c.SetCommand(ctx, "PAGESIZE", fmt.Sprintf("%d", size))
}

// SetLineSize sets the line width
func (c *Client) SetLineSize(ctx context.Context, size int) error {
	return c.SetCommand(ctx, "LINESIZE", fmt.Sprintf("%d", size))
}

// SetFeedback enables or disables command feedback
func (c *Client) SetFeedback(ctx context.Context, enabled bool) error {
	value := "ON"
	if !enabled {
		value = "OFF"
	}
	return c.SetCommand(ctx, "FEEDBACK", value)
}

// SetTiming enables or disables timing information
func (c *Client) SetTiming(ctx context.Context, enabled bool) error {
	value := "ON"
	if !enabled {
		value = "OFF"
	}
	return c.SetCommand(ctx, "TIMING", value)
}

// SetSQLTerminator sets the SQL terminator character
func (c *Client) SetSQLTerminator(ctx context.Context, terminator string) error {
	return c.SetCommand(ctx, "SQLTERMINATOR", terminator)
}

// SetHeading enables or disables column headings
func (c *Client) SetHeading(ctx context.Context, enabled bool) error {
	value := "ON"
	if !enabled {
		value = "OFF"
	}
	return c.SetCommand(ctx, "HEADING", value)
}

// SetTNSAdmin sets the TNS_ADMIN directory
func (c *Client) SetTNSAdmin(ctx context.Context, directory string) error {
	return c.SetCommand(ctx, "TNS_ADMIN", directory)
}

// Helper functions to parse output

// parseDescribeOutput parses the output of the DESCRIBE command
func parseDescribeOutput(output string, objectName string) (*DescribeResult, error) {
	lines := strings.Split(output, "\n")
	if len(lines) < 3 {
		return nil, fmt.Errorf("invalid DESCRIBE output format")
	}

	result := &DescribeResult{
		ObjectName: objectName,
		Columns:    []DescribeColumn{},
		Properties: make(map[string]string),
	}

	// Default to TABLE type for output matching table format
	result.ObjectType = "TABLE"

	// Try to determine specific object type
	for _, line := range lines {
		if strings.Contains(line, "TABLE") {
			result.ObjectType = "TABLE"
			break
		} else if strings.Contains(line, "VIEW") {
			result.ObjectType = "VIEW"
			break
		} else if strings.Contains(line, "PROCEDURE") || strings.Contains(line, "FUNCTION") {
			result.ObjectType = "PROCEDURE"
			break
		} else if strings.Contains(line, "PACKAGE") {
			result.ObjectType = "PACKAGE"
			break
		} else if strings.Contains(line, "SEQUENCE") {
			result.ObjectType = "SEQUENCE"
			break
		}
	}

	// Parse columns
	var columnStartIndex int
	for i, line := range lines {
		if strings.Contains(line, "Name") && strings.Contains(line, "Null?") && strings.Contains(line, "Type") {
			columnStartIndex = i + 2 // Skip the header and separator line
			break
		}
	}

	if columnStartIndex > 0 && columnStartIndex < len(lines) {
		for i := columnStartIndex; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if line == "" || strings.Contains(line, "SQL>") {
				break
			}

			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}

			column := DescribeColumn{
				Name:     parts[0],
				Nullable: true,
			}

			// Process type and nullable info
			if len(parts) >= 3 {
				if parts[1] == "NOT" && parts[2] == "NULL" {
					column.Nullable = false
					if len(parts) >= 4 {
						column.Type = strings.Join(parts[3:], " ")
					}
				} else {
					column.Type = strings.Join(parts[1:], " ")
				}
			}

			result.Columns = append(result.Columns, column)
		}
	}

	return result, nil
}

// parseHistoryOutput parses the output of the HISTORY command
func parseHistoryOutput(output string) []string {
	lines := strings.Split(output, "\n")
	history := []string{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "HISTORY") || strings.Contains(line, "SQL>") {
			continue
		}

		// Extract command from history entry (format is usually "1  command")
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) == 2 {
			command := strings.TrimSpace(parts[1])
			if command != "" {
				history = append(history, command)
			}
		}
	}

	return history
}
