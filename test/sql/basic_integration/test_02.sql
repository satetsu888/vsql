-- Test 2: Insert basic data
-- Expected: success

-- Setup
CREATE TABLE users (id int, name text, email text);

-- Test Query
INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@example.com');

-- Cleanup
DROP TABLE users;