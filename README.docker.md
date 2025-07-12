# VSQL - PostgreSQL-Compatible In-Memory Database

VSQL is a lightweight, PostgreSQL-compatible, schema-less, in-memory database written in Go. It provides full SQL syntax support through PostgreSQL's official parser while maintaining NoSQL-like flexibility.

## Quick Start

```bash
# Run VSQL server
docker run -d -p 5432:5432 satetsu888/vsql:latest

# Connect with psql
psql -h localhost -p 5432 -U any_user -d any_database
```

## Features

- ✅ PostgreSQL wire protocol compatible
- ✅ Schema-less design (NoSQL flexibility)
- ✅ Full SQL syntax support
- ✅ In-memory storage (fast performance)
- ✅ Multi-platform support
- ✅ Automatic seed data loading

## Usage Examples

### Basic Server
```bash
docker run -d -p 5432:5432 satetsu888/vsql:latest
```

### With Seed Data
Mount a directory containing `.sql` files to automatically load them on startup:
```bash
docker run -d -p 5432:5432 -v ./sql-files:/seed:ro satetsu888/vsql:latest
```

### One-time Execution
Execute commands and exit:
```bash
docker run satetsu888/vsql:latest -c "CREATE TABLE users (id int); SELECT * FROM users;" -q
```

### Custom Port
```bash
docker run -d -p 5433:5432 satetsu888/vsql:latest
```

## Docker Compose

```yaml
version: '3.8'
services:
  vsql:
    image: satetsu888/vsql:latest
    ports:
      - "5432:5432"
    volumes:
      - ./seeds:/seed:ro
```

## Environment Variables

- `SEED_DIR`: Directory containing seed SQL files (default: `/seed`)

## Supported SQL Features

- Basic operations: CREATE TABLE, INSERT, SELECT, UPDATE, DELETE, DROP TABLE
- JOINs: INNER, LEFT, RIGHT, FULL OUTER, CROSS
- Aggregations: COUNT, SUM, AVG, MAX, MIN, GROUP BY, HAVING
- Subqueries: IN, EXISTS, scalar subqueries
- And much more!

## Links

- [GitHub Repository](https://github.com/satetsu888/vsql)
- [Documentation](https://github.com/satetsu888/vsql/blob/main/README.md)

## License

MIT License