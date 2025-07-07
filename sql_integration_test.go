package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

var (
	vsqlCmd *exec.Cmd
	db      *sql.DB
)

func TestMain(m *testing.M) {
	// Build VSQL first
	buildCmd := exec.Command("go", "build", "-o", "vsql", ".")
	if err := buildCmd.Run(); err != nil {
		fmt.Printf("Failed to build VSQL: %v\n", err)
		os.Exit(1)
	}

	// Start VSQL server
	vsqlCmd = exec.Command("./vsql", "-port", "5434")
	if err := vsqlCmd.Start(); err != nil {
		fmt.Printf("Failed to start VSQL: %v\n", err)
		os.Exit(1)
	}

	// Wait for server to start
	time.Sleep(2 * time.Second)

	// Connect to database
	var err error
	db, err = sql.Open("postgres", "host=localhost port=5434 user=test dbname=test sslmode=disable")
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		vsqlCmd.Process.Kill()
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	db.Close()
	if vsqlCmd.Process != nil {
		vsqlCmd.Process.Kill()
		vsqlCmd.Wait()
	}
	
	os.Exit(code)
}

func TestNullHandling(t *testing.T) {
	// Run null_handling_test.sql
	runSQLFile(t, "test/null_handling_test.sql")
}

func TestTypeComparison(t *testing.T) {
	// Run type_comparison_test.sql
	runSQLFile(t, "test/type_comparison_test.sql")
}

func TestComplexQueries(t *testing.T) {
	// Run complex_queries_test.sql if it exists
	if _, err := os.Stat("test/complex_queries_test.sql"); err == nil {
		runSQLFile(t, "test/complex_queries_test.sql")
	}
}

func TestErrorHandling(t *testing.T) {
	// Run error_handling_test.sql if it exists
	if _, err := os.Stat("test/error_handling_test.sql"); err == nil {
		runSQLFile(t, "test/error_handling_test.sql")
	}
}

func runSQLFile(t *testing.T, filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read SQL file %s: %v", filePath, err)
	}

	// Split content into individual statements
	statements := splitSQLStatements(string(content))
	
	// Track test results
	testName := ""
	expectedRows := -1
	expectError := false
	expectNoRows := false

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		// Parse test expectations from comments
		if strings.HasPrefix(stmt, "--") {
			comment := strings.TrimPrefix(stmt, "--")
			comment = strings.TrimSpace(comment)
			
			if strings.HasPrefix(comment, "Test ") {
				testName = comment
			} else if strings.Contains(comment, "Expected:") {
				if strings.Contains(comment, "0 rows") || strings.Contains(comment, "no rows") {
					expectNoRows = true
					expectedRows = 0
				} else if strings.Contains(comment, " rows") {
					// Extract number of expected rows
					parts := strings.Split(comment, " ")
					for i, part := range parts {
						if part == "rows" && i > 0 {
							if num, err := parseNumber(parts[i-1]); err == nil {
								expectedRows = num
							}
						}
					}
				} else if strings.Contains(comment, "error") || strings.Contains(comment, "fail") {
					expectError = true
				}
			}
			continue
		}

		// Skip setup/teardown for individual test tracking
		if strings.HasPrefix(strings.ToUpper(stmt), "CREATE TABLE") ||
		   strings.HasPrefix(strings.ToUpper(stmt), "INSERT INTO") ||
		   strings.HasPrefix(strings.ToUpper(stmt), "DROP TABLE") {
			// Execute but don't track as a test
			_, err := db.Exec(stmt)
			if err != nil && !strings.Contains(err.Error(), "does not exist") {
				t.Logf("Setup/teardown query failed: %v", err)
			}
			continue
		}

		// Execute test query
		if testName != "" {
			t.Run(testName, func(t *testing.T) {
				rows, err := db.Query(stmt)
				
				// Check error expectation
				if expectError {
					if err == nil {
						t.Errorf("Expected error but query succeeded")
					}
					return
				}
				
				if err != nil {
					t.Errorf("Query failed: %v", err)
					return
				}
				defer rows.Close()

				// Count actual rows
				actualRows := 0
				for rows.Next() {
					actualRows++
				}

				// Verify row count if expected
				if expectedRows >= 0 {
					if actualRows != expectedRows {
						t.Errorf("Expected %d rows, got %d", expectedRows, actualRows)
					}
				} else if expectNoRows && actualRows > 0 {
					t.Errorf("Expected no rows, got %d", actualRows)
				}
			})

			// Reset for next test
			testName = ""
			expectedRows = -1
			expectError = false
			expectNoRows = false
		}
	}
}

func splitSQLStatements(content string) []string {
	lines := strings.Split(content, "\n")
	statements := []string{}
	current := []string{}
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Comment or empty line ends current statement
		if (trimmed == "" || strings.HasPrefix(trimmed, "--")) && len(current) > 0 {
			statements = append(statements, strings.Join(current, "\n"))
			current = []string{}
		}
		
		// Add line to statements
		if trimmed != "" {
			if strings.HasPrefix(trimmed, "--") {
				// Add comment as its own statement
				statements = append(statements, trimmed)
			} else {
				current = append(current, line)
			}
		}
	}
	
	// Add last statement if any
	if len(current) > 0 {
		statements = append(statements, strings.Join(current, "\n"))
	}
	
	return statements
}

func parseNumber(s string) (int, error) {
	s = strings.TrimSpace(s)
	var num int
	_, err := fmt.Sscanf(s, "%d", &num)
	return num, err
}