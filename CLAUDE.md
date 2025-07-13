# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

VSQL is a PostgreSQL-compatible, schema-less, in-memory database written in Go. It uses PostgreSQL's official parser (pg_query_go) to provide full SQL syntax support while maintaining NoSQL-like flexibility.

**⚠️ IMPORTANT: This project is currently PENDING**
- Development has been suspended due to a fundamental architectural conflict between schema-less design and properly generating PostgreSQL protocol's RowDescription messages
- Basic queries from psql work, but queries from ORMs (ActiveRecord, Prisma, SQLAlchemy, etc.) often fail
- Client libraries that require accurate type information do not work as expected

## Essential Commands

### Build and Run
```bash
# Build the project
go build -o vsql

# Run the server (default port 5432)
./vsql

# Run on custom port
./vsql -port 5433

# Execute query and exit (test mode)
./vsql -c "SELECT * FROM users;" -q

# Execute query then start server (seed data)
./vsql -c "CREATE TABLE users (id int, name text); INSERT INTO users (id, name) VALUES (1, 'Alice');"

# Execute SQL from file and exit
./vsql -f queries.sql -q

# Execute SQL from file as seed data then start server
./vsql -f seed.sql

# Execute both file and command (file runs first)
./vsql -f schema.sql -c "INSERT INTO users (id, name) VALUES (1, 'Alice');" -q

# Execute multiple SQL files in order
./vsql -f schema.sql -f data.sql -f indexes.sql -q

# Execute multiple commands
./vsql -c "CREATE TABLE users (id int)" -c "INSERT INTO users (id) VALUES (1)" -q

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
# - functions/         # String functions (UPPER, LOWER), COALESCE
# - data_types/        # Data type specific tests (boolean, numeric, text, etc.)

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

### Docker

VSQL is available as a Docker image, making it easy to deploy and use in containerized environments.

#### Building and Publishing the Docker Image
```bash
# Build multi-architecture image (local build only)
./build-and-push.sh

# Build and push to DockerHub
./build-and-push.sh --push

# Build with a specific tag
./build-and-push.sh --push --tag v1.0.0

# Note: The build script now builds for linux/amd64 and linux/arm64 only
# (linux/arm/v7 was removed due to build time constraints)
```

#### Running VSQL in Docker

##### Basic Server Mode
```bash
# Run VSQL server on default port
docker run -d -p 5432:5432 satetsu888/vsql:latest

# Run on custom port
docker run -d -p 5433:5432 satetsu888/vsql:latest
```

##### With Seed Data Directory
The Docker image automatically loads all `.sql` files from the `/seed` directory on startup:

```bash
# Mount local directory with SQL files as seed data
docker run -d -p 5432:5432 -v ./my-sql-files:/seed:ro satetsu888/vsql:latest

# The container will execute all *.sql files in alphabetical order
```

##### One-time Execution Mode
```bash
# Execute commands and exit
docker run satetsu888/vsql:latest -c "SELECT 1" -q

# Execute seed files and exit
docker run -v ./migrations:/seed:ro satetsu888/vsql:latest -q
```

##### Custom Commands
```bash
# Run with custom commands (files execute first, then commands)
docker run -d -p 5432:5432 \
  -v ./schema:/seed:ro \
  satetsu888/vsql:latest \
  -c "INSERT INTO users (id, name) VALUES (999, 'Docker User')"
```

#### Docker Compose Examples
See `docker-compose.example.yml` for various configuration examples:

```yaml
# Basic server
services:
  vsql:
    image: satetsu888/vsql:latest
    ports:
      - "5432:5432"
    volumes:
      - ./sql-seeds:/seed:ro
```

#### Environment Variables
- `SEED_DIR`: Directory containing seed SQL files (default: `/seed`)

#### DockerHub
```bash
# Pull from DockerHub
docker pull satetsu888/vsql:latest

