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
```

### Testing
```bash
# Run unit tests
go test ./...

# Run tests with race detector
go test ./... -race

# Run tests with coverage
go test ./... -cover

# Run all tests (convenience script)
./run_tests.sh

# Run integration tests (starts server, runs SQL tests via psql)
./test_vsql.sh

# Run enhanced integration tests with colored output
# This includes comprehensive tests for JOINs, subqueries, aggregations, etc.
./test_vsql_enhanced.sh

# Run memory stress test
./test/memory_stress_test.sh

# Manual testing with psql
psql -h localhost -p 5432 -U any_user -d any_database

# Run advanced SQL tests
psql -h localhost -p 5432 -U test -d test < test_advanced.sql
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

1. **Parser Module** (`parser/`)
   - `pg_parser.go`: Basic SQL operations (CREATE, INSERT, SELECT, UPDATE, DELETE, DROP)
   - `pg_parser_advanced.go`: Advanced features (JOINs, subqueries, aggregations)
   - Uses `github.com/pganalyze/pg_query_go/v5` for PostgreSQL-compatible parsing

2. **Server Module** (`server/`)
   - `server.go`: Main server logic, handles client connections
   - `protocol.go`: PostgreSQL wire protocol implementation
   - Listens on port 5432 by default, compatible with psql and other PostgreSQL clients

3. **Storage Module** (`storage/`)
   - `datastore.go`: In-memory table and row storage using `sync.RWMutex` for thread safety
   - `metastore.go`: Metadata storage for table schemas and column ordering
   - Schema-less design: rows are `map[string]interface{}`, non-existent columns return NULL

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

- Basic operations: CREATE TABLE, INSERT, SELECT, UPDATE, DELETE, DROP TABLE
- Complex WHERE clauses with AND, OR, NOT
- JOINs: INNER, LEFT, RIGHT, FULL OUTER
- Aggregate functions: COUNT, SUM, AVG, MAX, MIN
- GROUP BY / HAVING
- ORDER BY, LIMIT, OFFSET
- Subqueries: IN and EXISTS clauses fully supported
- SELECT * with consistent column ordering across schema-less data
- NULL value handling
