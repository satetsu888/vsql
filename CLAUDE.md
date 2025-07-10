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
# - joins/             # All JOIN types (INNER, LEFT, RIGHT, FULL OUTER, CROSS)
# - aggregates/        # Aggregate functions (COUNT, SUM, AVG, MAX, MIN)
# - grouping/          # GROUP BY, HAVING, DISTINCT operations
# - subqueries/        # Subqueries (IN, EXISTS, correlated, scalar)
# - null_handling/     # Comprehensive NULL value tests
# - type_conversion/   # Type conversion and comparison tests
# - operators/         # SQL operators (BETWEEN, LIKE, IN/NOT IN)
# - ordering/          # ORDER BY, LIMIT, OFFSET tests
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
   - Command-line interface with `-port`, `-c`, and `-h` options
   - Can run as server or execute queries directly
   - Supports executing multiple SQL statements separated by semicolons
   - Intelligent SQL statement splitter that respects comments and string literals

2. **Parser Module** (`parser/`)
   - `pg_parser.go`: Basic SQL operations (CREATE, INSERT, SELECT, UPDATE, DELETE, DROP)
     - Handles IN/NOT IN clauses with value lists
     - Implements SQL three-valued logic for NULL comparisons
     - IS NULL/IS NOT NULL operators
     - BETWEEN/NOT BETWEEN operators
     - LIKE/NOT LIKE operators with pattern matching
   - `pg_parser_advanced.go`: Advanced features (JOINs, subqueries, aggregations)
     - All types of JOINs (INNER, LEFT, RIGHT, FULL OUTER, CROSS)
     - Table aliases and qualified column references
     - GROUP BY/HAVING with aggregate functions (including COUNT DISTINCT)
     - ORDER BY with LIMIT/OFFSET
     - Subqueries (IN, NOT IN, EXISTS, scalar subqueries in WHERE)
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
- COUNT(DISTINCT column) - counts unique non-NULL values
- GROUP BY / HAVING
- Subqueries: IN with subqueries, scalar subqueries in WHERE clause
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

### Partially Implemented
- EXISTS/NOT EXISTS subqueries: Basic structure exists but correlated subqueries not supported
- UNION/UNION ALL: Basic structure exists
- Complex multi-table JOINs: Two-table joins work well, 3+ tables need more testing
- OFFSET: Basic implementation (test failures indicate partial support)

### Not Yet Implemented
- ILIKE operator (case-insensitive LIKE)
- CASE expressions
- COALESCE function
- Window functions
- CTEs (WITH clause)
- Scalar subqueries in SELECT clause
- Correlated subqueries (for EXISTS and other contexts)
- Transactions
- Indexes
- Constraints
- Data persistence
- Foreign keys
- Views
- Stored procedures/functions

## Recent Updates (2025-07-10)

### Refactoring
- Created `pg_parser_utils.go` to consolidate shared utility functions
- Eliminated duplicate code between `pg_parser.go` and `pg_parser_advanced.go`
- Reduced overall codebase by ~141 lines while maintaining functionality

### New Features Implemented
- **NOT LIKE operator**: Pattern matching with negation
- **LIKE wildcards**: Fixed support for complex patterns with % and _
- **COUNT(DISTINCT)**: Properly counts unique non-NULL values
- **CROSS JOIN**: Cartesian product of two tables
- **Scalar subqueries in WHERE**: Support for comparisons like `WHERE age > (SELECT AVG(age) FROM users)`

### Test Infrastructure
- Fixed test parser to correctly extract expected row counts from comments
- Multiple previously failing tests now pass
