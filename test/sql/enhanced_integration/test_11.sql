-- Test 11: NOT condition
-- Expected: 1 row (Charlie)

-- Setup
CREATE TABLE users (id int, name text, email text);
INSERT INTO users (id, name, age) VALUES (2, 'Bob', 30);
INSERT INTO users (id, name, email, country) VALUES (3, 'Charlie', 'charlie@example.com', 'USA');

-- Test Query
SELECT * FROM users WHERE NOT (id = 2);

-- Cleanup
DROP TABLE users;