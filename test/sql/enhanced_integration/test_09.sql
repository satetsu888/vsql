-- Test 9: AND condition
-- Expected: 1 row (Bob)

-- Setup
CREATE TABLE users (id int, name text, email text);
INSERT INTO users (id, name, age) VALUES (2, 'Bob', 30);
INSERT INTO users (id, name, email, country) VALUES (3, 'Charlie', 'charlie@example.com', 'USA');

-- Test Query
SELECT * FROM users WHERE id > 1 AND name = 'Bob';

-- Cleanup
DROP TABLE users;