-- Test 5: Delete data
-- Expected: DELETE 1

-- Setup
CREATE TABLE users (id int, name text, email text);
INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@example.com');

-- Test Query
DELETE FROM users WHERE id = 1;

-- Cleanup
DROP TABLE users;