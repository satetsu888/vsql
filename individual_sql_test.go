package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestIndividualSQLFiles(t *testing.T) {
	// Build VSQL first
	buildCmd := exec.Command("go", "build", "-o", "vsql", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build VSQL: %v", err)
	}

	// Test categories and their subdirectories
	testCategories := []string{
		"crud",
		"joins",
		"aggregates",
		"grouping",
		"subqueries",
		"null_handling",
		"type_conversion",
		"operators",
		"ordering",
		"error_cases",
		"comments",
		"data_types",
		"functions",
		"type_safety",
	}

	for _, category := range testCategories {
		t.Run(category, func(t *testing.T) {
			testDir := filepath.Join("test", "sql", category)
			
			// Get all SQL files in the directory
			sqlFiles, err := filepath.Glob(filepath.Join(testDir, "*.sql"))
			if err != nil {
				t.Fatalf("Failed to list SQL files in %s: %v", testDir, err)
			}

			// Sort files to ensure consistent order
			for _, sqlFile := range sqlFiles {
				testName := filepath.Base(sqlFile)
				t.Run(testName, func(t *testing.T) {
					runIndividualSQLFile(t, sqlFile)
				})
			}
		})
	}
}

func runIndividualSQLFile(t *testing.T, filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read SQL file %s: %v", filePath, err)
	}

	// Parse the file to extract test metadata
	lines := strings.Split(string(content), "\n")
	var expectedRows int = -1
	var expectError bool = false
	var statusFailing bool = false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "--") {
			comment := strings.TrimPrefix(line, "--")
			comment = strings.TrimSpace(comment)
			
			// Check for expected row count
			if strings.Contains(comment, "Expected:") {
				if strings.Contains(comment, " rows") {
					// Extract number before "rows"
					parts := strings.Fields(comment)
					for i, part := range parts {
						if part == "rows" && i > 0 {
							var num int
							if _, err := fmt.Sscanf(parts[i-1], "%d", &num); err == nil {
								expectedRows = num
							}
						}
					}
				} else if strings.Contains(comment, "no rows") {
					expectedRows = 0
				} else if strings.Contains(comment, "error") || strings.Contains(comment, "fail") {
					expectError = true
				}
			}
			
			// Check for Status: FAILING or FAILING:
			if strings.Contains(comment, "Status: FAILING") || strings.Contains(comment, "FAILING:") {
				statusFailing = true
			}
		}
	}

	// Use the original content without stripping comments
	cmd := exec.Command("./vsql", "-c", string(content), "-q")
	output, err := cmd.CombinedOutput()

	// If test is marked as failing, skip validation
	if statusFailing {
		t.Skipf("Test marked as FAILING")
		return
	}

	// Check error expectation
	if expectError {
		if err == nil {
			t.Errorf("Expected error but query succeeded. Output:\n%s", output)
		}
		return
	}

	if err != nil {
		t.Errorf("Query failed: %v\nOutput:\n%s", err, output)
		return
	}

	// If we have expected row count, verify it
	if expectedRows >= 0 {
		outputStr := string(output)
		actualRows := extractRowCount(outputStr)
		
		if actualRows != expectedRows {
			t.Errorf("Expected %d rows, got %d\nOutput:\n%s", expectedRows, actualRows, outputStr)
		}
	}
}

func extractRowCount(output string) int {
	lines := strings.Split(output, "\n")
	
	// Look for "SELECT N" pattern (vsql's output format)
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "SELECT ") {
			// Extract number from "SELECT N"
			parts := strings.Fields(line)
			if len(parts) == 2 {
				var num int
				if _, err := fmt.Sscanf(parts[1], "%d", &num); err == nil {
					return num
				}
			}
		}
	}
	
	// Fallback: Look for the last "(N rows)" pattern (psql format)
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "(") && strings.HasSuffix(line, " rows)") {
			// Extract number from "(N rows)"
			line = strings.TrimPrefix(line, "(")
			line = strings.TrimSuffix(line, " rows)")
			line = strings.TrimSpace(line)
			
			var num int
			if _, err := fmt.Sscanf(line, "%d", &num); err == nil {
				return num
			}
		}
	}
	
	return -1
}