// This example demonstrates how to use the config package to customize the SQLcl client
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/zodimo/go-sqlcl/pkg/config"
	"github.com/zodimo/go-sqlcl/pkg/sqlcl"
)

func main() {
	// Create a custom configuration using functional options
	clientConfig := config.New(
		config.WithSQLclPath("/usr/local/bin/sql"), // Custom SQLcl path
		config.WithTimeout(45*time.Second),         // Custom timeout
		config.WithConnectTimeout(15*time.Second),  // Custom connect timeout
		config.WithQueryTimeout(90*time.Second),    // Custom query timeout
		config.WithColorOutput(true),               // Enable color output
		config.WithOutputFormat("json"),            // Set output format to JSON
		config.WithMaxBufferSize(2*1024*1024),      // 2MB buffer size
	)

	// Create a new client with the custom configuration
	client, err := sqlcl.NewClient(clientConfig)
	if err != nil {
		log.Fatalf("Failed to create SQL client: %v", err)
	}
	defer client.Close()

	fmt.Println("Created SQLcl client with custom configuration")

	// Create connection options
	connOpts := config.NewConnection(
		config.WithUsername("scott"),
		config.WithPassword("tiger"),
		config.WithConnectString("localhost:1521/orclpdb"),
		config.WithRole("SYSDBA"),
	)

	// Connect to the database
	ctx := context.Background()
	err = client.ConnectWithOptions(ctx, *connOpts)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Connected to database")

	// Execute a query
	result, err := client.ExecuteSQL(ctx, "SELECT * FROM emp WHERE rownum <= 5")
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}

	// Print results
	fmt.Printf("Query returned %d rows\n", len(result.Rows))
	for i, row := range result.Rows {
		fmt.Printf("Row %d: %v\n", i+1, row.Values)
	}
}
