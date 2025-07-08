-- Test 3: Select all
-- Expected: 1 row

-- Setup
CREATE TABLE users (id int, name text, email text);
INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@example.com');

-- Test Query
SELECT * FROM users;

-- Cleanup
DROP TABLE users;