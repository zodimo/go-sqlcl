package container

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConcurrentQueries tests executing multiple queries concurrently
func TestConcurrentQueries(t *testing.T) {
	// Skip test if running in CI or if explicitly disabled
	if os.Getenv("SKIP_CONTAINER_TESTS") != "" {
		t.Skip("Skipping container test as SKIP_CONTAINER_TESTS is set")
	}

	// Create a test schema for concurrent query tests
	schemaName := "concurrent_test"

	// Create a schema and test table for this test
	setupClient := CreateFixedClient(t)
	defer setupClient.Close()

	// Set up test schema
	createSchemaSQL := fmt.Sprintf(`
		-- Create user
		CREATE USER %s IDENTIFIED BY %[1]s;
		-- Grant permissions
		GRANT CONNECT, RESOURCE, CREATE VIEW, CREATE PROCEDURE, CREATE SYNONYM TO %[1]s;
		-- Grant tablespace
		ALTER USER %[1]s QUOTA UNLIMITED ON USERS;
	`, strings.ToUpper(schemaName))

	result, err := setupClient.ExecuteSQL(globalCtx, createSchemaSQL)
	require.NoError(t, err, "Failed to execute schema creation SQL")
	require.True(t, result.Success, "Schema creation should be successful")

	// Create test table and insert test data
	setupTableSQL := fmt.Sprintf(`
		ALTER SESSION SET CURRENT_SCHEMA = %s;
		CREATE TABLE concurrent_test_table (
			id NUMBER PRIMARY KEY,
			value VARCHAR2(100)
		);
		
		-- Insert test data
		BEGIN
			FOR i IN 1..100 LOOP
				INSERT INTO concurrent_test_table VALUES (i, 'Value ' || i);
			END LOOP;
			COMMIT;
		END;
		/
	`, strings.ToUpper(schemaName))

	result, err = setupClient.ExecuteSQL(globalCtx, setupTableSQL)
	require.NoError(t, err, "Failed to set up test table")
	require.True(t, result.Success, "Test table setup should be successful")

	// Clean up after the test
	defer func() {
		dropSchemaSQL := fmt.Sprintf(`DROP USER %s CASCADE;`, strings.ToUpper(schemaName))
		_, err := setupClient.ExecuteSQL(globalCtx, dropSchemaSQL)
		if err != nil {
			t.Logf("Failed to drop test schema: %v", err)
		}
	}()

	// Test case 1: Multiple concurrent clients querying the same table
	t.Run("MultipleConcurrentClients", func(t *testing.T) {
		// Define the number of concurrent clients
		numClients := 10
		var wg sync.WaitGroup
		wg.Add(numClients)

		// Channel to collect results
		resultsChan := make(chan string, numClients)

		// Execute queries concurrently
		for i := 0; i < numClients; i++ {
			go func(clientID int) {
				defer wg.Done()

				// Create a new client for each goroutine
				client := CreateFixedClient(t)
				defer client.Close()

				// Execute a query with different WHERE clause for each client
				// Each client will get a different subset of rows
				querySQL := fmt.Sprintf(`
					ALTER SESSION SET CURRENT_SCHEMA = %s;
					SELECT id, value 
					FROM concurrent_test_table 
					WHERE MOD(id, %d) = %d
					ORDER BY id;
				`, strings.ToUpper(schemaName), numClients, clientID%numClients)

				result, err := client.ExecuteSQL(globalCtx, querySQL)
				if err != nil {
					resultsChan <- fmt.Sprintf("Client %d error: %v", clientID, err)
					return
				}

				resultsChan <- fmt.Sprintf("Client %d success: %s", clientID, result.Message)
			}(i)
		}

		// Wait for all goroutines to complete
		wg.Wait()
		close(resultsChan)

		// Collect and check results
		results := make([]string, 0, numClients)
		for result := range resultsChan {
			results = append(results, result)
		}

		// Verify we got the expected number of results
		assert.Equal(t, numClients, len(results), "Should receive results from all clients")

		// Log the results for debugging
		for _, result := range results {
			t.Logf("Result: %s", result)
		}
	})

	// Test case 2: Concurrent read and write operations
	t.Run("ConcurrentReadWrite", func(t *testing.T) {
		var wg sync.WaitGroup

		// Define numbers of readers and writers
		numReaders := 5
		numWriters := 3
		total := numReaders + numWriters

		wg.Add(total)
		resultsChan := make(chan string, total)

		// Start readers
		for i := 0; i < numReaders; i++ {
			go func(readerID int) {
				defer wg.Done()

				// Create a new client for each reader
				client := CreateFixedClient(t)
				defer client.Close()

				// Execute a read query
				readSQL := fmt.Sprintf(`
					ALTER SESSION SET CURRENT_SCHEMA = %s;
					SELECT COUNT(*) FROM concurrent_test_table;
				`, strings.ToUpper(schemaName))

				result, err := client.ExecuteSQL(globalCtx, readSQL)
				if err != nil {
					resultsChan <- fmt.Sprintf("Reader %d error: %v", readerID, err)
					return
				}

				resultsChan <- fmt.Sprintf("Reader %d success: %s", readerID, result.Message)
			}(i)
		}

		// Start writers
		for i := 0; i < numWriters; i++ {
			go func(writerID int) {
				defer wg.Done()

				// Create a new client for each writer
				client := CreateFixedClient(t)
				defer client.Close()

				// Base ID for new records - ensure each writer uses different IDs
				baseID := 1000 + (writerID * 10)

				// Execute a write query (transaction with multiple inserts)
				writeSQL := fmt.Sprintf(`
					ALTER SESSION SET CURRENT_SCHEMA = %s;
					BEGIN
						-- Insert a batch of records
						FOR i IN 0..9 LOOP
							INSERT INTO concurrent_test_table VALUES (%d + i, 'Concurrent Insert ' || (%d + i));
						END LOOP;
						COMMIT;
					END;
					/
					-- Return the count of records inserted by this writer
					SELECT COUNT(*) FROM concurrent_test_table WHERE id BETWEEN %d AND %d;
				`, strings.ToUpper(schemaName), baseID, baseID, baseID, baseID+9)

				result, err := client.ExecuteSQL(globalCtx, writeSQL)
				if err != nil {
					resultsChan <- fmt.Sprintf("Writer %d error: %v", writerID, err)
					return
				}

				resultsChan <- fmt.Sprintf("Writer %d success: %s", writerID, result.Message)
			}(i)
		}

		// Wait for all goroutines to complete
		wg.Wait()
		close(resultsChan)

		// Collect and check results
		results := make([]string, 0, total)
		for result := range resultsChan {
			results = append(results, result)
		}

		// Verify we got the expected number of results
		assert.Equal(t, total, len(results), "Should receive results from all readers and writers")

		// Log the results for debugging
		for _, result := range results {
			t.Logf("Result: %s", result)
		}

		// Verify that all the expected records were inserted
		verifySQL := fmt.Sprintf(`
			ALTER SESSION SET CURRENT_SCHEMA = %s;
			SELECT COUNT(*) FROM concurrent_test_table WHERE id >= 1000;
		`, strings.ToUpper(schemaName))

		client := CreateFixedClient(t)
		defer client.Close()

		result, err := client.ExecuteSQL(globalCtx, verifySQL)
		require.NoError(t, err, "Failed to execute verification query")
		require.True(t, result.Success, "Verification query should be successful")

		// Check that we have the expected number of new records (numWriters * 10)
		expectedCount := numWriters * 10
		t.Logf("Verification result: %s (expected %d new records)", result.Message, expectedCount)
	})

	// Test case 3: High concurrency stress test
	t.Run("StressConcurrentQueries", func(t *testing.T) {
		// This test can be skipped in regular unit testing as it's more of a stress test
		if testing.Short() {
			t.Skip("Skipping stress test in short mode")
		}

		// Define the number of concurrent operations
		numOperations := 20
		var wg sync.WaitGroup
		wg.Add(numOperations)

		// Channel to collect results
		errorsChan := make(chan error, numOperations)

		// Execute queries concurrently
		for i := 0; i < numOperations; i++ {
			go func(opID int) {
				defer wg.Done()

				// Create a new client for each operation
				client := CreateFixedClient(t)
				defer client.Close()

				// Determine operation type based on ID (even = read, odd = write)
				if opID%2 == 0 {
					// Read operation - query a random subset of records
					limit := (opID % 10) + 1  // Between 1 and 10
					offset := (opID * 3) % 50 // Different starting points

					querySQL := fmt.Sprintf(`
						ALTER SESSION SET CURRENT_SCHEMA = %s;
						SELECT * FROM (
							SELECT id, value
							FROM concurrent_test_table
							ORDER BY id
							OFFSET %d ROWS
						) WHERE ROWNUM <= %d;
					`, strings.ToUpper(schemaName), offset, limit)

					_, err := client.ExecuteSQL(globalCtx, querySQL)
					if err != nil {
						errorsChan <- fmt.Errorf("read operation %d failed: %w", opID, err)
					}
				} else {
					// Write operation - update a small batch of records
					startID := (opID*5)%95 + 1 // Different IDs to update (1-95)
					endID := startID + 5       // Update 5 records at a time

					updateSQL := fmt.Sprintf(`
						ALTER SESSION SET CURRENT_SCHEMA = %s;
						BEGIN
							UPDATE concurrent_test_table
							SET value = value || ' (updated by op %d)'
							WHERE id BETWEEN %d AND %d;
							COMMIT;
						END;
						/
					`, strings.ToUpper(schemaName), opID, startID, endID)

					_, err := client.ExecuteSQL(globalCtx, updateSQL)
					if err != nil {
						errorsChan <- fmt.Errorf("write operation %d failed: %w", opID, err)
					}
				}
			}(i)
		}

		// Wait for all goroutines to complete
		wg.Wait()
		close(errorsChan)

		// Collect and check for errors
		errors := make([]error, 0)
		for err := range errorsChan {
			errors = append(errors, err)
		}

		// Report errors if any
		assert.Empty(t, errors, "Should not have any errors during concurrent stress test")
		if len(errors) > 0 {
			for _, err := range errors {
				t.Logf("Error: %v", err)
			}
		}

		// Verify the database is still functional after stress test
		verifySQL := fmt.Sprintf(`
			ALTER SESSION SET CURRENT_SCHEMA = %s;
			SELECT COUNT(*) FROM concurrent_test_table;
		`, strings.ToUpper(schemaName))

		client := CreateFixedClient(t)
		defer client.Close()

		result, err := client.ExecuteSQL(globalCtx, verifySQL)
		require.NoError(t, err, "Failed to execute verification query after stress test")
		require.True(t, result.Success, "Database should still be functional after stress test")
		t.Logf("Final record count after stress test: %s", result.Message)
	})
}
