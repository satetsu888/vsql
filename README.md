# VSQL - The Migration-Free Development Database

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/satetsu888/vsql.svg)](https://hub.docker.com/r/satetsu888/vsql)

**Skip writing migrations during development. Start vibing with your code.**

VSQL is a PostgreSQL-compatible in-memory database that lets you modify your schema on the fly. Perfect for rapid prototyping, experimentation, and those moments when you just want to code without schema management overhead.

## 🎯 Why VSQL?

### The Problem
You're building a new feature. You need a new column. In common databases:

1. Update your ORM models
2. Generate a migration file
3. Run the migration
4. Finally write the actual feature code

### The VSQL Way
```sql
INSERT INTO users (id, name, additional_column) VALUES (1, 'Alice', 'This just works!');
```
Done. No migrations. No restarts. Just pure development flow.

## ✨ Features

- 🚀 **Drop-in PostgreSQL Replacement** - Your app can't tell the difference
- 🔄 **No Migrations Ever** - Add columns by just using them
- ⚡ **Zero Setup Time** - Run one command and start coding
- 🎨 **Perfect for Prototyping** - Try ideas without commitment
- 🔍 **Real SQL** - VSQL is a full-featured PostgreSQL-compatible database with: JOINs, subqueries, aggregations
- 🐳 **Docker Ready** - `docker run -p 5432:5432 satetsu888/vsql`
- 🛡️ **Smart Type Inference** - Automatically figures out your data types

## 🚀 Quick Start

### 30 Seconds to Get Running in Your Environment

```bash
# Start VSQL (replaces PostgreSQL on port 5432)
docker run -d -p 5432:5432 satetsu888/vsql:latest

```

or

```yaml
# Replace your existing PostgreSQL service in docker-compose.yml
services:
  db:
-   image: postgres:latest
+   image: satetsu888/vsql:latest
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
```

Your existing PostgreSQL app will connect without any changes.

## 🎨 Vibe Coding Examples

### The Magic of Schema-Free Development

```sql
-- Monday: Start with a simple idea
CREATE TABLE users (id int, name text);
INSERT INTO users VALUES (1, 'Alice');

-- Tuesday: Need email? Just add it!
INSERT INTO users (id, name, email) VALUES (2, 'Bob', 'bob@example.com');

-- Wednesday: Actually, let's add profile data
INSERT INTO users (id, name, profile_pic, bio, is_premium) 
VALUES (3, 'Charlie', 'https://...', 'Love coding!', true);

-- Thursday: All your data is still there, schema evolved naturally
SELECT * FROM users;
-- Returns all users with whatever columns they have
/*
 id |  name   |      email      | profile_pic |     bio      | is_premium 
----+---------+-----------------+-------------+--------------+------------
 1  | Alice   |                 |             |              | 
 2  | Bob     | bob@example.com |             |              | 
 3  | Charlie |                 | https://... | Love coding! | true
(3 rows)
*/
```

## 🛠️ Perfect For

- **Hackathons & Prototypes** - Focus on features, not database setup
- **Learning SQL** - Experiment without fear of breaking things
- **Testing New Ideas** - Try that crazy schema change instantly
- **Local Development** - Replace PostgreSQL/MySQL during development
- **CI/CD Pipelines** - Spin up a full database in milliseconds
- **Microservice Development** - Each service gets its own schema-free database

## 🚫 Not For

- Production use (it's in-memory, data doesn't persist!)
- Large datasets (remember, it's all in RAM)
- Applications requiring ACID transactions
- Systems needing user authentication

## 💡 Common Patterns

### Start with Seed Data

```bash
# Create a seed file with your test data
echo "CREATE TABLE users (id int, name text);
INSERT INTO users VALUES (1, 'Test User');
CREATE TABLE products (id int, name text, price float);
INSERT INTO products VALUES (1, 'Widget', 9.99);" > seed.sql

# Run VSQL with seed data
docker run -d -p 5432:5432 -v $(pwd):/seed:ro satetsu888/vsql:latest
```

### Quick CLI Testing

```bash
# Test queries without starting a server
vsql -c "CREATE TABLE test (id int); INSERT INTO test VALUES (1), (2), (3); SELECT COUNT(*) FROM test;" -q

# Output: 3
```

## 🔧 Real SQL Support

VSQL is not a toy - it's a real PostgreSQL-compatible database with:

✅ **Full SQL**: JOINs (all types), subqueries, aggregations, GROUP BY/HAVING  
✅ **Complex Queries**: Multi-table joins, correlated subqueries, CTEs (soon)  
✅ **All Data Types**: Integers, floats, strings, booleans with automatic inference  
✅ **Proper NULL Handling**: Three-valued logic, IS NULL/IS NOT NULL  
✅ **Advanced Features**: Table aliases, qualified columns, DISTINCT, ORDER BY/LIMIT  

## 🤔 FAQ

**Q: Can I use this with my ORM (Django, Rails, Prisma)?**  
A: Yes! VSQL speaks PostgreSQL protocol. Your ORM won't know the difference.

**Q: What happens to my data when I restart?**  
A: It's gone. That's the point - no state to manage during development. Just use a seed file to quickly recreate your test data when needed.

**Q: Can I migrate to a real database later?**  
A: Absolutely. When you're ready for production, export your final schema and create proper migration files for your target database.

**Q: Why not just use SQLite?**  
A: SQLite is great for many use cases! VSQL offers schema-free development and PostgreSQL compatibility for rapid prototyping scenarios.

## 🤝 Contributing

We love contributions! VSQL is built with Go and uses PostgreSQL's parser. Check out our [contributing guide](CONTRIBUTING.md) to get started.

## 📝 License

MIT License

---

**Remember**: VSQL is for development, not for production deployment. When you're ready to ship, use a real database. But until then... enjoy the freedom! 🚀