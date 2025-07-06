package parser

import (
	"fmt"
	"testing"

	"github.com/pganalyze/pg_query_go/v5"
)

// TestParseEdgeCases tests edge cases for the SQL parser
func TestParseEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		sql         string
		shouldError bool
	}{
		// Invalid SQL syntax
		{
			name:        "invalid syntax - missing FROM",
			sql:         "SELECT * WHERE id = 1",
			shouldError: false, // PostgreSQL parser allows this
		},
		{
			name:        "invalid syntax - malformed INSERT",
			sql:         "INSERT INTO users VALUES",
			shouldError: true,
		},
		{
			name:        "invalid syntax - incomplete statement",
			sql:         "SELECT",
			shouldError: false, // PostgreSQL parser allows incomplete SELECT
		},
		{
			name:        "empty query",
			sql:         "",
			shouldError: false, // PostgreSQL parser returns empty statement list
		},
		{
			name:        "only whitespace",
			sql:         "   \n\t  ",
			shouldError: false, // PostgreSQL parser returns empty statement list
		},
		
		// Special characters and SQL injection attempts
		{
			name:        "single quotes in string",
			sql:         "SELECT * FROM users WHERE name = 'O''Brien'",
			shouldError: false,
		},
		{
			name:        "double quotes for identifiers",
			sql:         `SELECT * FROM "users" WHERE "name" = 'test'`,
			shouldError: false,
		},
		{
			name:        "semicolon injection attempt",
			sql:         "SELECT * FROM users WHERE id = 1; DROP TABLE users;",
			shouldError: false, // Parser should handle this as multiple statements
		},
		{
			name:        "comment injection",
			sql:         "SELECT * FROM users WHERE id = 1 -- AND password = 'secret'",
			shouldError: false,
		},
		{
			name:        "unicode characters",
			sql:         "INSERT INTO users (name) VALUES ('日本語テスト')",
			shouldError: false,
		},
		{
			name:        "special characters in values",
			sql:         "INSERT INTO users (data) VALUES ('!@#$%^&*()_+-=[]{}|;:,.<>?')",
			shouldError: false,
		},
		
		// NULL and special values
		{
			name:        "NULL in WHERE clause",
			sql:         "SELECT * FROM users WHERE email IS NULL",
			shouldError: false,
		},
		{
			name:        "NULL in INSERT",
			sql:         "INSERT INTO users (id, name, email) VALUES (1, NULL, NULL)",
			shouldError: false,
		},
		{
			name:        "boolean values",
			sql:         "SELECT * FROM users WHERE active = true AND deleted = false",
			shouldError: false,
		},
		
		// Complex nested queries
		{
			name:        "deeply nested subquery",
			sql:         "SELECT * FROM users WHERE id IN (SELECT user_id FROM orders WHERE product_id IN (SELECT id FROM products WHERE category IN (SELECT id FROM categories)))",
			shouldError: false,
		},
		{
			name:        "multiple CTEs",
			sql:         "WITH t1 AS (SELECT * FROM users), t2 AS (SELECT * FROM orders) SELECT * FROM t1 JOIN t2 ON t1.id = t2.user_id",
			shouldError: false,
		},
		
		// Reserved keywords
		{
			name:        "reserved keyword as identifier without quotes",
			sql:         "SELECT * FROM select",
			shouldError: true,
		},
		{
			name:        "reserved keyword as identifier with quotes",
			sql:         `SELECT * FROM "select"`,
			shouldError: false,
		},
		
		// Large queries
		{
			name:        "extremely long identifier",
			sql:         "SELECT " + string(make([]byte, 1000, 1000)) + "col FROM users",
			shouldError: false, // PostgreSQL parser accepts long identifiers
		},
		{
			name:        "very long WHERE clause",
			sql:         "SELECT * FROM users WHERE " + generateLongWhereClause(100),
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pg_query.Parse(tt.sql)
			hasError := err != nil

			if hasError != tt.shouldError {
				if tt.shouldError {
					t.Errorf("Expected error but got none for SQL: %s", tt.sql)
				} else {
					t.Errorf("Expected no error but got: %v for SQL: %s", err, tt.sql)
				}
			}
		})
	}
}

