# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

VSQL is a PostgreSQL-compatible, schema-less, in-memory database written in Go. It uses PostgreSQL's official parser (pg_query_go) to provide full SQL syntax support while maintaining NoSQL-like flexibility.

## Essential Commands

### Build and Run
```bash
# Build the project
go build -o vsql

# Run the server (default port 5432)
./vsql

# Run on custom port
./vsql -port 5433

# Execute query directly without starting server
./vsql -c "SELECT * FROM users;"

# Execute multiple queries
./vsql -c "CREATE TABLE users (id int, name text); INSERT INTO users VALUES (1, 'Alice');"

# Execute SQL from file
./vsql -f queries.sql

# Show help
./vsql -h
```

### Testing
```bash
# Run all tests
go test -v

# Run specific test category
go test -v -run TestIndividualSQLFiles/crud
go test -v -run TestIndividualSQLFiles/joins
go test -v -run TestIndividualSQLFiles/null_handling

# Run tests with coverage
go test -v -cover

# Test files are organized by feature in test/sql/ directory:
# - crud/              # Basic CRUD operations (CREATE, INSERT, SELECT, UPDATE, DELETE)
# - joins/             # All JOIN types including multi-table joins (3+ tables)
# - aggregates/        # Aggregate functions (COUNT, SUM, AVG, MAX, MIN)
# - grouping/          # GROUP BY, HAVING (with aggregates and OR), DISTINCT
# - subqueries/        # IN, EXISTS, scalar subqueries in SELECT/WHERE
# - null_handling/     # Comprehensive NULL value tests with three-valued logic
# - type_conversion/   # Numeric comparisons, arithmetic expressions, proper type handling
# - operators/         # BETWEEN, LIKE, IN/NOT IN with NULL handling
# - ordering/          # ORDER BY with numeric sorting, LIMIT, OFFSET
# - error_cases/       # Error handling and edge cases
# - comments/          # SQL comment handling tests

# Manual testing with psql
psql -h localhost -p 5432 -U any_user -d any_database
```

### Development
```bash
# Install dependencies
go mod download

# Update dependencies
go get -u ./...

# Format code
go fmt ./...

# Run go vet
go vet ./...
```

## Architecture

### Core Components

1. **Main Module** (`main.go`)
   - Command-line interface with `-port`, `-c`, `-f`, and `-h` options
   - Can run as server or execute queries directly
   - Supports executing multiple SQL statements separated by semicolons
   - Intelligent SQL statement splitter that respects comments and string literals
   - `-f` option to execute SQL from files

2. **Parser Module** (`parser/`)
   - `pg_parser.go`: Basic SQL operations (CREATE, INSERT, SELECT, UPDATE, DELETE, DROP)
     - Handles IN/NOT IN clauses with value lists
     - Implements SQL three-valued logic for NULL comparisons
     - IS NULL/IS NOT NULL operators
     - BETWEEN/NOT BETWEEN operators
     - LIKE/NOT LIKE operators with pattern matching
   - `pg_parser_advanced.go`: Advanced features (JOINs, subqueries, aggregations)
     - All types of JOINs (INNER, LEFT, RIGHT, FULL OUTER, CROSS)
     - Multi-table JOINs (3+ tables) with proper qualified column resolution
     - Table aliases and qualified column references
     - GROUP BY/HAVING with aggregate functions (including COUNT DISTINCT)
     - ORDER BY with LIMIT/OFFSET
     - Subqueries (IN, NOT IN, EXISTS, scalar subqueries in SELECT and WHERE)
   - `pg_parser_utils.go`: Shared utility functions
     - Value conversion and comparison functions
     - Pattern matching for LIKE operator
     - Function name extraction and aggregate function detection
   - Uses `github.com/pganalyze/pg_query_go/v5` for PostgreSQL-compatible parsing

3. **Server Module** (`server/`)
   - `server.go`: Main server logic, handles client connections
   - `protocol.go`: PostgreSQL wire protocol implementation
   - Listens on port 5432 by default, compatible with psql and other PostgreSQL clients

4. **Storage Module** (`storage/`)
   - `datastore.go`: In-memory table and row storage using `sync.RWMutex` for thread safety
   - `metastore.go`: Metadata storage for table schemas and column ordering
   - Schema-less design: rows are `map[string]interface{}`, non-existent columns return NULL

5. **Test Module** (`individual_sql_test.go`)
   - Automated test runner using `-c` option for direct query execution
   - Parses SQL test files from `test/sql/` directory
   - Validates expected row counts and error conditions
   - Preserves SQL comments during testing for accurate validation

### Request Flow

1. Client connects using PostgreSQL protocol
2. Server receives SQL query string
3. Parser converts SQL to AST using pg_query_go
4. Executor processes AST:
   - For queries: evaluates WHERE clauses, performs JOINs, applies aggregations
   - For mutations: updates in-memory storage
5. Results formatted and returned via PostgreSQL wire protocol

### Key Design Decisions

- **Schema-less Storage**: Tables accept any columns at runtime, providing NoSQL flexibility
- **PostgreSQL Compatibility**: Full syntax support through official parser, wire protocol compatibility
- **In-memory Only**: No persistence, all data lost on restart
- **Thread-safe**: All storage operations protected by RWMutex
- **Column Ordering**: Maintains consistent column order from CREATE TABLE statements, with additional columns appended as needed

## Supported SQL Features

### Fully Implemented
- Basic operations: CREATE TABLE, INSERT, SELECT, UPDATE, DELETE, DROP TABLE
- Schema-less table design (can insert new columns at any time)
- Complex WHERE clauses with AND, OR, NOT
- NULL handling with SQL three-valued logic
- IS NULL / IS NOT NULL operators
- IN clause with value lists (including proper NULL handling)
- NOT IN clause with value lists (including proper NULL handling)
- Aggregate functions: COUNT, SUM, AVG, MAX, MIN
  - SUM returns NULL for empty result sets (SQL standard compliant)
  - MAX/MIN work correctly with numeric comparisons
