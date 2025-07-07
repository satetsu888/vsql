package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	// Build VSQL first
	buildCmd := exec.Command("go", "build", "-o", "vsql", ".")
	if err := buildCmd.Run(); err != nil {
		fmt.Printf("Failed to build VSQL: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()
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

func TestAdvancedQueries(t *testing.T) {
	// Run advanced_queries_test.sql
	runSQLFile(t, "test/advanced_queries_test.sql")
}

func TestBasicAdvanced(t *testing.T) {
	// Run basic_advanced_test.sql
	runSQLFile(t, "test/basic_advanced_test.sql")
}

func TestNullComparisons(t *testing.T) {
	// Run null_comparisons_test.sql
	runSQLFile(t, "test/null_comparisons_test.sql")
}

func TestBasicIntegration(t *testing.T) {
	// Run basic_integration_test.sql
	runSQLFile(t, "test/basic_integration_test.sql")
}

func TestEnhancedIntegration(t *testing.T) {
	// Run enhanced_integration_test.sql
	runSQLFile(t, "test/enhanced_integration_test.sql")
}

func runSQLFile(t *testing.T, filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read SQL file %s: %v", filePath, err)
	}

	// First, collect all setup queries (CREATE TABLE, INSERT) 
	statements := splitSQLStatements(string(content))
	setupQueries := []string{}
	cleanupQueries := []string{}
	
	for _, stmt := range statements {
		trimmed := strings.TrimSpace(stmt)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		
		upperStmt := strings.ToUpper(trimmed)
		if strings.HasPrefix(upperStmt, "CREATE TABLE") || strings.HasPrefix(upperStmt, "INSERT INTO") {
			setupQueries = append(setupQueries, trimmed)
		} else if strings.HasPrefix(upperStmt, "DROP TABLE") {
			cleanupQueries = append(cleanupQueries, trimmed)
		}
	}
	
	// Execute all setup queries in one command
	if len(setupQueries) > 0 {
		setupCmd := strings.Join(setupQueries, "; ")
		cmd := exec.Command("./vsql", "-c", setupCmd)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Setup failed: %v, output: %s", err, output)
		}
	}
	
	// Now run individual test queries
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

		// Skip setup/teardown statements
		upperStmt := strings.ToUpper(stmt)
		if strings.HasPrefix(upperStmt, "CREATE TABLE") ||
		   strings.HasPrefix(upperStmt, "INSERT INTO") ||
		   strings.HasPrefix(upperStmt, "DROP TABLE") {
			continue
		}

		// Execute test query
		if testName != "" {
			t.Run(testName, func(t *testing.T) {
				// Combine setup + test query
				fullCmd := strings.Join(append(setupQueries, stmt), "; ")
				cmd := exec.Command("./vsql", "-c", fullCmd)
				output, err := cmd.CombinedOutput()
				
				// Check error expectation
				if expectError {
					if err == nil {
						t.Errorf("Expected error but query succeeded")
					}
					return
				}
				
				if err != nil {
					t.Errorf("Query failed: %v, output: %s", err, output)
					return
				}

				// Parse output to find the result of the test query
				outputStr := string(output)
				outputs := strings.Split(outputStr, "\n")
				
				// Find the last SELECT result in the output
				lastSelectStart := -1
				for i := len(outputs) - 1; i >= 0; i-- {
					if strings.Contains(outputs[i], "----") {
						// Found separator line, the SELECT result starts before this
						for j := i - 1; j >= 0; j-- {
							if outputs[j] != "" && !strings.HasPrefix(outputs[j], "INSERT") && 
							   !strings.HasPrefix(outputs[j], "CREATE") && !strings.HasPrefix(outputs[j], "DROP") {
								lastSelectStart = j
								break
							}
						}
						if lastSelectStart >= 0 {
							break
						}
					}
				}
				
				actualRows := 0
				
				// Look for "(N rows)" pattern after the last SELECT
				for i := len(outputs) - 1; i >= 0; i-- {
					line := outputs[i]
					if strings.HasPrefix(line, "(") && strings.HasSuffix(line, " rows)") {
						// Extract number from "(N rows)"
						line = strings.TrimPrefix(line, "(")
						line = strings.TrimSuffix(line, " rows)")
						if num, err := parseNumber(line); err == nil {
							actualRows = num
							break
						}
					}
				}

				// Verify row count if expected
				if expectedRows >= 0 {
					if actualRows != expectedRows {
						t.Errorf("Expected %d rows, got %d. Output:\n%s", expectedRows, actualRows, outputStr)
					}
				} else if expectNoRows && actualRows > 0 {
					t.Errorf("Expected no rows, got %d. Output:\n%s", actualRows, outputStr)
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