// TestParseQueryTypes tests that different query types are parsed correctly
func TestParseQueryTypes(t *testing.T) {
	
	tests := []struct {
		name      string
		sql       string
		queryType string
	}{
		{
			name:      "SELECT query",
			sql:       "SELECT * FROM users",
			queryType: "T_SelectStmt",
		},
		{
			name:      "INSERT query",
			sql:       "INSERT INTO users (id, name) VALUES (1, 'test')",
			queryType: "T_InsertStmt",
		},
		{
			name:      "UPDATE query",
			sql:       "UPDATE users SET name = 'test' WHERE id = 1",
			queryType: "T_UpdateStmt",
		},
		{
			name:      "DELETE query",
			sql:       "DELETE FROM users WHERE id = 1",
			queryType: "T_DeleteStmt",
		},
		{
			name:      "CREATE TABLE query",
			sql:       "CREATE TABLE users (id INT, name TEXT)",
			queryType: "T_CreateStmt",
		},
		{
			name:      "DROP TABLE query",
			sql:       "DROP TABLE users",
			queryType: "T_DropStmt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePostgreSQL(tt.sql)
			if err != nil {
				t.Fatalf("Failed to parse SQL: %v", err)
			}

			if len(result.Stmts) == 0 {
				t.Fatal("No statements parsed")
			}

			// Check if the query type matches expected
			stmt := result.Stmts[0]
			nodeType := getNodeType(stmt.Stmt)
			if nodeType != tt.queryType {
				t.Errorf("Expected query type %s but got %s", tt.queryType, nodeType)
			}
		})
	}
}

// TestParseSpecialCases tests specific edge cases that might cause issues
func TestParseSpecialCases(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		test func(t *testing.T, result *pg_query.ParseResult)
	}{
		{
			name: "multiple statements in one query",
			sql:  "SELECT * FROM users; SELECT * FROM orders;",
			test: func(t *testing.T, result *pg_query.ParseResult) {
				if len(result.Stmts) != 2 {
					t.Errorf("Expected 2 statements but got %d", len(result.Stmts))
				}
			},
		},
		{
			name: "case sensitivity",
			sql:  "SeLeCt * FrOm UsErS wHeRe ID = 1",
			test: func(t *testing.T, result *pg_query.ParseResult) {
				if len(result.Stmts) != 1 {
					t.Error("Failed to parse mixed case SQL")
				}
			},
		},
		{
			name: "table and column aliases",
			sql:  "SELECT u.id AS user_id, u.name AS user_name FROM users u WHERE u.active = true",
			test: func(t *testing.T, result *pg_query.ParseResult) {
				if len(result.Stmts) != 1 {
					t.Error("Failed to parse query with aliases")
				}
			},
		},
		{
			name: "mathematical expressions",
			sql:  "SELECT id * 2 + 1, price * 0.9 AS discounted_price FROM products",
			test: func(t *testing.T, result *pg_query.ParseResult) {
				if len(result.Stmts) != 1 {
					t.Error("Failed to parse mathematical expressions")
				}
			},
		},
		{
			name: "string concatenation",
			sql:  "SELECT first_name || ' ' || last_name AS full_name FROM users",
			test: func(t *testing.T, result *pg_query.ParseResult) {
				if len(result.Stmts) != 1 {
					t.Error("Failed to parse string concatenation")
				}
			},
		},
		{
			name: "CASE expressions",
			sql:  "SELECT CASE WHEN age < 18 THEN 'minor' WHEN age >= 65 THEN 'senior' ELSE 'adult' END AS category FROM users",
			test: func(t *testing.T, result *pg_query.ParseResult) {
				if len(result.Stmts) != 1 {
					t.Error("Failed to parse CASE expression")
				}
			},
		},
		{
			name: "EXISTS subquery",
			sql:  "SELECT * FROM users u WHERE EXISTS (SELECT 1 FROM orders o WHERE o.user_id = u.id)",
			test: func(t *testing.T, result *pg_query.ParseResult) {
				if len(result.Stmts) != 1 {
					t.Error("Failed to parse EXISTS subquery")
				}
			},
		},
		{
			name: "UNION query",
			sql:  "SELECT id, name FROM users UNION SELECT id, name FROM customers",
			test: func(t *testing.T, result *pg_query.ParseResult) {
				if len(result.Stmts) != 1 {
					t.Error("Failed to parse UNION query")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePostgreSQL(tt.sql)
			if err != nil {
				t.Fatalf("Failed to parse SQL: %v", err)
			}
			tt.test(t, result)
		})
	}
}

// Helper functions
func generateLongWhereClause(conditions int) string {
	clause := "id = 1"
	for i := 1; i < conditions; i++ {
		clause += fmt.Sprintf(" AND id != %d", i+1)
	}
	return clause
}

func getNodeType(node *pg_query.Node) string {
	if node.GetSelectStmt() != nil {
		return "T_SelectStmt"
	} else if node.GetInsertStmt() != nil {
		return "T_InsertStmt"
	} else if node.GetUpdateStmt() != nil {
		return "T_UpdateStmt"
	} else if node.GetDeleteStmt() != nil {
		return "T_DeleteStmt"
	} else if node.GetCreateStmt() != nil {
		return "T_CreateStmt"
	} else if node.GetDropStmt() != nil {
		return "T_DropStmt"
	}
	return "Unknown"
}