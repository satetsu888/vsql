-- Test 16: Single quotes in data
-- Expected: success

-- Setup
CREATE TABLE users (id int, name text, email text);

-- Test Query
INSERT INTO users (id, name) VALUES (5, 'O''Brien');

-- Cleanup
DROP TABLE users;