-- Test 5: Select specific columns with WHERE
-- Expected: 1 row (Alice)

-- Setup
CREATE TABLE users (id int, name text, email text);
INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@example.com');
INSERT INTO users (id, name, age) VALUES (2, 'Bob', 30);

-- Test Query
SELECT name, email FROM users WHERE id = 1;

-- Cleanup
DROP TABLE users;