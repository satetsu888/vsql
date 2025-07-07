-- Basic Integration Test
-- Extracted from test_vsql.sh

-- Test 1: Create table with schema-less design
-- Expected: success
CREATE TABLE users (id int, name text, email text);

-- Test 2: Insert basic data
-- Expected: success
INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@example.com');

-- Test 3: Insert with different schema (demonstrating schema-less feature)
-- Expected: success
INSERT INTO users (id, name, age) VALUES (2, 'Bob', 30);

-- Test 4: Select all data
-- Expected: 2 rows
SELECT * FROM users;

-- Test 5: Select specific columns with WHERE
-- Expected: 1 row (Alice)
SELECT name, email FROM users WHERE id = 1;

-- Test 6: Select with NULL column comparison
-- Expected: 1 row (Bob)
SELECT * FROM users WHERE age > 25;

-- Cleanup
DROP TABLE users;