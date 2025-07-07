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

# Run specific test suite
go test -v -run TestNullHandling
go test -v -run TestTypeComparison

# Run tests with coverage
go test -v -cover

# Test files are located in test/ directory:
# - null_handling_test.sql          # NULL value handling tests
# - null_comparisons_test.sql       # Basic NULL comparison tests
# - type_comparison_test.sql        # Type conversion and comparison tests
# - complex_queries_test.sql        # Advanced SQL features (JOINs, subqueries)
# - error_handling_test.sql         # Error cases and edge conditions
# - advanced_queries_test.sql       # Comprehensive advanced SQL tests
# - basic_advanced_test.sql         # Basic to intermediate SQL tests
# - basic_integration_test.sql      # Basic integration tests
# - enhanced_integration_test.sql   # Comprehensive integration tests

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

2. **Parser Module** (`parser/`)
   - `pg_parser.go`: Basic SQL operations (CREATE, INSERT, SELECT, UPDATE, DELETE, DROP)
     - Handles IN/NOT IN clauses with value lists
     - Implements SQL three-valued logic for NULL comparisons
     - IS NULL/IS NOT NULL operators
   - `pg_parser_advanced.go`: Advanced features (JOINs, subqueries, aggregations)
   - Uses `github.com/pganalyze/pg_query_go/v5` for PostgreSQL-compatible parsing

3. **Server Module** (`server/`)
   - `server.go`: Main server logic, handles client connections
   - `protocol.go`: PostgreSQL wire protocol implementation
   - Listens on port 5432 by default, compatible with psql and other PostgreSQL clients

4. **Storage Module** (`storage/`)
   - `datastore.go`: In-memory table and row storage using `sync.RWMutex` for thread safety
   - `metastore.go`: Metadata storage for table schemas and column ordering
   - Schema-less design: rows are `map[string]interface{}`, non-existent columns return NULL

5. **Test Module** (`sql_integration_test.go`)
   - Automated test runner using `-c` option for direct query execution
   - Parses SQL test files from `test/` directory
   - Validates expected row counts and error conditions

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
- GROUP BY / HAVING
- Subqueries: IN with subqueries
- DISTINCT queries
- Column ordering consistency (from CREATE TABLE + dynamic columns)

### Partially Implemented
- JOINs: Basic structure exists but needs fixes
- EXISTS/NOT EXISTS subqueries: Structure exists but needs fixes
- ORDER BY: Works but LIMIT/OFFSET need fixes
- Complex WHERE conditions: Most work except LIKE operator

### Not Yet Implemented
- BETWEEN operator
- LIKE/ILIKE operators
- CASE expressions
- COALESCE function
- Window functions
- CTEs (WITH clause)
- UNION/UNION ALL
- Transactions
- Indexes
- Constraints
- Data persistence

## Recent Changes

### 2024-01 Updates
- Added `-c` command-line option for direct query execution
- Fixed IN clause handling with value lists
- Fixed NOT IN clause with proper NULL handling
- Improved SQL test framework to use `-c` option instead of server connection
- Moved all example queries to test directory with expected values
- Removed shell test scripts in favor of Go test integration
