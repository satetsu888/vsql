package storage

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestConcurrentWrite tests concurrent writes to the same table
func TestConcurrentWrite(t *testing.T) {
	ds := NewDataStore()
	tableName := "test_table"
	
	// Create table
	err := ds.CreateTable(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	
	table, exists := ds.GetTable(tableName)
	if !exists {
		t.Fatal("Table does not exist after creation")
	}
	
	numGoroutines := 100
	numWritesPerGoroutine := 100
	
	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	
	// Track all written IDs to verify no data loss
	writtenIDs := make(map[string]bool)
	var mu sync.Mutex
	
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			
			for j := 0; j < numWritesPerGoroutine; j++ {
				id := fmt.Sprintf("g%d_row%d", goroutineID, j)
				row := Row{
					"id":   id,
					"data": fmt.Sprintf("data_%d_%d", goroutineID, j),
				}
				
				table.Insert(row)
				
				mu.Lock()
				writtenIDs[id] = true
				mu.Unlock()
			}
		}(i)
	}
	
	wg.Wait()
	
	// Verify all rows were written
	rows := table.GetRows()
	if len(rows) != numGoroutines*numWritesPerGoroutine {
		t.Errorf("Expected %d rows, got %d", numGoroutines*numWritesPerGoroutine, len(rows))
	}
	
	// Verify each written ID exists
	rowMap := make(map[string]Row)
	for _, row := range rows {
		if id, ok := row["id"].(string); ok {
			rowMap[id] = row
		}
	}
	
	for id := range writtenIDs {
		if _, exists := rowMap[id]; !exists {
			t.Errorf("Row with ID %s was not found in table", id)
		}
	}
}

// TestConcurrentReadWrite tests concurrent reads and writes
func TestConcurrentReadWrite(t *testing.T) {
	ds := NewDataStore()
	tableName := "test_table"
	
	// Create table with initial data
	err := ds.CreateTable(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	
	table, exists := ds.GetTable(tableName)
	if !exists {
		t.Fatal("Table does not exist after creation")
	}
	
	// Insert initial data
	for i := 0; i < 100; i++ {
		table.Insert(Row{
			"id":   fmt.Sprintf("row_%d", i),
			"value": i,
		})
	}
	
	var wg sync.WaitGroup
	numReaders := 50
	numWriters := 50
	
	// Start readers
	wg.Add(numReaders)
	for i := 0; i < numReaders; i++ {
		go func(readerID int) {
			defer wg.Done()
			
			for j := 0; j < 100; j++ {
				rows := table.GetRows()
				if len(rows) == 0 {
					t.Errorf("Reader %d: no rows found", readerID)
					return
				}
				
				// Read and verify data integrity
				for _, row := range rows {
					if id, ok := row["id"].(string); ok {
						// Just verify the data can be read
						_ = id
					}
				}
			}
		}(i)
	}
	
	// Start writers
	wg.Add(numWriters)
	for i := 0; i < numWriters; i++ {
		go func(writerID int) {
			defer wg.Done()
			
			for j := 0; j < 100; j++ {
				// Insert new rows
				table.Insert(Row{
					"id":   fmt.Sprintf("new_row_%d_%d", writerID, j),
					"data": fmt.Sprintf("data_%d", j),
				})
			}
		}(i)
	}
	
	wg.Wait()
	
	// Final verification
	finalRows := table.GetRows()
	if len(finalRows) < 100 {
		t.Errorf("Lost original rows: expected at least 100, got %d", len(finalRows))
	}
}

// TestRaceCondition tests for race conditions using Go's race detector
func TestRaceCondition(t *testing.T) {
	ds := NewDataStore()
	tableName := "race_test"
	
	ds.CreateTable(tableName)
	table, _ := ds.GetTable(tableName)
	
	// Concurrent operations that might cause races
	var wg sync.WaitGroup
	
	// Writer 1: Inserts
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			table.Insert(Row{
				"id": fmt.Sprintf("row_%d", i),
				"value": i,
			})
		}
	}()
	
	// Reader 1: Full table scan
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			rows := table.GetRows()
			_ = len(rows) // Force read
		}
	}()
	
	// Reader 2: Check table existence
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			_, exists := ds.GetTable(tableName)
			if !exists {
				t.Errorf("Table disappeared during concurrent access")
			}
		}
	}()
	
	wg.Wait()
}

