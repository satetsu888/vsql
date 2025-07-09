-- Test 6: Insert with new column
-- Expected: success

-- Setup
CREATE TABLE users (id int, name text, email text);

-- Test Query
INSERT INTO users (id, name, age) VALUES (2, 'Bob', 30);

-- Cleanup
DROP TABLE users;