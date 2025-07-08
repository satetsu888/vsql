-- Test 8: Mixed schema insert
-- Expected: success

-- Setup
CREATE TABLE users (id int, name text, email text);
INSERT INTO users (id, name, age) VALUES (2, 'Bob', 30);

-- Test Query
INSERT INTO users (id, name, email, country) VALUES (3, 'Charlie', 'charlie@example.com', 'USA');

-- Cleanup
DROP TABLE users;