- COUNT(DISTINCT column) - counts unique non-NULL values
- GROUP BY / HAVING (including HAVING with aggregate functions and OR conditions)
- Subqueries: IN with subqueries, scalar subqueries in WHERE clause, scalar subqueries in SELECT clause
- DISTINCT queries
- Column ordering consistency (from CREATE TABLE + dynamic columns)
- BETWEEN / NOT BETWEEN operators
- LIKE operator (with % and _ wildcards)
- NOT LIKE operator
- JOINs: INNER JOIN, LEFT JOIN, RIGHT JOIN, FULL OUTER JOIN, CROSS JOIN
- ORDER BY with LIMIT and OFFSET (including ORDER BY with aggregates)
- Table aliases and qualified column references (e.g., t1.id)
- SQL comments (single-line -- and multi-line /* */)
- Scalar subqueries in WHERE clause (e.g., WHERE age > (SELECT AVG(age) FROM users))
- Arithmetic expressions in WHERE clause (e.g., WHERE price * 2 > 100)
- Numeric comparisons with proper type handling (not string comparison)

### Partially Implemented
- EXISTS/NOT EXISTS subqueries: Basic and correlated subqueries work, including:
  - ✓ Nested EXISTS (EXISTS within EXISTS)
  - ✓ GROUP BY/HAVING in the subquery (including aggregate functions not in SELECT)
  - ✓ Complex OR conditions referencing outer table (works with proper data types, see note on boolean handling)
- UNION/UNION ALL: Basic structure exists
- OFFSET: Basic implementation (some edge cases may not work)
- Boolean type: Boolean literals (true/false) are not properly stored or compared
  - Workaround: Use integer (0/1) or text ('true'/'false') instead

### Not Yet Implemented
- ILIKE operator (case-insensitive LIKE)
- CASE expressions
- COALESCE function
- Window functions
- CTEs (WITH clause)
- Correlated subqueries (for contexts other than EXISTS/NOT EXISTS)
- Transactions
- Indexes
- Constraints
- Data persistence
- Foreign keys
- Views
- Stored procedures/functions

## Writing SQL Test Files

### Overview
SQL test files are located in `test/sql/` directory, organized by feature categories. The test runner (`individual_sql_test.go`) automatically executes these files and validates their output.

### Test File Structure Rules

1. **One Test Query Per File**
   - Each `.sql` file must contain exactly ONE main test query
   - Setup queries (CREATE TABLE, INSERT) and cleanup queries (DROP TABLE) are allowed
   - Do NOT put multiple test scenarios in a single file

2. **File Naming Convention**
   - Use descriptive names: `test_XX_feature_description.sql`
   - For split files, use suffixes: `test_XX a_specific_case.sql`, `test_XXb_another_case.sql`
   - Group related tests with the same number prefix

3. **Required Comments**
   - Always include a comment describing what the test is testing
   - Use `-- Expected: N rows` to specify expected row count
   - Use `-- Expected: no rows` for queries that should return 0 rows
   - Use `-- Expected: error` for queries that should fail
   - Mark known failing tests with `-- Status: FAILING`

### Comment Metadata Format

```sql
-- Test: Description of what this test validates
-- Expected: 3 rows
-- Status: FAILING (optional, only for known failures)

-- Setup
CREATE TABLE test_table (...);
INSERT INTO test_table VALUES (...);

-- Main test query (ONLY ONE per file)
SELECT * FROM test_table WHERE condition;

-- Cleanup
DROP TABLE test_table;
```

### Examples

#### Working Test Example
```sql
-- Test: Basic SELECT with WHERE clause
-- Expected: 2 rows

CREATE TABLE users (id int, name text);
INSERT INTO users VALUES (1, 'Alice'), (2, 'Bob'), (3, 'Charlie');

SELECT * FROM users WHERE id < 3;

DROP TABLE users;
```

#### Expected Error Example
```sql
-- Test: SELECT from non-existent table
-- Expected: error

SELECT * FROM non_existent_table;
```

#### Known Failing Test Example
```sql
-- Test: CASE expression (not yet implemented)
-- Expected: 1 rows
-- Status: FAILING

CREATE TABLE test (id int, value int);
INSERT INTO test VALUES (1, 10);

SELECT CASE WHEN value > 5 THEN 'high' ELSE 'low' END FROM test;

DROP TABLE test;
```

### Best Practices

1. **Keep tests focused**: Each file should test one specific feature or edge case
2. **Use clear test data**: Make it obvious why the expected result is correct
3. **Clean up after tests**: Always DROP tables created during the test
4. **Document limitations**: If a test demonstrates a limitation or bug, explain it in comments
5. **Order matters**: The test runner preserves SQL comments, so place metadata comments before the SQL

### Test Categories

Place test files in the appropriate subdirectory:
- `crud/` - Basic CREATE, INSERT, SELECT, UPDATE, DELETE operations
- `joins/` - All types of JOINs
- `aggregates/` - Aggregate functions (COUNT, SUM, AVG, etc.)
- `grouping/` - GROUP BY, HAVING, DISTINCT
- `subqueries/` - IN, EXISTS, scalar subqueries
- `null_handling/` - NULL value behavior and three-valued logic
- `type_conversion/` - Type handling and conversions
- `operators/` - BETWEEN, LIKE, IN/NOT IN, etc.
- `ordering/` - ORDER BY, LIMIT, OFFSET
- `error_cases/` - Error conditions and edge cases
- `comments/` - SQL comment handling
- `data_types/` - Data type specific tests (boolean, numeric, text, etc.)