// TestDeadlockPrevention tests that the implementation prevents deadlocks
func TestDeadlockPrevention(t *testing.T) {
	ds := NewDataStore()
	table1 := "table1"
	table2 := "table2"
	
	ds.CreateTable(table1)
	ds.CreateTable(table2)
	
	t1, _ := ds.GetTable(table1)
	t2, _ := ds.GetTable(table2)
	
	// Insert initial data
	for i := 0; i < 10; i++ {
		t1.Insert(Row{"id": i, "data": fmt.Sprintf("data1_%d", i)})
		t2.Insert(Row{"id": i, "data": fmt.Sprintf("data2_%d", i)})
	}
	
	done := make(chan bool, 2)
	
	// Goroutine 1: Access table1 then table2
	go func() {
		for i := 0; i < 100; i++ {
			// Read from table1
			rows1 := t1.GetRows()
			
			// Small delay to increase chance of interleaving
			time.Sleep(time.Microsecond)
			
			// Read from table2
			rows2 := t2.GetRows()
			
			// Write to table1 based on table2
			if len(rows2) > 0 {
				t1.Insert(Row{
					"id": fmt.Sprintf("cross_%d", i),
					"source": "table2",
				})
			}
			
			_ = len(rows1) // Use the variable
		}
		done <- true
	}()
	
	// Goroutine 2: Access table2 then table1 (opposite order)
	go func() {
		for i := 0; i < 100; i++ {
			// Read from table2
			rows2 := t2.GetRows()
			
			// Small delay to increase chance of interleaving
			time.Sleep(time.Microsecond)
			
			// Read from table1
			rows1 := t1.GetRows()
			
			// Write to table2 based on table1
			if len(rows1) > 0 {
				t2.Insert(Row{
					"id": fmt.Sprintf("cross_%d", i),
					"source": "table1",
				})
			}
			
			_ = len(rows2) // Use the variable
		}
		done <- true
	}()
	
	// Wait for both goroutines with timeout
	timeout := time.After(5 * time.Second)
	for i := 0; i < 2; i++ {
		select {
		case <-done:
			// Success
		case <-timeout:
			t.Fatal("Deadlock detected: goroutines did not complete within timeout")
		}
	}
}

// TestConcurrentTableCreation tests concurrent table creation
func TestConcurrentTableCreation(t *testing.T) {
	ds := NewDataStore()
	
	var wg sync.WaitGroup
	numGoroutines := 50
	
	wg.Add(numGoroutines)
	
	errors := make(chan error, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			
			tableName := fmt.Sprintf("table_%d", id%10) // Intentionally create conflicts
			
			// Try to create table
			err := ds.CreateTable(tableName)
			if err != nil {
				errors <- err
				return
			}
			
			// Get table and insert data
			table, exists := ds.GetTable(tableName)
			if !exists {
				errors <- fmt.Errorf("table %s not found after creation", tableName)
				return
			}
			
			// Insert data
			for j := 0; j < 10; j++ {
				table.Insert(Row{
					"id":   fmt.Sprintf("row_%d_%d", id, j),
					"data": fmt.Sprintf("data_%d_%d", id, j),
				})
			}
		}(i)
	}
	
	wg.Wait()
	close(errors)
	
	// Check for errors
	for err := range errors {
		t.Errorf("Error during concurrent table creation: %v", err)
	}
	
	// Verify tables exist
	for i := 0; i < 10; i++ {
		tableName := fmt.Sprintf("table_%d", i)
		table, exists := ds.GetTable(tableName)
		if !exists {
			t.Errorf("Table %s does not exist", tableName)
		} else {
			rows := table.GetRows()
			if len(rows) == 0 {
				t.Errorf("Table %s has no rows", tableName)
			}
		}
	}
}

// TestMemoryVisibility tests that changes are visible across goroutines
func TestMemoryVisibility(t *testing.T) {
	ds := NewDataStore()
	tableName := "visibility_test"
	
	ds.CreateTable(tableName)
	table, _ := ds.GetTable(tableName)
	
	// Channel to synchronize
	written := make(chan bool)
	
	// Writer goroutine
	go func() {
		table.Insert(Row{
			"id":   "test_row",
			"value": 42,
		})
		written <- true
	}()
	
	// Wait for write to complete
	<-written
	
	// Reader goroutine
	readDone := make(chan bool)
	go func() {
		rows := table.GetRows()
		
		found := false
		for _, row := range rows {
			if row["id"] == "test_row" {
				found = true
				if row["value"] != 42 {
					t.Errorf("Expected value 42, got %v", row["value"])
				}
				break
			}
		}
		
		if !found {
			t.Error("Row not found after write")
		}
		
		readDone <- true
	}()
	
	select {
	case <-readDone:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Read operation timed out - possible visibility issue")
	}
}

// TestTableListConcurrency tests concurrent operations on table list
func TestTableListConcurrency(t *testing.T) {
	ds := NewDataStore()
	
	var wg sync.WaitGroup
	
	// Creator goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				tableName := fmt.Sprintf("table_%d_%d", id, j)
				ds.CreateTable(tableName)
			}
		}(i)
	}
	
	// Lister goroutines
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				tables := ds.ListTables()
				// Just verify we can list tables without panic
				_ = len(tables)
			}
		}(i)
	}
	
	// Dropper goroutines
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond) // Let some tables be created first
			for j := 0; j < 25; j++ {
				tableName := fmt.Sprintf("table_%d_%d", id, j)
				ds.DropTable(tableName)
			}
		}(i)
	}
	
	wg.Wait()
}