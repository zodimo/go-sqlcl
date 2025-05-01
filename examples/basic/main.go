// Basic Example - go-sqlcl
//
// This example demonstrates how to use the go-sqlcl package to:
// 1. Create a new client with custom configuration
// 2. Connect to a database
// 3. Execute a simple query
// 4. Process and display the query results
// 5. Handle errors at each step
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/zodimo/go-sqlcl/pkg/sqlcl"
	"github.com/zodimo/go-sqlcl/pkg/types"
)

func main() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting go-sqlcl basic example")

	// 1. Create a client with custom configuration
	// Create a configuration with custom settings
	config := &types.ClientConfig{
		SQLclPath:      "sql",            // Path to SQLcl executable (assuming it's in PATH)
		Timeout:        30 * time.Second, // Default operation timeout
		ConnectTimeout: 15 * time.Second, // Connection timeout
		QueryTimeout:   60 * time.Second, // Query execution timeout
		ColorOutput:    false,            // Disable color output for easier parsing
		Format:         "table",          // Default output format
		LogLevel:       "info",           // Log level
	}

	// Create a new client with the configuration
	client, err := sqlcl.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close() // Ensure resources are cleaned up when done

	// 2. Connect to a database
	// You can use either a simple connection string or detailed options

	// Option 1: Simple connection string (username/password@host:port/service)
	// ctx := context.Background()
	// err = client.Connect(ctx, "username/password@localhost:1521/XEPDB1")

	// Option 2: Detailed connection options (recommended for production use)
	ctx := context.Background()
	connectionOpts := types.ConnectionOptions{
		Username:   os.Getenv("DB_USER"),     // Get from environment variable
		Password:   os.Getenv("DB_PASSWORD"), // Get from environment variable
		ConnectStr: os.Getenv("DB_CONNECT"),  // Could be like "localhost:1521/XEPDB1"
		// Optional: Add TNS_ADMIN directory or Wallet for secure connections
		// TNSAdmin:   "/path/to/tnsadmin",
		// Wallet:     "/path/to/wallet",
		// Role:       "SYSDBA", // Uncomment to connect with SYSDBA role
	}

	// Check if environment variables are set
	if connectionOpts.Username == "" || connectionOpts.Password == "" || connectionOpts.ConnectStr == "" {
		log.Println("Environment variables not set, using demo mode with fake connection")
		fmt.Println("To run with a real database, set these environment variables:")
		fmt.Println("  DB_USER - Database username")
		fmt.Println("  DB_PASSWORD - Database password")
		fmt.Println("  DB_CONNECT - Connection string (host:port/service)")

		// This is just to demonstrate the code structure
		// In a real scenario, we would exit here
		fmt.Println("\n--- Demo Mode (no actual database connection) ---")
	} else {
		// Connect with the provided options
		err = client.ConnectWithOptions(ctx, connectionOpts)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		log.Println("Successfully connected to database")
	}

	// 3. Execute a simple query
	// Add a timeout to the query context
	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Execute a simple SQL query
	sqlQuery := "SELECT user, sysdate FROM dual"
	result, err := client.ExecuteSQL(queryCtx, sqlQuery)

	// 4. Handle errors and process results
	if err != nil {
		// Check for specific Oracle errors
		if oraErr, ok := err.(*types.Error); ok {
			switch oraErr.Code {
			case types.ORA01017:
				log.Fatalf("Authentication failed: %v", err)
			case types.ORA12514:
				log.Fatalf("Database service not found: %v", err)
			default:
				log.Fatalf("Database error: %v", err)
			}
		} else {
			log.Fatalf("Failed to execute query: %v", err)
		}
	}

	// 5. Process and display the results
	if result != nil {
		fmt.Println("Query executed successfully!")
		fmt.Println("Summary:", result.Summary)

		// Print the column headers
		for i, col := range result.Columns {
			if i > 0 {
				fmt.Print(" | ")
			}
			fmt.Print(col.Name)
		}
		fmt.Println("\n" + strings.Repeat("-", 30))

		// Print each row
		for _, row := range result.Rows {
			for i, val := range row.Values {
				if i > 0 {
					fmt.Print(" | ")
				}
				fmt.Print(val)
			}
			fmt.Println()
		}
	} else {
		fmt.Println("No results returned or query didn't produce a result set")
	}

	// 6. Execute a SQLCL-specific command (as an additional example)
	fmt.Println("\n--- Executing a SQLCL-specific command ---")
	cmd := types.Command{
		SQL:           "SHOW USER",
		ReturnResults: true,
	}

	cmdResult, err := client.ExecuteCommand(ctx, cmd)
	if err != nil {
		log.Printf("Failed to execute SQLCL command: %v", err)
	} else if cmdResult != nil {
		fmt.Println(cmdResult.Summary)
	}

	fmt.Println("\nBasic example completed successfully")
}
