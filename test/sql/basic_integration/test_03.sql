-- Test 3: Insert with different schema (demonstrating schema-less feature)
-- Expected: success

-- Setup
CREATE TABLE users (id int, name text, email text);
INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@example.com');

-- Test Query
INSERT INTO users (id, name, age) VALUES (2, 'Bob', 30);

-- Cleanup
DROP TABLE users;