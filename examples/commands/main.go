package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/zodimo/go-sqlcl/pkg/sqlcl"
	"github.com/zodimo/go-sqlcl/pkg/types"
)

func main() {
	// Get connection credentials from environment variables or use demo mode
	username := os.Getenv("ORACLE_USERNAME")
	password := os.Getenv("ORACLE_PASSWORD")
	connectString := os.Getenv("ORACLE_CONNECT_STRING")
	demoMode := false

	if username == "" || password == "" || connectString == "" {
		fmt.Println("Running in demo mode (no actual database connection)")
		fmt.Println("To connect to a real database, set ORACLE_USERNAME, ORACLE_PASSWORD, and ORACLE_CONNECT_STRING environment variables")
		demoMode = true
	}

	// Create a client configuration
	config := &types.ClientConfig{
		SQLclPath:      "sql", // Adjust path to your SQLcl executable
		Timeout:        30 * time.Second,
		QueryTimeout:   15 * time.Second,
		ConnectTimeout: 20 * time.Second,
	}

	// Create a new client
	client, err := sqlcl.NewClient(config)
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to the database (only in real mode)
	if !demoMode {
		fmt.Println("Connecting to the database...")
		connOpts := types.ConnectionOptions{
			Username:   username,
			Password:   password,
			ConnectStr: connectString,
		}
		if err := client.ConnectWithOptions(ctx, connOpts); err != nil {
			fmt.Printf("Error connecting to database: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Connected successfully!")
	} else {
		fmt.Println("Demo mode: Skipping database connection")
		fmt.Println("The following examples show how the command wrappers would be used with a real connection.")
		return
	}

	// Example 1: SET commands
	fmt.Println("\n=== Example 1: SET commands ===")

	// Set pagesize to 50
	if err := client.SetPageSize(ctx, 50); err != nil {
		fmt.Printf("Error setting pagesize: %v\n", err)
	} else {
		fmt.Println("Pagesize set to 50")
	}

	// Set linesize to 120
	if err := client.SetLineSize(ctx, 120); err != nil {
		fmt.Printf("Error setting linesize: %v\n", err)
	} else {
		fmt.Println("Linesize set to 120")
	}

	// Enable timing
	if err := client.SetTiming(ctx, true); err != nil {
		fmt.Printf("Error enabling timing: %v\n", err)
	} else {
		fmt.Println("Timing enabled")
	}

	// Example 2: SHOW command
	fmt.Println("\n=== Example 2: SHOW command ===")

	// Show the current pagesize
	pagesize, err := client.ShowCommand(ctx, "PAGESIZE")
	if err != nil {
		fmt.Printf("Error showing pagesize: %v\n", err)
	} else {
		fmt.Printf("Current pagesize: %s\n", pagesize)
	}

	// Show the current linesize
	linesize, err := client.ShowCommand(ctx, "LINESIZE")
	if err != nil {
		fmt.Printf("Error showing linesize: %v\n", err)
	} else {
		fmt.Printf("Current linesize: %s\n", linesize)
	}

	// Example 3: DESCRIBE command
	fmt.Println("\n=== Example 3: DESCRIBE command ===")

	// Describe a table
	tableName := "EMPLOYEES" // Adjust to a table that exists in your database
	fmt.Printf("Describing table %s...\n", tableName)
	descResult, err := client.DescribeObject(ctx, tableName)
	if err != nil {
		fmt.Printf("Error describing table: %v\n", err)
	} else {
		fmt.Printf("Object Name: %s\n", descResult.ObjectName)
		fmt.Printf("Object Type: %s\n", descResult.ObjectType)
		fmt.Println("Columns:")
		for _, col := range descResult.Columns {
			nullable := "NULL"
			if !col.Nullable {
				nullable = "NOT NULL"
			}
			fmt.Printf("  - %s (%s) %s\n", col.Name, col.Type, nullable)
		}
	}

	// Example 4: HELP command
	fmt.Println("\n=== Example 4: HELP command ===")

	// Get help on SELECT
	helpText, err := client.GetHelp(ctx, "SELECT")
	if err != nil {
		fmt.Printf("Error getting help: %v\n", err)
	} else {
		fmt.Println("Help for SELECT:")
		// Print just the first few lines to avoid flooding the output
		lines := strings.Split(helpText, "\n")
		displayLines := 5
		if len(lines) < displayLines {
			displayLines = len(lines)
		}
		for i := 0; i < displayLines; i++ {
			fmt.Println(lines[i])
		}
		if len(lines) > displayLines {
			fmt.Println("...")
		}
	}

	// Example 5: HISTORY command
	fmt.Println("\n=== Example 5: HISTORY command ===")

	// Get command history
	history, err := client.GetHistory(ctx)
	if err != nil {
		fmt.Printf("Error getting history: %v\n", err)
	} else {
		fmt.Println("Command history:")
		displayCount := 5
		if len(history) < displayCount {
			displayCount = len(history)
		}
		for i := 0; i < displayCount; i++ {
			fmt.Printf("  %d. %s\n", i+1, history[i])
		}
		if len(history) > displayCount {
			fmt.Println("...")
		}
	}

	fmt.Println("\nExample completed successfully!")
}