# Run the pulled image
docker run -d -p 5432:5432 satetsu888/vsql:latest
```

## Architecture

### Core Components

1. **Main Module** (`main.go`)
   - Command-line interface with `-port`, `-c`, `-f`, `-q`, and `-h` options
   - Can run as server or execute queries directly
   - `-q` flag: quit after executing commands (test mode)
   - Without `-q`: executes commands then starts server (seed data mode)
   - `-f` flag: can be specified multiple times to execute multiple files in order
   - `-c` flag: can be specified multiple times to execute multiple commands in order
   - When both `-f` and `-c` are specified, all `-f` files execute first, then all `-c` commands
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
     - Multi-table JOINs (3+ tables) with proper qualified column resolution
     - Table aliases and qualified column references
     - GROUP BY/HAVING with aggregate functions (including COUNT DISTINCT)
     - ORDER BY with LIMIT/OFFSET
     - Subqueries (IN, NOT IN, EXISTS, scalar subqueries in SELECT and WHERE)
   - `pg_parser_utils.go`: Shared utility functions
     - Value conversion and comparison functions
     - Pattern matching for LIKE operator
     - Function name extraction and aggregate function detection
   - `prepared_statements.go`: Manages prepared statement storage
   - `prepare_execute.go`: Handles PREPARE/EXECUTE/DEALLOCATE SQL commands
   - Uses `github.com/pganalyze/pg_query_go/v5` for PostgreSQL-compatible parsing

3. **Server Module** (`server/`)
   - `server.go`: Main server logic, handles client connections
   - `protocol.go`: PostgreSQL wire protocol implementation (Simple Query Protocol)
   - `extended_protocol.go`: Extended Query Protocol implementation
     - Handles Parse, Bind, Execute, Describe, Close messages
     - Maps VSQL types to PostgreSQL OIDs
     - Limited parameter type inference
   - Listens on port 5432 by default, compatible with psql and other PostgreSQL clients

4. **Storage Module** (`storage/`)
   - `datastore.go`: In-memory table and row storage using `sync.RWMutex` for thread safety
   - `metastore.go`: Metadata storage for table schemas, column ordering, and type information
     - Tracks column types per table
     - Maintains column order from CREATE TABLE
     - Validates type compatibility on INSERT/UPDATE
   - `types.go`: Type system definitions and type inference
     - Basic types: Integer, Float, String, Boolean, Timestamp
     - Type inference from values
     - Type compatibility rules (e.g., Integer→Float promotion)
     - PostgreSQL OID mapping for Extended Query Protocol
   - Schema-less design: rows are `map[string]interface{}`, non-existent columns return NULL
   - Type safety: Automatic type inference with validation to prevent incompatible type changes

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
- **Graceful Handling of Non-existent Tables**: Queries on non-existent tables return empty results rather than errors, allowing for more flexible application development

## Supported SQL Features

### Fully Implemented
- Basic operations: CREATE TABLE, INSERT, SELECT, UPDATE, DELETE, DROP TABLE
- Schema-less table design (can insert new columns at any time)
- Type inference and validation system:
  - Automatic type inference from INSERT values
  - Type safety: prevents incompatible type changes
  - Supports: Integer, Float, String, Boolean types
  - Integer to Float promotion is allowed
  - NULL values don't affect type determination
- Complex WHERE clauses with AND, OR, NOT
- NULL handling with SQL three-valued logic
- IS NULL / IS NOT NULL operators (including in SELECT and WHERE clauses)
- IN clause with value lists (including proper NULL handling)
- NOT IN clause with value lists (including proper NULL handling)
- Aggregate functions: COUNT, SUM, AVG, MAX, MIN
  - SUM returns NULL for empty result sets (SQL standard compliant)
  - MAX/MIN work correctly with numeric comparisons
- COUNT(DISTINCT column) - counts unique non-NULL values
- GROUP BY / HAVING (including HAVING with aggregate functions and OR conditions)
- Subqueries: IN with subqueries, scalar subqueries in WHERE clause, scalar subqueries in SELECT clause
- EXISTS/NOT EXISTS subqueries: Both basic and correlated subqueries work, including:
  - ✓ Nested EXISTS (EXISTS within EXISTS)
  - ✓ GROUP BY/HAVING in the subquery (including aggregate functions not in SELECT)
  - ✓ Complex OR conditions referencing outer table
  - ✓ Proper handling of qualified column references in correlated subqueries
- DISTINCT queries
- Column ordering consistency (from CREATE TABLE + dynamic columns)
- BETWEEN / NOT BETWEEN operators
- LIKE operator (with % and _ wildcards)
- NOT LIKE operator
- JOINs: INNER JOIN, LEFT JOIN, RIGHT JOIN, FULL OUTER JOIN, CROSS JOIN
  - Proper NULL handling for non-matching rows in outer joins
  - Qualified column references work correctly in JOIN queries
  - Multi-table JOINs with proper column resolution
- ORDER BY with LIMIT and OFFSET (including ORDER BY with aggregates)
- Table aliases and qualified column references (e.g., t1.id)
- SQL comments (single-line -- and multi-line /* */)
- Scalar subqueries in WHERE clause (e.g., WHERE age > (SELECT AVG(age) FROM users))
- Arithmetic expressions in WHERE clause (e.g., WHERE price * 2 > 100)
- Numeric comparisons with proper type handling (not string comparison)
- String functions: UPPER, LOWER
- COALESCE function (returns first non-NULL argument)
- Boolean type: Now fully implemented with PostgreSQL-compatible behavior
  - Boolean literals (true/false) work correctly
  - WHERE clause with boolean columns works as expected
  - NOT operator on boolean values is supported
- Non-existent table handling:
  - SELECT from non-existent tables returns empty result set (0 rows)
  - UPDATE on non-existent tables returns "UPDATE 0" 
  - DELETE from non-existent tables returns "DELETE 0"
  - JOINs with non-existent tables return appropriate empty results

### Partially Implemented
- UNION/UNION ALL: Basic structure exists but not fully tested
- SELECT without FROM clause: Now supported for PostgreSQL compatibility
- Extended Query Protocol: Basic implementation exists but has limitations:
  - Parse, Bind, Execute, Describe, Close messages are implemented
  - Parameter type inference is limited (only OFFSET clauses)
  - RowDescription messages lack accurate type information for schema-less columns
  - Prepared statements (PREPARE/EXECUTE/DEALLOCATE) are supported

### Not Yet Implemented
- ILIKE operator (case-insensitive LIKE)
- CASE expressions
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
- information_schema tables (required by many ORMs)
- System functions like current_database(), current_schema()
- Full ORM compatibility (due to type information requirements)

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

4. **Expected Behavior Must Follow PostgreSQL**
   - The expected behavior defined in tests should conform to PostgreSQL behavior
   - When PostgreSQL behavior differs from SQL standard, follow PostgreSQL
   - This project aims for PostgreSQL compatibility, not generic SQL standard compliance
   - For SQL three-valued logic, NULL comparisons, and boolean operations, follow PostgreSQL rules

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

#### Empty Result Example
```sql
-- Test: SELECT from non-existent table returns empty result
-- Expected: 0 rows

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
- `functions/` - SQL functions (UPPER, LOWER, COALESCE, etc.)
- `type_safety/` - Type inference and type mismatch tests

## Project Status and Future Directions

### Current Limitations
The fundamental challenge is that PostgreSQL's Extended Query Protocol requires accurate type information in RowDescription messages, which conflicts with the schema-less design philosophy. This makes it difficult to support ORMs and advanced client libraries that rely on this type information.

### Potential Future Directions
1. **Hybrid Schema-less Approach**: Maintain type information from first INSERT while allowing new columns
2. **Development Environment Focus**: Position as a rapid prototyping tool rather than full PostgreSQL replacement
3. **Enhanced Type Inference**: Improve context-based type inference for better ORM compatibility
4. **information_schema Implementation**: Add virtual tables for basic ORM support

### Testing Extended Query Protocol
```bash
# Test basic ORM-style queries
./test_protocol.sh

# Test prepared statements
./test_orm_query.sh
```

These scripts test the Extended Query Protocol implementation and help identify compatibility issues with ORMs.
