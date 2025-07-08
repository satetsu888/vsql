-- Test 6: Select with NULL column comparison
-- Expected: 1 row (Bob)

-- Setup
CREATE TABLE users (id int, name text, email text);
INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@example.com');
INSERT INTO users (id, name, age) VALUES (2, 'Bob', 30);

-- Test Query
SELECT * FROM users WHERE age > 25;

-- Cleanup
DROP TABLE users;