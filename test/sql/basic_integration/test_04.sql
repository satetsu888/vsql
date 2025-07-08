-- Test 4: Select all data
-- Expected: 2 rows

-- Setup
CREATE TABLE users (id int, name text, email text);
INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@example.com');
INSERT INTO users (id, name, age) VALUES (2, 'Bob', 30);

-- Test Query
SELECT * FROM users;

-- Cleanup
DROP TABLE users;